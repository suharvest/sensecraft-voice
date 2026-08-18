package openai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// TitleGenerator 标题生成器
type TitleGenerator struct {
	client Client
	config *Config
}

// NewTitleGenerator 创建标题生成器
func NewTitleGenerator(client Client, config *Config) *TitleGenerator {
	return &TitleGenerator{
		client: client,
		config: config,
	}
}

// GenerateTitle 生成会话标题
func (tg *TitleGenerator) GenerateTitle(ctx context.Context, content string) (string, error) {
	klog.Infof("开始生成标题，原始内容长度: %d, 内容: '%s'", len(content), content)

	// 截取前100个字符，但尽量保持完整句子
	truncatedContent := tg.truncateContent(content, 100)
	klog.Infof("截取后内容长度: %d, 内容: '%s'", len(truncatedContent), truncatedContent)

	// 构建标题生成提示词
	prompt := fmt.Sprintf(`请为以下对话内容生成一个简洁的标题（不超过20个字符）：

%s

要求：
1. 标题要简洁明了，能概括对话的主要内容
2. 使用中文
3. 不要包含标点符号
`, truncatedContent)

	klog.Infof("生成的提示词: '%s'", prompt)

	// 创建OpenAI请求
	req := &ChatCompletionRequest{
		Model: tg.config.Model, // 使用配置中的模型
		Messages: []Message{
			{
				Role:    "system",
				Content: "你是一个专业的标题生成助手，擅长为对话内容生成简洁准确的标题。",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   50,  // 限制token数量
		Temperature: 0.3, // 降低随机性，确保标题稳定
	}

	// 调用OpenAI API
	resp, err := tg.client.CreateChatCompletion(req)
	if err != nil {
		klog.Errorf("调用OpenAI生成标题失败: %v", err)
		return "", fmt.Errorf("生成标题失败: %w", err)
	}

	klog.Infof("OpenAI API调用成功，choices数量: %d", len(resp.Choices))
	if len(resp.Choices) == 0 {
		klog.Warningf("OpenAI API返回的choices为空")
		return "新对话", nil
	}

	rawContent := resp.Choices[0].Message.Content
	klog.Infof("OpenAI API返回的原始内容: '%s'", rawContent)

	title := strings.TrimSpace(rawContent)

	// 清理标题，移除可能的引号和其他符号
	title = strings.Trim(title, `"'""''`)
	title = strings.TrimSpace(title)

	// 限制标题长度
	if len(title) > 20 {
		title = title[:20]
	}

	// 如果标题为空或太短，使用默认值
	if title == "" || len(title) < 2 {
		title = "新对话"
	}

	klog.Infof("成功生成标题: '%s' (原内容长度: %d)", title, len(content))
	return title, nil
}

// truncateContent 智能截取内容，尽量保持完整句子
func (tg *TitleGenerator) truncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}

	// 截取前maxLength个字符
	truncated := content[:maxLength]

	// 尝试找到最后一个句号、问号或感叹号
	lastSentenceEnd := -1
	for i := len(truncated) - 1; i >= 0; i-- {
		char := truncated[i]
		if char == '.' || char == '?' || char == '!' {
			lastSentenceEnd = i
			break
		}
	}

	// 检查中文字符（需要按rune处理）
	runes := []rune(truncated)
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] == '。' || runes[i] == '？' || runes[i] == '！' {
			// 将rune索引转换为字节索引
			lastSentenceEnd = len(string(runes[:i+1]))
			break
		}
	}

	// 如果找到句子结束符，截取到该位置
	if lastSentenceEnd > 0 && lastSentenceEnd > maxLength/2 {
		return truncated[:lastSentenceEnd+1]
	}

	// 否则尝试找到最后一个空格
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLength/2 {
		return truncated[:lastSpace]
	}

	// 直接截取
	return truncated
}

// GenerateTitleAsync 异步生成标题
func (tg *TitleGenerator) GenerateTitleAsync(ctx context.Context, content string, callback func(string, error)) {
	go func() {
		// 设置超时上下文
		timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		title, err := tg.GenerateTitle(timeoutCtx, content)
		if callback != nil {
			callback(title, err)
		}
	}()
}
