package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type AuthConfig struct {
	Username     string
	Password     string
	Environment  string
	AuthToken    string
	BaseURL      string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type AuthClient struct {
	httpClient *http.Client
	config     *AuthConfig
	token      string
	expiresAt  time.Time
}

func NewAuthClient(config *AuthConfig) *AuthClient {
	return &AuthClient{
		httpClient: createHTTPClient(nil, 0),
		config:     config,
	}
}

func (ac *AuthClient) GetToken(ctx context.Context) (string, error) {
	if ac.config.AuthToken != "" {
		return ac.config.AuthToken, nil
	}

	if ac.token != "" && time.Now().Before(ac.expiresAt) {
		return ac.token, nil
	}

	return ac.authenticateWithCredentials(ctx)
}

func (ac *AuthClient) authenticateWithCredentials(ctx context.Context) (string, error) {
	authURL := ac.getAuthURL()
	
	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("username", ac.config.Username)
	formData.Set("password", ac.config.Password)
	formData.Set("scope", "openid profile email")
	config, err := LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %v", err)
	}
	
	formData.Set("audience", config.Audience)
	formData.Set("client_id", config.ClientID)

	req, err := http.NewRequestWithContext(ctx, "POST", authURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	tflog.Debug(ctx, "Authenticating with credentials", map[string]interface{}{
		"url":      authURL,
		"username": ac.config.Username,
	})

	resp, err := ac.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute auth request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	ac.token = tokenResp.AccessToken
	ac.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	tflog.Debug(ctx, "Authentication successful", map[string]interface{}{
		"token_type": tokenResp.TokenType,
		"expires_in": tokenResp.ExpiresIn,
	})

	return ac.token, nil
}

func (ac *AuthClient) getAuthURL() string {
	baseURL := ac.config.BaseURL
	if baseURL == "" {
		switch ac.config.Environment {
		case "qa":
			baseURL = "https://pace-app-qa.us.auth0.com"
		case "staging":
			baseURL = "https://pace-app-staging.us.auth0.com"
		case "prod":
			baseURL = "https://pace-app.us.auth0.com"
		default:
			baseURL = "https://pace-app-qa.us.auth0.com"
		}
	}
	return fmt.Sprintf("%s/oauth/token", baseURL)
}

func (ac *AuthClient) IsTokenValid() bool {
	return ac.token != "" && time.Now().Before(ac.expiresAt)
}

func (ac *AuthClient) RefreshToken(ctx context.Context) error {
	ac.token = ""
	ac.expiresAt = time.Time{}
	_, err := ac.GetToken(ctx)
	return err
}
