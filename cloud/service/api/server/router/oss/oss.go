package oss

import (
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/options"
)

// NewOSSRouter 创建OSS路由注册函数
func NewOSSRouter(o *options.Options) {
	// 创建OSS路由
	ossRouter := newOSSRouter(o.Controller)

	// 注册路由到API组
	api := o.HttpEngine.Group("/api/v1")
	ossRouter.RegisterRoutes(api)
}
