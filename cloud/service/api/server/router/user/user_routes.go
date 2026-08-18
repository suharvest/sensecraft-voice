package user

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/user"
)

func (r *userRouter) register(c *gin.Context) {
	resp := httputils.NewResponse()

	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.User().Register(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *userRouter) login(c *gin.Context) {
	resp := httputils.NewResponse()

	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.User().Login(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *userRouter) list(c *gin.Context) {
	resp := httputils.NewResponse()

	var req user.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.User().List(c.Request.Context(), req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *userRouter) getById(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.User().GetById(c.Request.Context(), id)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *userRouter) update(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req user.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	result, err := r.c.User().Update(c.Request.Context(), id, req)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = result
	httputils.SetSuccess(c, resp)
}

func (r *userRouter) delete(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.User().Delete(c.Request.Context(), id); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{"message": "删除成功"}
	httputils.SetSuccess(c, resp)
}

func (r *userRouter) changePassword(c *gin.Context) {
	resp := httputils.NewResponse()

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	var req user.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	if err := r.c.User().ChangePassword(c.Request.Context(), id, req); err != nil {
		httputils.SetFailed(c, resp, err)
		return
	}

	resp.Result = gin.H{"message": "密码修改成功"}
	httputils.SetSuccess(c, resp)
}
