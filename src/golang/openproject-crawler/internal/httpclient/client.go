package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"openproject-crawler/internal/core"
	"time"
)

type APIClient struct {
	*core.URLHandler
	authToken string
	client    *http.Client
	headers   map[string]string
}

func NewAPIClient(baseURL, authToken string) (*APIClient, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	if authToken != "" {
		headers["Authorization"] = "Basic " + authToken
	}

	urlHandler, err := core.NewURLHandler(baseURL, "")
	if err != nil {
		return nil, err
	}

	return &APIClient{
		URLHandler: urlHandler,
		authToken:  authToken,
		headers:    headers,
		client:     &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (api *APIClient) GetRequest(customURI string, params map[string]interface{}) (string, error) {
	if customURI != "" {
		api.SetURIPath(customURI)
	}

	fullURL := api.GetFullURL()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range api.headers {
		req.Header.Set(key, value)
	}

	if params == nil {
		params = make(map[string]interface{})
	}
	if len(params) == 0 {
		params["pageSize"] = "1000"
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, fmt.Sprintf("%v", value))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
