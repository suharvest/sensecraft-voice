package recording

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	ctrl "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/recording"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (r *recordingRouter) wsStream(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Status(500)
		return
	}
	defer conn.Close()

	for {
		var msg ctrl.SaveRequest
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		if _, err := r.c.Recording().Save(c, msg); err != nil {
			// 失败也返回 ack=false，便于端重试
			_ = conn.WriteJSON(gin.H{"ack": false})
			continue
		}
		_ = conn.WriteJSON(gin.H{"ack": true})
	}
}

func (r *recordingRouter) save(c *gin.Context) {
	resp := httputils.NewResponse()

	var req ctrl.SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	klog.Infof("save req: %+v", req)
	result, err := r.c.Recording().Save(c, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *recordingRouter) list(c *gin.Context) {
	resp := httputils.NewResponse()

	var req ctrl.ListRequest
	// 忽略绑定错误，使用宽松的参数处理
	c.ShouldBindQuery(&req)

	// 手动处理状态参数，因为需要转换为指针类型
	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.ParseInt(statusStr, 10, 8); err == nil {
			statusInt8 := int8(status)
			req.Status = &statusInt8
		}
	}

	result, err := r.c.Recording().List(c, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *recordingRouter) query(c *gin.Context) {
	resp := httputils.NewResponse()

	// 验证 token
	authHeader := c.GetHeader("authorization")
	if authHeader == "" {
		httputils.SetFailed(c, resp, fmt.Errorf("missing authorization header"))
		return
	}

	// 检查 token 是否匹配固定值
	expectedToken := "Bearer voiceai_secure_token_2025"
	if authHeader != expectedToken {
		httputils.SetFailed(c, resp, fmt.Errorf("invalid token"))
		return
	}

	// 绑定请求参数
	var req ctrl.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 调用 controller 查询录音记录
	result, err := r.c.Recording().Query(c, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 返回查询结果
	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *recordingRouter) getKeywordMatches(c *gin.Context) {
	resp := httputils.NewResponse()

	// 手动解析参数，避免验证问题
	req := ctrl.KeywordMatchRequest{
		Offset:     0,
		Limit:      0, // 不设置默认值，如果没有传 limit 就查询所有记录
		MacAddress: c.Query("mac_address"),
	}

	// 解析 offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	// 解析 limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 10000 {
			req.Limit = limit
		}
	}

	// 解析 keyword_id
	if keywordIDStr := c.Query("keyword_id"); keywordIDStr != "" {
		if keywordID, err := strconv.ParseInt(keywordIDStr, 10, 64); err == nil {
			req.KeywordID = keywordID
		}
	}

	// 解析 recording_id
	if recordingIDStr := c.Query("recording_id"); recordingIDStr != "" {
		if recordingID, err := strconv.ParseInt(recordingIDStr, 10, 64); err == nil {
			req.RecordingID = recordingID
		}
	}

	// 解析 start_time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if startTime, err := strconv.ParseInt(startTimeStr, 10, 64); err == nil && startTime > 0 {
			req.StartTime = startTime
		}
	}

	// 解析 end_time
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if endTime, err := strconv.ParseInt(endTimeStr, 10, 64); err == nil && endTime > 0 {
			req.EndTime = endTime
		}
	}

	// 解析 store_id
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if storeID, err := strconv.ParseInt(storeIDStr, 10, 64); err == nil && storeID > 0 {
			req.StoreID = storeID
		}
	}

	result, err := r.c.Recording().GetKeywordMatches(c, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}
