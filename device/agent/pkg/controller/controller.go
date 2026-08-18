package controller

import (
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/controller/mqtt"
	"github.com/casbin/casbin/v2"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/controller/audit"
	vcontroller "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/controller/voice"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/db"
	pluMq "github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/plugins/mqtt"
)

type SensecraftVoiceInterface interface {
	audit.AuditGetter
	mqtt.MqttGetter
	vcontroller.Getter
	UpdateRemoteConfig(baseURL string) error
}

type sensecraftVoice struct {
	cc       config.Config
	factory  db.ShareDaoFactory
	enforcer *casbin.SyncedEnforcer
	mqCli    pluMq.Client
}

func (p *sensecraftVoice) Mqtt() mqtt.Interface         { return mqtt.NewMq(p.cc, p.factory, p.mqCli) }
func (p *sensecraftVoice) Audit() audit.Interface       { return audit.NewAudit(p.cc, p.factory) }
func (p *sensecraftVoice) Voice() vcontroller.Interface { return vcontroller.New(p.cc) }

// UpdateRemoteConfig 更新远程配置
func (p *sensecraftVoice) UpdateRemoteConfig(baseURL string) error {
	return p.Voice().UpdateRemoteConfig(baseURL)
}

func New(cfg config.Config, f db.ShareDaoFactory, enforcer *casbin.SyncedEnforcer, mqCli pluMq.Client) SensecraftVoiceInterface {
	return &sensecraftVoice{
		cc:       cfg,
		factory:  f,
		enforcer: enforcer,
		mqCli:    mqCli,
	}
}
