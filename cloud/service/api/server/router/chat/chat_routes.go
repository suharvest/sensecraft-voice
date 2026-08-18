package chat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

// streamMessage 流式发送聊天消息
func (r *chatRouter) streamMessage(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("streamMessage panic: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		}
	}()

	var req types.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 验证必填字段
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "查询内容不能为空"})
		return
	}

	if req.User == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户标识不能为空"})
		return
	}

	// 设置默认值
	if req.ResponseMode == "" {
		req.ResponseMode = "streaming"
	}

	// klog.Infof("开始流式聊天请求，用户: %s, 查询: %s, conversation_id: '%s'",
	// 	req.User, req.Query, req.ConversationID)

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 创建流式响应通道
	streamChan := make(chan interface{}, 100)
	defer close(streamChan)

	// 在goroutine中处理流式响应
	go func() {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("流式处理goroutine panic: %v", r)
				// 发送错误事件
				select {
				case streamChan <- gin.H{"event": "error", "data": fmt.Sprintf("内部错误: %v", r)}:
				default:
					klog.Warningf("流式通道已满，丢弃panic错误事件")
				}
			}
		}()

		ctx := c.Request.Context()
		if err := r.c.Chat().StreamMessage(ctx, &req, streamChan); err != nil {
			klog.Errorf("流式处理失败: %v", err)
			// 发送错误事件
			select {
			case streamChan <- gin.H{"event": "error", "data": err.Error()}:
			default:
				klog.Warningf("流式通道已满，丢弃处理失败错误事件")
			}
		}
	}()

	// 流式响应处理
	c.Stream(func(w io.Writer) bool {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("流式响应处理panic: %v", r)
			}
		}()

		select {
		case data, ok := <-streamChan:
			if !ok {
				klog.Infof("流式通道关闭，结束响应")
				return false // 通道关闭，结束流
			}

			// 发送SSE格式数据
			jsonData, err := json.Marshal(data)
			if err != nil {
				klog.Errorf("JSON序列化失败: %v", err)
				// 发送错误事件
				fmt.Fprintf(w, "data: {\"event\": \"error\", \"data\": \"JSON序列化失败\"}\n\n")
				return true
			}

			// 写入SSE格式
			if _, err := fmt.Fprintf(w, "data: %s\n\n", jsonData); err != nil {
				klog.Errorf("写入SSE数据失败: %v", err)
				return false
			}
			return true

		case <-c.Request.Context().Done():
			klog.Infof("客户端断开连接")
			return false // 客户端断开连接
		}
	})
}

// sendMessage 发送聊天消息
func (r *chatRouter) sendMessage(c *gin.Context) {
	resp := httputils.NewResponse()

	var req types.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 验证必填字段
	if req.Query == "" {
		resp.Message = "查询内容不能为空"
		httputils.SetFailed(c, resp, fmt.Errorf("查询内容不能为空"))
		return
	}

	if req.User == "" {
		resp.Message = "用户标识不能为空"
		httputils.SetFailed(c, resp, fmt.Errorf("用户标识不能为空"))
		return
	}

	// 设置默认值
	if req.ResponseMode == "" {
		req.ResponseMode = "streaming"
	}

	// 调用控制器发送消息
	ctx := c.Request.Context()
	if err := r.c.Chat().SendMessage(ctx, &req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"conversation_id": req.ConversationID,
		"status":          "sent",
	}
	httputils.SetSuccess(c, resp)
}

// getChatHistory 获取聊天历史
func (r *chatRouter) getChatHistory(c *gin.Context) {
	resp := httputils.NewResponse()

	sessionID := c.Param("session_id")
	if sessionID == "" {
		resp.Message = "会话ID不能为空"
		httputils.SetFailed(c, resp, fmt.Errorf("会话ID不能为空"))
		return
	}

	// 获取分页参数
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	// 调用控制器获取聊天历史
	ctx := c.Request.Context()
	messages, err := r.c.Chat().GetChatHistory(ctx, sessionID, limit)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"session_id": sessionID,
		"messages":   messages,
		"total":      len(messages),
	}
	httputils.SetSuccess(c, resp)
}

// getChatSession 获取聊天会话
func (r *chatRouter) getChatSession(c *gin.Context) {
	resp := httputils.NewResponse()

	sessionID := c.Param("session_id")
	if sessionID == "" {
		resp.Message = "会话ID不能为空"
		httputils.SetFailed(c, resp, fmt.Errorf("会话ID不能为空"))
		return
	}

	// 调用控制器获取聊天会话
	ctx := c.Request.Context()
	session, err := r.c.Chat().GetChatSession(ctx, sessionID)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = session
	httputils.SetSuccess(c, resp)
}

// getChatSessions 获取聊天会话列表
func (r *chatRouter) getChatSessions(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取查询参数
	userID := c.Query("user_id")

	// 调用控制器获取聊天会话列表
	ctx := c.Request.Context()
	sessions, err := r.c.Chat().GetChatSessions(ctx, userID)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	}
	httputils.SetSuccess(c, resp)
}

// deleteChatSession 删除单个聊天会话
func (r *chatRouter) deleteChatSession(c *gin.Context) {
	resp := httputils.NewResponse()

	sessionID := c.Param("session_id")
	if sessionID == "" {
		resp.Message = "会话ID不能为空"
		httputils.SetFailed(c, resp, fmt.Errorf("会话ID不能为空"))
		return
	}

	// 调用控制器删除聊天会话
	ctx := c.Request.Context()
	if err := r.c.Chat().DeleteChatSession(ctx, sessionID); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"session_id": sessionID,
		"status":     "deleted",
	}
	httputils.SetSuccess(c, resp)
}

// deleteChatSessions 批量删除聊天会话
func (r *chatRouter) deleteChatSessions(c *gin.Context) {
	resp := httputils.NewResponse()

	// 绑定请求参数
	var req struct {
		SessionIDs []string `json:"session_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 验证参数
	if len(req.SessionIDs) == 0 {
		resp.Message = "会话ID列表不能为空"
		httputils.SetFailed(c, resp, fmt.Errorf("会话ID列表不能为空"))
		return
	}

	// 调用控制器批量删除聊天会话
	ctx := c.Request.Context()
	if err := r.c.Chat().DeleteChatSessions(ctx, req.SessionIDs); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"session_ids": req.SessionIDs,
		"count":       len(req.SessionIDs),
		"status":      "deleted",
	}
	httputils.SetSuccess(c, resp)
}

// updateChatSessionTitle 更新聊天会话标题
func (r *chatRouter) updateChatSessionTitle(c *gin.Context) {
	resp := httputils.NewResponse()

	sessionID := c.Param("session_id")
	if sessionID == "" {
		resp.Message = "会话ID不能为空"
		httputils.SetFailed(c, resp, fmt.Errorf("会话ID不能为空"))
		return
	}

	// 绑定请求参数
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 验证标题长度
	if len(req.Title) > 255 {
		resp.Message = "标题长度不能超过255个字符"
		httputils.SetFailed(c, resp, fmt.Errorf("标题长度不能超过255个字符"))
		return
	}

	// 调用控制器更新会话标题
	ctx := c.Request.Context()
	if err := r.c.Chat().UpdateChatSessionTitle(ctx, sessionID, req.Title); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"session_id": sessionID,
		"title":      req.Title,
		"status":     "updated",
	}
	httputils.SetSuccess(c, resp)
}
