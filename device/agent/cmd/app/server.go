package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"

	"github.com/spf13/cobra"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/api/server/router"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/options"
)

func NewServerCommand(version string) *cobra.Command {
	opts, err := options.NewOptions()
	if err != nil {
		log.Fatalf("unable to initialize command options: %v", err)
	}

	cmd := &cobra.Command{
		Use:  "sensecraftVoice-server",
		Long: "The sensecraftVoice server controller is a daemon that embeds the core control loops.",
		Run: func(cmd *cobra.Command, args []string) {
			if err = opts.Complete(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			if err = opts.Validate(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			if err = Run(opts); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	// 绑定命令行参数
	opts.BindFlags(cmd)

	verCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Long:  "Print version and exit.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
	cmd.AddCommand(verCmd)
	return cmd
}

// Run 优雅启动貔貅服务
func Run(opt *options.Options) error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", opt.ComponentConfig.Default.Listen),
		Handler: opt.HttpEngine,
	}

	// 启动部署计划
	// TODO: 暂未设置优雅退出

	// 同步sensecraftVoice异常退出后的任务状态
	// 安装 http 路由
	router.InstallRouters(opt)

	// 根据配置自动开始录音（若开启）
	if opt.ComponentConfig.Voice.AutoStart {
		go func() {
			// 延迟启动，给WebSocket目标服务一些准备时间
			time.Sleep(3 * time.Second)
			if err := opt.Controller.Voice().StartByConfig(context.Background()); err != nil {
				log.Error("auto start voice recording failed: ", err)
				// 如果启动失败，尝试重试一次
				time.Sleep(5 * time.Second)
				if retryErr := opt.Controller.Voice().StartByConfig(context.Background()); retryErr != nil {
					log.Error("auto start voice recording retry failed: ", retryErr)
				} else {
					log.Info("auto start voice recording succeeded on retry")
				}
			} else {
				log.Info("auto start voice recording succeeded")
			}
		}()
	}

	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		log.Info("starting sensecraftVoice server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to listen sensecraftVoice server: ", err)
		}
	}()

	log.Warn("starting job manager")
	opt.JobManager.Run()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting sensecraftVoice server down ...")

	// The context is used to inform the server it has 5 seconds to finish the request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("sensecraftVoice server forced to shutdown: %v", err)
	}

	log.Info("shutting job manager down ...")
	opt.JobManager.Stop()

	return nil
}
