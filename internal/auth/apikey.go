package auth

import (
	"net/http"
)

type ApiKeyAuth struct{}

func (a *ApiKeyAuth) Type() string               { return "api_key" }
func (a *ApiKeyAuth) NeedsInteractiveAuth() bool  { return false }
func (a *ApiKeyAuth) ApplyAuth(req *http.Request, token Token) {
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
}

func (a *ApiKeyAuth) Authenticate(info ProviderInfo, store *TokenStore) <-chan AuthStatus {
	ch := make(chan AuthStatus, 1)
	go func() {
		defer close(ch)
		ch <- AuthStatus{Stage: StageAuthorized}
	}()
	return ch
}

func (a *ApiKeyAuth) Refresh(info ProviderInfo, token Token, store *TokenStore) (Token, error) {
	return Token{}, ErrNoRefresh
}
