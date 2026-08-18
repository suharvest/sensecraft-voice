package mqtt

import (
	"context"
	"fmt"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/db"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/plugins/mqtt"
)

type MqttGetter interface {
	Mqtt() Interface
}

type Interface interface {
	Push(ctx context.Context) error
}

type mqttClient struct {
	cc      config.Config
	factory db.ShareDaoFactory
	mqCli   mqtt.Client
}

func (a *mqttClient) Push(ctx context.Context) error {
	err := a.mqCli.Publish("/hello", []byte("Hello World"))
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func NewMq(cfg config.Config, f db.ShareDaoFactory, mqCli mqtt.Client) *mqttClient {
	return &mqttClient{
		cc:      cfg,
		factory: f,
		mqCli:   mqCli,
	}
}
