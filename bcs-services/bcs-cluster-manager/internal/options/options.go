/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package options xxx
package options

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/encryptv2" // nolint

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// EtcdOption option for etcd
type EtcdOption struct {
	EtcdEndpoints string `json:"endpoints" value:"" usage:"endpoints of etcd"`
	EtcdCert      string `json:"cert" value:"" usage:"cert file of etcd"`
	EtcdKey       string `json:"key" value:"" usage:"key file for etcd"`
	EtcdCa        string `json:"ca" value:"" usage:"ca file for etcd"`
}

// LogConfig option for log
type LogConfig struct {
	LogDir          string `json:"dir"`
	LogMaxSize      uint64 `json:"maxsize"`
	LogMaxNum       int    `json:"maxnum"`
	ToStdErr        bool   `json:"tostderr"`
	AlsoToStdErr    bool   `json:"alsotostderr"`
	Verbosity       int32  `json:"v"`
	StdErrThreshold string `json:"stderrthreshold"`
	VModule         string `json:"vmodule"`
	TraceLocation   string `json:"backtraceat"`
}

// SwaggerConfig option for swagger
type SwaggerConfig struct {
	Dir string `json:"dir"`
}

// ServerConfig option for server
type ServerConfig struct {
	Address         string `json:"address"`
	Ipv6Address     string `json:"ipv6Address"`
	InsecureAddress string `json:"insecureaddress"`
	Port            uint   `json:"port"`
	HTTPPort        uint   `json:"httpport"`
	MetricPort      uint   `json:"metricport"`
	ServerCert      string `json:"servercert"`
	ServerKey       string `json:"serverkey"`
	ServerCa        string `json:"serverca"`
}

// ClientConfig option for bcs-cluster-manager as client
type ClientConfig struct {
	ClientCert string `json:"clientcert"`
	ClientKey  string `json:"clientkey"`
	ClientCa   string `json:"clientca"`
}

// TunnelConfig option for tunnel
type TunnelConfig struct {
	PeerToken string `json:"peertoken"`
}

// MongoConfig option for mongo
type MongoConfig struct {
	Address        string `json:"address"`
	ConnectTimeout uint   `json:"connecttimeout"`
	Database       string `json:"database"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	MaxPoolSize    uint   `json:"maxpoolsize"`
	MinPoolSize    uint   `json:"minpoolsize"`
}

// BrokerConfig option for dispatch task broker
type BrokerConfig struct {
	QueueAddress string `json:"address"`
	Exchange     string `json:"exchange"`
}

// BKOpsConfig for call bkops job
type BKOpsConfig struct {
	EsbServer   string `json:"esbServer"`
	Server      string `json:"server"`
	AppCode     string `json:"appCode"`
	AppSecret   string `json:"appSecret"`
	BkUserName  string `json:"bkUserName"`
	Debug       bool   `json:"debug"`
	TemplateURL string `json:"templateURL"`
	FrontURL    string `json:"frontURL"`
}

// CmdbConfig for cloud different tags info
type CmdbConfig struct {
	Enable     bool   `json:"enable"`
	AppCode    string `json:"appCode"`
	AppSecret  string `json:"appSecret"`
	BkUserName string `json:"bkUserName"`
	Server     string `json:"server"`
	Debug      bool   `json:"debug"`
}

// NodeManConfig for nodeman
type NodeManConfig struct {
	Enable     bool   `json:"enable"`
	AppCode    string `json:"appCode"`
	AppSecret  string `json:"appSecret"`
	BkUserName string `json:"bkUserName"`
	Server     string `json:"server"`
	Debug      bool   `json:"debug"`
}

// ResourceManagerConfig init resource_module
type ResourceManagerConfig struct {
	Enable bool   `json:"enable"`
	Module string `json:"module"`
}

// ProjectManagerConfig init project_module
type ProjectManagerConfig struct {
	Enable bool   `json:"enable"`
	Module string `json:"module"`
}

// CidrManagerConfig init cidr_module
type CidrManagerConfig struct {
	Enable bool   `json:"enable"`
	TLS    bool   `json:"tls"`
	Module string `json:"module"`
}

// AccessConfig for auth
type AccessConfig struct {
	Server string `json:"server"`
	Debug  bool   `json:"debug"`
}

// PassConfig pass-cc config
type PassConfig struct {
	Server    string `json:"server"`
	AppCode   string `json:"appCode"`
	AppSecret string `json:"appSecret"`
	Enable    bool   `json:"enable"`
	Debug     bool   `json:"debug"`
}

// UserConfig userManager config
type UserConfig struct {
	Enable      bool   `json:"enable"`
	GateWay     string `json:"gateWay"`
	IsVerifyTLS bool   `json:"isVerifyTLS"`
	Token       string `json:"token"`
}

// AlarmConfig for alarm interface
type AlarmConfig struct {
	Server        string `json:"server"`
	MonitorServer string `json:"monitorServer"`
	AppCode       string `json:"appCode"`
	AppSecret     string `json:"appSecret"`
	BkUserName    string `json:"bkUserName"`
	Enable        bool   `json:"enable"`
	Debug         bool   `json:"debug"`
}

// IAMConfig for perm interface
type IAMConfig struct {
	SystemID         string `json:"systemID"`
	AppCode          string `json:"appCode"`
	AppSecret        string `json:"appSecret"`
	External         bool   `json:"external"`
	GatewayServer    string `json:"gateWayServer"`
	IAMServer        string `json:"iamServer"`
	BkiIAMServer     string `json:"bkiIamServer"`
	Metric           bool   `json:"metric"`
	Debug            bool   `json:"debug"`
	ApplyPermAddress string `json:"applyPermAddress"`
}

// HelmConfig for helm
type HelmConfig struct {
	Enable bool `json:"enable"`
	// GateWay address
	GateWay string `json:"gateWay"`
	Token   string `json:"token"`
	Module  string `json:"module"`
}

// AutoScaler Config for autoscaler
type AutoScaler struct {
	ChartName        string `json:"chartName"`
	ReleaseName      string `json:"releaseName"`
	ReleaseNamespace string `json:"releaseNamespace"`
	IsPublicRepo     bool   `json:"isPublicRepo"`
	CaImageRegistry  string `json:"caImageRegistry"`
}

// BcsWatch config
type BcsWatch struct {
	ChartName        string `json:"chartName"`
	ReleaseName      string `json:"releaseName"`
	ReleaseNamespace string `json:"releaseNamespace"`
	IsPublicRepo     bool   `json:"isPublicRepo"`
	StorageServer    string `json:"storageServer"`
}

// VirtualCluster config
type VirtualCluster struct {
	HttpServer    string `json:"httpServer"`
	WsServer      string `json:"wsServer"`
	DebugWsServer string `json:"debugWsServer"`
	ChartName     string `json:"chartName"`
	ReleaseName   string `json:"releaseName"`
	IsPublicRepo  bool   `json:"isPublicRepo"`
}

// AddonData addon data
type AddonData struct {
	AddonName string `json:"addonName"`
}

// ComponentDeploy config
type ComponentDeploy struct {
	AutoScaler      AutoScaler     `json:"autoScaler"`
	Watch           BcsWatch       `json:"watch"`
	Vcluster        VirtualCluster `json:"vcluster"`
	ImagePullSecret AddonData      `json:"imagePullSecret"`
	Registry        string         `json:"registry"`
	BCSAPIGateway   string         `json:"bcsApiGateway"`
	Token           string         `json:"token"`
	BcsClusterUrl   string         `json:"bcsClusterUrl"`
}

// AuthConfig config for auth
type AuthConfig struct {
	Enable bool `json:"enable"`
	// jwt key
	PublicKeyFile  string `json:"publicKeyFile"`
	PrivateKeyFile string `json:"privateKeyFile"`
	// client 类型用户权限，使用 json 格式，key 为 client 名称，values 为拥有的权限，'*' 表示所有
	// 如：`{"admin": ["*"], "client_a": ["ClusterManager.CreateCluster"]}`
	ClientPermissions string `json:"clientPermissions"`
	// 不鉴权接口，使用逗号分隔，格式 `ClusterManager.Health,ClusterManager.Health`
	NoAuthMethod string `json:"noAuthMethod"`
}

// GseConfig for gse
type GseConfig struct {
	Enable        bool   `json:"enable"`
	AppCode       string `json:"appCode"`
	AppSecret     string `json:"appSecret"`
	BkUserName    string `json:"bkUserName"`
	EsbServer     string `json:"server"`
	GatewayServer string `json:"gatewayServer"`
	Debug         bool   `json:"debug"`
}

// JobConfig for job
type JobConfig struct {
	AppCode     string `json:"appCode"`
	AppSecret   string `json:"appSecret"`
	BkUserName  string `json:"bkUserName"`
	Server      string `json:"server"`
	Debug       bool   `json:"debug"`
	JobTaskLink string `json:"jobTaskLink"`
}

// DaemonConfig for daemon
type DaemonConfig struct {
	Enable bool `json:"enable"`
}

// ClusterManagerOptions options of cluster manager
type ClusterManagerOptions struct {
	Etcd               EtcdOption            `json:"etcd"`
	Swagger            SwaggerConfig         `json:"swagger"`
	Tunnel             TunnelConfig          `json:"tunnel"`
	BcsLog             LogConfig             `json:"bcslog"`
	Mongo              MongoConfig           `json:"mongo"`
	Broker             BrokerConfig          `json:"broker"`
	BKOps              BKOpsConfig           `json:"bkOps"`
	Cmdb               CmdbConfig            `json:"cmdb"`
	NodeMan            NodeManConfig         `json:"nodeman"`
	ResourceManager    ResourceManagerConfig `json:"resource"`
	ProjectManager     ProjectManagerConfig  `json:"project"`
	CidrManager        CidrManagerConfig     `json:"cidr"`
	CloudTemplatePath  string                `json:"cloudTemplatePath"`
	Access             AccessConfig          `json:"access"`
	Passcc             PassConfig            `json:"passcc"`
	UserManager        UserConfig            `json:"user"`
	Alarm              AlarmConfig           `json:"alarm"`
	IAM                IAMConfig             `json:"iam_config"`
	Auth               AuthConfig            `json:"auth"`
	Gse                GseConfig             `json:"gse"`
	Job                JobConfig             `json:"job"`
	Debug              bool                  `json:"debug"`
	Version            BCSEdition            `json:"version"`
	Helm               HelmConfig            `json:"helm"`
	ComponentDeploy    ComponentDeploy       `json:"componentDeploy"`
	ResourceSchemaPath string                `json:"resourceSchemaPath"`
	TagDepart          string                `json:"tagDepart"`
	PrefixVcluster     string                `json:"prefixVcluster"`
	TracingConfig      conf.TracingConfig    `json:"tracingConfig"`
	Encrypt            encryptv2.Config      `json:"encrypt"`
	Daemon             DaemonConfig          `json:"daemon"`
	ServerConfig
	ClientConfig
}

var globalClusterManagerOption *ClusterManagerOptions

// SetGlobalCMOptions set global CM options
func SetGlobalCMOptions(opts *ClusterManagerOptions) {
	if globalClusterManagerOption != nil {
		return
	}
	globalClusterManagerOption = opts
}

// GetGlobalCMOptions get global CM options
func GetGlobalCMOptions() *ClusterManagerOptions {
	return globalClusterManagerOption
}

// GetEditionInfo get edition
func GetEditionInfo() BCSEdition {
	if globalClusterManagerOption.Version == "" {
		globalClusterManagerOption.Version = InnerEdition
	}

	return globalClusterManagerOption.Version
}

// CloudTemplateList cloud template init config
type CloudTemplateList struct {
	CloudList []*cmproto.Cloud `json:"cloudList"`
}
