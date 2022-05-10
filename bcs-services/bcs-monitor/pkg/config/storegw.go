package config

import "time"

type StoreGWConf struct {
	HTTP    *EndpointConfig `yaml:"http" mapstructure:"http"`
	GRPC    *EndpointConfig `yaml:"grpc" mapstructure:"grpc"`
	DataDir string          `yaml:"data_dir" mapstructure:"data_dir"`
}

func (s *StoreGWConf) Init() error {
	s.DataDir = "./data/store"

	s.HTTP = &EndpointConfig{
		Address:     "127.0.0.1:10212",
		GracePeriod: time.Minute * 2,
	}

	s.GRPC = &EndpointConfig{
		Address:     "127.0.0.1:10213",
		GracePeriod: time.Minute * 2,
	}

	return nil
}
