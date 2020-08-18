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
	BcsAPIHost        string                        `json:"bcs_api_host" value:"" usage:"BcsApi Host"`
	AuthToken         string                        `json:"api_auth_token" value:"" usage:"BcsApi authentication token"`
	Gateway           bool                          `json:"use_gateway" value:"true" usage:"whether use api gateway"`
	KubeConfig        string                        `json:"kubeconfig" value:"" usage:"k8s config file path"`
	BCSSystemDataID   string                        `json:"bcs_system_dataid" value:"" usage:"DataID used to upload logs of bcs system modules with standard output"`
	K8SSystemDataID   string                        `json:"k8s_system_dataid" value:"" usage:"DataID used to upload logs of k8s system modules with standard output"`
}

func NewLogManagerOption() *LogManagerOption {
	return &LogManagerOption{}
}
