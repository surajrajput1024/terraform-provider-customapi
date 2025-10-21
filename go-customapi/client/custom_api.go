package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-customapi/go-customapi/client/types"
)

type CustomAPIClient struct {
	*Client
}

func NewCustomAPIClient(authConfig *AuthConfig, baseURL string) *CustomAPIClient {
	return &CustomAPIClient{
		Client: NewClient(authConfig, baseURL),
	}
}

func (c *CustomAPIClient) MakeRequest(ctx context.Context, req *types.CustomAPIRequest) (*types.CustomAPIResponse, error) {
	token, err := c.GetToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %v", err)
	}

	fullURL := c.buildURL(req.URL, req.QueryParams)
	
	var requestData interface{}
	if len(req.Body) > 0 {
		requestData = req.Body
	}

	httpReq, err := createRequest(ctx, req.Method, fullURL, requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	c.setHeaders(httpReq, req.Headers, token)

	tflog.Debug(ctx, "Making API request", map[string]interface{}{
		"method": req.Method,
		"url":    fullURL,
		"headers": httpReq.Header,
	})

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	responseBody := respToString(resp)
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	apiResponse := &types.CustomAPIResponse{
		StatusCode: resp.StatusCode,
		Headers:    responseHeaders,
		Body:       []byte(responseBody),
		Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	if !apiResponse.Success {
		apiResponse.Error = fmt.Sprintf("Request failed with status %d", resp.StatusCode)
	}

	tflog.Debug(ctx, "API response received", map[string]interface{}{
		"status_code": resp.StatusCode,
		"success":     apiResponse.Success,
	})

	return apiResponse, nil
}

func (c *CustomAPIClient) GetUserProfile(ctx context.Context, orgID string) (*types.UserProfile, error) {
	req := &types.CustomAPIRequest{
		Method: "GET",
		URL:    "/api/users/profile/me",
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	if orgID != "" {
		req.Headers["current-organization"] = orgID
	}

	resp, err := c.MakeRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("failed to get user profile: %s", resp.Error)
	}

	var profile types.UserProfile
	if err := json.Unmarshal(resp.Body, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user profile: %v", err)
	}

	return &profile, nil
}

func (c *CustomAPIClient) buildURL(endpoint string, queryParams map[string]string) string {
	baseURL := c.GetBaseURL()
	if baseURL == "" {
		config, err := LoadConfig()
		if err != nil {
			panic("CUSTOMAPI_BASE_URL environment variable is required")
		}
		baseURL = config.BaseURL
		if baseURL == "" {
			panic("CUSTOMAPI_BASE_URL environment variable is required")
		}
	}

	fullURL := baseURL + endpoint
	
	if len(queryParams) > 0 {
		params := url.Values{}
		for key, value := range queryParams {
			params.Add(key, value)
		}
		fullURL += "?" + params.Encode()
	}

	return fullURL
}

func (c *CustomAPIClient) setHeaders(req *http.Request, customHeaders map[string]string, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("User-Agent", "Terraform-Provider-CustomAPI/1.0")

	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}
}


func (c *CustomAPIClient) CreateResource(ctx context.Context, endpoint string, data interface{}, orgID string) (*types.CustomAPIResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %v", err)
	}

	req := &types.CustomAPIRequest{
		Method: "POST",
		URL:    endpoint,
		Body:   body,
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	if orgID != "" {
		req.Headers["current-organization"] = orgID
	}

	return c.MakeRequest(ctx, req)
}

func (c *CustomAPIClient) ReadResource(ctx context.Context, endpoint string, orgID string) (*types.CustomAPIResponse, error) {
	req := &types.CustomAPIRequest{
		Method: "GET",
		URL:    endpoint,
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	if orgID != "" {
		req.Headers["current-organization"] = orgID
	}

	return c.MakeRequest(ctx, req)
}

func (c *CustomAPIClient) UpdateResource(ctx context.Context, endpoint string, data interface{}, orgID string) (*types.CustomAPIResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %v", err)
	}

	req := &types.CustomAPIRequest{
		Method: "PUT",
		URL:    endpoint,
		Body:   body,
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	if orgID != "" {
		req.Headers["current-organization"] = orgID
	}

	return c.MakeRequest(ctx, req)
}

func (c *CustomAPIClient) DeleteResource(ctx context.Context, endpoint string, orgID string) (*types.CustomAPIResponse, error) {
	req := &types.CustomAPIRequest{
		Method: "DELETE",
		URL:    endpoint,
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	if orgID != "" {
		req.Headers["current-organization"] = orgID
	}

	return c.MakeRequest(ctx, req)
}
