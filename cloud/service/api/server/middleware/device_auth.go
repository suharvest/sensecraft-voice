package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db/model"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/token"
)

const (
	// HeaderEnrollmentKey 设备首次注册携带的共享密钥头
	HeaderEnrollmentKey = "X-Enrollment-Key"

	// DeviceContextKey gin context 里认证通过的设备对象
	DeviceContextKey = "auth_device"
	// EnrollmentContextKey gin context 里标记本次请求通过 enrollment key 认证
	EnrollmentContextKey = "auth_enrollment"
)

var (
	errDeviceTokenMissing = errors.New("缺少设备凭证：请携带 Authorization: Bearer <device_token>")
	errDeviceTokenInvalid = errors.New("设备凭证无效或已吊销")
	errEnrollmentInvalid  = errors.New("enrollment key 无效")
)

// bearerToken 提取 Authorization: Bearer <token>，WS 额外允许 ?token= 传参
func bearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth != "" {
		if len(auth) > 7 && strings.EqualFold(auth[:7], "bearer ") {
			return strings.TrimSpace(auth[7:])
		}
		return strings.TrimSpace(auth)
	}
	// WebSocket 握手无法自定义头的客户端可用 query 传递
	return strings.TrimSpace(c.Query("token"))
}

// GetAuthDevice 取出中间件写入的设备对象
func GetAuthDevice(c *gin.Context) (*model.Device, bool) {
	v, ok := c.Get(DeviceContextKey)
	if !ok {
		return nil, false
	}
	d, ok := v.(*model.Device)
	return d, ok
}

// IsEnrollmentRequest 本次请求是否通过 enrollment key 认证
func IsEnrollmentRequest(c *gin.Context) bool {
	v, ok := c.Get(EnrollmentContextKey)
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// DeviceAuth 设备侧接口鉴权：必须携带有效 device token。
// device_auth_enforce=false 时无凭证放行并记 warning（固件全量升级前的过渡开关）。
//
// 挂载方式是按单条路由，不是按 group —— 设备侧路由与 web 管理端路由混在同一 group。
func DeviceAuth(o *options.Options) gin.HandlerFunc {
	enforce := o.ComponentConfig.Default.DeviceAuthEnforce
	factory := o.Factory

	return func(c *gin.Context) {
		tk := bearerToken(c)
		if tk != "" {
			dev, err := factory.Device().GetByTokenHash(c.Request.Context(), token.HashDeviceToken(tk))
			if err == nil && dev != nil {
				c.Set(DeviceContextKey, dev)
				c.Next()
				return
			}
			if enforce {
				abortUnauthorized(c, errDeviceTokenInvalid)
				return
			}
			klog.Warningf("device auth: invalid token on %s %s (enforce=false, allowed)",
				c.Request.Method, c.Request.URL.Path)
			c.Next()
			return
		}

		if enforce {
			abortUnauthorized(c, errDeviceTokenMissing)
			return
		}
		klog.Warningf("device auth: no token on %s %s (enforce=false, allowed)",
			c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}

// DeviceRegisterAuth 注册/心跳接口鉴权：接受 device token（已注册设备）或 X-Enrollment-Key（首次注册）。
func DeviceRegisterAuth(o *options.Options) gin.HandlerFunc {
	enforce := o.ComponentConfig.Default.DeviceAuthEnforce
	enrollmentKey := o.ComponentConfig.Default.EnrollmentKey
	factory := o.Factory

	return func(c *gin.Context) {
		// 1. 已注册设备：token 优先
		if tk := bearerToken(c); tk != "" {
			dev, err := factory.Device().GetByTokenHash(c.Request.Context(), token.HashDeviceToken(tk))
			if err == nil && dev != nil {
				c.Set(DeviceContextKey, dev)
				c.Next()
				return
			}
			klog.Warningf("device register: token not recognized on %s", c.Request.URL.Path)
		}

		// 2. 首次注册：enrollment key
		provided := c.GetHeader(HeaderEnrollmentKey)
		if provided != "" && enrollmentKey != "" && token.SecureCompare(provided, enrollmentKey) {
			c.Set(EnrollmentContextKey, true)
			c.Next()
			return
		}

		if enforce {
			if provided != "" {
				abortUnauthorized(c, errEnrollmentInvalid)
				return
			}
			abortUnauthorized(c, errDeviceTokenMissing)
			return
		}

		klog.Warningf("device register: unauthenticated %s %s (enforce=false, allowed)",
			c.Request.Method, c.Request.URL.Path)
		// 过渡期：无凭证也允许签发首个 token
		c.Set(EnrollmentContextKey, true)
		c.Next()
	}
}
