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
	SystemDataID      string                        `json:"system_dataid" value:"" usage:"DataID used to upload logs of k8s and bcs system modules with standard output"`
	BkUsername        string                        `json:"bk_username" value:"" usage:"User to request bkdata api"`
	BkAppCode         string                        `json:"bk_appcode" value:"" usage:"BK app code"`
	BkAppSecret       string                        `json:"bk_appsecret" value:"" usage:"BK app secret"`
	BkBizID           int                           `json:"bk_bizid" value:"-1" usage:"BK business id"`
}

func NewLogManagerOption() *LogManagerOption {
	return &LogManagerOption{}
}
