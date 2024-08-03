package core

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type URLHandler struct {
	baseURL string
	uriPath string
}

func NewURLHandler(baseURL, uriPath string) (*URLHandler, error) {
	baseURL, err := addSchema(baseURL)
	if err != nil {
		return nil, err
	}
	return &URLHandler{
		baseURL: baseURL,
		uriPath: uriPath,
	}, nil
}

func addSchema(rawURL string) (string, error) {
	if rawURL == "" {
		return "", errors.New("URL was empty")
	}
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsedURL.Scheme == "" {
		rawURL = "http://" + rawURL
		parsedURL, err = url.Parse(rawURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL after adding schema: %w", err)
		}
	}
	if parsedURL.Host == "" {
		return "", errors.New("URL must include a host")
	}
	return rawURL, nil
}

func (u *URLHandler) GetBaseURL() string {
	return u.baseURL
}

func (u *URLHandler) SetBaseURL(rawURL string) error {
	url, err := addSchema(rawURL)
	if err != nil {
		return err
	}
	u.baseURL = url
	return nil
}

func (u *URLHandler) GetURIPath() string {
	return u.uriPath
}

func (u *URLHandler) SetURIPath(uri string) {
	u.uriPath = uri
}

func (u *URLHandler) GetFullURL() string {
	if u.uriPath != "" {
		return strings.TrimRight(u.baseURL, "/") + "/" + strings.TrimLeft(u.uriPath, "/")
	}
	return u.baseURL
}
