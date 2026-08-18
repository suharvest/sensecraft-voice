package controller

import (
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/mqtt"
	"github.com/casbin/casbin/v2"

	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/asr_server"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/audio_recordings"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/audit"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/chat"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/device"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/file_upload"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/keywords"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/location"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/openai"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/recording"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/stats"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/store"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/controller/user"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/plugins/minio"
	pluMq "github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/plugins/mqtt"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/pkg/types"
)

type SensecraftVoiceInterface interface {
	audit.AuditGetter
	mqtt.MqttGetter
	device.DeviceGetter
	recording.RecordingGetter
	stats.StatsGetter
	store.StoreGetter
	location.LocationGetter
	user.UserGetter
	chat.ChatGetter
	openai.OpenAIGetter
	file_upload.FileUploadGetter
	audio_recordings.AudioRecordingGetter
	keywords.KeywordGetter
	asr_server.AsrServerGetter
}

type sensecraftVoice struct {
	cc          config.Config
	factory     db.ShareDaoFactory
	enforcer    *casbin.SyncedEnforcer
	mqCli       pluMq.Client
	minioClient minio.Client
}

func (p *sensecraftVoice) Mqtt() mqtt.Interface           { return mqtt.NewMq(p.cc, p.factory, p.mqCli) }
func (p *sensecraftVoice) Audit() audit.Interface         { return audit.NewAudit(p.cc, p.factory) }
func (p *sensecraftVoice) Device() device.Interface       { return device.NewDevice(p.cc, p.factory) }
func (p *sensecraftVoice) Recording() recording.Interface { return recording.NewRecording(p.cc, p.factory) }
func (p *sensecraftVoice) Stats() stats.Interface         { return stats.NewStats(p.cc, p.factory) }
func (p *sensecraftVoice) Store() store.Interface         { return store.NewStore(p.cc, p.factory) }
func (p *sensecraftVoice) Location() location.Interface   { return location.NewLocation(p.cc, p.factory) }
func (p *sensecraftVoice) User() user.Interface           { return user.NewUser(p.cc, p.factory) }
func (p *sensecraftVoice) Chat() chat.Interface {
	return chat.NewController(p.factory, getChatConfig(p.cc), getOpenAIConfig(p.cc))
}
func (p *sensecraftVoice) OpenAI() openai.Interface {
	return openai.NewController(p.factory, getOpenAIConfig(p.cc))
}
func (p *sensecraftVoice) FileUpload() file_upload.Interface {
	return file_upload.NewFileUploadController(p.cc, p.factory, p.minioClient)
}
func (p *sensecraftVoice) AudioRecording() audio_recordings.AudioRecordingInterface {
	return audio_recordings.NewAudioRecordingController(p.cc, p.factory, p.minioClient)
}
func (p *sensecraftVoice) AsrServer() asr_server.Interface {
	return asr_server.NewAsrServer(p.cc, p.factory)
}
func (p *sensecraftVoice) Keyword() keywords.Interface {
	return keywords.NewController(p.cc, p.factory)
}

func New(cfg config.Config, f db.ShareDaoFactory, enforcer *casbin.SyncedEnforcer, mqCli pluMq.Client, minioClient minio.Client) SensecraftVoiceInterface {
	return &sensecraftVoice{
		cc:          cfg,
		factory:     f,
		enforcer:    enforcer,
		mqCli:       mqCli,
		minioClient: minioClient,
	}
}

// getChatConfig 从配置中获取聊天配置
func getChatConfig(cfg config.Config) *types.ChatConfig {
	return &types.ChatConfig{
		BaseURL:     cfg.Chat.BaseURL,
		APIKey:      cfg.Chat.APIKey,
		Timeout:     cfg.Chat.Timeout,
		EnableDebug: cfg.Chat.EnableDebug,
	}
}

// getOpenAIConfig 从配置中获取OpenAI配置
func getOpenAIConfig(cfg config.Config) *types.OpenAIConfig {
	return &types.OpenAIConfig{
		APIKey:      cfg.OpenAI.APIKey,
		BaseURL:     cfg.OpenAI.BaseURL,
		Timeout:     cfg.OpenAI.Timeout,
		MaxTokens:   cfg.OpenAI.MaxTokens,
		Temperature: cfg.OpenAI.Temperature,
		Model:       cfg.OpenAI.Model,
	}
}
