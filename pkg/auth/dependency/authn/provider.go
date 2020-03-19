package authn

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type Provider struct {
	OAuth                   *OAuthCoordinator
	Authn                   *AuthenticateProcess
	Signup                  *SignupProcess
	AuthnSession            *SessionProvider
	Session                 session.Provider
	SessionCookieConfig     session.CookieConfiguration
	BearerTokenCookieConfig mfa.BearerTokenCookieConfiguration
}

func (p *Provider) SignupWithLoginIDs(
	client config.OAuthClientConfiguration,
	loginIDs []loginid.LoginID,
	plainPassword string,
	metadata map[string]interface{},
	onUserDuplicate model.OnUserDuplicate,
) (Result, error) {
	pr, err := p.Signup.SignupWithLoginIDs(loginIDs, plainPassword, metadata, onUserDuplicate)
	if err != nil {
		return nil, err
	}

	s, err := p.AuthnSession.BeginSession(client, pr.PrincipalUserID(), pr, session.CreateReasonSignup)
	if err != nil {
		return nil, err
	}

	return p.AuthnSession.StepSession(s)
}

func (p *Provider) LoginWithLoginID(
	client config.OAuthClientConfiguration,
	loginID loginid.LoginID,
	plainPassword string,
) (Result, error) {
	pr, err := p.Authn.AuthenticateWithLoginID(loginID, plainPassword)
	if err != nil {
		return nil, err
	}

	s, err := p.AuthnSession.BeginSession(client, pr.PrincipalUserID(), pr, session.CreateReasonLogin)
	if err != nil {
		return nil, err
	}

	return p.AuthnSession.StepSession(s)
}

func (p *Provider) OAuthAuthenticate(
	authInfo sso.AuthInfo,
	codeChallenge string,
	loginState sso.LoginState,
) (*sso.SkygearAuthorizationCode, error) {
	return p.OAuth.Authenticate(authInfo, codeChallenge, loginState)
}

func (p *Provider) OAuthLink(
	authInfo sso.AuthInfo,
	codeChallenge string,
	linkState sso.LinkState,
) (*sso.SkygearAuthorizationCode, error) {
	return p.OAuth.Link(authInfo, codeChallenge, linkState)
}

func (p *Provider) OAuthExchangeCode(
	client config.OAuthClientConfiguration,
	s auth.AuthSession,
	code *sso.SkygearAuthorizationCode,
) (Result, error) {
	pr, err := p.OAuth.ExchangeCode(code)
	if err != nil {
		return nil, err
	}

	if code.Action == "link" {
		if s == nil {
			return nil, authz.ErrNotAuthenticated
		}
		return p.AuthnSession.MakeResult(client, s, "")
	}

	// code.Action == "login"
	reason := session.CreateReason(code.SessionCreateReason)
	as, err := p.AuthnSession.BeginSession(client, pr.PrincipalUserID(), pr, reason)
	if err != nil {
		return nil, err
	}

	return p.AuthnSession.StepSession(as)
}

func (p *Provider) WriteResult(rw http.ResponseWriter, result Result) {
	r, err := result.result()
	if err == nil {
		useCookie := r.Client == nil || r.Client.AuthAPIUseCookie()
		resp := model.AuthResponse{
			User:     *r.User,
			Identity: r.Principal,
		}

		if r.Session != nil {
			resp.SessionID = r.Session.ID
		}
		if r.SessionToken != "" && useCookie {
			p.SessionCookieConfig.WriteTo(rw, r.SessionToken)
		}
		if r.MFABearerToken != "" {
			if useCookie {
				p.BearerTokenCookieConfig.WriteTo(rw, r.MFABearerToken)
			} else {
				resp.MFABearerToken = r.MFABearerToken
			}
		}

		handler.WriteResponse(rw, handler.APIResponse{Result: resp})
	} else {
		handler.WriteResponse(rw, handler.APIResponse{Error: err})
	}
}

func (p *Provider) Resolve(
	client config.OAuthClientConfiguration,
	authnSessionToken string,
	stepPredicate func(SessionStep) bool,
) (*AuthnSession, error) {
	s, err := p.AuthnSession.ResolveSession(authnSessionToken)
	if err != nil {
		return nil, err
	}

	step, ok := s.NextStep()
	if !ok {
		return nil, ErrInvalidAuthenticationSession
	}

	if !stepPredicate(step) {
		return nil, authz.ErrNotAuthenticated
	}

	return s, nil
}

func (p *Provider) StepSession(
	client config.OAuthClientConfiguration,
	s authn.Attributer,
	mfaBearerToken string,
) (Result, error) {
	switch s := s.(type) {
	case *AuthnSession:
		return p.AuthnSession.StepSession(s)
	case *session.IDPSession:
		err := p.Session.Update(s)
		if err != nil {
			return nil, err
		}
		return p.AuthnSession.MakeResult(client, s, mfaBearerToken)
	default:
		panic("authn: unexpected session container type")
	}
}