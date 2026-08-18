package mqtt

import (
	"github.com/YOUR-ORG/sensecraft-voice/cloud/service/cmd/app/config"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test that default client has random ID
func TestRandomClientID(t *testing.T) {
	var err error

	cfg := &config.MqttInfo{
		Servers: []string{"tcp://localhost:1883"},
	}

	client1, err := NewMQTTv311Client(cfg)
	require.NoError(t, err)

	client2, err := NewMQTTv311Client(cfg)
	require.NoError(t, err)

	options1 := client1.client.OptionsReader()
	options2 := client2.client.OptionsReader()
	require.NotEqual(t, options1.ClientID(), options2.ClientID())
}
