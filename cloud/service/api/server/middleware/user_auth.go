package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/api/server/httputils"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/token"
)

const (
	// UserIdContextKey / UserNameContextKey 管理端登录用户信息
	UserIdContextKey   = "auth_user_id"
	UserNameContextKey = "auth_user_name"
)

var (
	errUserTokenMissing = errors.New("未登录：缺少 Authorization: Bearer <token>")
)

// abortUnauthorized 以真实 HTTP 401 终止请求。
// 不用 httputils.AbortFailedWithCode——它把 HTTP status 写成 200，
// 前端按 HTTP status 判断是否跳登录（voice-web/src/services/api.ts:63-65）。
func abortUnauthorized(c *gin.Context, err error) {
	resp := httputils.NewResponse()
	resp.SetMessageWithCode(err, http.StatusUnauthorized)
	c.AbortWithStatusJSON(http.StatusUnauthorized, resp)
}

// UserAuth 管理端（web）JWT 鉴权。
// 前端已在发 Bearer 头且 401 时跳登录，后端接上即可。
func UserAuth(o *options.Options) gin.HandlerFunc {
	jwtKey := []byte(o.ComponentConfig.Default.JWTKey)

	return func(c *gin.Context) {
		tk := bearerToken(c)
		if tk == "" {
			abortUnauthorized(c, errUserTokenMissing)
			return
		}
		claims, err := token.ParseToken(tk, jwtKey)
		if err != nil {
			abortUnauthorized(c, err)
			return
		}
		c.Set(UserIdContextKey, claims.Id)
		c.Set(UserNameContextKey, claims.Name)
		c.Next()
	}
}

// GetAuthUserId 取出登录用户 id
func GetAuthUserId(c *gin.Context) (int64, bool) {
	v, ok := c.Get(UserIdContextKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
