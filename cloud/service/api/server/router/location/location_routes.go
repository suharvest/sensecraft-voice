package location

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/location"
)

func (r *locationRouter) create(c *gin.Context) {
	resp := httputils.NewResponse()

	var req location.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Location().Create(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *locationRouter) list(c *gin.Context) {
	resp := httputils.NewResponse()

	var req location.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Location().List(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *locationRouter) listByStoreId(c *gin.Context) {
	resp := httputils.NewResponse()

	storeIdStr := c.Param("id")
	storeId, err := strconv.ParseInt(storeIdStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req location.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Location().ListByStoreId(c.Request.Context(), storeId, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *locationRouter) getById(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Location().GetById(c.Request.Context(), id)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *locationRouter) update(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req location.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Location().Update(c.Request.Context(), id, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *locationRouter) delete(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.Location().Delete(c.Request.Context(), id); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{"message": "删除成功"}
	httputils.SetSuccess(c, resp)
}
