package client 

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Client struct {
	httpClient *http.Client
	token      string
	baseURL    string
	authClient *AuthClient
}

func createHTTPClient(transport *http.Transport, timeoutInSec int) *http.Client {
	client := &http.Client{}

	// If a transport is provided, use it
	if transport != nil {
		client.Transport = transport
	}

	// Set timeout
	if timeoutInSec > 0 {
		client.Timeout = time.Duration(timeoutInSec) * time.Second
	} else {
		client.Timeout = 120 * time.Second // 2 minutes default timeout
	}

	return client
}


func AddBasicAuthHeader(req *http.Request, username, password string) {
	auth := username + ":" + password
	bAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+bAuth)
}

func AddBearerAuthHeader(req *http.Request, token string) {
	req.Header.Set("Authorization", "Bearer "+token)
}

func createRequest(ctx context.Context, method string, url string, data interface{}) (*http.Request, error) {
	var req *http.Request
	var err error

	debugLog("HTTP Request: %s %s data: %+v", method, url, data)

	if data != nil {
		tflog.Debug(ctx, "HTTP Request", map[string]interface{}{
			"method": method,
			"url":    url,
			"data":   data,
		})
		encodedData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %v", err)
		}
		tflog.Debug(ctx, "HTTP Request with data", map[string]interface{}{
			"method": method,
			"url":    url,
			"data":   string(encodedData),
		})
		req, err = http.NewRequest(method, url, bytes.NewBuffer(encodedData))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		tflog.Debug(ctx, "Failed to create request", map[string]interface{}{
			"error": err,
		})
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

func respToString(resp *http.Response) string {
	decodedData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Failed to read response body: %v", err)
	}

	debugLog("Response Body: %s", string(decodedData))
	return string(decodedData)
}

func jsonToObj(jsonStr string, target any) error {
	debugLog("JSON to Object: %s", jsonStr)
	err := json.Unmarshal([]byte(jsonStr), target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	debugLog("JSON to Object: %+v", target)
	return nil
}

func objToJson(obj any) (string, error) {
	debugLog("Object to JSON: %+v", obj)
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal object: %v", err)
	}
	debugLog("Object to JSON: %s", string(jsonData))
	return string(jsonData), nil
}

type httpRequestOptions struct {
	Method string 
	Url string 
	Data any
	ApiResponseTarget any
	ObjectResponseTarget any
	AuthRequired bool
}

func (c *Client) httpRequest(ctx context.Context, opts httpRequestOptions) (int, error) {
	req, err := createRequest(ctx, opts.Method, opts.Url, opts.Data)

	if err != nil {
		tflog.Debug(ctx, "Failed to create request", map[string]interface{}{
			"error": err,
		})
		return 0, err
	}

	if opts.AuthRequired {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	req.Header.Set("Content-Type", "application/json")

	if c.httpClient == nil {
		return 0, fmt.Errorf("HTTP client not initialized")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		tflog.Debug(ctx, "Failed to execute request", map[string]interface{}{
			"error": err,
		})
		return 0, err
	}
	defer resp.Body.Close()

	respString := respToString(resp)
	debugLog("Response: %s", respString)
	tflog.Debug(ctx, "Response", map[string]interface{}{
		"body": respString,
	})
	tflog.Trace(ctx, "Response", map[string]interface{}{
		"body": respString,
	})

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp.StatusCode == 401 {
			return resp.StatusCode, fmt.Errorf("Unauthorized: %s", respString)
		} else {
			if isJson(respString) {
				if opts.ApiResponseTarget != nil {
					err = jsonToObj(respString, opts.ApiResponseTarget)
					if err != nil {
						return resp.StatusCode, fmt.Errorf("Failed to unmarshal API response: %v", err)
					}
				}
			} else {
				return resp.StatusCode, fmt.Errorf("Response is not JSON: %s", respString)
			}
		}
	} else {
		isJson := isJson(respString)
		if isJson {
			if opts.ApiResponseTarget != nil {
				err = jsonToObj(respString, opts.ApiResponseTarget)
				if err != nil {
					return resp.StatusCode, fmt.Errorf("Failed to unmarshal API response: %v", err)
				}
			}
			if opts.ObjectResponseTarget != nil {
				err = jsonToObj(respString, opts.ObjectResponseTarget)
				if err != nil {
					return resp.StatusCode, fmt.Errorf("Failed to unmarshal object response: %v", err)
				}
			}
		} else {
			return resp.StatusCode, fmt.Errorf("Response is not JSON: %s", respString)
		}
	}
	return resp.StatusCode, nil
}

func debugLog(format string, args ...any) {
	tfLogValue := os.Getenv("TF_LOG")

	debugMode := false 
	if strings.ToLower(tfLogValue) == "debug" || strings.ToLower(tfLogValue) == "trace" {
		debugMode = true
	}

	if debugMode {
		format = fmt.Sprintf("DEBUG: %s", format)
		log.Printf(format, args...)
		tflog.Debug(context.Background(), format, map[string]interface{}{
			"args": args,
		})
	}
}


func isJson(str string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(str), &js) == nil
}
	