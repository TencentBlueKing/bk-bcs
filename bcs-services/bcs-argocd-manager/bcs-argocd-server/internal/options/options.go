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

// ClientConfig option for bcs-argocd-server as client
type ClientConfig struct {
	ClientCert string `json:"clientcert"`
	ClientKey  string `json:"clientkey"`
	ClientCa   string `json:"clientca"`
}

// ArgocdServerOptions options of bcs argocd server
type ArgocdServerOptions struct {
	Etcd       EtcdOption    `json:"etcd"`
	BcsLog     LogConfig     `json:"bcslog"`
	Swagger    SwaggerConfig `json:"swagger"`
	MasterURL  string        `json:"masterurl"`
	KubeConfig string        `json:"kubeconfig"`
	Debug      bool          `json:"debug"`
	ServerConfig
	ClientConfig
}
