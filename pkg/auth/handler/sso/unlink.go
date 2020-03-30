package sso

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authz"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	authprincipal "github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"

	pkg "github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/sso"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	coreauthz "github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/inject"
	"github.com/skygeario/skygear-server/pkg/core/server"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

func AttachUnlinkHandler(
	router *mux.Router,
	authDependency pkg.DependencyMap,
) {
	router.NewRoute().
		Path("/sso/{provider}/unlink").
		Handler(server.FactoryToHandler(&UnlinkHandlerFactory{
			Dependency: authDependency,
		})).
		Methods("OPTIONS", "POST")
}

type UnlinkHandlerFactory struct {
	Dependency pkg.DependencyMap
}

func (f UnlinkHandlerFactory) NewHandler(request *http.Request) http.Handler {
	h := &UnlinkHandler{}
	inject.DefaultRequestInject(h, f.Dependency, request)
	vars := mux.Vars(request)
	h.ProviderID = vars["provider"]
	return h.RequireAuthz(h, h)
}

/*
	@Operation POST /sso/{provider_id}/unlink - Unlink SSO provider
		Unlink the specified SSO provider from the current user.

		@Tag SSO
		@SecurityRequirement access_key
		@SecurityRequirement access_token

		@Parameter {SSOProviderID}
		@Response 200 {EmptyResponse}

		@Callback identity_delete {UserSyncEvent}
		@Callback user_sync {UserSyncEvent}
*/
type UnlinkHandler struct {
	TxContext         db.TxContext                   `dependency:"TxContext"`
	RequireAuthz      handler.RequireAuthz           `dependency:"RequireAuthz"`
	SessionProvider   session.Provider               `dependency:"SessionProvider"`
	OAuthAuthProvider oauth.Provider                 `dependency:"OAuthAuthProvider"`
	IdentityProvider  authprincipal.IdentityProvider `dependency:"IdentityProvider"`
	AuthInfoStore     authinfo.Store                 `dependency:"AuthInfoStore"`
	UserProfileStore  userprofile.Store              `dependency:"UserProfileStore"`
	HookProvider      hook.Provider                  `dependency:"HookProvider"`
	ProviderFactory   *sso.OAuthProviderFactory      `dependency:"SSOOAuthProviderFactory"`
	ProviderID        string
}

func (h UnlinkHandler) ProvideAuthzPolicy() coreauthz.Policy {
	return authz.AuthAPIRequireValidUser
}

func (h UnlinkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response handler.APIResponse
	var payload struct{}
	if err := handler.DecodeJSONBody(r, w, &payload); err != nil {
		response.Error = err
	} else {
		result, err := h.Handle(r)
		if err != nil {
			response.Error = err
		} else {
			response.Result = result
		}
	}
	handler.WriteResponse(w, response)
}

func (h UnlinkHandler) Handle(r *http.Request) (resp interface{}, err error) {
	err = db.WithTx(h.TxContext, func() error {
		providerConfig, ok := h.ProviderFactory.GetOAuthProviderConfig(h.ProviderID)
		if !ok {
			return skyerr.NewNotFound("unknown SSO provider")
		}

		sess := auth.GetSession(r.Context())
		userID := sess.AuthnAttrs().UserID
		principal, err := h.OAuthAuthProvider.GetPrincipalByUser(oauth.GetByUserOptions{
			ProviderType: string(providerConfig.Type),
			ProviderKeys: oauth.ProviderKeysFromProviderConfig(providerConfig),
			UserID:       userID,
		})
		if err != nil {
			return err
		}

		// principalID can be missing
		principalID := sess.AuthnAttrs().PrincipalID
		if principalID != "" && principalID == principal.ID {
			err = authprincipal.ErrCurrentIdentityBeingDeleted
			return err
		}

		err = h.OAuthAuthProvider.DeletePrincipal(principal)
		if err != nil {
			return err
		}

		sessions, err := h.SessionProvider.List(userID)
		if err != nil {
			return err
		}

		// filter sessions of deleted principal
		n := 0
		for _, session := range sessions {
			if session.PrincipalID == principal.ID {
				sessions[n] = session
				n++
			}
		}
		sessions = sessions[:n]

		err = h.SessionProvider.InvalidateBatch(sessions)
		if err != nil {
			return err
		}

		authInfo := &authinfo.AuthInfo{}
		if err := h.AuthInfoStore.GetAuth(userID, authInfo); err != nil {
			return err
		}

		var userProfile userprofile.UserProfile
		userProfile, err = h.UserProfileStore.GetUserProfile(userID)
		if err != nil {
			return err
		}

		user := model.NewUser(*authInfo, userProfile)
		identity := model.NewIdentity(h.IdentityProvider, principal)
		err = h.HookProvider.DispatchEvent(
			event.IdentityDeleteEvent{
				User:     user,
				Identity: identity,
			},
			&user,
		)
		if err != nil {
			return err
		}

		resp = struct{}{}
		return nil
	})
	return
}
