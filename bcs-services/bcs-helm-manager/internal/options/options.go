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

import "fmt"

// EtcdOption option for etcd
type EtcdOption struct {
	EtcdEndpoints string `json:"endpoints" yaml:"endpoints" value:"" usage:"endpoints of etcd"`
	EtcdCert      string `json:"cert" yaml:"cert" value:"" usage:"cert file of etcd"`
	EtcdKey       string `json:"key" yaml:"key" value:"" usage:"key file for etcd"`
	EtcdCa        string `json:"ca" yaml:"ca" value:"" usage:"ca file for etcd"`
}

// LogConfig option for log
type LogConfig struct {
	LogDir          string `json:"dir" yaml:"dir"`
	LogMaxSize      uint64 `json:"maxSize" yaml:"maxSize"`
	LogMaxNum       int    `json:"maxNum" yaml:"maxNum"`
	ToStdErr        bool   `json:"tostderr" yaml:"tostderr"`
	AlsoToStdErr    bool   `json:"alsotostderr" yaml:"alsotostderr"`
	Verbosity       int32  `json:"verbosity" yaml:"verbosity"`
	StdErrThreshold string `json:"stderrthreshold" yaml:"stderrthreshold"`
	VModule         string `json:"vmodule" yaml:"vmodule"`
	TraceLocation   string `json:"backtraceat" yaml:"backtraceat"`
}

// SwaggerConfig option for swagger
type SwaggerConfig struct {
	Dir string `json:"dir" yaml:"dir"`
}

// ServerConfig option for server
type ServerConfig struct {
	UseLocalIP      bool   `json:"uselocalip" yaml:"uselocalip"`
	Address         string `json:"address" yaml:"address"`
	IPv6Address     string `json:"ipv6Address" yaml:"ipv6Address"`
	InsecureAddress string `json:"insecureaddress" yaml:"insecureaddress"`
	Port            uint   `json:"port" yaml:"port"`
	HTTPPort        uint   `json:"httpport" yaml:"httpport"`
	MetricPort      uint   `json:"metricport" yaml:"metricport"`
}

// MongoConfig option for mongo
type MongoConfig struct {
	Address        string `json:"address" yaml:"address"`
	ConnectTimeout uint   `json:"connectTimeout" yaml:"connectTimeout"`
	AuthDatabase   string `json:"authDatabase" yaml:"authDatabase"`
	Database       string `json:"database" yaml:"database"`
	Username       string `json:"username" yaml:"username"`
	Password       string `json:"password" yaml:"password"`
	MaxPoolSize    uint   `json:"maxPoolSize" yaml:"maxPoolSize"`
	MinPoolSize    uint   `json:"minPoolSize" yaml:"minPoolSize"`
	Encrypted      bool   `json:"encrypted" yaml:"encrypted"`
}

// RepoConfig option for repo platform
type RepoConfig struct {
	// bkrepo api url
	URL string `json:"url" yaml:"url"`
	// repo base url
	BaseURL           string `json:"baseurl" yaml:"baseurl"`
	UID               string `json:"uid" yaml:"uid"`
	Username          string `json:"username" yaml:"username"`
	Password          string `json:"password" yaml:"password"`
	OciURL            string `json:"ociurl" yaml:"ociurl"`
	PublicRepoProject string `json:"publicRepoProject" yaml:"publicRepoProject"`
	PublicRepoName    string `json:"publicRepoName" yaml:"publicRepoName"`
	Encrypted         bool   `json:"encrypted" yaml:"encrypted"`
}

// GetRepoBaseURL get repo base url
func (r *RepoConfig) GetRepoBaseURL() string {
	if r.BaseURL != "" {
		return r.BaseURL
	}
	return fmt.Sprintf("%s/helm", r.URL)
}

// ReleaseConfig option for helm release handler
type ReleaseConfig struct {
	APIServer        string `json:"api" yaml:"api"`
	Token            string `json:"token" yaml:"token"`
	PatchDir         string `json:"patchdir" yaml:"patchdir"`
	AddonsConfigFile string `json:"addonsConfigFile" yaml:"addonsConfigFile"`
}

// JWTConfig option for jwt config
type JWTConfig struct {
	Enable         bool   `json:"enable" yaml:"enable"`
	PublicKey      string `json:"publickey" yaml:"publickey"`
	PublicKeyFile  string `json:"publickeyfile" yaml:"publickeyfile"`
	PrivateKey     string `json:"privatekey" yaml:"privatekey"`
	PrivateKeyFile string `json:"privatekeyfile" yaml:"privatekeyfile"`
}

// IAMConfig for perm interface
type IAMConfig struct {
	SystemID      string `json:"systemid" yaml:"systemid"`
	AppCode       string `json:"appcode" yaml:"appcode"`
	AppSecret     string `json:"appsecret" yaml:"appsecret"`
	External      bool   `json:"external" yaml:"external"`
	GatewayServer string `json:"gatewayserver" yaml:"gatewayserver"`
	IAMServer     string `json:"iamserver" yaml:"iamserver"`
	BkiIAMServer  string `json:"bkiiamserver" yaml:"bkiiamserver"`
	Metric        bool   `json:"metric" yaml:"metric"`
	Debug         bool   `json:"debug" yaml:"debug"`
}

// TLS option for tls
type TLS struct {
	ServerCert string `json:"serverCert" yaml:"serverCert"`
	ServerKey  string `json:"serverKey" yaml:"serverKey"`
	ServerCa   string `json:"serverCA" yaml:"serverCA"`
	ClientCert string `json:"clientCert" yaml:"clientCert"`
	ClientKey  string `json:"clientKey" yaml:"clientKey"`
	ClientCa   string `json:"clientCA" yaml:"clientCA"`
}

// Credential define client permissions config
type Credential struct {
	Name   string          `json:"name" yaml:"name"`
	Enable bool            `json:"enable" yaml:"enable"`
	Scopes CredentialScope `json:"scopes" yaml:"scopes"`
}

// CredentialScope define credentials scope
type CredentialScope struct {
	ProjectCode string `json:"projectCode" yaml:"projectCode"`
	ClusterID   string `json:"clusterID" yaml:"clusterID"`
	ProjectID   string `json:"projectID" yaml:"projectID"`
	Namespace   string `json:"namespace" yaml:"namespace"`
}

// Encrypt define encrypt config
type Encrypt struct {
	Enable    bool          `json:"enable" yaml:"enable"`
	Algorithm string        `json:"algorithm" yaml:"algorithm"`
	Secret    EncryptSecret `json:"secret" yaml:"secret"`
}

// EncryptSecret define encrypt secret
type EncryptSecret struct {
	Key    string `json:"key" yaml:"key"`
	Secret string `json:"secret" yaml:"secret"`
}

// HelmManagerOptions options of helm manager
type HelmManagerOptions struct {
	Etcd        EtcdOption    `json:"etcd" yaml:"etcd"`
	BcsLog      LogConfig     `json:"log" yaml:"log"`
	Swagger     SwaggerConfig `json:"swagger" yaml:"swagger"`
	Mongo       MongoConfig   `json:"mongo" yaml:"mongo"`
	Repo        RepoConfig    `json:"repo" yaml:"repo"`
	Release     ReleaseConfig `json:"release" yaml:"release"`
	IAM         IAMConfig     `json:"iam" yaml:"iam"`
	JWT         JWTConfig     `json:"jwt" yaml:"jwt"`
	Credentials []Credential  `json:"credentials" yaml:"credentials"`
	Encrypt     Encrypt       `json:"encrypt" yaml:"encrypt"`
	Debug       bool          `json:"debug" yaml:"debug"`
	TLS         TLS           `json:"tls" yaml:"tls"`
	ServerConfig
}

// GlobalOptions global helm manager options
var GlobalOptions *HelmManagerOptions
