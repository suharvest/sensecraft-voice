package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	dbmodel "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

// sendMessage 发送聊天消息
func (r *openaiRouter) sendMessage(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("sendMessage panic: %v", r)
		}
	}()

	resp := httputils.NewResponse()

	var req types.OpenAIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 验证必填字段
	if req.Message == "" {
		httputils.SetFailed(c, resp, fmt.Errorf("消息内容不能为空"))
		return
	}
	if req.UserID == "" {
		httputils.SetFailed(c, resp, fmt.Errorf("用户ID不能为空"))
		return
	}

	// 调用控制器
	result, err := r.c.OpenAI().SendMessage(c.Request.Context(), &req)
	if err != nil {
		klog.Errorf("发送OpenAI消息失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

// streamMessage 流式发送聊天消息
func (r *openaiRouter) streamMessage(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			klog.Errorf("streamMessage panic: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "内部服务器错误"})
		}
	}()

	var req types.OpenAIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 验证必填字段
	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息内容不能为空"})
		return
	}
	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户ID不能为空"})
		return
	}

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
			}
		}()

		if err := r.c.OpenAI().StreamMessage(c.Request.Context(), &req, streamChan); err != nil {
			klog.Errorf("流式OpenAI消息失败: %v", err)
			select {
			case streamChan <- gin.H{"event": "error", "data": fmt.Sprintf("AI服务调用失败: %v", err)}:
			case <-c.Request.Context().Done():
			default:
			}
		}
	}()

	// 流式输出响应
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-streamChan:
			if !ok {
				return false
			}

			// 发送SSE格式的数据
			eventData, err := json.Marshal(event)
			if err != nil {
				klog.Errorf("序列化事件失败: %v", err)
				return false
			}

			fmt.Fprintf(w, "data: %s\n\n", string(eventData))
			return true

		case <-c.Request.Context().Done():
			klog.Infof("客户端断开连接")
			return false
		}
	})
}

// getChatHistory 获取聊天历史
func (r *openaiRouter) getChatHistory(c *gin.Context) {
	resp := httputils.NewResponse()

	sessionID := c.Param("session_id")
	if sessionID == "" {
		httputils.SetFailed(c, resp, fmt.Errorf("会话ID不能为空"))
		return
	}

	// 获取limit参数
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // 限制最大数量
	}

	// 调用控制器
	messages, err := r.c.OpenAI().GetChatHistory(c.Request.Context(), sessionID, limit)
	if err != nil {
		klog.Errorf("获取聊天历史失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = messages
	httputils.SetSuccess(c, resp)
}

// createSession 创建新会话
func (r *openaiRouter) createSession(c *gin.Context) {
	resp := httputils.NewResponse()

	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 调用控制器
	sessionID, err := r.c.OpenAI().CreateSession(c.Request.Context(), req.UserID)
	if err != nil {
		klog.Errorf("创建会话失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"session_id": sessionID,
	}
	httputils.SetSuccess(c, resp)
}

// closeSession 关闭会话
func (r *openaiRouter) closeSession(c *gin.Context) {
	resp := httputils.NewResponse()

	sessionID := c.Param("session_id")
	if sessionID == "" {
		httputils.SetFailed(c, resp, fmt.Errorf("会话ID不能为空"))
		return
	}

	// 调用控制器
	if err := r.c.OpenAI().CloseSession(c.Request.Context(), sessionID); err != nil {
		klog.Errorf("关闭会话失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"message": "会话已关闭",
	}
	httputils.SetSuccess(c, resp)
}

// ---- SystemPrompt CRUD ----

// createSystemPrompt 创建系统提示词
func (r *openaiRouter) createSystemPrompt(c *gin.Context) {
	resp := httputils.NewResponse()

	var req struct {
		Name      string `json:"name" binding:"required"`
		Role      string `json:"role"`
		Content   string `json:"content" binding:"required"`
		Tags      string `json:"tags"`
		Active    *bool  `json:"is_active"`
		IsActive  *bool  `json:"IsActive"` // 支持大写字段名
		Default   *bool  `json:"is_default"`
		IsDefault *bool  `json:"IsDefault"` // 支持大写字段名
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 处理激活状态字段，支持两种格式
	var activeValue *bool
	if req.Active != nil {
		activeValue = req.Active
		klog.Infof("创建时使用小写字段 is_active: %v", *req.Active)
	} else if req.IsActive != nil {
		activeValue = req.IsActive
		klog.Infof("创建时使用大写字段 IsActive: %v", *req.IsActive)
	}

	// 处理默认状态字段，支持两种格式
	var defaultValue *bool
	if req.Default != nil {
		defaultValue = req.Default
		klog.Infof("创建时使用小写字段 is_default: %v", *req.Default)
	} else if req.IsDefault != nil {
		defaultValue = req.IsDefault
		klog.Infof("创建时使用大写字段 IsDefault: %v", *req.IsDefault)
	}

	sp := &dbmodel.SystemPrompt{
		Name:      req.Name,
		Role:      req.Role,
		Content:   req.Content,
		Tags:      req.Tags,
		IsActive:  true,  // 默认值
		IsDefault: false, // 默认值
		Version:   1,
	}
	if activeValue != nil {
		sp.IsActive = *activeValue
		klog.Infof("创建时设置 IsActive = %v", sp.IsActive)
	}
	if defaultValue != nil {
		sp.IsDefault = *defaultValue
		klog.Infof("创建时设置 IsDefault = %v", sp.IsDefault)
	}

	if err := r.c.OpenAI().CreateSystemPrompt(c.Request.Context(), sp); err != nil {
		klog.Errorf("创建系统提示词失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = sp
	httputils.SetSuccess(c, resp)
}

// listSystemPrompts 列表系统提示词
func (r *openaiRouter) listSystemPrompts(c *gin.Context) {
	resp := httputils.NewResponse()

	name := c.Query("name")
	role := c.Query("role")
	activeParam := c.Query("active")

	var activePtr *bool
	if activeParam != "" {
		v := activeParam == "true" || activeParam == "1"
		activePtr = &v
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	sps, total, err := r.c.OpenAI().ListSystemPrompts(c.Request.Context(), name, role, activePtr, offset, limit)
	if err != nil {
		klog.Errorf("获取系统提示词列表失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"items":  sps,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	}
	httputils.SetSuccess(c, resp)
}

// searchSystemPromptsByName 按名称模糊搜索系统提示词
func (r *openaiRouter) searchSystemPromptsByName(c *gin.Context) {
	resp := httputils.NewResponse()

	name := c.Query("name")
	if name == "" {
		httputils.SetFailed(c, resp, fmt.Errorf("搜索名称不能为空"))
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	sps, err := r.c.OpenAI().SearchSystemPromptsByName(c.Request.Context(), name, limit)
	if err != nil {
		klog.Errorf("搜索系统提示词失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"items": sps,
		"count": len(sps),
		"name":  name,
		"limit": limit,
	}
	httputils.SetSuccess(c, resp)
}

// getSystemPrompt 获取系统提示词
func (r *openaiRouter) getSystemPrompt(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httputils.SetFailed(c, resp, err)
		return
	}

	sp, err := r.c.OpenAI().GetSystemPrompt(c.Request.Context(), id)
	if err != nil {
		klog.Errorf("获取系统提示词失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = sp
	httputils.SetSuccess(c, resp)
}

// updateSystemPrompt 更新系统提示词
func (r *openaiRouter) updateSystemPrompt(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req struct {
		Name      string `json:"name"`
		Role      string `json:"role"`
		Content   string `json:"content"`
		Tags      string `json:"tags"`
		Active    *bool  `json:"is_active"`
		IsActive  *bool  `json:"IsActive"` // 支持大写字段名
		Default   *bool  `json:"is_default"`
		IsDefault *bool  `json:"IsDefault"` // 支持大写字段名
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	sp := &dbmodel.SystemPrompt{
		ID:      id,
		Name:    req.Name,
		Role:    req.Role,
		Content: req.Content,
		Tags:    req.Tags,
	}
	// 处理激活状态字段，支持两种格式
	var activeValue *bool
	if req.Active != nil {
		activeValue = req.Active
		klog.Infof("使用小写字段 is_active: %v", *req.Active)
	} else if req.IsActive != nil {
		activeValue = req.IsActive
		klog.Infof("使用大写字段 IsActive: %v", *req.IsActive)
	}

	// 处理默认状态字段，支持两种格式
	var defaultValue *bool
	if req.Default != nil {
		defaultValue = req.Default
		klog.Infof("使用小写字段 is_default: %v", *req.Default)
	} else if req.IsDefault != nil {
		defaultValue = req.IsDefault
		klog.Infof("使用大写字段 IsDefault: %v", *req.IsDefault)
	}

	// 只有当用户明确提供 is_active 或 is_default 字段时才更新状态
	if activeValue != nil {
		sp.IsActive = *activeValue
		klog.Infof("设置 IsActive = %v", sp.IsActive)
	}
	if defaultValue != nil {
		sp.IsDefault = *defaultValue
		klog.Infof("设置 IsDefault = %v", sp.IsDefault)
	}

	// 检查是否提供了状态字段
	updateActive := activeValue != nil || defaultValue != nil
	klog.Infof("updateActive = %v (包含 is_active 或 is_default 更新)", updateActive)

	if err := r.c.OpenAI().UpdateSystemPrompt(c.Request.Context(), sp, updateActive); err != nil {
		klog.Errorf("更新系统提示词失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = sp
	httputils.SetSuccess(c, resp)
}

// updateSystemPromptStatus 更新系统提示词激活状态
func (r *openaiRouter) updateSystemPromptStatus(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req struct {
		IsActive *bool `json:"is_active" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.OpenAI().UpdateSystemPromptStatus(c.Request.Context(), id, *req.IsActive); err != nil {
		klog.Errorf("更新系统提示词状态失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	statusText := "激活"
	if !*req.IsActive {
		statusText = "停用"
	}

	resp.Result = gin.H{
		"message":   fmt.Sprintf("系统提示词状态已更新为%s", statusText),
		"id":        id,
		"is_active": *req.IsActive,
	}
	httputils.SetSuccess(c, resp)
}

// setSystemPromptAsDefault 设置系统提示词为默认
func (r *openaiRouter) setSystemPromptAsDefault(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.OpenAI().SetSystemPromptAsDefault(c.Request.Context(), id); err != nil {
		klog.Errorf("设置系统提示词为默认失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"message": "设置成功",
	}
	httputils.SetSuccess(c, resp)
}

// deleteSystemPrompt 删除系统提示词
func (r *openaiRouter) deleteSystemPrompt(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.OpenAI().DeleteSystemPrompt(c.Request.Context(), id); err != nil {
		klog.Errorf("删除系统提示词失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"message": "删除成功",
	}
	httputils.SetSuccess(c, resp)
}

// batchDeleteSystemPrompts 批量删除系统提示词
func (r *openaiRouter) batchDeleteSystemPrompts(c *gin.Context) {
	resp := httputils.NewResponse()

	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 验证ID列表
	if len(req.IDs) == 0 {
		httputils.SetFailed(c, resp, fmt.Errorf("ID列表不能为空"))
		return
	}

	// 限制批量删除的数量，防止误操作
	if len(req.IDs) > 100 {
		httputils.SetFailed(c, resp, fmt.Errorf("批量删除数量不能超过100个"))
		return
	}

	if err := r.c.OpenAI().BatchDeleteSystemPrompts(c.Request.Context(), req.IDs); err != nil {
		klog.Errorf("批量删除系统提示词失败: %v", err)
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{
		"message": fmt.Sprintf("成功删除 %d 个系统提示词", len(req.IDs)),
		"count":   len(req.IDs),
	}
	httputils.SetSuccess(c, resp)
}
