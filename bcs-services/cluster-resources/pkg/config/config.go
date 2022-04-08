/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"crypto/rsa"
	"io/ioutil"

	jwtGo "github.com/dgrijalva/jwt-go"
	"gopkg.in/yaml.v3"
)

// EtcdConf Etcd 相关配置
type EtcdConf struct {
	EtcdEndpoints string `yaml:"endpoints" usage:"Etcd Endpoints"`
	EtcdCert      string `yaml:"cert" usage:"Etcd Cert"`
	EtcdKey       string `yaml:"key" usage:"Etcd Key"`
	EtcdCa        string `yaml:"ca" usage:"Etcd CA"`
}

// ServerConf Server 配置
type ServerConf struct {
	Address          string `yaml:"address" usage:"服务启动地址"`
	InsecureAddress  string `yaml:"insecureAddress" usage:"服务启动地址（非安全）"`
	Port             int    `yaml:"port" usage:"GRPC 服务端口"`
	HTTPPort         int    `yaml:"httpPort" usage:"HTTP 服务端口"`
	MetricPort       int    `yaml:"metricPort" usage:"Metric 服务端口"`
	RegisterTTL      int    `yaml:"registerTTL" usage:"注册TTL"` //nolint:tagliatelle
	RegisterInterval int    `yaml:"registerInterval" usage:"注册间隔"`
	Cert             string `yaml:"cert" usage:"Server Cert"`
	CertPwd          string `yaml:"certPwd" usage:"Server Cert Password"`
	Key              string `yaml:"key" usage:"Server Key"`
	Ca               string `yaml:"ca" usage:"Server CA"`
}

// ClientConf Client 配置
type ClientConf struct {
	Cert    string `yaml:"cert" usage:"Client Cert"`
	CertPwd string `yaml:"certPwd" usage:"Client Cert Password"`
	Key     string `yaml:"key" usage:"Client Key"`
	Ca      string `yaml:"ca" usage:"Client CA"`
}

// SwaggerConf Swagger 配置
type SwaggerConf struct {
	Enabled bool   `yaml:"enabled" usage:"是否启用 swagger 服务"`
	Dir     string `yaml:"dir" usage:"swagger.json 存放目录"`
}

// LogConf 日志配置
type LogConf struct {
	Level         string `yaml:"level" usage:"日志级别"`
	FlushInterval int    `yaml:"flushInterval" usage:"刷新数据的间隔"`
	Path          string `yaml:"path" usage:"日志文件的绝对路径，如 /tmp/logs"`
	Name          string `yaml:"name" usage:"日志文件的名称，如 cr.log"`
	Size          int    `yaml:"size" usage:"文件的大小，单位 MB"`
	Age           int    `yaml:"age" usage:"日志的保存时间，单位天"`
	Backups       int    `yaml:"backups" usage:"历史文件保留数量"`
}

// RedisConf Redis 配置
type RedisConf struct {
	Address      string `yaml:"address" usage:"Redis Server Address"`
	DB           int    `yaml:"db" usage:"Redis DB"`
	Password     string `yaml:"password" usage:"Redis Password"`
	DialTimeout  int    `yaml:"dialTimeout" usage:"Redis Dial Timeout"`
	ReadTimeout  int    `yaml:"readTimeout" usage:"Redis Read Timeout(s)"`
	WriteTimeout int    `yaml:"writeTimeout" usage:"Redis Write Timeout(s)"`
	PoolSize     int    `yaml:"poolSize" usage:"Redis Pool Size"`
	MinIdleConns int    `yaml:"minIdleConns" usage:"Redis Min Idle Conns"`
	IdleTimeout  int    `yaml:"idleTimeout" usage:"Redis Idle Timeout(min)"`
}

// GlobalConf 全局配置，可在业务逻辑中使用
type GlobalConf struct {
	Auth          AuthConf          `yaml:"auth"`
	BCSAPIGW      BCSAPIGatewayConf `yaml:"bcsApiGW"` // nolint:tagliatelle
	SharedCluster SharedClusterConf `yaml:"sharedCluster"`
}

// AuthConf 认证相关配置
type AuthConf struct {
	Disabled     bool           `yaml:"disabled" usage:"是否禁用身份认证"`
	JWTPubKey    string         `yaml:"jwtPublicKey" usage:"jwt 公钥"`
	JWTPubKeyObj *rsa.PublicKey `yaml:"-" usage:"jwt 公钥对象（自动生成）"`
}

// BCSAPIGatewayConf 容器服务网关配置
type BCSAPIGatewayConf struct {
	Host      string `yaml:"host" usage:"容器服务网关 Host"`
	AuthToken string `yaml:"authToken" usage:"网关 Auth Token"`
}

// SharedClusterConf 共享集群相关配置
type SharedClusterConf struct {
	EnabledCObjKinds []string `yaml:"enabledCObjKinds" usage:"共享集群中支持的自定义对象 Kind"`
	EnabledCRDs      []string `yaml:"enabledCRDs" usage:"共享集群中支持的 CRD"` // nolint:tagliatelle
	ClusterIDs       []string `yaml:"clusterIDs" usage:"共享集群 ID 列表"`    // TODO 对接 ClusterMgr 后去除
}

// ClusterResourcesConf ClusterResources 服务启动配置
type ClusterResourcesConf struct {
	Debug   bool        `yaml:"debug"`
	Etcd    EtcdConf    `yaml:"etcd"`
	Server  ServerConf  `yaml:"server"`
	Client  ClientConf  `yaml:"client"`
	Swagger SwaggerConf `yaml:"swagger"`
	Log     LogConf     `yaml:"log"`
	Redis   RedisConf   `yaml:"redis"`
	Global  GlobalConf  `yaml:"crGlobal"`
}

// InitJWTPubKey ...
func (c *ClusterResourcesConf) InitJWTPubKey() (err error) {
	if c.Global.Auth.JWTPubKey == "" {
		return nil
	}
	c.Global.Auth.JWTPubKeyObj, err = jwtGo.ParseRSAPublicKeyFromPEM([]byte(c.Global.Auth.JWTPubKey))
	return err
}

// G 全局配置，可在业务逻辑中使用
var G = &GlobalConf{}

// LoadConf 加载配置信息
func LoadConf(filePath string) (*ClusterResourcesConf, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	conf := &ClusterResourcesConf{}
	if err = yaml.Unmarshal(yamlFile, conf); err != nil {
		return nil, err
	}
	// 初始化 jwt 配置
	if err = conf.InitJWTPubKey(); err != nil {
		return nil, err
	}
	// 初始化全局配置
	G = &conf.Global
	return conf, nil
}
