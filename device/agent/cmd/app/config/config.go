package config

import (
	"fmt"
	"time"

	logutil "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/log"
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
	Default     DefaultOptions     `yaml:"default"`
	Mysql       MysqlOptions       `yaml:"mysql"`
	Worker      WorkerOptions      `yaml:"worker"`
	Audit       AuditOptions       `yaml:"audit"`
	MqttInfo    MqttInfo           `yaml:"mqtt"`
	Voice       VoiceOptions       `yaml:"voice"`
	Remote      RemoteOptions      `yaml:"remote"`
	AudioUpload AudioUploadOptions `yaml:"audio_upload"`
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

	// 自动创建指定模型的数据库表结构，不会更新已存在的数据库表
	AutoMigrate bool `yaml:"auto_migrate"`

	logutil.LogOptions `yaml:",inline"`
}

func (o DefaultOptions) Valid() error {
	if err := o.LogOptions.Valid(); err != nil {
		return err
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

// AuditOptions 审计配置选项
type AuditOptions struct {
	Enabled      bool   `yaml:"enabled"`
	Schedule     string `yaml:"schedule"`
	DaysReserved int    `yaml:"days_reserved"`
}

type Engine struct {
	Image       string   `yaml:"image"`
	OSSupported []string `yaml:"os_supported"`
}

func (w WorkerOptions) Valid() error {
	// TODO
	return nil
}

type VoiceOptions struct {
	AutoStart           bool              `yaml:"autoStart"`
	DeviceID            string            `yaml:"deviceId"`
	SampleRate          int               `yaml:"sampleRate"`
	Channels            int               `yaml:"channels"`
	Format              string            `yaml:"format"` // pcm16 | opus | wav
	Output              string            `yaml:"output"` // file | stream | both
	FilePath            string            `yaml:"filePath"`
	SoftMuteOnInit      bool              `yaml:"softMuteOnStart"`
	OnDeviceLost        string            `yaml:"onDeviceLost"` // stop | switch-default
	SegmentSeconds      time.Duration     `yaml:"segmentSeconds"`
	WSUrl               string            `yaml:"wsUrl"`
	WSHeaders           map[string]string `yaml:"wsHeaders"`
	WSChunkBytes        int               `yaml:"wsChunkBytes"`
	WSMaxQueue          int               `yaml:"wsMaxQueue"`
	WSMaxReconnectDelay time.Duration     `yaml:"wsMaxReconnectDelay"`

	// ASR缓存配置
	ASRCache ASRCacheOptions `yaml:"asr_cache"`
}

// ASRCacheOptions ASR缓存配置选项
type ASRCacheOptions struct {
	Enabled         bool          `yaml:"enabled"`
	CacheDir        string        `yaml:"cache_dir"`
	MaxRetries      int           `yaml:"max_retries"`
	RetryInterval   time.Duration `yaml:"retry_interval"`
	CacheExpiry     time.Duration `yaml:"cache_expiry"`
	MaxCacheSize    int           `yaml:"max_cache_size"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`

	// HTTP批量上传配置
	HTTPBatch HTTPBatchOptions `yaml:"http_batch"`
}

// HTTPBatchOptions HTTP批量上传配置选项
type HTTPBatchOptions struct {
	Enabled          bool          `yaml:"enabled"`
	BatchSize        int           `yaml:"batch_size"`
	UploadInterval   time.Duration `yaml:"upload_interval"`
	MaxRetryAttempts int           `yaml:"max_retry_attempts"`
	Timeout          time.Duration `yaml:"timeout"`
}

func (o VoiceOptions) Valid() error {
	// 基础校验，允许为空使用默认值
	if o.SampleRate != 0 && o.SampleRate != 8000 && o.SampleRate != 16000 && o.SampleRate != 24000 && o.SampleRate != 44100 && o.SampleRate != 48000 {
		return fmt.Errorf("unsupported sampleRate: %d", o.SampleRate)
	}
	if o.Channels != 0 && o.Channels != 1 && o.Channels != 2 {
		return fmt.Errorf("unsupported channels: %d", o.Channels)
	}
	if o.Format != "" && o.Format != "pcm16" && o.Format != "opus" && o.Format != "wav" {
		return fmt.Errorf("unsupported format: %s", o.Format)
	}
	if o.Output != "" && o.Output != "file" && o.Output != "stream" && o.Output != "both" {
		return fmt.Errorf("unsupported output: %s", o.Output)
	}

	// 验证ASR缓存配置
	if err := o.ASRCache.Valid(); err != nil {
		return fmt.Errorf("invalid asr_cache config: %w", err)
	}

	return nil
}

// Valid 验证ASR缓存配置
func (o ASRCacheOptions) Valid() error {
	if !o.Enabled {
		return nil // 如果未启用，跳过验证
	}

	if o.CacheDir == "" {
		return fmt.Errorf("cache_dir is required when asr_cache is enabled")
	}

	if o.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative")
	}

	if o.RetryInterval <= 0 {
		return fmt.Errorf("retry_interval must be positive")
	}

	if o.CacheExpiry <= 0 {
		return fmt.Errorf("cache_expiry must be positive")
	}

	if o.MaxCacheSize < 0 {
		return fmt.Errorf("max_cache_size must be non-negative")
	}

	if o.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup_interval must be positive")
	}

	// 验证HTTP批量上传配置
	if err := o.HTTPBatch.Valid(); err != nil {
		return fmt.Errorf("invalid http_batch config: %w", err)
	}

	return nil
}

// Valid 验证HTTP批量上传配置
func (o HTTPBatchOptions) Valid() error {
	if !o.Enabled {
		return nil // 如果未启用，跳过验证
	}

	if o.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be positive")
	}

	if o.UploadInterval <= 0 {
		return fmt.Errorf("upload_interval must be positive")
	}

	if o.MaxRetryAttempts < 0 {
		return fmt.Errorf("max_retry_attempts must be non-negative")
	}

	if o.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// ToVoiceASRCacheConfig 转换为voice包的ASRCacheConfig
func (o ASRCacheOptions) ToVoiceASRCacheConfig() interface{} {
	// 这里需要导入voice包，但为了避免循环导入，我们返回interface{}
	// 实际使用时需要类型断言
	return map[string]interface{}{
		"enabled":          o.Enabled,
		"cache_dir":        o.CacheDir,
		"max_retries":      o.MaxRetries,
		"retry_interval":   o.RetryInterval,
		"cache_expiry":     o.CacheExpiry,
		"max_cache_size":   o.MaxCacheSize,
		"cleanup_interval": o.CleanupInterval,
		"http_batch": map[string]interface{}{
			"enabled":            o.HTTPBatch.Enabled,
			"batch_size":         o.HTTPBatch.BatchSize,
			"upload_interval":    o.HTTPBatch.UploadInterval,
			"max_retry_attempts": o.HTTPBatch.MaxRetryAttempts,
			"timeout":            o.HTTPBatch.Timeout,
		},
	}
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
	if err = c.Voice.Valid(); err != nil {
		return
	}
	if err = c.Remote.Valid(); err != nil {
		return
	}
	if err = c.AudioUpload.Valid(); err != nil {
		return
	}

	return
}

// RemoteOptions 远程服务配置
type RemoteOptions struct {
	BaseURL     string                   `yaml:"base_url"`
	AudioStream RemoteAudioStreamOptions `yaml:"audio_stream"`
}

// RemoteAudioStreamOptions 远程音频流配置
type RemoteAudioStreamOptions struct {
	Enabled           bool              `yaml:"enabled"`
	MacAddress        string            `yaml:"mac_address"`
	Headers           map[string]string `yaml:"headers"`
	ChunkBytes        int               `yaml:"chunk_bytes"`
	MaxQueue          int               `yaml:"max_queue"`
	MaxReconnectDelay time.Duration     `yaml:"max_reconnect_delay"`
}

func (o RemoteOptions) Valid() error {
	// TODO: 可以添加 URL 格式验证
	return nil
}

// AudioUploadOptions 音频文件上传配置选项
type AudioUploadOptions struct {
	Enabled       bool   `yaml:"enabled"`
	ScanDir       string `yaml:"scan_dir"`
	MacAddress    string `yaml:"mac_address"`
	Timeout       string `yaml:"timeout"`
	MaxFileSize   int64  `yaml:"max_file_size"`
	MaxConcurrent int    `yaml:"max_concurrent"`
}

func (o AudioUploadOptions) Valid() error {
	if !o.Enabled {
		return nil // 如果未启用，跳过验证
	}

	if o.ScanDir == "" {
		return fmt.Errorf("scan_dir is required when audio_upload is enabled")
	}

	// MAC地址可以为空，系统会自动获取

	if o.Timeout == "" {
		o.Timeout = "30s" // 设置默认值
	}

	if o.MaxFileSize <= 0 {
		o.MaxFileSize = 50 // 默认50MB
	}

	if o.MaxConcurrent <= 0 {
		o.MaxConcurrent = 3 // 默认3个并发
	}

	return nil
}
