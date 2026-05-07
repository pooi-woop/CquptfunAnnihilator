package client

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"CquptFunAnnihilator/logger"
)

type HttpClient struct {
	client  *resty.Client
	baseURL string
	token   string
	cookie  string
}

func NewHttpClient(baseURL string, timeout int, userAgent string) *HttpClient {
	client := resty.New().
		SetTimeout(time.Duration(timeout)*time.Second).
		SetHeader("User-Agent", userAgent).
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	logger.Info("HTTP client initialized",
		zap.String("baseURL", baseURL),
		zap.Int("timeout", timeout),
	)

	return &HttpClient{
		client:  client,
		baseURL: baseURL,
	}
}

func (c *HttpClient) SetToken(token string) {
	c.token = token
	logger.Debug("Token set for HTTP client")
}

func (c *HttpClient) SetCookie(cookie string) {
	c.cookie = cookie
	logger.Debug("Cookie set for HTTP client")
}

func (c *HttpClient) SetHeader(key, value string) {
	c.client.SetHeader(key, value)
}

func (c *HttpClient) Get(path string) (*resty.Response, error) {
	url := c.buildURL(path)
	req := c.client.R()

	if c.cookie != "" {
		req.SetHeader("Cookie", c.cookie)
	} else if c.token != "" {
		req.SetHeader("Authorization", "Bearer "+c.token)
	}

	logger.Debug("HTTP GET request",
		zap.String("url", url),
		zap.String("path", path),
	)

	resp, err := req.Get(url)

	if err != nil {
		logger.Error("HTTP GET request failed",
			zap.String("url", url),
			zap.Error(err),
		)
		return resp, err
	}

	logger.Debug("HTTP GET response received",
		zap.String("url", url),
		zap.Int("status", resp.StatusCode()),
		zap.Int64("bodySize", resp.Size()),
	)

	return resp, nil
}

func (c *HttpClient) Post(path string, body interface{}) (*resty.Response, error) {
	url := c.buildURL(path)
	req := c.client.R()

	if c.cookie != "" {
		req.SetHeader("Cookie", c.cookie)
	} else if c.token != "" {
		req.SetHeader("Authorization", "Bearer "+c.token)
	}

	if body != nil {
		req.SetBody(body)
	}

	logger.Debug("HTTP POST request",
		zap.String("url", url),
		zap.String("path", path),
	)

	resp, err := req.Post(url)

	if err != nil {
		logger.Error("HTTP POST request failed",
			zap.String("url", url),
			zap.Error(err),
		)
		return resp, err
	}

	logger.Debug("HTTP POST response received",
		zap.String("url", url),
		zap.Int("status", resp.StatusCode()),
		zap.Int64("bodySize", resp.Size()),
	)

	return resp, nil
}

func (c *HttpClient) Put(path string, body interface{}) (*resty.Response, error) {
	url := c.buildURL(path)
	req := c.client.R()

	if c.cookie != "" {
		req.SetHeader("Cookie", c.cookie)
	} else if c.token != "" {
		req.SetHeader("Authorization", "Bearer "+c.token)
	}

	if body != nil {
		req.SetBody(body)
	}

	logger.Debug("HTTP PUT request",
		zap.String("url", url),
		zap.String("path", path),
	)

	resp, err := req.Put(url)

	if err != nil {
		logger.Error("HTTP PUT request failed",
			zap.String("url", url),
			zap.Error(err),
		)
		return resp, err
	}

	logger.Debug("HTTP PUT response received",
		zap.String("url", url),
		zap.Int("status", resp.StatusCode()),
	)

	return resp, nil
}

func (c *HttpClient) Delete(path string) (*resty.Response, error) {
	url := c.buildURL(path)
	req := c.client.R()

	if c.cookie != "" {
		req.SetHeader("Cookie", c.cookie)
	} else if c.token != "" {
		req.SetHeader("Authorization", "Bearer "+c.token)
	}

	logger.Debug("HTTP DELETE request",
		zap.String("url", url),
	)

	resp, err := req.Delete(url)

	if err != nil {
		logger.Error("HTTP DELETE request failed",
			zap.String("url", url),
			zap.Error(err),
		)
		return resp, err
	}

	return resp, nil
}

func (c *HttpClient) buildURL(path string) string {
	if path == "" {
		return c.baseURL
	}
	return fmt.Sprintf("%s%s", c.baseURL, path)
}

func (c *HttpClient) GetClient() *resty.Client {
	return c.client
}

func (c *HttpClient) GetBaseURL() string {
	return c.baseURL
}

func (c *HttpClient) GetCookie() string {
	return c.cookie
}
