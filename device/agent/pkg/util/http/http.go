package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// HTTPClient HTTP 客户端接口
type HTTPClient interface {
	Get(url string) (*Response, error)
	Post(url string, body interface{}) (*Response, error)
	Put(url string, body interface{}) (*Response, error)
	Delete(url string) (*Response, error)
	Patch(url string, body interface{}) (*Response, error)
	Head(url string) (*Response, error)
	Options(url string) (*Response, error)

	// 带上下文的请求方法
	GetWithContext(ctx context.Context, url string) (*Response, error)
	PostWithContext(ctx context.Context, url string, body interface{}) (*Response, error)
	PutWithContext(ctx context.Context, url string, body interface{}) (*Response, error)
	DeleteWithContext(ctx context.Context, url string) (*Response, error)
	PatchWithContext(ctx context.Context, url string, body interface{}) (*Response, error)

	// 设置请求头
	SetHeader(key, value string) HTTPClient
	SetHeaders(headers map[string]string) HTTPClient

	// 设置查询参数
	SetQueryParam(key, value string) HTTPClient
	SetQueryParams(params map[string]string) HTTPClient

	// 设置认证
	SetAuthToken(token string) HTTPClient
	SetBasicAuth(username, password string) HTTPClient

	// 设置超时
	SetTimeout(timeout time.Duration) HTTPClient

	// 设置重试
	SetRetryCount(count int) HTTPClient
	SetRetryWaitTime(waitTime time.Duration) HTTPClient
	SetRetryMaxWaitTime(maxWaitTime time.Duration) HTTPClient

	// 获取底层 resty 客户端
	GetRestyClient() *resty.Client
}

// Response HTTP 响应包装
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
	Raw        *resty.Response
}

// Client HTTP 客户端实现
type Client struct {
	client *resty.Client
}

// NewClient 创建新的 HTTP 客户端
func NewClient() HTTPClient {
	client := resty.New()

	// 设置默认配置
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	return &Client{client: client}
}

// NewClientWithConfig 使用自定义配置创建 HTTP 客户端
func NewClientWithConfig(config *Config) HTTPClient {
	client := resty.New()

	// 应用配置
	if config.Timeout > 0 {
		client.SetTimeout(config.Timeout)
	}

	if config.RetryCount > 0 {
		client.SetRetryCount(config.RetryCount)
	}

	if config.RetryWaitTime > 0 {
		client.SetRetryWaitTime(config.RetryWaitTime)
	}

	if config.RetryMaxWaitTime > 0 {
		client.SetRetryMaxWaitTime(config.RetryMaxWaitTime)
	}

	// 设置 TLS 配置
	if config.TLSConfig != nil {
		client.SetTLSClientConfig(config.TLSConfig)
	}

	// 设置代理
	if config.Proxy != "" {
		client.SetProxy(config.Proxy)
	}

	// 设置用户代理
	if config.UserAgent != "" {
		client.SetHeader("User-Agent", config.UserAgent)
	}

	return &Client{client: client}
}

// Config HTTP 客户端配置
type Config struct {
	Timeout          time.Duration
	RetryCount       int
	RetryWaitTime    time.Duration
	RetryMaxWaitTime time.Duration
	TLSConfig        *tls.Config
	Proxy            string
	UserAgent        string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:          30 * time.Second,
		RetryCount:       3,
		RetryWaitTime:    1 * time.Second,
		RetryMaxWaitTime: 5 * time.Second,
		UserAgent:        "SenseCraft-Voice-Client/1.0",
	}
}

// Get 执行 GET 请求
func (c *Client) Get(url string) (*Response, error) {
	return c.GetWithContext(context.Background(), url)
}

// Post 执行 POST 请求
func (c *Client) Post(url string, body interface{}) (*Response, error) {
	return c.PostWithContext(context.Background(), url, body)
}

// Put 执行 PUT 请求
func (c *Client) Put(url string, body interface{}) (*Response, error) {
	return c.PutWithContext(context.Background(), url, body)
}

// Delete 执行 DELETE 请求
func (c *Client) Delete(url string) (*Response, error) {
	return c.DeleteWithContext(context.Background(), url)
}

// Patch 执行 PATCH 请求
func (c *Client) Patch(url string, body interface{}) (*Response, error) {
	return c.PatchWithContext(context.Background(), url, body)
}

// Head 执行 HEAD 请求
func (c *Client) Head(url string) (*Response, error) {
	return c.HeadWithContext(context.Background(), url)
}

// Options 执行 OPTIONS 请求
func (c *Client) Options(url string) (*Response, error) {
	return c.OptionsWithContext(context.Background(), url)
}

// GetWithContext 带上下文的 GET 请求
func (c *Client) GetWithContext(ctx context.Context, url string) (*Response, error) {
	resp, err := c.client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}
	return c.wrapResponse(resp), nil
}

// PostWithContext 带上下文的 POST 请求
func (c *Client) PostWithContext(ctx context.Context, url string, body interface{}) (*Response, error) {
	resp, err := c.client.R().SetContext(ctx).SetBody(body).Post(url)
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %w", err)
	}
	return c.wrapResponse(resp), nil
}

// PutWithContext 带上下文的 PUT 请求
func (c *Client) PutWithContext(ctx context.Context, url string, body interface{}) (*Response, error) {
	resp, err := c.client.R().SetContext(ctx).SetBody(body).Put(url)
	if err != nil {
		return nil, fmt.Errorf("PUT request failed: %w", err)
	}
	return c.wrapResponse(resp), nil
}

// DeleteWithContext 带上下文的 DELETE 请求
func (c *Client) DeleteWithContext(ctx context.Context, url string) (*Response, error) {
	resp, err := c.client.R().SetContext(ctx).Delete(url)
	if err != nil {
		return nil, fmt.Errorf("DELETE request failed: %w", err)
	}
	return c.wrapResponse(resp), nil
}

// PatchWithContext 带上下文的 PATCH 请求
func (c *Client) PatchWithContext(ctx context.Context, url string, body interface{}) (*Response, error) {
	resp, err := c.client.R().SetContext(ctx).SetBody(body).Patch(url)
	if err != nil {
		return nil, fmt.Errorf("PATCH request failed: %w", err)
	}
	return c.wrapResponse(resp), nil
}

// HeadWithContext 带上下文的 HEAD 请求
func (c *Client) HeadWithContext(ctx context.Context, url string) (*Response, error) {
	resp, err := c.client.R().SetContext(ctx).Head(url)
	if err != nil {
		return nil, fmt.Errorf("HEAD request failed: %w", err)
	}
	return c.wrapResponse(resp), nil
}

// OptionsWithContext 带上下文的 OPTIONS 请求
func (c *Client) OptionsWithContext(ctx context.Context, url string) (*Response, error) {
	resp, err := c.client.R().SetContext(ctx).Options(url)
	if err != nil {
		return nil, fmt.Errorf("OPTIONS request failed: %w", err)
	}
	return c.wrapResponse(resp), nil
}

// SetHeader 设置请求头
func (c *Client) SetHeader(key, value string) HTTPClient {
	c.client.SetHeader(key, value)
	return c
}

// SetHeaders 批量设置请求头
func (c *Client) SetHeaders(headers map[string]string) HTTPClient {
	c.client.SetHeaders(headers)
	return c
}

// SetQueryParam 设置查询参数
func (c *Client) SetQueryParam(key, value string) HTTPClient {
	c.client.SetQueryParam(key, value)
	return c
}

// SetQueryParams 批量设置查询参数
func (c *Client) SetQueryParams(params map[string]string) HTTPClient {
	c.client.SetQueryParams(params)
	return c
}

// SetAuthToken 设置认证令牌
func (c *Client) SetAuthToken(token string) HTTPClient {
	c.client.SetAuthToken(token)
	return c
}

// SetBasicAuth 设置基本认证
func (c *Client) SetBasicAuth(username, password string) HTTPClient {
	c.client.SetBasicAuth(username, password)
	return c
}

// SetTimeout 设置超时时间
func (c *Client) SetTimeout(timeout time.Duration) HTTPClient {
	c.client.SetTimeout(timeout)
	return c
}

// SetRetryCount 设置重试次数
func (c *Client) SetRetryCount(count int) HTTPClient {
	c.client.SetRetryCount(count)
	return c
}

// SetRetryWaitTime 设置重试等待时间
func (c *Client) SetRetryWaitTime(waitTime time.Duration) HTTPClient {
	c.client.SetRetryWaitTime(waitTime)
	return c
}

// SetRetryMaxWaitTime 设置最大重试等待时间
func (c *Client) SetRetryMaxWaitTime(maxWaitTime time.Duration) HTTPClient {
	c.client.SetRetryMaxWaitTime(maxWaitTime)
	return c
}

// GetRestyClient 获取底层 resty 客户端
func (c *Client) GetRestyClient() *resty.Client {
	return c.client
}

// wrapResponse 包装 resty 响应
func (c *Client) wrapResponse(resp *resty.Response) *Response {
	return &Response{
		StatusCode: resp.StatusCode(),
		Headers:    resp.Header(),
		Body:       resp.Body(),
		Raw:        resp,
	}
}

// IsSuccess 检查响应是否成功
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError 检查是否为客户端错误
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError 检查是否为服务器错误
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// GetHeader 获取响应头
func (r *Response) GetHeader(key string) string {
	if headers, exists := r.Headers[key]; exists && len(headers) > 0 {
		return headers[0]
	}
	return ""
}

// GetBodyString 获取响应体字符串
func (r *Response) GetBodyString() string {
	return string(r.Body)
}

// UnmarshalJSON 将响应体解析为 JSON
func (r *Response) UnmarshalJSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// UnmarshalXML 将响应体解析为 XML
func (r *Response) UnmarshalXML(v interface{}) error {
	return xml.Unmarshal(r.Body, v)
}
