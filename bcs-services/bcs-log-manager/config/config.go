package config

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
)

type Config struct {
	CollectionConfigs []CollectionConfig
	BcsApiConfig      bcsapi.Config
}

// CollectionConfig defines some customed infomation of log collection.
// For example, customed dataid of some Cluster.
// TODO: Customization of all kinds of log collections.
type CollectionConfig struct {
	StdDataId    string `json:"std_dataid"`
	NonStdDataId string `json:"non_std_dataid"`
	ClusterID    string `json:"cluster_id"`
}

func NewConfig() *Config {
	return &Config{}
}
