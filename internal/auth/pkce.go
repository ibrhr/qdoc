package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type PkceAuth struct {
	AuthURL     string
	TokenURL    string
	ClientID    string
	Scope       string
	RedirectURI string
	Port        int
}

func (a *PkceAuth) Type() string               { return "oauth_pkce" }
func (a *PkceAuth) NeedsInteractiveAuth() bool  { return true }
func (a *PkceAuth) ApplyAuth(req *http.Request, token Token) {
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
}

func (a *PkceAuth) Refresh(info ProviderInfo, token Token, store *TokenStore) (Token, error) {
	if token.RefreshToken == "" {
		return Token{}, ErrNoRefresh
	}

	form := url.Values{}
	form.Set("client_id", a.ClientID)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", token.RefreshToken)

	req, err := http.NewRequest("POST", a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := deviceHTTPClient.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("refreshing token: %w", err)
	}
	defer resp.Body.Close()

	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return Token{}, fmt.Errorf("parsing refresh response: %w", err)
	}

	if tr.Error != "" {
		return Token{}, &deviceError{ErrorCode: tr.Error, ErrorMessage: tr.ErrorDesc}
	}

	tkn := Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		TokenType:    tr.TokenType,
		Scope:        tr.Scope,
		Extra:        token.Extra,
	}
	if tkn.TokenType == "" {
		tkn.TokenType = "bearer"
	}
	if tr.ExpiresIn > 0 {
		tkn.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	if store != nil {
		tsKey := info.Name
		store.Set(tsKey, tkn)
	}

	return tkn, nil
}

func (a *PkceAuth) Authenticate(info ProviderInfo, store *TokenStore) <-chan AuthStatus {
	ch := make(chan AuthStatus, 1)

	go func() {
		defer close(ch)

		codeVerifier, err := generateCodeVerifier()
		if err != nil {
			ch <- AuthStatus{Stage: StageError, Err: fmt.Errorf("generating code verifier: %w", err)}
			return
		}

		codeChallenge := generateCodeChallenge(codeVerifier)
		state, err := generateRandomString(32)
		if err != nil {
			ch <- AuthStatus{Stage: StageError, Err: fmt.Errorf("generating state: %w", err)}
			return
		}

		authURL := buildAuthURL(a.AuthURL, a.ClientID, a.RedirectURI, a.Scope, codeChallenge, state)
		if authURL == "" {
			ch <- AuthStatus{Stage: StageError, Err: errors.New("failed to build authorization URL")}
			return
		}

		ch <- AuthStatus{
			Stage:           StageAwaitingUser,
			Message:         "Opening browser for authorization...",
			VerificationURI: authURL,
			UserCode:        "",
		}

		codeCh := make(chan string, 1)
		errCh := make(chan error, 1)

		srv := startCallbackServer(a.Port, codeCh, errCh, state)
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			srv.Shutdown(ctx)
		}()

		ch <- AuthStatus{
			Stage:   StagePolling,
			Message: "Waiting for authorization...",
		}

		select {
		case code := <-codeCh:
			if code == "" {
				ch <- AuthStatus{Stage: StageTimeout, Message: "Authorization timed out", Err: errors.New("no authorization code received")}
				return
			}

			token, err := a.exchangeCode(code, codeVerifier)
			if err != nil {
				ch <- AuthStatus{Stage: StageError, Err: fmt.Errorf("exchanging code: %w", err)}
				return
			}

			if store != nil {
				tsKey := info.Name
				store.Set(tsKey, token)
			}

			ch <- AuthStatus{Stage: StageAuthorized, Token: &token}

		case err := <-errCh:
			ch <- AuthStatus{Stage: StageError, Err: err}

		case <-time.After(5 * time.Minute):
			ch <- AuthStatus{
				Stage:   StageTimeout,
				Message: "Authorization timed out",
				Err:     errors.New("authorization timed out after 5 minutes"),
			}
		}
	}()

	return ch
}

func (a *PkceAuth) exchangeCode(code, codeVerifier string) (Token, error) {
	form := url.Values{}
	form.Set("client_id", a.ClientID)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", a.RedirectURI)

	req, err := http.NewRequest("POST", a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := deviceHTTPClient.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("token exchange: %w", err)
	}
	defer resp.Body.Close()

	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return Token{}, fmt.Errorf("parsing token response: %w", err)
	}

	if tr.Error != "" {
		return Token{}, &deviceError{ErrorCode: tr.Error, ErrorMessage: tr.ErrorDesc}
	}

	if tr.AccessToken == "" {
		return Token{}, errors.New("empty access token in response")
	}

	tkn := Token{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		TokenType:    tr.TokenType,
	}
	if tkn.TokenType == "" {
		tkn.TokenType = "bearer"
	}

	if tr.ExpiresIn > 0 {
		tkn.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	if tr.IDToken != "" || tr.RefreshToken != "" {
		tkn.Extra = make(map[string]string)
		if tr.IDToken != "" {
			tkn.Extra["id_token"] = tr.IDToken
			accountID := extractSubFromJWT(tr.IDToken)
			if accountID != "" {
				tkn.Extra["account_id"] = accountID
			}
		}
	}

	return tkn, nil
}

func generateCodeVerifier() (string, error) {
	return generateRandomString(128)
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func extractSubFromJWT(idToken string) string {
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}

	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}

	return claims.Sub
}

func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

func buildAuthURL(authURL, clientID, redirectURI, scope, codeChallenge, state string) string {
	u, err := url.Parse(authURL)
	if err != nil {
		return ""
	}
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	if scope != "" {
		q.Set("scope", scope)
	}
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

func startCallbackServer(port int, codeCh chan<- string, errCh chan<- error, expectedState string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		state := q.Get("state")
		if state != expectedState {
			errCh <- errors.New("state mismatch in OAuth callback")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid state parameter"))
			return
		}

		code := q.Get("code")
		if code == "" {
			errCh <- errors.New("no authorization code in callback")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing authorization code"))
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>qdoc — Authorization Complete</title></head>
<body style="font-family:system-ui,sans-serif;text-align:center;padding-top:3rem;background:#0d1117;color:#c9d1d9">
<h1 style="color:#58a6ff">Authorization complete</h1>
<p>You may close this window and return to qdoc.</p>
</body>
</html>`))

		codeCh <- code
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			errCh <- fmt.Errorf("starting callback server: %w", err)
			return
		}
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	return srv
}
