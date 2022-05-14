/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package options

import (
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
	AppCode       string `json:"appCode"`
	AppSecret     string `json:"appSecret"`
	Debug         bool   `json:"debug"`
	External      bool   `json:"external"`
	CreateTaskURL string `json:"createTaskURL"`
	TaskStatusURL string `json:"taskStatusURL"`
	StartTaskURL  string `json:"startTaskURL"`
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

// ResourceManagerConfig init resource_module
type ResourceManagerConfig struct {
	Enable bool   `json:"enable"`
	Module string `json:"module"`
}

// SsmConfig for perm
type SsmConfig struct {
	Server    string `json:"server"`
	AppCode   string `json:"appCode"`
	AppSecret string `json:"appSecret"`
	Enable    bool   `json:"enable"`
	Debug     bool   `json:"debug"`
}

// PassConfig pass-cc config
type PassConfig struct {
	Server string `json:"server"`
	Enable bool   `json:"enable"`
	Debug  bool   `json:"debug"`
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
	Server     string `json:"server"`
	AppCode    string `json:"appCode"`
	AppSecret  string `json:"appSecret"`
	BkUserName string `json:"bkUserName"`
	Enable     bool   `json:"enable"`
	Debug      bool   `json:"debug"`
}

// IAMConfig for perm interface
type IAMConfig struct {
	SystemID      string `json:"systemID"`
	AppCode       string `json:"appCode"`
	AppSecret     string `json:"appSecret"`
	External      bool   `json:"external"`
	GatewayServer string `json:"gateWayServer"`
	IAMServer     string `json:"iamServer"`
	BkiIAMServer  string `json:"bkiIamServer"`
	Metric        bool   `json:"metric"`
	Debug         bool   `json:"debug"`
}

// ClusterManagerOptions options of cluster manager
type ClusterManagerOptions struct {
	Etcd              EtcdOption            `json:"etcd"`
	Swagger           SwaggerConfig         `json:"swagger"`
	Tunnel            TunnelConfig          `json:"tunnel"`
	BcsLog            LogConfig             `json:"bcslog"`
	Mongo             MongoConfig           `json:"mongo"`
	Broker            BrokerConfig          `json:"broker"`
	BKOps             BKOpsConfig           `json:"bkOps"`
	Cmdb              CmdbConfig            `json:"cmdb"`
	ResourceManager   ResourceManagerConfig `json:"resource"`
	CloudTemplatePath string                `json:"cloudTemplatePath"`
	Ssm               SsmConfig             `json:"ssm"`
	Passcc            PassConfig            `json:"passcc"`
	UserManager       UserConfig            `json:"user"`
	Alarm             AlarmConfig           `json:"alarm"`
	IAM               IAMConfig             `json:"iam_config"`
	Debug             bool                  `json:"debug"`
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

// CloudTemplateList cloud template init config
type CloudTemplateList struct {
	CloudList []*cmproto.Cloud `json:"cloudList"`
}
