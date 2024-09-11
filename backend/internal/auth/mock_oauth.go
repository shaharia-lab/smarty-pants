package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/shaharia-lab/smarty-pants/backend/internal/types"
)

// Important Information:
// - The MockOAuthProvider struct is a mock OAuth provider for testing purposes.
// - This Mock OAuth provider is used to simulate an OAuth provider for testing purposes. Compatibility tested with
//   https://github.com/shaharia-lab/oauth-mock-server
//
// Only intended to run in a test (CI/CD) environment.

// MockOAuthProvider is a mock OAuth provider for testing purposes
type MockOAuthProvider struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// NewMockOAuthProvider creates a new MockOAuthProvider
func NewMockOAuthProvider(baseURL, clientID, clientSecret, redirectURL string) *MockOAuthProvider {
	return &MockOAuthProvider{
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}
}

// GetAuthURL returns the URL to redirect the user to for authorization
func (m *MockOAuthProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", m.ClientID)
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("redirect_uri", m.RedirectURL)

	return fmt.Sprintf("%s/authorize?%s", m.BaseURL, params.Encode())
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (m *MockOAuthProvider) ExchangeCodeForToken(ctx context.Context, code string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", m.ClientID)
	data.Set("client_secret", m.ClientSecret)
	data.Set("redirect_uri", m.RedirectURL)

	resp, err := http.PostForm(m.BaseURL+"/token", data)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed: %s - %s", resp.Status, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	return result.AccessToken, nil
}

// GetUserInfo retrieves user information using an access token
func (m *MockOAuthProvider) GetUserInfo(ctx context.Context, token string) (*types.OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", m.BaseURL+"/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s - %s", resp.Status, string(body))
	}

	var userInfo types.OAuthUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %v", err)
	}

	return &userInfo, nil
}
