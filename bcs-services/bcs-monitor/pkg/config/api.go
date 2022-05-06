package config

import "time"

type APIConf struct {
	HTTP      *EndpointConfig `yaml:"http" mapstructure:"http"`
	GRPC      *EndpointConfig `yaml:"grpc" mapstructure:"grpc"`
	StoreList []string        `yaml:"store" mapstructure:"store"`
}

func (c *APIConf) Init() error {
	c.StoreList = []string{"127.0.0.1:10210"}

	c.HTTP = &EndpointConfig{
		Address:     "127.0.0.1:10214",
		GracePeriod: time.Minute * 2,
	}

	c.GRPC = &EndpointConfig{
		Address:     "127.0.0.1:10215",
		GracePeriod: time.Minute * 2,
	}

	return nil
}
