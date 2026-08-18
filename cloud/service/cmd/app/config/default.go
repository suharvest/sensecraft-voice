package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	yamlConfig = "yaml"
)

type ConfigInit struct {
	configFile string
	configType string

	data []byte
}

func New() *ConfigInit {
	return &ConfigInit{}
}

func (c *ConfigInit) SetConfigFile(configFile string) {
	c.configFile = configFile
}

func (c *ConfigInit) SetConfigType(in string) {
	c.configType = in
}

func (c *ConfigInit) readInConfig() error {
	var err error
	c.data, err = ioutil.ReadFile(c.configFile)
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigInit) Binding(out interface{}) error {
	if err := c.readInConfig(); err != nil {
		return err
	}
	switch c.configType {
	case yamlConfig:
		if err := yaml.Unmarshal(c.data, out); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported config type %s", c.configType)
	}

	return nil
}
