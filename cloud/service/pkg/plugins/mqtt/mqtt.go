package mqtt

import (
	"errors"
	"fmt"
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"net/url"
	"strings"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// Client is a protocol neutral MQTT client for connecting,
// disconnecting, and publishing data to a topic.
// The protocol specific clients must implement this interface
type Client interface {
	Connect() (bool, error)
	Publish(topic string, data []byte) error
	SubscribeMultiple(filters map[string]byte, callback paho.MessageHandler) error
	AddRoute(topic string, callback paho.MessageHandler)
	Close() error
}

func NewClient(cfg *config.MqttInfo) (Client, error) {
	if len(cfg.Servers) == 0 {
		return nil, errors.New("no servers specified")
	}

	if cfg.PersistentSession && cfg.ClientID == "" {
		return nil, errors.New("persistent_session requires client_id")
	}

	if cfg.QoS > 2 || cfg.QoS < 0 {
		return nil, fmt.Errorf("invalid QoS value %d; must be 0, 1 or 2", cfg.QoS)
	}

	switch cfg.Protocol {
	case "", "3.1.1":
		return NewMQTTv311Client(cfg)
	case "5":
		return NewMQTTv5Client(cfg)
	}
	return nil, fmt.Errorf("unsupported protocol %q: must be \"3.1.1\" or \"5\"", cfg.Protocol)
}

func parseServers(servers []string) ([]*url.URL, error) {
	urls := make([]*url.URL, 0, len(servers))
	for _, svr := range servers {
		// Preserve support for host:port style servers; deprecated in Telegraf 1.4.4
		if !strings.Contains(svr, "://") {
			urls = append(urls, &url.URL{Scheme: "tcp", Host: svr})
			continue
		}

		u, err := url.Parse(svr)
		if err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}
