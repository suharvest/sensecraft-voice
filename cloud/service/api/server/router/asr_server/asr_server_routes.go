package asr_server

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/asr_server"
)

func (r *asrServerRouter) create(c *gin.Context) {
	resp := httputils.NewResponse()

	var req asr_server.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	out, err := r.c.AsrServer().Create(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}

func (r *asrServerRouter) update(c *gin.Context) {
	resp := httputils.NewResponse()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req asr_server.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	out, err := r.c.AsrServer().Update(c.Request.Context(), id, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}

func (r *asrServerRouter) delete(c *gin.Context) {
	resp := httputils.NewResponse()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.AsrServer().Delete(c.Request.Context(), id); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = gin.H{"message": "删除成功"}
	httputils.SetSuccess(c, resp)
}

func (r *asrServerRouter) getById(c *gin.Context) {
	resp := httputils.NewResponse()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	out, err := r.c.AsrServer().GetById(c.Request.Context(), id)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}

func (r *asrServerRouter) list(c *gin.Context) {
	resp := httputils.NewResponse()

	var req asr_server.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	out, err := r.c.AsrServer().List(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}

// probe 手工触发一次健康探测与能力回填
func (r *asrServerRouter) probe(c *gin.Context) {
	resp := httputils.NewResponse()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	out, err := r.c.AsrServer().Probe(c.Request.Context(), id)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}
	resp.Result = out
	httputils.SetSuccess(c, resp)
}
