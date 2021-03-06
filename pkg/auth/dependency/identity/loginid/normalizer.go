package loginid

import (
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/text/secure/precis"
	"golang.org/x/text/unicode/norm"

	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type Normalizer interface {
	Normalize(loginID string) (string, error)
	ComputeUniqueKey(normalizeLoginID string) (string, error)
}

type NormalizerFactory struct {
	Keys  []config.LoginIDKeyConfiguration
	Types *config.LoginIDTypesConfiguration
}

func (f *NormalizerFactory) NormalizerWithLoginIDKey(loginIDKey string) Normalizer {
	for _, c := range f.Keys {
		if c.Key == loginIDKey {
			return f.NormalizerWithLoginIDType(c.Type)
		}
	}

	panic("loginid: invalid login id key: " + loginIDKey)
}

func (f *NormalizerFactory) NormalizerWithLoginIDType(loginIDKeyType config.LoginIDKeyType) Normalizer {
	metadataKey, _ := loginIDKeyType.MetadataKey()
	switch metadataKey {
	case metadata.Email:
		return &EmailNormalizer{
			Config: f.Types.Email,
		}
	case metadata.Username:
		return &UsernameNormalizer{
			Config: f.Types.Username,
		}
	}

	return &NullNormalizer{}
}

type EmailNormalizer struct {
	Config *config.LoginIDTypeEmailConfiguration
}

func (n *EmailNormalizer) Normalize(loginID string) (string, error) {
	// refs from stdlib
	// https://golang.org/src/net/mail/message.go?s=5217:5250#L172
	at := strings.LastIndex(loginID, "@")
	if at < 0 {
		panic("loginid: malformed address, should be rejected by the email format checker")
	}
	local, domain := loginID[:at], loginID[at+1:]

	// convert the domain part
	var err error
	p := precis.NewFreeform(precis.FoldCase())
	domain, err = p.String(domain)
	if err != nil {
		return "", errors.HandledWithMessage(err, "failed to case fold email")
	}

	// convert the local part
	local = norm.NFKC.String(local)

	if !*n.Config.CaseSensitive {
		local, err = p.String(local)
		if err != nil {
			return "", errors.HandledWithMessage(err, "failed to case fold email")
		}
	}

	if *n.Config.IgnoreDotSign {
		local = strings.Replace(local, ".", "", -1)
	}

	return local + "@" + domain, nil
}

func (n *EmailNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	at := strings.LastIndex(normalizeLoginID, "@")
	if at < 0 {
		panic("loginid: malformed address, should be rejected by the email format checker")
	}
	local, domain := normalizeLoginID[:at], normalizeLoginID[at+1:]
	punycode, err := idna.ToASCII(domain)
	if err != nil {
		return "", err
	}
	domain = punycode
	return local + "@" + domain, nil
}

type UsernameNormalizer struct {
	Config *config.LoginIDTypeUsernameConfiguration
}

func (n *UsernameNormalizer) Normalize(loginID string) (string, error) {
	loginID = norm.NFKC.String(loginID)

	var err error
	if !*n.Config.CaseSensitive {
		p := precis.NewIdentifier(precis.FoldCase())
		loginID, err = p.String(loginID)
		if err != nil {
			return "", errors.HandledWithMessage(err, "failed to case fold username")
		}
	}

	return loginID, nil
}

func (n *UsernameNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	return normalizeLoginID, nil
}

type NullNormalizer struct{}

func (n *NullNormalizer) Normalize(loginID string) (string, error) {
	return loginID, nil
}

func (n *NullNormalizer) ComputeUniqueKey(normalizeLoginID string) (string, error) {
	return normalizeLoginID, nil
}
