package config

import (
	"fmt"
	"github.com/skygeario/skygear-server/pkg/validation"
	"strings"
)

var _ = SecretConfigSchema.Add("DatabaseCredentials", `
{
	"type": "object",
	"properties": {
		"database_url": { "type": "string" },
		"database_schema": { "type": "string" }
	},
	"required": ["database_url"]
}
`)

type DatabaseCredentials struct {
	DatabaseURL    string `json:"database_url,omitempty"`
	DatabaseSchema string `json:"database_schema,omitempty"`
}

func (c *DatabaseCredentials) SensitiveStrings() []string {
	return []string{
		c.DatabaseURL,
		c.DatabaseSchema,
	}
}

func (c *DatabaseCredentials) SetDefaults() {
	if c.DatabaseSchema == "" {
		c.DatabaseSchema = "public"
	}
}

var _ = SecretConfigSchema.Add("RedisCredentials", `
{
	"type": "object",
	"properties": {
		"host": { "type": "string" },
		"port": { "type": "integer" },
		"password": { "type": "string" },
		"db": { "type": "integer" },
		"sentinel": { "$ref": "#/$defs/RedisSentinelConfig" }
	}
}
`)

type RedisCredentials struct {
	Host     string               `json:"host,omitempty"`
	Port     int                  `json:"port,omitempty"`
	Password string               `json:"password,omitempty"`
	DB       int                  `json:"db,omitempty"`
	Sentinel *RedisSentinelConfig `json:"sentinel,omitempty"`
}

func (c *RedisCredentials) SensitiveStrings() []string {
	return append([]string{
		c.Host,
		c.Password,
		c.Sentinel.MasterName,
	}, c.Sentinel.Addrs...)
}

func (c *RedisCredentials) SetDefaults() {
	if c.Port == 0 {
		c.Port = 6379
	}
}

func (c *RedisCredentials) Validate(ctx *validation.Context) {
	if c.Sentinel.Enabled {
		if len(c.Sentinel.Addrs) == 0 {
			ctx.Child("sentinel", "addrs").EmitErrorMessage("redis sentinel addrs are not provided")
		}
	} else {
		if c.Host == "" {
			ctx.Child("host").EmitErrorMessage("redis host is not provided")
		}
	}
}

func (c *RedisCredentials) ConnKey() string {
	if c.Sentinel.Enabled {
		return fmt.Sprintf("sentinel/%s;%s/%d",
			c.Sentinel.MasterName,
			strings.Join(c.Sentinel.Addrs, ";"),
			c.DB,
		)
	}
	return fmt.Sprintf("redis/%s:%s/%d",
		c.Host,
		c.Password,
		c.DB,
	)
}

var _ = SecretConfigSchema.Add("RedisSentinelConfig", `
{
	"type": "object",
	"properties": {
		"enabled": { "type": "boolean" },
		"addrs": { "type": "array", "items": { "type": "string" } },
		"master_name": { "type": "string" }
	}
}
`)

type RedisSentinelConfig struct {
	Enabled    bool     `json:"enabled,omitempty"`
	Addrs      []string `json:"addrs,omitempty"`
	MasterName string   `json:"master_name,omitempty"`
}

var _ = SecretConfigSchema.Add("SMTPMode", `
{
	"type": "string",
	"enum": ["normal", "ssl"]
}
`)

type SMTPMode string

const (
	SMTPModeNormal SMTPMode = "normal"
	SMTPModeSSL    SMTPMode = "ssl"
)

var _ = SecretConfigSchema.Add("SMTPServerCredentials", `
{
	"type": "object",
	"properties": {
		"host": { "type": "string" },
		"port": { "type": "integer", "minimum": 1, "maximum": 65535 },
		"mode": { "$ref": "#/$defs/SMTPMode" },
		"username": { "type": "string" },
		"password": { "type": "string" }
	},
	"required": ["host", "port"]
}
`)

type SMTPServerCredentials struct {
	Host     string   `json:"host,omitempty"`
	Port     int      `json:"port,omitempty"`
	Mode     SMTPMode `json:"mode,omitempty"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
}

func (c *SMTPServerCredentials) SensitiveStrings() []string {
	return []string{
		c.Host,
		c.Username,
		c.Password,
	}
}

func (c *SMTPServerCredentials) SetDefaults() {
	if c.Mode == "" {
		c.Mode = SMTPModeNormal
	}
}

var _ = SecretConfigSchema.Add("TwilioCredentials", `
{
	"type": "object",
	"properties": {
		"account_sid": { "type": "string" },
		"auth_token": { "type": "string" }
	},
	"required": ["account_sid", "auth_token"]
}
`)

type TwilioCredentials struct {
	AccountSID string `json:"account_sid,omitempty"`
	AuthToken  string `json:"auth_token,omitempty"`
}

func (c *TwilioCredentials) SensitiveStrings() []string {
	return []string{
		c.AccountSID,
		c.AuthToken,
	}
}

var _ = SecretConfigSchema.Add("NexmoCredentials", `
{
	"type": "object",
	"properties": {
		"api_key": { "type": "string" },
		"api_secret": { "type": "string" }
	},
	"required": ["api_key", "api_secret"]
}
`)

type NexmoCredentials struct {
	APIKey    string `json:"api_key,omitempty"`
	APISecret string `json:"api_secret,omitempty"`
}

func (c *NexmoCredentials) SensitiveStrings() []string {
	return []string{
		c.APIKey,
		c.APISecret,
	}
}

var _ = SecretConfigSchema.Add("JWK", `
{
	"type": "object",
	"properties": {
		"kid": { "type": "string" },
		"kty": { "type": "string" }
	},
	"required": ["kid", "kty"]
}
`)

var _ = SecretConfigSchema.Add("JWS", `
{
	"type": "object",
	"properties": {
		"keys": {
			"type": "array",
			"items": { "$ref": "#/$defs/JWK" },
			"minItems": 1
		}
	},
	"required": ["keys"]
}
`)

var _ = SecretConfigSchema.Add("JWTKeyMaterials", `{ "$ref": "#/$defs/JWS" }`)

type JWTKeyMaterials struct {
	Keys []interface{} `json:"keys"`
}

func (c *JWTKeyMaterials) SensitiveStrings() []string {
	return extractJWKSecrets(c.Keys)
}

var _ = SecretConfigSchema.Add("OIDCKeyMaterials", `{ "$ref": "#/$defs/JWS" }`)

type OIDCKeyMaterials struct {
	Keys []interface{} `json:"keys"`
}

func (c *OIDCKeyMaterials) SensitiveStrings() []string {
	return extractJWKSecrets(c.Keys)
}

var _ = SecretConfigSchema.Add("CSRFKeyMaterials", `{ "$ref": "#/$defs/JWS" }`)

type CSRFKeyMaterials struct {
	Keys []interface{} `json:"keys"`
}

func (c *CSRFKeyMaterials) SensitiveStrings() []string {
	return extractJWKSecrets(c.Keys)
}

var _ = SecretConfigSchema.Add("WebhookKeyMaterials", `{ "$ref": "#/$defs/JWS" }`)

type WebhookKeyMaterials struct {
	Keys []interface{} `json:"keys"`
}

func (c *WebhookKeyMaterials) SensitiveStrings() []string {
	return extractJWKSecrets(c.Keys)
}