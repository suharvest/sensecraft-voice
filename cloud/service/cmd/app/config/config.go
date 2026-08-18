package config

import (
	"errors"
	"time"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/jobmanager"
	logutil "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/util/log"
)

type Mode string

const (
	DebugMode   Mode = "debug"
	ReleaseMode Mode = "release"
)

func (m Mode) InDebug() bool {
	return m == DebugMode
}

type Config struct {
	Default  DefaultOptions             `yaml:"default"`
	Mysql    MysqlOptions               `yaml:"mysql"`
	Worker   WorkerOptions              `yaml:"worker"`
	Audit    jobmanager.AuditOptions    `yaml:"audit"`
	Keywords KeywordsOptions            `yaml:"keywords"`
	MqttInfo MqttInfo                   `yaml:"mqtt"`
	Chat     ChatOptions                `yaml:"chat"`
	OpenAI   OpenAIOptions              `yaml:"openai"`
	Seeed    SeeedOptions               `yaml:"seeed"`
	OSS      OSSOptions                 `yaml:"oss"`
	ASR      jobmanager.AsrProbeOptions `yaml:"asr"`
}

// KeywordsOptions 关键词缓存配置
type KeywordsOptions struct {
	CacheSchedule string `yaml:"cache_schedule"` // cron 表达式，如 */1 * * * *
}
type MqttInfo struct {
	Servers             []string           `yaml:"servers"`
	Protocol            string             `yaml:"protocol"`
	Username            string             `yaml:"username"`
	Password            string             `yaml:"password"`
	Timeout             time.Duration      `yaml:"timeout"`
	ConnectionTimeout   time.Duration      `yaml:"connection_timeout"`
	QoS                 int                `yaml:"qos"`
	ClientID            string             `yaml:"client_id"`
	Retain              bool               `yaml:"retain"`
	KeepAlive           int64              `yaml:"keep_alive"`
	PersistentSession   bool               `yaml:"persistent_session"`
	PublishPropertiesV5 *PublishProperties `yaml:"v5"`
	ClientTrace         bool               `yaml:"client_trace"`
	TLSConfig           TLSConfig          `yaml:"tls_config"`
	AutoReconnect       bool               `yaml:"-"`
	OnConnectionLost    func(error)        `yaml:"-"`
}

type TLSConfig struct {
	Enable  bool   `yaml:"tls_enable"`
	TLSCert string `yaml:"tls_cert"`
	TLSKey  string `yaml:"tls_key"`
}

/*
	type ClientConfig struct {
		TLSCA               string   `yaml:"tls_ca"`
		TLSCert             string   `yaml:"tls_cert"`
		TLSKey              string   `yaml:"tls_key"`
		TLSKeyPwd           string   `yaml:"tls_key_pwd"`
		TLSMinVersion       string   `yaml:"tls_min_version"`
		TLSCipherSuites     []string `yaml:"tls_cipher_suites"`
		InsecureSkipVerify  bool     `yaml:"insecure_skip_verify"`
		ServerName          string   `yaml:"tls_server_name"`
		RenegotiationMethod string   `yaml:"tls_renegotiation_method"`
		Enable              *bool    `yaml:"tls_enable"`

		SSLCA   string `yaml:"ssl_ca" deprecated:"1.7.0;1.35.0;use 'tls_ca' instead"`
		SSLCert string `yaml:"ssl_cert" deprecated:"1.7.0;1.35.0;use 'tls_cert' instead"`
		SSLKey  string `yaml:"ssl_key" deprecated:"1.7.0;1.35.0;use 'tls_key' instead"`
	}
*/
type PublishProperties struct {
	ContentType    string            `yaml:"content_type"`
	ResponseTopic  string            `yaml:"response_topic"`
	MessageExpiry  time.Duration     `yaml:"message_expiry"`
	TopicAlias     *uint16           `yaml:"topic_alias"`
	UserProperties map[string]string `yaml:"user_properties"`
}

type DefaultOptions struct {
	Mode   Mode   `yaml:"mode"`
	Listen int    `yaml:"listen"`
	JWTKey string `yaml:"jwt_key"`

	// EnrollmentKey 设备首次注册使用的共享密钥（请求头 X-Enrollment-Key）
	EnrollmentKey string `yaml:"enrollment_key"`
	// DeviceAuthEnforce 为 true 时设备侧接口强制校验凭证；false 时无凭证放行并记 warning
	DeviceAuthEnforce bool `yaml:"device_auth_enforce"`
	// CryptoMasterKey 敏感字段落库加密的主密钥（AES-GCM）
	CryptoMasterKey string `yaml:"crypto_master_key"`
	// DeviceHeartbeatIntervalSeconds 设备心跳间隔（秒），在线判定窗口为 3 倍该值
	DeviceHeartbeatIntervalSeconds int `yaml:"device_heartbeat_interval_seconds"`

	// 自动创建指定模型的数据库表结构，不会更新已存在的数据库表
	AutoMigrate bool `yaml:"auto_migrate"`

	// HTTP服务器配置
	HTTP HTTPOptions `yaml:"http"`

	// 服务器基础URL（用于生成完整的API链接）
	BaseURL string `yaml:"base_url"`

	logutil.LogOptions `yaml:",inline"`
}

// OnlineWindowMs 在线判定窗口（毫秒）：3 倍心跳间隔
func (o DefaultOptions) OnlineWindowMs() int64 {
	interval := o.DeviceHeartbeatIntervalSeconds
	if interval <= 0 {
		interval = 60
	}
	return int64(interval) * 3 * 1000
}

func (o DefaultOptions) Valid() error {
	if err := o.LogOptions.Valid(); err != nil {
		return err
	}
	if err := o.HTTP.Valid(); err != nil {
		return err
	}
	return nil
}

// HTTPOptions HTTP服务器配置
type HTTPOptions struct {
	ReadTimeout       int `yaml:"read_timeout"`        // 读取超时时间(秒)
	WriteTimeout      int `yaml:"write_timeout"`       // 写入超时时间(秒)
	IdleTimeout       int `yaml:"idle_timeout"`        // 空闲超时时间(秒)
	ReadHeaderTimeout int `yaml:"read_header_timeout"` // 读取头部超时时间(秒)
}

func (h HTTPOptions) Valid() error {
	if h.ReadTimeout <= 0 {
		h.ReadTimeout = 60 // 默认60秒
	}
	if h.WriteTimeout <= 0 {
		h.WriteTimeout = 60 // 默认60秒
	}
	if h.IdleTimeout <= 0 {
		h.IdleTimeout = 120 // 默认120秒
	}
	if h.ReadHeaderTimeout <= 0 {
		h.ReadHeaderTimeout = 10 // 默认10秒
	}
	return nil
}

// MysqlOptions 数据库具体配置
type MysqlOptions struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
}

func (o MysqlOptions) Valid() error {
	// TODO
	return nil
}

type WorkerOptions struct {
	WorkDir string   `yaml:"work_dir"`
	Engines []Engine `yaml:"engines"`
}

type Engine struct {
	Image       string   `yaml:"image"`
	OSSupported []string `yaml:"os_supported"`
}

func (w WorkerOptions) Valid() error {
	// TODO
	return nil
}

// ChatOptions 聊天服务配置
type ChatOptions struct {
	BaseURL     string `yaml:"base_url"`
	APIKey      string `yaml:"api_key"`
	Timeout     int    `yaml:"timeout"`
	EnableDebug bool   `yaml:"enable_debug"`
}

func (c ChatOptions) Valid() error {
	// TODO: 可以添加配置验证逻辑
	return nil
}

// OpenAIOptions OpenAI API配置
type OpenAIOptions struct {
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Timeout     int     `yaml:"timeout"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
	Model       string  `yaml:"model"`
}

func (o OpenAIOptions) Valid() error {
	if o.APIKey == "" {
		return errors.New("openai api_key is required")
	}
	if o.BaseURL == "" {
		return errors.New("openai base_url is required")
	}
	if o.Timeout <= 0 {
		return errors.New("openai timeout must be greater than 0")
	}
	if o.MaxTokens <= 0 {
		return errors.New("openai max_tokens must be greater than 0")
	}
	if o.Temperature < 0 || o.Temperature > 2 {
		return errors.New("openai temperature must be between 0 and 2")
	}
	if o.Model == "" {
		return errors.New("openai model is required")
	}
	return nil
}

// SeeedOptions Seeed API配置
type SeeedOptions struct {
	BaseURL   string `yaml:"base_url"`
	SecretKey string `yaml:"secret_key"`
	Timeout   int    `yaml:"timeout"`
}

func (s SeeedOptions) Valid() error {
	// TODO: 可以添加配置验证逻辑
	return nil
}

// OSSOptions OSS文件存储配置
type OSSOptions struct {
	MinIO      MinIOOptions      `yaml:"minio"`
	Processing ProcessingOptions `yaml:"processing"`
}

func (o OSSOptions) Valid() error {
	if err := o.MinIO.Valid(); err != nil {
		return err
	}
	if err := o.Processing.Valid(); err != nil {
		return err
	}
	return nil
}

// MinIOOptions MinIO存储配置
type MinIOOptions struct {
	Endpoint   string `yaml:"endpoint"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	UseSSL     bool   `yaml:"use_ssl"`
	BucketName string `yaml:"bucket_name"`
	Region     string `yaml:"region"`
	Timeout    int    `yaml:"timeout"` // 秒
}

func (m MinIOOptions) Valid() error {
	if m.Endpoint == "" {
		return errors.New("minio endpoint is required")
	}
	if m.AccessKey == "" {
		return errors.New("minio access_key is required")
	}
	if m.SecretKey == "" {
		return errors.New("minio secret_key is required")
	}
	if m.BucketName == "" {
		return errors.New("minio bucket_name is required")
	}
	return nil
}

// ProcessingOptions 文件处理配置
type ProcessingOptions struct {
	MaxFileSize int64 `yaml:"max_file_size"` // 最大文件大小(字节)
}

func (p ProcessingOptions) Valid() error {
	if p.MaxFileSize <= 0 {
		return errors.New("max_file_size must be greater than 0")
	}
	return nil
}

func (c *Config) Valid() (err error) {
	if err = c.Default.Valid(); err != nil {
		return
	}
	if err = c.Mysql.Valid(); err != nil {
		return
	}
	if err = c.Worker.Valid(); err != nil {
		return
	}
	if err = c.Chat.Valid(); err != nil {
		return
	}
	if err = c.OpenAI.Valid(); err != nil {
		return
	}
	if err = c.Seeed.Valid(); err != nil {
		return
	}
	if err = c.OSS.Valid(); err != nil {
		return
	}

	return
}
