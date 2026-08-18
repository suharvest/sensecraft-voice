package options

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/plugins/minio"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/plugins/mqtt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller"
	appdb "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/jobmanager"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/service"
	logutil "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/log"
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

	// 设备心跳间隔（秒），在线判定窗口为 3 倍该值
	defaultHeartbeatInterval = 60
	// ASR 服务器健康探测：每分钟一次，单次超时 5 s，连续 5 次失败才判 down
	// （OVS 推理同步阻塞 event loop，转写高峰期探测超时属正常现象）
	defaultASRProbeSchedule      = "*/1 * * * *"
	defaultASRProbeTimeout       = 5
	defaultASRProbeFailThreshold = 5

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

	JobManager  *jobmanager.Manager
	MqttClient  mqtt.Client
	MinIOClient minio.Client
}

func NewOptions() (*Options, error) {
	engine := gin.Default()
	// 设置最大内存限制为32MB，用于处理multipart表单
	engine.MaxMultipartMemory = 32 << 20 // 32 MB
	return &Options{
		HttpEngine: engine,
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
	if o.ComponentConfig.Default.DeviceHeartbeatIntervalSeconds <= 0 {
		o.ComponentConfig.Default.DeviceHeartbeatIntervalSeconds = defaultHeartbeatInterval
	}
	if len(o.ComponentConfig.Default.CryptoMasterKey) == 0 {
		// 未配置时退化为 JWT key，保证服务可启动；生产必须显式配置 crypto_master_key
		o.ComponentConfig.Default.CryptoMasterKey = o.ComponentConfig.Default.JWTKey
	}
	if o.ComponentConfig.ASR.Schedule == "" {
		o.ComponentConfig.ASR.Schedule = defaultASRProbeSchedule
	}
	if o.ComponentConfig.ASR.TimeoutSeconds <= 0 {
		o.ComponentConfig.ASR.TimeoutSeconds = defaultASRProbeTimeout
	}
	if o.ComponentConfig.ASR.FailThreshold <= 0 {
		o.ComponentConfig.ASR.FailThreshold = defaultASRProbeFailThreshold
	}

	if err := o.ComponentConfig.Valid(); err != nil {
		return err
	}

	o.ComponentConfig.Default.LogOptions.Init()

	// 注册依赖组件
	if err := o.register(); err != nil {
		return err
	}

	o.Controller = controller.New(o.ComponentConfig, o.Factory, o.Enforcer, o.MqttClient, o.MinIOClient)

	// 初始化关键词缓存并冷启动一次
	kc := service.NewKeywordCache(o.Factory)
	if err := kc.Refresh(context.Background()); err != nil {
		logutil.Warnf("keyword cache cold refresh failed: %v", err)
	}
	service.SetGlobalKeywordCache(kc)

	// 注册定时任务管理器（审计清理 + 关键词缓存刷新）
	schedule := o.ComponentConfig.Keywords.CacheSchedule
	if schedule == "" {
		schedule = "*/1 * * * *" // 默认每分钟
	}
	// ASR 服务器健康探测参数（主密钥单独注入，config 不直传给 job）
	asrProbeOptions := o.ComponentConfig.ASR
	asrProbeOptions.MasterKey = o.ComponentConfig.Default.CryptoMasterKey

	o.JobManager = jobmanager.NewManager(
		&o.ComponentConfig.Default.LogOptions,
		jobmanager.NewAuditsCleaner(o.ComponentConfig.Audit, o.Factory),
		jobmanager.NewKeywordsRefresher(schedule, kc),
		jobmanager.NewAsrServerProber(asrProbeOptions, o.Factory),
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

	// 注册数据库
	if err := o.registerDatabase(); err != nil {
		return err
	}

	// 注册Mqtt
	// if err := o.registerMqtt(); err != nil {
	// 	fmt.Println("register mqtt failed")
	// 	return err
	// }

	// 注册MinIO
	if err := o.registerMinIO(); err != nil {
		fmt.Println("register minio failed")
		return err
	}

	// TODO: 注册其他依赖
	if err := o.registerEnforcer(); err != nil {
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=False&loc=Local",
		sqlConfig.User,
		sqlConfig.Password,
		sqlConfig.Host,
		sqlConfig.Port,
		sqlConfig.Name)

	opt := &gorm.Config{
		Logger: appdb.NewLogger(logger.Info, defaultSlowSQLDuration),
		// 禁用GORM的自动时间戳处理
		DisableForeignKeyConstraintWhenMigrating: true,
		// 禁用自动创建/更新时间戳
		NowFunc: func() time.Time {
			return time.Now()
		},
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

// registerMinIO 注册MinIO客户端
func (o *Options) registerMinIO() error {
	minioConfig := o.ComponentConfig.OSS.MinIO
	client, err := minio.NewClient(&minioConfig)
	if err != nil {
		return fmt.Errorf("failed to create minio client: %v", err)
	}
	o.MinIOClient = client
	return nil
}

// Validate va1lidates all the required options.
func (o *Options) Validate() error {
	// TODO
	return nil
}
