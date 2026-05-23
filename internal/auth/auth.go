package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Scope        string    `json:"scope,omitempty"`
}

func (t Token) IsZero() bool { return t.AccessToken == "" }

type AuthStage string

const (
	StageAwaitingUser AuthStage = "awaiting_user"
	StagePolling      AuthStage = "polling"
	StageAuthorized   AuthStage = "authorized"
	StageError        AuthStage = "error"
	StageTimeout      AuthStage = "timeout"
)

type AuthStatus struct {
	Stage   AuthStage
	Message string
	Token   *Token
	Err     error

	VerificationURI string
	UserCode        string
}

var (
	ErrNoToken   = errors.New("no token available")
	ErrNoRefresh = errors.New("auth method does not support refresh")
	ErrExpired   = errors.New("token expired or revoked")
)

type OAuthConfig struct {
	DeviceAuthURL string `json:"device_auth_url,omitempty"`
	TokenURL      string `json:"token_url,omitempty"`
	AuthURL       string `json:"auth_url,omitempty"`
	ClientID      string `json:"client_id,omitempty"`
	Scope         string `json:"scope,omitempty"`
}

type ProviderInfo struct {
	Name        string
	AuthType    string
	OAuthConfig *OAuthConfig
	EnvKey      string
}

type AuthMethod interface {
	Type() string

	NeedsInteractiveAuth() bool

	Authenticate(info ProviderInfo, store *TokenStore) <-chan AuthStatus

	Refresh(info ProviderInfo, token Token, store *TokenStore) (Token, error)

	ApplyAuth(req *http.Request, token Token)
}

func ForProvider(info ProviderInfo) (AuthMethod, error) {
	if info.AuthType == "" {
		info.AuthType = "api_key"
	}
	switch info.AuthType {
	case "api_key":
		return &ApiKeyAuth{}, nil
	case "oauth_device":
		if info.OAuthConfig == nil {
			return nil, errors.New("oauth_device provider missing oauth_config")
		}
		return &DeviceFlowAuth{
			DeviceAuthURL: info.OAuthConfig.DeviceAuthURL,
			TokenURL:      info.OAuthConfig.TokenURL,
			ClientID:      info.OAuthConfig.ClientID,
			Scope:         info.OAuthConfig.Scope,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", info.AuthType)
	}
}
