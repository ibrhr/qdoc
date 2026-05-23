package auth

import (
	"net/http"

	"github.com/ibrhr/qdoc/internal/provider"
)

type ApiKeyAuth struct{}

func (a *ApiKeyAuth) Type() string               { return "api_key" }
func (a *ApiKeyAuth) NeedsInteractiveAuth() bool  { return false }
func (a *ApiKeyAuth) ApplyAuth(req *http.Request, token Token) {
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
}

func (a *ApiKeyAuth) Authenticate(prov provider.Provider, store *TokenStore) <-chan AuthStatus {
	ch := make(chan AuthStatus, 1)
	go func() {
		defer close(ch)
		ch <- AuthStatus{Stage: StageAuthorized}
	}()
	return ch
}

func (a *ApiKeyAuth) Refresh(prov provider.Provider, token Token, store *TokenStore) (Token, error) {
	return Token{}, ErrNoRefresh
}
