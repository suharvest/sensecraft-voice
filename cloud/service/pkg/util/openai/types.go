package openai

// OpenAI API 基础配置
type Config struct {
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	Timeout     int     `json:"timeout"` // 秒
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	Model       string  `json:"model"`
}

// 默认配置
func DefaultConfig() *Config {
	return &Config{
		BaseURL:     "https://api.openai.com/v1",
		Timeout:     120,
		MaxTokens:   4096,
		Temperature: 0.7,
		Model:       "gpt-3.5-turbo",
	}
}

// 消息结构
type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// 工具调用结构
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function map[string]interface{} `json:"function"`
}

// 聊天补全请求
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
}

// 聊天补全响应
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// 选择项
type Choice struct {
	Index        int        `json:"index"`
	Message      Message    `json:"message,omitempty"`
	Delta        Message    `json:"delta,omitempty"`
	FinishReason string     `json:"finish_reason"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
}

// 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// 工具定义
type Tool struct {
	Type     string                 `json:"type"`
	Function map[string]interface{} `json:"function"`
}

// 流式响应事件
type StreamEvent struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// 错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// 上下文管理器接口
type ContextManager interface {
	// 保存消息到数据库
	SaveMessage(sessionID, role, content string, metadata map[string]interface{}) error

	// 获取会话上下文
	GetConversationContext(sessionID string, maxMessages int) ([]Message, error)

	// 创建新会话
	CreateSession(userID string) (string, error)

	// 关闭会话
	CloseSession(sessionID string) error
}

// 客户端接口
type Client interface {
	// 创建聊天补全
	CreateChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error)

	// 创建流式聊天补全
	CreateChatCompletionStream(req *ChatCompletionRequest) (<-chan *StreamEvent, error)

	// 带上下文的聊天
	ChatWithContext(sessionID, userMessage, userID string) (*ChatCompletionResponse, error)

	// 流式聊天
	ChatWithContextStream(sessionID, userMessage, userID string) (<-chan *StreamEvent, error)
}
