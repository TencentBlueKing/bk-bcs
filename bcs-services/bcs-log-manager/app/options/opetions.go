package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	logmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
)

//LogManagerOption is option in flags
type LogManagerOption struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.MetricConfig
	conf.ZkConfig
	conf.CertConfig
	conf.LicenseServerConfig
	conf.LogConfig
	conf.ProcessConfig

	CollectionConfigs []logmanager.CollectionConfig `json:"collection_configs" usage:"Custom configs of log collections"`
	BcsApiHost        string                        `json:"bcs_api_host"`
	AuthToken         string                        `json:"api_auth_token"`
	Gateway           bool                          `json:"use_gateway" value:"true"`
}

func NewLogManagerOption() *LogManagerOption {
	return &LogManagerOption{}
}
