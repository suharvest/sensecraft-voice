package store

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/store"
)

func (r *storeRouter) create(c *gin.Context) {
	resp := httputils.NewResponse()

	var req store.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Store().Create(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *storeRouter) list(c *gin.Context) {
	resp := httputils.NewResponse()

	var req store.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Store().List(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *storeRouter) getById(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Store().GetById(c.Request.Context(), id)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *storeRouter) update(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req store.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Store().Update(c.Request.Context(), id, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *storeRouter) delete(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.Store().Delete(c.Request.Context(), id); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{"message": "删除成功"}
	httputils.SetSuccess(c, resp)
}
