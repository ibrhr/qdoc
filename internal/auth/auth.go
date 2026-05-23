package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ibrhr/qdoc/internal/provider"
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

type AuthMethod interface {
	Type() string

	NeedsInteractiveAuth() bool

	Authenticate(prov provider.Provider, store *TokenStore) <-chan AuthStatus

	Refresh(prov provider.Provider, token Token, store *TokenStore) (Token, error)

	ApplyAuth(req *http.Request, token Token)
}

func ForProvider(prov provider.Provider) (AuthMethod, error) {
	switch prov.EffectiveAuthType() {
	case "api_key":
		return &ApiKeyAuth{}, nil
	case "oauth_device":
		if prov.OAuthConfig == nil {
			return nil, errors.New("oauth_device provider missing oauth_config")
		}
		return &DeviceFlowAuth{
			DeviceAuthURL: prov.OAuthConfig.DeviceAuthURL,
			TokenURL:      prov.OAuthConfig.TokenURL,
			ClientID:      prov.OAuthConfig.ClientID,
			Scope:         prov.OAuthConfig.Scope,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", prov.EffectiveAuthType())
	}
}
