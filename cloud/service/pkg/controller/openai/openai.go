package openai

import (
	"context"
	"fmt"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	dbmodel "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/openai"
	"k8s.io/klog/v2"
)

// OpenAI控制器接口
type Interface interface {
	// 发送聊天消息
	SendMessage(ctx context.Context, req *types.OpenAIChatRequest) (*types.OpenAIChatResponse, error)

	// 流式发送聊天消息
	StreamMessage(ctx context.Context, req *types.OpenAIChatRequest, streamChan chan<- interface{}) error

	// 获取聊天历史
	GetChatHistory(ctx context.Context, sessionID string, limit int) ([]*types.ChatMessage, error)

	// 创建新会话
	CreateSession(ctx context.Context, userID string) (string, error)

	// 关闭会话
	CloseSession(ctx context.Context, sessionID string) error

	// System Prompt CRUD
	CreateSystemPrompt(ctx context.Context, sp *dbmodel.SystemPrompt) error
	UpdateSystemPrompt(ctx context.Context, sp *dbmodel.SystemPrompt, updateActive bool) error
	UpdateSystemPromptStatus(ctx context.Context, id int64, isActive bool) error
	SetSystemPromptAsDefault(ctx context.Context, id int64) error
	DeleteSystemPrompt(ctx context.Context, id int64) error
	BatchDeleteSystemPrompts(ctx context.Context, ids []int64) error
	GetSystemPrompt(ctx context.Context, id int64) (*dbmodel.SystemPrompt, error)
	SearchSystemPromptsByName(ctx context.Context, name string, limit int) ([]*dbmodel.SystemPrompt, error)
	ListSystemPrompts(ctx context.Context, name, role string, active *bool, offset, limit int) ([]*dbmodel.SystemPrompt, int64, error)
}

// OpenAIGetter OpenAI获取器接口
type OpenAIGetter interface {
	OpenAI() Interface
}

// Controller OpenAI控制器实现
type Controller struct {
	factory      db.ShareDaoFactory
	config       *types.OpenAIConfig
	openaiClient openai.Client
}

// NewController 创建OpenAI控制器
func NewController(factory db.ShareDaoFactory, config *types.OpenAIConfig) Interface {
	// 创建上下文管理器
	contextManager := openai.NewContextManager(factory)

	// 创建OpenAI客户端
	openaiConfig := &openai.Config{
		APIKey:      config.APIKey,
		BaseURL:     config.BaseURL,
		Timeout:     config.Timeout,
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		Model:       config.Model,
	}

	openaiClient := openai.NewClient(openaiConfig, contextManager)

	return &Controller{
		factory:      factory,
		config:       config,
		openaiClient: openaiClient,
	}
}

// SendMessage 发送聊天消息
func (c *Controller) SendMessage(ctx context.Context, req *types.OpenAIChatRequest) (*types.OpenAIChatResponse, error) {
	// 验证必填字段
	if req.Message == "" {
		return nil, fmt.Errorf("消息内容不能为空")
	}
	if req.UserID == "" {
		return nil, fmt.Errorf("用户ID不能为空")
	}

	// 如果没有提供会话ID，创建新会话
	sessionID := req.SessionID
	isNewSession := false

	if sessionID == "" {
		var err error
		sessionID, err = c.openaiClient.(*openai.OpenAIClient).GetContextManager().(*openai.ChatContextManager).CreateSession(req.UserID)
		if err != nil {
			return nil, fmt.Errorf("创建会话失败: %w", err)
		}
		isNewSession = true
	} else {
		klog.Infof("使用现有会话: sessionID=%s", sessionID)
	}

	// 只在新会话时注入系统提示词：优先级为直接内容 > 指定ID > 默认提示词
	if isNewSession {
		var systemPromptContent string
		var systemPromptMeta map[string]interface{}

		klog.Infof("新会话，开始系统提示词注入检查: SystemPromptContent='%s', SystemPromptID=%d", req.SystemPromptContent, req.SystemPromptID)

		if req.SystemPromptContent != "" {
			// 1. 优先使用直接提供的系统提示词内容
			klog.Infof("使用直接提供的系统提示词内容")
			systemPromptContent = req.SystemPromptContent
			systemPromptMeta = map[string]interface{}{
				"system_prompt_source": "direct",
			}
		} else if req.SystemPromptID > 0 {
			// 2. 使用指定的系统提示词ID
			klog.Infof("尝试获取系统提示词ID: %d", req.SystemPromptID)
			sp, err := c.factory.SystemPrompt().Get(ctx, req.SystemPromptID)
			if err != nil {
				klog.Errorf("获取系统提示词失败: %v", err)
			} else if sp == nil {
				klog.Errorf("系统提示词不存在: ID=%d", req.SystemPromptID)
			} else if !sp.IsActive {
				klog.Errorf("系统提示词未激活: ID=%d, Name=%s, IsActive=%v", sp.ID, sp.Name, sp.IsActive)
			} else {
				klog.Infof("成功获取系统提示词: ID=%d, Name=%s, Content='%s'", sp.ID, sp.Name, sp.Content)
				systemPromptContent = sp.Content
				systemPromptMeta = map[string]interface{}{
					"system_prompt_id":     sp.ID,
					"system_prompt_name":   sp.Name,
					"system_prompt_source": "id",
				}
			}
		} else {
			// 3. 使用默认系统提示词
			klog.Infof("尝试获取默认系统提示词")
			sp, err := c.factory.SystemPrompt().GetActiveDefault(ctx)
			if err != nil {
				klog.Errorf("获取默认系统提示词失败: %v", err)
			} else if sp == nil {
				klog.Errorf("默认系统提示词不存在")
			} else if sp.Content == "" {
				klog.Errorf("默认系统提示词内容为空: ID=%d, Name=%s", sp.ID, sp.Name)
			} else {
				klog.Infof("成功获取默认系统提示词: ID=%d, Name=%s, Content='%s'", sp.ID, sp.Name, sp.Content)
				systemPromptContent = sp.Content
				systemPromptMeta = map[string]interface{}{
					"system_prompt_id":     sp.ID,
					"system_prompt_name":   sp.Name,
					"system_prompt_source": "default",
				}
			}
		}

		// 注入系统提示词到上下文
		if systemPromptContent != "" {
			klog.Infof("注入系统提示词到上下文: sessionID=%s, content='%s', meta=%+v", sessionID, systemPromptContent, systemPromptMeta)
			if err := c.openaiClient.(*openai.OpenAIClient).GetContextManager().SaveMessage(sessionID, "system", systemPromptContent, systemPromptMeta); err != nil {
				klog.Errorf("保存系统提示词到上下文失败: %v", err)
			} else {
				klog.Infof("系统提示词注入成功")
			}
		} else {
			klog.Warningf("没有可用的系统提示词，跳过注入")
		}
	} else {
		klog.Infof("使用现有会话，跳过系统提示词注入: sessionID=%s", sessionID)
	}

	// 调用OpenAI API
	resp, err := c.openaiClient.ChatWithContext(sessionID, req.Message, req.UserID)
	if err != nil {
		klog.Errorf("OpenAI API调用失败: %v", err)
		return nil, fmt.Errorf("AI服务调用失败: %w", err)
	}

	// 构建响应
	response := &types.OpenAIChatResponse{
		SessionID: sessionID,
		Message:   resp.Choices[0].Message.Content,
		Usage: types.OpenAIUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		CreatedAt: time.Now().UnixMilli(),
	}

	klog.Infof("OpenAI聊天完成: session=%s, tokens=%d", sessionID, resp.Usage.TotalTokens)
	return response, nil
}

// StreamMessage 流式发送聊天消息
func (c *Controller) StreamMessage(ctx context.Context, req *types.OpenAIChatRequest, streamChan chan<- interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("StreamMessage panic: %v", r)
			select {
			case streamChan <- map[string]interface{}{"event": "error", "data": fmt.Sprintf("内部错误: %v", r)}:
			case <-ctx.Done():
			default:
			}
		}
	}()

	// 验证必填字段
	if req.Message == "" {
		select {
		case streamChan <- map[string]interface{}{"event": "error", "data": "消息内容不能为空"}:
		case <-ctx.Done():
		}
		return fmt.Errorf("消息内容不能为空")
	}
	if req.UserID == "" {
		select {
		case streamChan <- map[string]interface{}{"event": "error", "data": "用户ID不能为空"}:
		case <-ctx.Done():
		}
		return fmt.Errorf("用户ID不能为空")
	}

	// 如果没有提供会话ID，创建新会话
	sessionID := req.SessionID
	isNewSession := false

	klog.Infof("会话检查: 请求中的sessionID='%s', 是否为空=%v", sessionID, sessionID == "")

	if sessionID == "" {
		var err error
		sessionID, err = c.openaiClient.(*openai.OpenAIClient).GetContextManager().(*openai.ChatContextManager).CreateSession(req.UserID)
		if err != nil {
			select {
			case streamChan <- map[string]interface{}{"event": "error", "data": fmt.Sprintf("创建会话失败: %v", err)}:
			case <-ctx.Done():
			}
			return fmt.Errorf("创建会话失败: %w", err)
		}
		isNewSession = true
		klog.Infof("创建新会话: sessionID=%s", sessionID)
	} else {
		klog.Infof("使用现有会话: sessionID=%s", sessionID)
	}

	// 只在新会话时注入系统提示词：优先级为直接内容 > 指定ID > 默认提示词
	if isNewSession {
		var systemPromptContent string
		var systemPromptMeta map[string]interface{}

		klog.Infof("新会话，开始系统提示词注入检查: SystemPromptContent='%s', SystemPromptID=%d", req.SystemPromptContent, req.SystemPromptID)

		if req.SystemPromptContent != "" {
			// 1. 优先使用直接提供的系统提示词内容
			klog.Infof("使用直接提供的系统提示词内容")
			systemPromptContent = req.SystemPromptContent
			systemPromptMeta = map[string]interface{}{
				"system_prompt_source": "direct",
			}
		} else if req.SystemPromptID > 0 {
			// 2. 使用指定的系统提示词ID
			klog.Infof("尝试获取系统提示词ID: %d", req.SystemPromptID)
			sp, err := c.factory.SystemPrompt().Get(ctx, req.SystemPromptID)
			if err != nil {
				klog.Errorf("获取系统提示词失败: %v", err)
			} else if sp == nil {
				klog.Errorf("系统提示词不存在: ID=%d", req.SystemPromptID)
			} else if !sp.IsActive {
				klog.Errorf("系统提示词未激活: ID=%d, Name=%s, IsActive=%v", sp.ID, sp.Name, sp.IsActive)
			} else {
				klog.Infof("成功获取系统提示词: ID=%d, Name=%s, Content='%s'", sp.ID, sp.Name, sp.Content)
				systemPromptContent = sp.Content
				systemPromptMeta = map[string]interface{}{
					"system_prompt_id":     sp.ID,
					"system_prompt_name":   sp.Name,
					"system_prompt_source": "id",
				}
			}
		} else {
			// 3. 使用默认系统提示词
			klog.Infof("尝试获取默认系统提示词")
			sp, err := c.factory.SystemPrompt().GetActiveDefault(ctx)
			if err != nil {
				klog.Errorf("获取默认系统提示词失败: %v", err)
			} else if sp == nil {
				klog.Errorf("默认系统提示词不存在")
			} else if sp.Content == "" {
				klog.Errorf("默认系统提示词内容为空: ID=%d, Name=%s", sp.ID, sp.Name)
			} else {
				klog.Infof("成功获取默认系统提示词: ID=%d, Name=%s, Content='%s'", sp.ID, sp.Name, sp.Content)
				systemPromptContent = sp.Content
				systemPromptMeta = map[string]interface{}{
					"system_prompt_id":     sp.ID,
					"system_prompt_name":   sp.Name,
					"system_prompt_source": "default",
				}
			}
		}

		// 注入系统提示词到上下文
		if systemPromptContent != "" {
			klog.Infof("注入系统提示词到上下文: sessionID=%s, content='%s', meta=%+v", sessionID, systemPromptContent, systemPromptMeta)
			if err := c.openaiClient.(*openai.OpenAIClient).GetContextManager().SaveMessage(sessionID, "system", systemPromptContent, systemPromptMeta); err != nil {
				klog.Errorf("保存系统提示词到上下文失败: %v", err)
			} else {
				klog.Infof("系统提示词注入成功")
			}
		} else {
			klog.Warningf("没有可用的系统提示词，跳过注入")
		}
	} else {
		klog.Infof("使用现有会话，跳过系统提示词注入: sessionID=%s", sessionID)
	}

	// 调用流式OpenAI API
	eventChan, err := c.openaiClient.ChatWithContextStream(sessionID, req.Message, req.UserID)
	if err != nil {
		klog.Errorf("OpenAI流式API调用失败: %v", err)
		select {
		case streamChan <- map[string]interface{}{"event": "error", "data": fmt.Sprintf("AI服务调用失败: %v", err)}:
		case <-ctx.Done():
		}
		return fmt.Errorf("AI服务调用失败: %w", err)
	}

	// 累积 AI 回复内容
	var fullAIContent string

	// 处理流式事件
	for event := range eventChan {
		// 调试日志：打印事件内容
		// 检查事件是否有内容
		if len(event.Choices) > 0 {
			choice := event.Choices[0]
			var content string

			// 优先使用delta字段（流式响应），其次使用message字段
			if choice.Delta.Content != "" {
				content = choice.Delta.Content
			} else if choice.Message.Content != "" {
				content = choice.Message.Content
			}

			if content != "" {
				// 累积 AI 回复内容
				fullAIContent += content

				select {
				case streamChan <- map[string]interface{}{
					"event": "message",
					"data": map[string]interface{}{
						"session_id": sessionID,
						"content":    content,
						"timestamp":  time.Now().UnixMilli(),
					},
				}:
				case <-ctx.Done():
					klog.Infof("上下文已取消，停止流式响应")
					return ctx.Err()
				default:
					klog.Warningf("流式通道已满，丢弃事件")
				}
			} else {
				klog.Infof("跳过空内容事件: delta_content='%s', message_content='%s'", choice.Delta.Content, choice.Message.Content)
			}
		} else {
			klog.Infof("跳过空choices事件")
		}
	}

	// 发送完成事件
	select {
	case streamChan <- map[string]interface{}{
		"event": "completed",
		"data": map[string]interface{}{
			"session_id": sessionID,
			"timestamp":  time.Now().UnixMilli(),
		},
	}:
	case <-ctx.Done():
		return ctx.Err()
	default:
		klog.Warningf("流式通道已满，丢弃完成事件")
	}

	// 存储 AI 回复消息
	if fullAIContent != "" {
		aiMessageID := generateMessageID()
		aiMessage := &types.ChatMessage{
			SessionID: sessionID,
			MessageID: aiMessageID,
			Event:     "assistant",   // AI 助手角色
			Content:   fullAIContent, // 完整的 AI 回复内容
			CreatedAt: time.Now().UnixMilli(),
		}

		klog.Infof("准备保存 OpenAI AI 回复: SessionID=%s, MessageID=%s, Content长度=%d",
			sessionID, aiMessageID, len(fullAIContent))

		if err := c.factory.Chat().SaveChatMessage(ctx, aiMessage); err != nil {
			klog.Errorf("保存 OpenAI AI 回复消息失败: %v，继续处理", err)
		} else {
			klog.Infof("成功保存 OpenAI AI 回复消息，SessionID=%s, MessageID=%s, 内容长度: %d",
				sessionID, aiMessageID, len(fullAIContent))
		}
	} else {
		klog.Warningf("AI 回复内容为空，跳过存储: sessionID=%s", sessionID)
	}

	// 异步生成会话标题（仅对新会话）
	if isNewSession && fullAIContent != "" {
		go c.generateSessionTitleAsync(ctx, sessionID, req.Message)
	}

	klog.Infof("OpenAI流式聊天完成: session=%s", sessionID)
	return nil
}

// 辅助函数
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixMilli())
}

// generateSessionTitleAsync 异步生成会话标题
func (c *Controller) generateSessionTitleAsync(ctx context.Context, sessionID, userQuery string) {
	// 创建标题生成器
	titleGenerator := openai.NewTitleGenerator(c.openaiClient, &openai.Config{
		APIKey:      c.config.APIKey,
		BaseURL:     c.config.BaseURL,
		Timeout:     c.config.Timeout,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Model:       c.config.Model,
	})

	// 使用标题生成器的异步方法
	titleGenerator.GenerateTitleAsync(ctx, userQuery, func(title string, err error) {
		if err != nil {
			klog.Errorf("生成会话标题失败: %v", err)
			return
		}

		// 更新会话标题
		if err := c.factory.Chat().UpdateChatSessionTitle(ctx, sessionID, title); err != nil {
			klog.Errorf("更新会话标题失败: %v", err)
		} else {
			klog.Infof("成功更新会话标题: sessionID=%s, title=%s", sessionID, title)
		}
	})
}

// GetChatHistory 获取聊天历史
func (c *Controller) GetChatHistory(ctx context.Context, sessionID string, limit int) ([]*types.ChatMessage, error) {
	return c.factory.Chat().GetChatHistory(ctx, sessionID, limit)
}

// CreateSession 创建新会话
func (c *Controller) CreateSession(ctx context.Context, userID string) (string, error) {
	return c.openaiClient.(*openai.OpenAIClient).GetContextManager().(*openai.ChatContextManager).CreateSession(userID)
}

// CloseSession 关闭会话
func (c *Controller) CloseSession(ctx context.Context, sessionID string) error {
	return c.openaiClient.(*openai.OpenAIClient).GetContextManager().(*openai.ChatContextManager).CloseSession(sessionID)
}

// ---- System Prompt CRUD ----

// CreateSystemPrompt 创建系统提示词
func (c *Controller) CreateSystemPrompt(ctx context.Context, sp *dbmodel.SystemPrompt) error {
	if sp == nil || sp.Name == "" || sp.Content == "" {
		return fmt.Errorf("参数不合法")
	}

	// 检查名称是否已存在
	existing, err := c.factory.SystemPrompt().GetByName(ctx, sp.Name)
	if err == nil && existing != nil {
		return fmt.Errorf("系统提示词名称 '%s' 已存在", sp.Name)
	}

	return c.factory.SystemPrompt().Create(ctx, sp)
}

// UpdateSystemPrompt 更新系统提示词
func (c *Controller) UpdateSystemPrompt(ctx context.Context, sp *dbmodel.SystemPrompt, updateActive bool) error {
	if sp == nil || sp.ID == 0 {
		return fmt.Errorf("参数不合法")
	}
	if sp.Name == "" && sp.Content == "" && sp.Role == "" && sp.Tags == "" && !updateActive {
		return fmt.Errorf("无更新内容")
	}

	// 如果要更新名称，检查新名称是否已存在
	if sp.Name != "" {
		existing, err := c.factory.SystemPrompt().GetByName(ctx, sp.Name)
		if err == nil && existing != nil && existing.ID != sp.ID {
			return fmt.Errorf("系统提示词名称 '%s' 已存在", sp.Name)
		}
	}

	sp.Version++
	return c.factory.SystemPrompt().UpdateWithActive(ctx, sp, updateActive)
}

// UpdateSystemPromptStatus 更新系统提示词激活状态
func (c *Controller) UpdateSystemPromptStatus(ctx context.Context, id int64, isActive bool) error {
	if id <= 0 {
		return fmt.Errorf("参数不合法：ID必须大于0")
	}
	return c.factory.SystemPrompt().UpdateStatus(ctx, id, isActive)
}

// SetSystemPromptAsDefault 设置系统提示词为默认
func (c *Controller) SetSystemPromptAsDefault(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("参数不合法：ID必须大于0")
	}

	// 检查系统提示词是否存在且激活
	sp, err := c.factory.SystemPrompt().Get(ctx, id)
	if err != nil {
		return fmt.Errorf("获取系统提示词失败: %w", err)
	}
	if sp == nil {
		return fmt.Errorf("系统提示词不存在")
	}
	if !sp.IsActive {
		return fmt.Errorf("系统提示词未激活，无法设为默认")
	}

	return c.factory.SystemPrompt().SetAsDefault(ctx, id)
}

// DeleteSystemPrompt 删除系统提示词
func (c *Controller) DeleteSystemPrompt(ctx context.Context, id int64) error {
	if id == 0 {
		return fmt.Errorf("参数不合法")
	}
	return c.factory.SystemPrompt().Delete(ctx, id)
}

// BatchDeleteSystemPrompts 批量删除系统提示词
func (c *Controller) BatchDeleteSystemPrompts(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("参数不合法：ID列表不能为空")
	}

	// 验证ID的有效性
	for _, id := range ids {
		if id <= 0 {
			return fmt.Errorf("参数不合法：ID必须大于0")
		}
	}

	return c.factory.SystemPrompt().BatchDelete(ctx, ids)
}

// GetSystemPrompt 获取系统提示词
func (c *Controller) GetSystemPrompt(ctx context.Context, id int64) (*dbmodel.SystemPrompt, error) {
	if id == 0 {
		return nil, fmt.Errorf("参数不合法")
	}
	return c.factory.SystemPrompt().Get(ctx, id)
}

// SearchSystemPromptsByName 按名称模糊搜索系统提示词
func (c *Controller) SearchSystemPromptsByName(ctx context.Context, name string, limit int) ([]*dbmodel.SystemPrompt, error) {
	if name == "" {
		return nil, fmt.Errorf("搜索名称不能为空")
	}
	if limit <= 0 {
		limit = 20 // 默认限制
	}
	if limit > 100 {
		limit = 100 // 最大限制
	}
	return c.factory.SystemPrompt().SearchByName(ctx, name, limit)
}

// ListSystemPrompts 列表系统提示词
func (c *Controller) ListSystemPrompts(ctx context.Context, name, role string, active *bool, offset, limit int) ([]*dbmodel.SystemPrompt, int64, error) {
	filter := &db.SystemPromptFilter{
		Name:   name,
		Role:   role,
		Active: active,
		Offset: offset,
		Limit:  limit,
	}
	return c.factory.SystemPrompt().List(ctx, filter)
}
