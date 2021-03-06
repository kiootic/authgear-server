package sso

import (
	"github.com/google/wire"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func ProvideStateCodec(config *config.TenantConfiguration) *StateCodec {
	return NewStateCodec(
		config.AppID,
		config.AppConfig.Identity.OAuth,
	)
}

func ProvideOAuthProviderFactory(
	cfg *config.TenantConfiguration,
	up urlprefix.Provider,
	tp time.Provider,
	nf *loginid.NormalizerFactory,
	rf RedirectURLFunc,
) *OAuthProviderFactory {
	return NewOAuthProviderFactory(*cfg, up, tp, NewUserInfoDecoder(nf), nf, rf)
}

var DependencySet = wire.NewSet(
	ProvideStateCodec,
	ProvideOAuthProviderFactory,
)
