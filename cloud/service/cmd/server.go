package main

import (
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app"
)

var version string

func main() {
	klog.InitFlags(nil)
	rand.Seed(time.Now().UnixNano())

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	cmd := app.NewServerCommand(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
