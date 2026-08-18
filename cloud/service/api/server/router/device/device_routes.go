package device

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/middleware"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/device"
)

var errDeviceUnidentified = errors.New("无法识别设备：请携带 device token 或用 MAC 地址访问")

func (r *deviceRouter) register(c *gin.Context) {
	resp := httputils.NewResponse()

	var req device.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	auth := device.RegisterAuth{Enrollment: middleware.IsEnrollmentRequest(c)}
	if dev, ok := middleware.GetAuthDevice(c); ok {
		auth.Device = dev
	}

	out, err := r.c.Device().Register(c, req, auth)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}

func (r *deviceRouter) assignToLocation(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	deviceId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req struct {
		LocationId int64 `json:"location_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.Device().AssignToLocation(c.Request.Context(), deviceId, req.LocationId); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{"message": "分配成功"}
	httputils.SetSuccess(c, resp)
}

func (r *deviceRouter) listByLocation(c *gin.Context) {
	resp := httputils.NewResponse()

	locationIdStr := c.Param("id")
	locationId, err := strconv.ParseInt(locationIdStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req device.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Device().ListByLocation(c.Request.Context(), locationId, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *deviceRouter) listByStore(c *gin.Context) {
	resp := httputils.NewResponse()

	storeIdStr := c.Param("id")
	storeId, err := strconv.ParseInt(storeIdStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req device.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Device().ListByStore(c.Request.Context(), storeId, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *deviceRouter) list(c *gin.Context) {
	resp := httputils.NewResponse()

	var req device.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	// 检查是否有location_id参数
	locationIdStr := c.Query("location_id")
	if locationIdStr != "" {
		locationId, err := strconv.ParseInt(locationIdStr, 10, 64)
		if err != nil {
			httputils.SetFailed(c, resp, err)
			return
		}
		// 调用按位置查询的方法
		result, err := r.c.Device().ListByLocation(c.Request.Context(), locationId, req)
		if err != nil {
			httputils.SetFailed(c, resp, err)
			return
		}
		resp.Result = result
	} else if storeIdStr := c.Query("store_id"); storeIdStr != "" {
		// 检查是否有store_id参数
		storeId, err := strconv.ParseInt(storeIdStr, 10, 64)
		if err != nil {
			httputils.SetFailed(c, resp, err)
			return
		}
		// 调用按门店查询的方法
		result, err := r.c.Device().ListByStore(c.Request.Context(), storeId, req)
		if err != nil {
			httputils.SetFailed(c, resp, err)
			return
		}
		resp.Result = result
	} else {
		// 原有的查询所有设备的逻辑
		result, err := r.c.Device().List(c, req)
		if err != nil {
			httputils.SetFailed(c, resp, err)
			return
		}
		resp.Result = result
	}

	httputils.SetSuccess(c, resp)
}

// getByMacOrId 兼容两种寻址：纯数字视为设备 id（web 在用 GET /devices/{id}），否则按 MAC 查询
func (r *deviceRouter) getByMacOrId(c *gin.Context) {
	resp := httputils.NewResponse()
	key := strings.ToLower(c.Param("mac"))

	if id, err := strconv.ParseInt(key, 10, 64); err == nil && id > 0 {
		res, err := r.c.Device().GetById(c, id)
		if err != nil {
			httputils.SetFailed(c, resp, err)
			return
		}
		resp.Result = res
		httputils.SetSuccess(c, resp)
		return
	}

	res, err := r.c.Device().GetByMac(c, key)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = res
	httputils.SetSuccess(c, resp)
}

func (r *deviceRouter) updateName(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	deviceId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	device, err := r.c.Device().UpdateName(c.Request.Context(), deviceId, req.Name)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = device
	httputils.SetSuccess(c, resp)
}

// update 管理端更新设备（名称 / 点位 / 门店）
func (r *deviceRouter) update(c *gin.Context) {
	resp := httputils.NewResponse()

	deviceId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req device.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	out, err := r.c.Device().Update(c.Request.Context(), deviceId, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}

// assignAsrServer 分配设备到 ASR 服务器，变更时 asr_config_version++
func (r *deviceRouter) assignAsrServer(c *gin.Context) {
	resp := httputils.NewResponse()

	deviceId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req struct {
		AsrServerId int64 `json:"asr_server_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	out, err := r.c.Device().AssignAsrServer(c.Request.Context(), deviceId, req.AsrServerId)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}

// getAsrConfig 设备用自身 token 拉取 ASR 配置（GET /api/v1/devices/me/asr-config）
func (r *deviceRouter) getAsrConfig(c *gin.Context) {
	resp := httputils.NewResponse()

	dev, ok := middleware.GetAuthDevice(c)
	if !ok {
		// device_auth_enforce=false 的过渡期：允许用 mac 定位设备
		key := strings.ToLower(c.Param("mac"))
		if key == "" || key == "me" {
			httputils.SetFailedWithCode(c, resp, 401, errDeviceUnidentified)
			return
		}
		found, err := r.c.Device().GetByMac(c.Request.Context(), key)
		if err != nil {
			httputils.SetFailed(c, resp, err)
			return
		}
		dev = found
	}

	out, err := r.c.Device().GetAsrConfig(c.Request.Context(), dev)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}
