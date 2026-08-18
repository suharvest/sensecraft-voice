package stats

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/errors"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/log"
)

func (r *statsRouter) getDashboardStats(c *gin.Context) {
	resp := httputils.NewResponse()

	// 获取可选的 store_id 参数
	var storeID *int64
	if storeIDStr := c.Query("store_id"); storeIDStr != "" {
		if id, err := strconv.ParseInt(storeIDStr, 10, 64); err == nil {
			storeID = &id
		} else {
			log.Errorf("解析store_id参数失败: %v, storeIDStr: %s", err, storeIDStr)
			httputils.SetFailedWithCode(c, resp, errors.ErrInvalidRequest.Code, errors.ErrInvalidRequest)
			return
		}
	}

	stats, err := r.c.Stats().GetDashboardStats(c, storeID)
	if err != nil {
		log.Errorf("获取仪表板统计数据失败: %v, storeID: %v", err, storeID)
		// 根据错误类型返回不同的错误码
		if err == errors.ErrServerInternal {
			httputils.SetFailedWithCode(c, resp, errors.ErrServerInternal.Code, errors.ErrServerInternal)
		} else {
			httputils.SetFailedWithCode(c, resp, errors.ErrServerInternal.Code, errors.ErrServerInternal)
		}
		return
	}

	resp.Result = stats
	httputils.SetSuccess(c, resp)
}
