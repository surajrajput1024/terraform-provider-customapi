package client

import (
	"context"
	"net/http"
	"time"
)


func NewClient(authConfig *AuthConfig, baseURL string) *Client {
	return &Client{
		httpClient: createHTTPClient(nil, 0),
		authClient: NewAuthClient(authConfig),
		baseURL:    baseURL,
	}
}

func (c *Client) GetHTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) GetAuthClient() *AuthClient {
	return c.authClient
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

func (c *Client) IsAuthenticated() bool {
	return c.authClient.IsTokenValid()
}

func (c *Client) RefreshAuth(ctx context.Context) error {
	return c.authClient.RefreshToken(ctx)
}

func (c *Client) GetToken(ctx context.Context) (string, error) {
	return c.authClient.GetToken(ctx)
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

func (c *Client) SetTransport(transport *http.Transport) {
	c.httpClient.Transport = transport
}
