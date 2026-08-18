package mqtt

import (
	"crypto/tls"
	"fmt"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/cmd/app/config"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util"
	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util/errors"
	"time"

	mqttv3 "github.com/eclipse/paho.mqtt.golang" // Library that supports v3.1.1
)

type mqttv311Client struct {
	client  mqttv3.Client
	timeout time.Duration
	qos     int
	retain  bool
}

func NewMQTTv311Client(cfg *config.MqttInfo) (*mqttv311Client, error) {
	opts := mqttv3.NewClientOptions()
	opts.KeepAlive = cfg.KeepAlive
	opts.WriteTimeout = cfg.Timeout
	if cfg.ConnectionTimeout >= 1*time.Second {
		opts.ConnectTimeout = cfg.ConnectionTimeout
	}
	opts.SetCleanSession(!cfg.PersistentSession)
	if cfg.OnConnectionLost != nil {
		onConnectionLost := func(_ mqttv3.Client, err error) {
			cfg.OnConnectionLost(err)
		}
		opts.SetConnectionLostHandler(onConnectionLost)
	}
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(time.Second * 10)

	if cfg.ClientID != "" {
		opts.SetClientID(cfg.ClientID)
	} else {
		id, err := util.RandomString(5)
		if err != nil {
			return nil, fmt.Errorf("generating random client ID failed: %w", err)
		}
		opts.SetClientID("Telegraf-Output-" + id)
	}
	if cfg.TLSConfig.Enable {
		cer, err := tls.LoadX509KeyPair(cfg.TLSConfig.TLSCert, cfg.TLSConfig.TLSKey)
		if err != nil {
			return nil, err
		}
		// 设置 MQTT 连接选项的 TLS 配置
		opts.SetTLSConfig(&tls.Config{Certificates: []tls.Certificate{cer}})
	} else {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}

	servers, err := parseServers(cfg.Servers)
	if err != nil {
		return nil, err
	}
	for _, server := range servers {
		if cfg.TLSConfig.Enable {
			server.Scheme = "tls"
		}
		broker := server.String()
		opts.AddBroker(broker)
	}

	/*	if cfg.ClientTrace {
			log := &mqttLogger{logger.New("paho", "", "")}
			mqttv3.ERROR = log
			mqttv3.CRITICAL = log
			mqttv3.WARN = log
			mqttv3.DEBUG = log
		}
	*/
	return &mqttv311Client{
		client:  mqttv3.NewClient(opts),
		timeout: time.Duration(cfg.Timeout),
		qos:     cfg.QoS,
		retain:  cfg.Retain,
	}, nil
}

func (m *mqttv311Client) Connect() (bool, error) {
	token := m.client.Connect()

	if token.Wait() && token.Error() != nil {
		return false, token.Error()
	}

	// Persistent sessions should skip subscription if a session is present, as
	// the subscriptions are stored by the server.
	type sessionPresent interface {
		SessionPresent() bool
	}
	if t, ok := token.(sessionPresent); ok {
		return t.SessionPresent(), nil
	}

	return false, nil
}

func (m *mqttv311Client) Publish(topic string, body []byte) error {
	token := m.client.Publish(topic, byte(m.qos), m.retain, body)
	if !token.WaitTimeout(m.timeout) {
		return errors.ErrTimeout
	}
	return token.Error()
}

func (m *mqttv311Client) SubscribeMultiple(filters map[string]byte, callback mqttv3.MessageHandler) error {
	token := m.client.SubscribeMultiple(filters, callback)
	token.Wait()
	return token.Error()
}

func (m *mqttv311Client) AddRoute(topic string, callback mqttv3.MessageHandler) {
	m.client.AddRoute(topic, callback)
}

func (m *mqttv311Client) Close() error {
	if m.client.IsConnected() {
		m.client.Disconnect(100)
	}
	return nil
}
