// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package handler

import (
	"github.com/skygeario/skygear-server/pkg/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	auth2 "github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	redis3 "github.com/skygeario/skygear-server/pkg/auth/dependency/auth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/bearertoken"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/oob"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/recoverycode"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/authenticator/totp"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/challenge"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/anonymous"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/adaptors"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/flows"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction/redis"
	oauth2 "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	handler2 "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/handler"
	pq3 "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/pq"
	redis2 "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oidc"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory/pq"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	redis4 "github.com/skygeario/skygear-server/pkg/auth/dependency/session/redis"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/async"
	pq2 "github.com/skygeario/skygear-server/pkg/core/auth/authinfo/pq"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
	"net/http"
)

// Injectors from wire.go:

func newLoginHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	requireAuthz := handler.NewRequireAuthzFactory(factory)
	validator := auth.ProvideValidator(m)
	provider := time.NewProvider()
	store := redis.ProvideStore(context, tenantConfiguration, provider)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration, reservedNameChecker)
	oauthProvider := oauth.ProvideProvider(sqlBuilder, sqlExecutor, provider)
	anonymousProvider := anonymous.ProvideProvider(sqlBuilder, sqlExecutor)
	identityAdaptor := &adaptors.IdentityAdaptor{
		LoginID:   loginidProvider,
		OAuth:     oauthProvider,
		Anonymous: anonymousProvider,
	}
	passwordhistoryStore := pq.ProvidePasswordHistoryStore(provider, sqlBuilder, sqlExecutor)
	passwordChecker := audit.ProvidePasswordChecker(tenantConfiguration, passwordhistoryStore)
	passwordProvider := password.ProvideProvider(sqlBuilder, sqlExecutor, provider, factory, passwordhistoryStore, passwordChecker, tenantConfiguration)
	totpProvider := totp.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	engine := auth.ProvideTemplateEngine(tenantConfiguration, m)
	urlprefixProvider := urlprefix.NewProvider(r)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	executor := auth.ProvideTaskExecutor(m)
	queue := async.ProvideTaskQueue(context, txContext, requestID, tenantConfiguration, executor)
	oobProvider := oob.ProvideProvider(tenantConfiguration, sqlBuilder, sqlExecutor, provider, engine, urlprefixProvider, queue)
	bearertokenProvider := bearertoken.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	recoverycodeProvider := recoverycode.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	authenticatorAdaptor := &adaptors.AuthenticatorAdaptor{
		Password:     passwordProvider,
		TOTP:         totpProvider,
		OOBOTP:       oobProvider,
		BearerToken:  bearertokenProvider,
		RecoveryCode: recoverycodeProvider,
	}
	authinfoStore := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, requestID, tenantConfiguration, txContext, provider, authinfoStore, userprofileStore, loginidProvider, factory)
	userProvider := interaction.ProvideUserProvider(authinfoStore, userprofileStore, provider, hookProvider, urlprefixProvider, queue, tenantConfiguration)
	interactionProvider := interaction.ProvideProvider(store, provider, factory, identityAdaptor, authenticatorAdaptor, userProvider, oobProvider, tenantConfiguration, hookProvider)
	authorizationStore := &pq3.AuthorizationStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	grantStore := redis2.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	eventStore := redis3.ProvideEventStore(context, tenantConfiguration)
	accessEventProvider := auth2.AccessEventProvider{
		Store: eventStore,
	}
	sessionStore := redis4.ProvideStore(context, tenantConfiguration, provider, factory)
	authAccessEventProvider := &auth2.AccessEventProvider{
		Store: eventStore,
	}
	sessionProvider := session.ProvideSessionProvider(r, sessionStore, authAccessEventProvider, tenantConfiguration)
	challengeProvider := challenge.ProvideProvider(context, provider, tenantConfiguration)
	anonymousFlow := &flows.AnonymousFlow{
		Interactions: interactionProvider,
		Anonymous:    anonymousProvider,
		Challenges:   challengeProvider,
	}
	idTokenIssuer := oidc.ProvideIDTokenIssuer(tenantConfiguration, urlprefixProvider, authinfoStore, userprofileStore, provider)
	tokenGenerator := _wireTokenGeneratorValue
	tokenHandler := handler2.ProvideTokenHandler(r, tenantConfiguration, factory, authorizationStore, grantStore, grantStore, grantStore, accessEventProvider, sessionProvider, anonymousFlow, idTokenIssuer, tokenGenerator, provider)
	insecureCookieConfig := auth.ProvideSessionInsecureCookieConfig(m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	userController := flows.ProvideUserController(authinfoStore, userprofileStore, tokenHandler, cookieConfiguration, sessionProvider, hookProvider, provider, tenantConfiguration)
	authAPIFlow := &flows.AuthAPIFlow{
		Interactions:   interactionProvider,
		UserController: userController,
	}
	httpHandler := provideLoginHandler(requireAuthz, validator, authAPIFlow, txContext)
	return httpHandler
}

var (
	_wireTokenGeneratorValue = handler2.TokenGenerator(oauth2.GenerateToken)
)

func newSignupHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	requireAuthz := handler.NewRequireAuthzFactory(factory)
	validator := auth.ProvideValidator(m)
	provider := time.NewProvider()
	store := redis.ProvideStore(context, tenantConfiguration, provider)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration, reservedNameChecker)
	oauthProvider := oauth.ProvideProvider(sqlBuilder, sqlExecutor, provider)
	anonymousProvider := anonymous.ProvideProvider(sqlBuilder, sqlExecutor)
	identityAdaptor := &adaptors.IdentityAdaptor{
		LoginID:   loginidProvider,
		OAuth:     oauthProvider,
		Anonymous: anonymousProvider,
	}
	passwordhistoryStore := pq.ProvidePasswordHistoryStore(provider, sqlBuilder, sqlExecutor)
	passwordChecker := audit.ProvidePasswordChecker(tenantConfiguration, passwordhistoryStore)
	passwordProvider := password.ProvideProvider(sqlBuilder, sqlExecutor, provider, factory, passwordhistoryStore, passwordChecker, tenantConfiguration)
	totpProvider := totp.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	engine := auth.ProvideTemplateEngine(tenantConfiguration, m)
	urlprefixProvider := urlprefix.NewProvider(r)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	executor := auth.ProvideTaskExecutor(m)
	queue := async.ProvideTaskQueue(context, txContext, requestID, tenantConfiguration, executor)
	oobProvider := oob.ProvideProvider(tenantConfiguration, sqlBuilder, sqlExecutor, provider, engine, urlprefixProvider, queue)
	bearertokenProvider := bearertoken.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	recoverycodeProvider := recoverycode.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	authenticatorAdaptor := &adaptors.AuthenticatorAdaptor{
		Password:     passwordProvider,
		TOTP:         totpProvider,
		OOBOTP:       oobProvider,
		BearerToken:  bearertokenProvider,
		RecoveryCode: recoverycodeProvider,
	}
	authinfoStore := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, requestID, tenantConfiguration, txContext, provider, authinfoStore, userprofileStore, loginidProvider, factory)
	userProvider := interaction.ProvideUserProvider(authinfoStore, userprofileStore, provider, hookProvider, urlprefixProvider, queue, tenantConfiguration)
	interactionProvider := interaction.ProvideProvider(store, provider, factory, identityAdaptor, authenticatorAdaptor, userProvider, oobProvider, tenantConfiguration, hookProvider)
	authorizationStore := &pq3.AuthorizationStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	grantStore := redis2.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	eventStore := redis3.ProvideEventStore(context, tenantConfiguration)
	accessEventProvider := auth2.AccessEventProvider{
		Store: eventStore,
	}
	sessionStore := redis4.ProvideStore(context, tenantConfiguration, provider, factory)
	authAccessEventProvider := &auth2.AccessEventProvider{
		Store: eventStore,
	}
	sessionProvider := session.ProvideSessionProvider(r, sessionStore, authAccessEventProvider, tenantConfiguration)
	challengeProvider := challenge.ProvideProvider(context, provider, tenantConfiguration)
	anonymousFlow := &flows.AnonymousFlow{
		Interactions: interactionProvider,
		Anonymous:    anonymousProvider,
		Challenges:   challengeProvider,
	}
	idTokenIssuer := oidc.ProvideIDTokenIssuer(tenantConfiguration, urlprefixProvider, authinfoStore, userprofileStore, provider)
	tokenGenerator := _wireTokenGeneratorValue
	tokenHandler := handler2.ProvideTokenHandler(r, tenantConfiguration, factory, authorizationStore, grantStore, grantStore, grantStore, accessEventProvider, sessionProvider, anonymousFlow, idTokenIssuer, tokenGenerator, provider)
	insecureCookieConfig := auth.ProvideSessionInsecureCookieConfig(m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	userController := flows.ProvideUserController(authinfoStore, userprofileStore, tokenHandler, cookieConfiguration, sessionProvider, hookProvider, provider, tenantConfiguration)
	authAPIFlow := &flows.AuthAPIFlow{
		Interactions:   interactionProvider,
		UserController: userController,
	}
	httpHandler := provideSignupHandler(requireAuthz, validator, authAPIFlow, txContext)
	return httpHandler
}

func newLogoutHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	requireAuthz := handler.NewRequireAuthzFactory(factory)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	store := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	provider := time.NewProvider()
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration, reservedNameChecker)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, requestID, tenantConfiguration, txContext, provider, store, userprofileStore, loginidProvider, factory)
	sessionStore := redis4.ProvideStore(context, tenantConfiguration, provider, factory)
	insecureCookieConfig := auth.ProvideSessionInsecureCookieConfig(m)
	cookieConfiguration := session.ProvideSessionCookieConfiguration(r, insecureCookieConfig, tenantConfiguration)
	manager := session.ProvideSessionManager(sessionStore, provider, tenantConfiguration, cookieConfiguration)
	grantStore := redis2.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	sessionManager := &oauth2.SessionManager{
		Store: grantStore,
		Time:  provider,
	}
	authSessionManager := &auth2.SessionManager{
		AuthInfoStore:       store,
		UserProfileStore:    userprofileStore,
		Hooks:               hookProvider,
		IDPSessions:         manager,
		AccessTokenSessions: sessionManager,
	}
	httpHandler := provideLogoutHandler(requireAuthz, authSessionManager, txContext)
	return httpHandler
}

func newRefreshHandler(r *http.Request, m auth.DependencyMap) http.Handler {
	context := auth.ProvideContext(r)
	requestID := auth.ProvideLoggingRequestID(r)
	tenantConfiguration := auth.ProvideTenantConfig(context, m)
	factory := logging.ProvideLoggerFactory(context, requestID, tenantConfiguration)
	requireAuthz := handler.NewRequireAuthzFactory(factory)
	validator := auth.ProvideValidator(m)
	sqlBuilderFactory := db.ProvideSQLBuilderFactory(tenantConfiguration)
	sqlBuilder := auth.ProvideAuthSQLBuilder(sqlBuilderFactory)
	sqlExecutor := db.ProvideSQLExecutor(context, tenantConfiguration)
	authorizationStore := &pq3.AuthorizationStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	provider := time.NewProvider()
	grantStore := redis2.ProvideGrantStore(context, factory, tenantConfiguration, sqlBuilder, sqlExecutor, provider)
	eventStore := redis3.ProvideEventStore(context, tenantConfiguration)
	accessEventProvider := auth2.AccessEventProvider{
		Store: eventStore,
	}
	store := redis4.ProvideStore(context, tenantConfiguration, provider, factory)
	authAccessEventProvider := &auth2.AccessEventProvider{
		Store: eventStore,
	}
	sessionProvider := session.ProvideSessionProvider(r, store, authAccessEventProvider, tenantConfiguration)
	redisStore := redis.ProvideStore(context, tenantConfiguration, provider)
	reservedNameChecker := auth.ProvideReservedNameChecker(m)
	loginidProvider := loginid.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration, reservedNameChecker)
	oauthProvider := oauth.ProvideProvider(sqlBuilder, sqlExecutor, provider)
	anonymousProvider := anonymous.ProvideProvider(sqlBuilder, sqlExecutor)
	identityAdaptor := &adaptors.IdentityAdaptor{
		LoginID:   loginidProvider,
		OAuth:     oauthProvider,
		Anonymous: anonymousProvider,
	}
	passwordhistoryStore := pq.ProvidePasswordHistoryStore(provider, sqlBuilder, sqlExecutor)
	passwordChecker := audit.ProvidePasswordChecker(tenantConfiguration, passwordhistoryStore)
	passwordProvider := password.ProvideProvider(sqlBuilder, sqlExecutor, provider, factory, passwordhistoryStore, passwordChecker, tenantConfiguration)
	totpProvider := totp.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	engine := auth.ProvideTemplateEngine(tenantConfiguration, m)
	urlprefixProvider := urlprefix.NewProvider(r)
	txContext := db.ProvideTxContext(context, tenantConfiguration)
	executor := auth.ProvideTaskExecutor(m)
	queue := async.ProvideTaskQueue(context, txContext, requestID, tenantConfiguration, executor)
	oobProvider := oob.ProvideProvider(tenantConfiguration, sqlBuilder, sqlExecutor, provider, engine, urlprefixProvider, queue)
	bearertokenProvider := bearertoken.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	recoverycodeProvider := recoverycode.ProvideProvider(sqlBuilder, sqlExecutor, provider, tenantConfiguration)
	authenticatorAdaptor := &adaptors.AuthenticatorAdaptor{
		Password:     passwordProvider,
		TOTP:         totpProvider,
		OOBOTP:       oobProvider,
		BearerToken:  bearertokenProvider,
		RecoveryCode: recoverycodeProvider,
	}
	authinfoStore := pq2.ProvideStore(sqlBuilderFactory, sqlExecutor)
	userprofileStore := userprofile.ProvideStore(provider, sqlBuilder, sqlExecutor)
	hookProvider := hook.ProvideHookProvider(context, sqlBuilder, sqlExecutor, requestID, tenantConfiguration, txContext, provider, authinfoStore, userprofileStore, loginidProvider, factory)
	userProvider := interaction.ProvideUserProvider(authinfoStore, userprofileStore, provider, hookProvider, urlprefixProvider, queue, tenantConfiguration)
	interactionProvider := interaction.ProvideProvider(redisStore, provider, factory, identityAdaptor, authenticatorAdaptor, userProvider, oobProvider, tenantConfiguration, hookProvider)
	challengeProvider := challenge.ProvideProvider(context, provider, tenantConfiguration)
	anonymousFlow := &flows.AnonymousFlow{
		Interactions: interactionProvider,
		Anonymous:    anonymousProvider,
		Challenges:   challengeProvider,
	}
	idTokenIssuer := oidc.ProvideIDTokenIssuer(tenantConfiguration, urlprefixProvider, authinfoStore, userprofileStore, provider)
	tokenGenerator := _wireTokenGeneratorValue
	tokenHandler := handler2.ProvideTokenHandler(r, tenantConfiguration, factory, authorizationStore, grantStore, grantStore, grantStore, accessEventProvider, sessionProvider, anonymousFlow, idTokenIssuer, tokenGenerator, provider)
	httpHandler := provideRefreshHandler(requireAuthz, validator, tokenHandler, txContext)
	return httpHandler
}

// wire.go:

func provideLoginHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	f LoginInteractionFlow,
	tx db.TxContext,
) http.Handler {
	h := &LoginHandler{
		Validator:    v,
		Interactions: f,
		TxContext:    tx,
	}
	return requireAuthz(h, h)
}

func provideSignupHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	f SignupInteractionFlow,
	tx db.TxContext,
) http.Handler {
	h := &SignupHandler{
		Validator:    v,
		Interactions: f,
		TxContext:    tx,
	}
	return requireAuthz(h, h)
}

func provideLogoutHandler(
	requireAuthz handler.RequireAuthz,
	sm logoutSessionManager,
	tx db.TxContext,
) http.Handler {
	h := &LogoutHandler{
		SessionManager: sm,
		TxContext:      tx,
	}
	return requireAuthz(h, h)
}

func provideRefreshHandler(
	requireAuthz handler.RequireAuthz,
	v *validation.Validator,
	rp refreshProvider,
	tx db.TxContext,
) http.Handler {
	h := &RefreshHandler{
		validator:       v,
		refreshProvider: rp,
		txContext:       tx,
	}
	return requireAuthz(h, h)
}
