package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
	config *Config
}

type Config struct {
	Timeout          time.Duration
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
	EnableDebug      bool
	BaseURL          string
	Headers          map[string]string
}

func DefaultConfig() *Config {
	return &Config{
		Timeout:          30 * time.Second,
		RetryCount:       3,
		RetryWaitTime:    1 * time.Second,
		RetryMaxWaitTime: 10 * time.Second,
		EnableDebug:      false,
		Headers:          make(map[string]string),
	}
}

func NewClient(config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := resty.New().
		SetTimeout(config.Timeout).
		SetRetryCount(config.RetryCount).
		SetRetryWaitTime(config.RetryWaitTime).
		SetRetryMaxWaitTime(config.RetryMaxWaitTime).
		SetDebug(config.EnableDebug)

	if config.BaseURL != "" {
		client.SetBaseURL(config.BaseURL)
	}

	for k, v := range config.Headers {
		client.SetHeader(k, v)
	}

	return &Client{
		client: client,
		config: config,
	}
}

func (c *Client) SetAuthToken(token string) *Client {
	c.client.SetAuthToken(token)
	return c
}

func (c *Client) SetHeader(key, value string) *Client {
	c.client.SetHeader(key, value)
	return c
}

func (c *Client) SetHeaders(headers map[string]string) *Client {
	c.client.SetHeaders(headers)
	return c
}

func (c *Client) R() *resty.Request {
	return c.client.R()
}

func (c *Client) Get(ctx context.Context, url string, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(result).
		Get(url)

	if err != nil {
		return fmt.Errorf("GET request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func (c *Client) Post(ctx context.Context, url string, body interface{}, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(result).
		Post(url)

	if err != nil {
		return fmt.Errorf("POST request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func (c *Client) Put(ctx context.Context, url string, body interface{}, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(result).
		Put(url)

	if err != nil {
		return fmt.Errorf("PUT request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, url string, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(result).
		Delete(url)

	if err != nil {
		return fmt.Errorf("DELETE request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func (c *Client) PostJSON(ctx context.Context, url string, body interface{}, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(result).
		Post(url)

	if err != nil {
		return fmt.Errorf("POST JSON request failed: %w", err)
	}

	if resp.IsError() {
		var errorBody map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &errorBody); err == nil {
			return fmt.Errorf("request failed with status %d: %v", resp.StatusCode(), errorBody)
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func (c *Client) PostForm(ctx context.Context, url string, data map[string]string, result interface{}) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetFormData(data).
		SetResult(result).
		Post(url)

	if err != nil {
		return fmt.Errorf("POST Form request failed: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	return nil
}

func (c *Client) DownloadFile(ctx context.Context, url string, filePath string) error {
	_, err := c.client.R().
		SetContext(ctx).
		SetOutput(filePath).
		Get(url)

	if err != nil {
		return fmt.Errorf("download file failed: %w", err)
	}

	return nil
}

// SeeedModifyRecordingMeetingContent 调用Seeed API修改录音会议内容
func (c *Client) SeeedModifyRecordingMeetingContent(ctx context.Context, baseURL string, req *SeeedModifyRecordingRequest) (*SeeedModifyRecordingResponse, error) {
	url := baseURL + "/api/SF/ModifyRecordingMeetingContent"

	var result SeeedModifyRecordingResponse
	err := c.PostJSON(ctx, url, req, &result)
	if err != nil {
		return nil, fmt.Errorf("Seeed API call failed: %w", err)
	}

	return &result, nil
}
