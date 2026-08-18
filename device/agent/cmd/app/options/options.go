package options

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/plugins/mqtt"
	"gopkg.in/yaml.v3"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/controller"
	appdb "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/jobmanager"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/plugins/voice"
	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
)

const (
	maxIdleConns = 10
	maxOpenConns = 100

	defaultListen     = 8080
	defaultTokenKey   = "sensecraftVoice"
	defaultConfigFile = "/etc/sensecraft-voice/config.yaml"
	defaultLogFormat  = logutil.LogFormatJson
	defaultWorkDir    = "/etc/sensecraftVoice"

	defaultSlowSQLDuration = 1 * time.Second

	rulesTableName = "rules"
)

// Options has all the params needed to run a sensecraftVoice
type Options struct {
	// The default values.
	ComponentConfig config.Config
	HttpEngine      *gin.Engine

	// 数据库接口
	db      *gorm.DB
	Factory appdb.ShareDaoFactory
	// 貔貅主控制接口
	Controller controller.SensecraftVoiceInterface

	// ConfigFile is the location of the sensecraftVoice server's configuration file.
	ConfigFile string

	// Authorization enforcement and policy management
	Enforcer *casbin.SyncedEnforcer

	JobManager *jobmanager.Manager
	MqttClient mqtt.Client

	// 配置管理相关
	configMutex sync.RWMutex // 保护配置的读写锁
}

func NewOptions() (*Options, error) {
	return &Options{
		HttpEngine: gin.Default(), // 初始化默认 api 路由
		ConfigFile: defaultConfigFile,
	}, nil
}

// Complete completes all the required options
func (o *Options) Complete() error {
	// 配置文件优先级: 默认配置，环境变量，命令行
	if len(o.ConfigFile) == 0 {
		// Try to read config file path from env.
		if cfgFile := os.Getenv("ConfigFile"); cfgFile != "" {
			o.ConfigFile = cfgFile
		} else {
			o.ConfigFile = defaultConfigFile
		}
	}

	c := config.New()
	c.SetConfigFile(o.ConfigFile)
	c.SetConfigType("yaml")
	if err := c.Binding(&o.ComponentConfig); err != nil {
		return err
	}

	// TODO: move to config initialization?
	if o.ComponentConfig.Default.Listen == 0 {
		o.ComponentConfig.Default.Listen = defaultListen
	}
	if len(o.ComponentConfig.Default.JWTKey) == 0 {
		o.ComponentConfig.Default.JWTKey = defaultTokenKey
	}
	if o.ComponentConfig.Default.LogFormat == "" {
		o.ComponentConfig.Default.LogFormat = defaultLogFormat
	}
	if o.ComponentConfig.Worker.WorkDir == "" {
		o.ComponentConfig.Worker.WorkDir = defaultWorkDir
	}
	if o.ComponentConfig.Audit.Schedule == "" {
		o.ComponentConfig.Audit.Schedule = jobmanager.DefaultSchedule
	}
	if o.ComponentConfig.Audit.DaysReserved == 0 {
		o.ComponentConfig.Audit.DaysReserved = jobmanager.DefaultDaysReserved
	}

	if err := o.ComponentConfig.Valid(); err != nil {
		return err
	}

	o.ComponentConfig.Default.LogOptions.Init()

	// 注册依赖组件
	if err := o.register(); err != nil {
		return err
	}

	o.Controller = controller.New(o.ComponentConfig, o.Factory, o.Enforcer, o.MqttClient)

	// 转换AuditOptions为jobmanager.AuditOptions
	auditOptions := jobmanager.AuditOptions{
		Schedule:     o.ComponentConfig.Audit.Schedule,
		DaysReserved: o.ComponentConfig.Audit.DaysReserved,
	}

	// 转换ASRCacheOptions为voice.ASRCacheConfig
	asrCacheConfig := voice.ASRCacheConfig{
		Enabled:         o.ComponentConfig.Voice.ASRCache.Enabled,
		CacheDir:        o.ComponentConfig.Voice.ASRCache.CacheDir,
		MaxRetries:      o.ComponentConfig.Voice.ASRCache.MaxRetries,
		RetryInterval:   o.ComponentConfig.Voice.ASRCache.RetryInterval,
		CacheExpiry:     o.ComponentConfig.Voice.ASRCache.CacheExpiry,
		MaxCacheSize:    o.ComponentConfig.Voice.ASRCache.MaxCacheSize,
		CleanupInterval: o.ComponentConfig.Voice.ASRCache.CleanupInterval,
		HTTPBatch: voice.HTTPBatchConfig{
			Enabled:          o.ComponentConfig.Voice.ASRCache.HTTPBatch.Enabled,
			BatchSize:        o.ComponentConfig.Voice.ASRCache.HTTPBatch.BatchSize,
			UploadInterval:   o.ComponentConfig.Voice.ASRCache.HTTPBatch.UploadInterval,
			MaxRetryAttempts: o.ComponentConfig.Voice.ASRCache.HTTPBatch.MaxRetryAttempts,
			Timeout:          o.ComponentConfig.Voice.ASRCache.HTTPBatch.Timeout,
		},
	}

	// 创建麦克风监控任务
	microphoneMonitorJob := jobmanager.NewMicrophoneMonitorJobWithAutoStart(
		o.ComponentConfig.Voice.ASRCache.Enabled,
		o.ComponentConfig.Voice.DeviceID,
		o.Controller.Voice(),
		o.ComponentConfig.Voice.AutoStart,
	)

	// 创建ASR重试任务，传入获取baseURL的函数
	asrRetryJob := jobmanager.NewASRRetryJobWithBaseURL(asrCacheConfig, func() string {
		return o.ComponentConfig.Remote.BaseURL
	})

	// 创建音频上传任务
	audioUploadConfig := jobmanager.AudioUploadConfig{
		Enabled:       o.ComponentConfig.AudioUpload.Enabled,
		ScanDir:       o.ComponentConfig.AudioUpload.ScanDir,
		MacAddress:    o.ComponentConfig.AudioUpload.MacAddress,
		Timeout:       o.ComponentConfig.AudioUpload.Timeout,
		MaxFileSize:   o.ComponentConfig.AudioUpload.MaxFileSize,
		MaxConcurrent: o.ComponentConfig.AudioUpload.MaxConcurrent,
	}
	audioUploadJob := jobmanager.NewAudioUploadJobWithBaseURL(audioUploadConfig, func() string {
		return o.ComponentConfig.Remote.BaseURL
	})

	registerDeviceJob := jobmanager.NewRegisterDeviceJobWithBaseURL(o.ComponentConfig.Remote.BaseURL, func() string {
		return o.ComponentConfig.Remote.BaseURL
	})

	o.JobManager = jobmanager.NewManager(
		&o.ComponentConfig.Default.LogOptions,
		jobmanager.NewAuditsCleaner(auditOptions, o.Factory),
		registerDeviceJob,
		asrRetryJob,
		microphoneMonitorJob,
		audioUploadJob,
	)
	return nil
}

// BindFlags binds the sensecraftVoice Configuration struct fields
func (o *Options) BindFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.ConfigFile, "configfile", defaultConfigFile, "The location of the sensecraftVoice configuration file")
}

func (o *Options) register() error {
	// 注册日志系统
	if err := o.registerLogger(); err != nil {
		return err
	}

	return nil
}

// registerLogger 注册日志系统
func (o *Options) registerLogger() error {
	// 确保日志目录存在
	if o.ComponentConfig.Default.LogFile.Enabled {
		logDir := filepath.Dir(o.ComponentConfig.Default.LogFile.Path)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("create log directory failed: %v", err)
		}
	}

	// 初始化日志配置
	o.ComponentConfig.Default.LogOptions.Init()
	return nil
}

// This panics if o.db is nil.
func (o *Options) registerEnforcer() error {
	// Casbin
	_, err := gormadapter.NewAdapterByDBUseTableName(o.db, "", rulesTableName)
	if err != nil {
		return err
	}

	return err
}

func (o *Options) registerDatabase() error {
	sqlConfig := o.ComponentConfig.Mysql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		sqlConfig.User,
		sqlConfig.Password,
		sqlConfig.Host,
		sqlConfig.Port,
		sqlConfig.Name)

	opt := &gorm.Config{
		Logger: appdb.NewLogger(logger.Info, defaultSlowSQLDuration),
	}
	db, err := gorm.Open(mysql.Open(dsn), opt)
	if err != nil {
		return err
	}
	o.db = db

	// 设置数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)

	o.Factory, err = appdb.NewDaoFactory(db, o.ComponentConfig.Default.AutoMigrate)
	return err
}

func (o *Options) registerMqtt() error {
	mqttInfo := o.ComponentConfig.MqttInfo
	client, err := mqtt.NewClient(&mqttInfo)
	if err != nil {
		return err
	}
	_, err = client.Connect()
	if err != nil {
		return err
	}
	o.MqttClient = client

	//defer client.Close()
	return err
}

// Validate va1lidates all the required options.
func (o *Options) Validate() error {
	// TODO
	return nil
}

// UpdateRemoteConfig 动态更新远程配置
func (o *Options) UpdateRemoteConfig(newBaseURL string) error {
	// 验证新的base URL
	if err := validateRemoteURL(newBaseURL); err != nil {
		return fmt.Errorf("invalid remote URL: %w", err)
	}

	// 检查是否与当前配置相同
	o.configMutex.RLock()
	currentBaseURL := o.ComponentConfig.Remote.BaseURL
	o.configMutex.RUnlock()

	if currentBaseURL == newBaseURL {
		logutil.Infof("remote config unchanged: %s", newBaseURL)
		return nil
	}

	// 更新内存中的配置
	o.configMutex.Lock()
	oldBaseURL := o.ComponentConfig.Remote.BaseURL
	o.ComponentConfig.Remote.BaseURL = newBaseURL
	o.configMutex.Unlock()

	// 持久化配置到文件
	if err := o.persistConfigToFile(); err != nil {
		// 如果持久化失败，回滚内存中的配置
		o.configMutex.Lock()
		o.ComponentConfig.Remote.BaseURL = oldBaseURL
		o.configMutex.Unlock()
		return fmt.Errorf("failed to persist config to file: %w", err)
	}

	// 通知相关组件配置已更新
	if o.Controller != nil {
		if err := o.Controller.UpdateRemoteConfig(newBaseURL); err != nil {
			// 如果更新失败，回滚配置
			o.configMutex.Lock()
			o.ComponentConfig.Remote.BaseURL = oldBaseURL
			o.configMutex.Unlock()
			// 尝试回滚文件配置
			o.persistConfigToFile()
			return fmt.Errorf("failed to update remote config: %w", err)
		}
	}

	// 更新JobManager中的ASR重试任务、音频上传任务和设备注册任务的baseURL获取函数
	if o.JobManager != nil {
		// 获取ASR重试任务并更新其baseURL获取函数
		if asrRetryJob, ok := o.JobManager.GetJob("asr-retry-job").(*jobmanager.ASRRetryJob); ok {
			asrRetryJob.UpdateBaseURLGetter(func() string {
				return newBaseURL
			})
		}

		// 获取音频上传任务并更新其baseURL获取函数
		if audioUploadJob, ok := o.JobManager.GetJob("audio-upload-job").(*jobmanager.AudioUploadJob); ok {
			audioUploadJob.UpdateBaseURLGetter(func() string {
				return newBaseURL
			})
		}

		// 获取设备注册任务并更新其baseURL获取函数
		if registerDeviceJob, ok := o.JobManager.GetJob("register-device").(*jobmanager.RegisterDeviceJob); ok {
			registerDeviceJob.UpdateBaseURLGetter(func() string {
				return newBaseURL
			})
		}
	}

	logutil.Infof("remote config updated from %s to %s (persisted to file)", oldBaseURL, newBaseURL)
	return nil
}

// validateRemoteURL 验证远程URL格式
func validateRemoteURL(baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("remote base URL cannot be empty")
	}

	// 解析URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// 检查协议
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported protocol: %s, only http and https are supported", u.Scheme)
	}

	// 检查主机名
	if u.Host == "" {
		return fmt.Errorf("missing host in URL")
	}

	// 检查主机名格式
	hostParts := strings.Split(u.Host, ":")
	if len(hostParts) > 2 {
		return fmt.Errorf("invalid host format: %s", u.Host)
	}

	// 检查端口（如果存在）
	if len(hostParts) == 2 {
		port := hostParts[1]
		if port == "" {
			return fmt.Errorf("empty port in URL")
		}
		// 这里可以添加端口范围检查
	}

	return nil
}

// GetRemoteConfig 获取当前远程配置
func (o *Options) GetRemoteConfig() string {
	o.configMutex.RLock()
	defer o.configMutex.RUnlock()
	return o.ComponentConfig.Remote.BaseURL
}

// GetConfig 获取当前配置的副本（线程安全）
func (o *Options) GetConfig() config.Config {
	o.configMutex.RLock()
	defer o.configMutex.RUnlock()
	return o.ComponentConfig
}

// persistConfigToFile 将当前配置持久化到配置文件
func (o *Options) persistConfigToFile() error {
	// 创建配置文件的备份
	backupFile := o.ConfigFile + ".backup"
	if err := copyFile(o.ConfigFile, backupFile); err != nil {
		logutil.Warnf("failed to create config backup: %v", err)
	}

	// 读取当前配置文件内容
	configData, err := os.ReadFile(o.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML文件
	var configMap map[string]interface{}
	if err := yaml.Unmarshal(configData, &configMap); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// 更新remote.base_url字段
	if remote, ok := configMap["remote"].(map[string]interface{}); ok {
		remote["base_url"] = o.ComponentConfig.Remote.BaseURL
	} else {
		// 如果remote字段不存在，创建它
		configMap["remote"] = map[string]interface{}{
			"base_url": o.ComponentConfig.Remote.BaseURL,
		}
	}

	// 将更新后的配置序列化回YAML
	updatedData, err := yaml.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// 写入配置文件
	if err := os.WriteFile(o.ConfigFile, updatedData, 0644); err != nil {
		// 如果写入失败，尝试恢复备份
		if restoreErr := copyFile(backupFile, o.ConfigFile); restoreErr != nil {
			logutil.Errorf("failed to restore config backup: %v", restoreErr)
		}
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// 删除备份文件
	os.Remove(backupFile)

	logutil.Infof("config file updated successfully: %s", o.ConfigFile)
	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}
