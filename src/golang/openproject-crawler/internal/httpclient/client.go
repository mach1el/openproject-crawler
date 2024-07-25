package httpclient

import (
	"fmt"
	"io"
	"net/http"
	"openproject-crawler/internal/core"
	"strings"
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

func (api *APIClient) GetRequest(customURI string, params map[string]string) (string, error) {
	if customURI != "" {
		api.SetURIPath(customURI)
	}

	url := api.GetFullURL()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	for key, value := range api.headers {
		req.Header.Set(key, value)
	}

	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch URL: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (api *APIClient) PostRequest(customURI string, body string) (string, error) {
	if customURI != "" {
		api.SetURIPath(customURI)
	}

	url := api.GetFullURL()

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return "", err
	}

	for key, value := range api.headers {
		req.Header.Set(key, value)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to post to URL: %s", resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseBody), nil
}
