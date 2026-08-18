package jobmanager

import (
	"fmt"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/device"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/http"
	"github.com/sirupsen/logrus"
)

const (
	// 设备注册接口路径
	DeviceRegisterPath = "/api/v1/devices/register"
	// 同步间隔
	SyncInterval = "@every 1m"
)

// RegisterDeviceJob 设备注册任务
type RegisterDeviceJob struct {
	baseURL    string
	client     http.HTTPClient
	getBaseURL func() string
}

// NewRegisterDeviceJob 创建新的设备注册任务
func NewRegisterDeviceJob(baseURL string) *RegisterDeviceJob {
	client := http.NewClient().
		SetHeader("Content-Type", "application/json").
		SetTimeout(30 * time.Second).
		SetRetryCount(3)

	return &RegisterDeviceJob{
		baseURL:    baseURL,
		client:     client,
		getBaseURL: func() string { return baseURL }, // 默认使用传入的baseURL
	}
}

// NewRegisterDeviceJobWithBaseURL 创建带baseURL获取函数的设备注册任务
func NewRegisterDeviceJobWithBaseURL(baseURL string, getBaseURL func() string) *RegisterDeviceJob {
	client := http.NewClient().
		SetHeader("Content-Type", "application/json").
		SetTimeout(30 * time.Second).
		SetRetryCount(3)

	return &RegisterDeviceJob{
		baseURL:    baseURL,
		client:     client,
		getBaseURL: getBaseURL,
	}
}

// Name 返回任务名称
func (j *RegisterDeviceJob) Name() string {
	return "register-device"
}

// CronSpec 返回 cron 表达式
func (j *RegisterDeviceJob) CronSpec() string {
	return SyncInterval
}

// Do 执行任务
func (j *RegisterDeviceJob) Do(ctx *JobContext) error {
	// 获取设备信息管理器
	deviceManager := device.GetInstance()

	// 获取设备信息
	deviceInfo, err := deviceManager.GetDeviceInfo()
	if err != nil {
		logrus.Error(fmt.Errorf("收集设备信息失败: %w", err))
		return err
	}

	// 调用注册接口
	err = j.registerDevice(deviceInfo)
	if err != nil {
		logrus.Error(fmt.Errorf("设备注册失败: %w", err))
		return err
	}

	return nil
}

// registerDevice 调用设备注册接口
func (j *RegisterDeviceJob) registerDevice(deviceInfo *device.DeviceInfo) error {
	// 获取当前配置的baseURL
	baseURL := j.getBaseURL()
	if baseURL == "" {
		return fmt.Errorf("远程服务地址未配置")
	}

	url := baseURL + DeviceRegisterPath

	resp, err := j.client.Post(url, deviceInfo)
	if err != nil {
		return fmt.Errorf("HTTP 请求失败: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("设备注册失败，状态码: %d，响应: %s", resp.StatusCode, resp.GetBodyString())
	}

	return nil
}

// UpdateBaseURLGetter 更新baseURL获取函数
func (j *RegisterDeviceJob) UpdateBaseURLGetter(getBaseURL func() string) {
	j.getBaseURL = getBaseURL
}
