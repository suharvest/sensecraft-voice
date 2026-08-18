package keywords

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

func (r *keywordsRouter) create(c *gin.Context) {
	resp := httputils.NewResponse()

	var req types.CreateKeywordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Keyword().Create(c.Request.Context(), &req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *keywordsRouter) list(c *gin.Context) {
	resp := httputils.NewResponse()

	var req types.ListKeywordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Keyword().List(c.Request.Context(), &req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *keywordsRouter) getById(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Keyword().GetById(c.Request.Context(), id)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *keywordsRouter) update(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req types.UpdateKeywordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Keyword().Update(c.Request.Context(), id, &req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *keywordsRouter) delete(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.Keyword().Delete(c.Request.Context(), id); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{"message": "删除成功"}
	httputils.SetSuccess(c, resp)
}

func (r *keywordsRouter) batchDelete(c *gin.Context) {
	resp := httputils.NewResponse()

	var req types.BatchDeleteKeywordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.Keyword().BatchDelete(c.Request.Context(), &req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}
