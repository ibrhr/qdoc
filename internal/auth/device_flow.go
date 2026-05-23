package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ibrhr/qdoc/internal/provider"
)

var deviceHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}

type DeviceFlowAuth struct {
	DeviceAuthURL string
	TokenURL      string
	ClientID      string
	Scope         string
}

type deviceCodeResp struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type tokenResp struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	Error        string `json:"error"`
	ErrorDesc    string `json:"error_description"`
}

func (a *DeviceFlowAuth) Type() string               { return "oauth_device" }
func (a *DeviceFlowAuth) NeedsInteractiveAuth() bool  { return true }
func (a *DeviceFlowAuth) ApplyAuth(req *http.Request, token Token) {
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
}

func (a *DeviceFlowAuth) Refresh(prov provider.Provider, token Token, store *TokenStore) (Token, error) {
	return Token{}, ErrNoRefresh
}

func (a *DeviceFlowAuth) Authenticate(prov provider.Provider, store *TokenStore) <-chan AuthStatus {
	ch := make(chan AuthStatus, 1)

	go func() {
		defer close(ch)

		clientID := osLookupEnv("QDOC_GITHUB_CLIENT_ID")
		if clientID == "" {
			clientID = a.ClientID
		}
		if clientID == "" {
			ch <- AuthStatus{Stage: StageError, Err: errors.New("no client_id configured for device flow")}
			return
		}

		device, err := a.requestDeviceCode(clientID)
		if err != nil {
			ch <- AuthStatus{Stage: StageError, Err: fmt.Errorf("requesting device code: %w", err)}
			return
		}

		uri := device.VerificationURI
		if uri == "" {
			uri = "https://github.com/login/device"
		}

		ch <- AuthStatus{
			Stage:           StageAwaitingUser,
			Message:         fmt.Sprintf("Visit %s and enter the code", uri),
			VerificationURI: uri,
			UserCode:        device.UserCode,
		}

		interval := device.Interval
		if interval < 1 {
			interval = 5
		}

		expiry := time.Now().Add(time.Duration(device.ExpiresIn) * time.Second)

		pollCount := 0
		for {
			if time.Now().After(expiry) {
				ch <- AuthStatus{
					Stage:   StageTimeout,
					Message: "Authorization timed out",
					Err:     errors.New("device authorization expired"),
				}
				return
			}

			time.Sleep(time.Duration(interval) * time.Second)
			pollCount++

			tok, pollErr := a.pollToken(clientID, device.DeviceCode)
			if pollErr == nil {
				tkn := Token{
					AccessToken: tok.AccessToken,
					TokenType:   tok.TokenType,
					Scope:       tok.Scope,
				}
				if tkn.TokenType == "" {
					tkn.TokenType = "bearer"
				}

				if store != nil {
					if err := store.Set(prov.Name, tkn); err != nil {
						ch <- AuthStatus{Stage: StageError, Err: fmt.Errorf("saving token: %w", err)}
						return
					}
				}

				ch <- AuthStatus{Stage: StageAuthorized, Token: &tkn}
				return
			}

			var de *deviceError
			if errors.As(pollErr, &de) {
				switch de.ErrorCode {
				case "authorization_pending":
					ch <- AuthStatus{
						Stage:   StagePolling,
						Message: fmt.Sprintf("Waiting for authorization (attempt %d)...", pollCount),
					}
				case "slow_down":
					interval += 5
					ch <- AuthStatus{
						Stage:   StagePolling,
						Message: "Rate limited — slowing down...",
					}
				case "expired_token":
					ch <- AuthStatus{
						Stage:   StageTimeout,
						Message: "Device code expired",
						Err:     errors.New("device code expired"),
					}
					return
				case "access_denied":
					ch <- AuthStatus{
						Stage:   StageError,
						Message: "Access denied by user",
						Err:     errors.New("user denied authorization"),
					}
					return
				default:
					ch <- AuthStatus{
						Stage: StageError,
						Err:   de,
					}
					return
				}
			} else {
				ch <- AuthStatus{
					Stage: StageError,
					Err:   fmt.Errorf("polling token: %w", pollErr),
				}
				return
			}
		}
	}()

	return ch
}

type deviceError struct {
	ErrorCode    string
	ErrorMessage string
}

func (e *deviceError) Error() string {
	return fmt.Sprintf("device auth error: %s (%s)", e.ErrorCode, e.ErrorMessage)
}

func (a *DeviceFlowAuth) requestDeviceCode(clientID string) (*deviceCodeResp, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	if a.Scope != "" {
		form.Set("scope", a.Scope)
	}

	req, err := http.NewRequest("POST", a.DeviceAuthURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := deviceHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request returned %d", resp.StatusCode)
	}

	var dcr deviceCodeResp
	if err := json.NewDecoder(resp.Body).Decode(&dcr); err != nil {
		return nil, fmt.Errorf("parsing device code response: %w", err)
	}
	if dcr.UserCode == "" || dcr.DeviceCode == "" {
		return nil, errors.New("invalid device code response: missing required fields")
	}
	return &dcr, nil
}

func (a *DeviceFlowAuth) pollToken(clientID, deviceCode string) (*tokenResp, error) {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("device_code", deviceCode)
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequest("POST", a.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := deviceHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tr tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	if tr.Error != "" {
		return nil, &deviceError{ErrorCode: tr.Error, ErrorMessage: tr.ErrorDesc}
	}

	if tr.AccessToken == "" {
		return nil, errors.New("empty access token in response")
	}

	return &tr, nil
}
