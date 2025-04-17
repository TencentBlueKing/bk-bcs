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

// Package config xxx
package config

import (
	"crypto/rsa"
	"net"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	bkiam "github.com/TencentBlueKing/iam-go-sdk"
	"github.com/TencentBlueKing/iam-go-sdk/logger"
	"github.com/TencentBlueKing/iam-go-sdk/metric"
	jwtGo "github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	constant "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// G 全局配置，可在业务逻辑中使用
var G = &GlobalConf{}

// LoadConf 加载配置信息
func LoadConf(filePath string) (*ClusterResourcesConf, error) {
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	conf := &ClusterResourcesConf{}
	if err = yaml.Unmarshal(yamlFile, conf); err != nil {
		return nil, err
	}

	//  变量 fallback to env
	if conf.Global.Basic.AppCode == "" {
		conf.Global.Basic.AppCode = envs.BKAppCode
	}
	if conf.Global.Basic.AppSecret == "" {
		conf.Global.Basic.AppSecret = envs.BKAppSecret
	}
	if conf.Global.Basic.BKAPIGWHost == "" {
		conf.Global.Basic.BKAPIGWHost = envs.BKIAMGatewayHost
	}
	if conf.Global.Basic.BKPaaSHost == "" {
		conf.Global.Basic.BKPaaSHost = envs.BKPaaSHost
	}

	if conf.Global.IAM.SystemID == "" {
		conf.Global.IAM.SystemID = envs.BKIAMSystemID
	}
	if conf.Global.IAM.Host == "" {
		conf.Global.IAM.Host = envs.BKIAMHost
	}

	if conf.Global.SharedCluster.AnnotationKeyProjectCode == "" {
		conf.Global.SharedCluster.AnnotationKeyProjectCode = constant.ProjectCodeAnnoKey
	}

	if conf.Redis.Password == "" {
		conf.Redis.Password = envs.RedisPassword
	}

	// mongo env
	if conf.Mongo.Address == "" {
		conf.Mongo.Address = envs.MongoAddress
	}
	if conf.Mongo.Replicaset == "" {
		conf.Mongo.Replicaset = envs.MongoReplicaset
	}
	if conf.Mongo.Username == "" {
		conf.Mongo.Username = envs.MongoUsername
	}
	if conf.Mongo.Password == "" {
		conf.Mongo.Password = envs.MongoPassword
	}

	for _, f := range []func() error{
		// 初始化 Server.Address
		conf.initServerAddress,
		// 初始化 AuthToken
		conf.initAuthToken,
		// 初始化 jwt 公钥
		conf.initJWTPubKey,
		// 初始化 iam
		conf.initIAM,
		conf.initCompoment,
	} {
		if initErr := f(); initErr != nil {
			return nil, initErr
		}
	}
	// 初始化全局配置
	G = &conf.Global
	return conf, nil
}

// getIPv6AddrFromEnv 解析ipv6
func getIPv6AddrFromEnv() string {
	ipv6 := util.GetIPv6Address(envs.PodIPs)
	if ipv6 == "" {
		return ""
	}

	// 在实际中，ipv6不能是回环地址
	if v := net.ParseIP(ipv6); v == nil || v.IsLoopback() || v.IsUnspecified() {
		return ""
	}

	// local link ipv6 需要带上 interface， 格式如::%eth0
	if envs.IPv6Interface != "" {
		ipv6 = ipv6 + "%" + envs.IPv6Interface
	}

	return ipv6
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
	Mongo   MongoConfig `yaml:"mongo"`
	Global  GlobalConf  `yaml:"crGlobal"`
	Tracing TracingConf `yaml:"tracing"`
}

// TracingConf 链路追踪配置
type TracingConf struct {
	Enabled       bool              `yaml:"enabled" usage:"enable trace"`
	Endpoint      string            `yaml:"endpoint" usage:"Collector service endpoint"`
	Token         string            `yaml:"token" usage:"token for collector service"`
	ResourceAttrs map[string]string `yaml:"resourceAttrs" usage:"attributes of traced service"`
}

func (c *ClusterResourcesConf) initServerAddress() error {
	// 若指定使用 LOCAL_IP 且环境变量中 LOCAL_IP 有值，则替换掉 Server.Address
	if c.Server.UseLocalIP && envs.LocalIP != "" {
		c.Server.Address = envs.LocalIP
		c.Server.InsecureAddress = envs.LocalIP
	}
	c.Server.AddressIPv6 = getIPv6AddrFromEnv()
	return nil
}

// initAuthToken 初始化 BCS AuthToken
func (c *ClusterResourcesConf) initAuthToken() error {
	// 若指定从环境变量读取 BCS AuthToken，则丢弃配置文件中的值
	if c.Global.BCSAPIGW.ReadAuthTokenFromEnv {
		c.Global.BCSAPIGW.AuthToken = envs.BCSApiGWAuthToken
	}
	return nil
}

// initJWTPubKey 初始化 jwt 公钥
func (c *ClusterResourcesConf) initJWTPubKey() error {
	if c.Global.Auth.JWTPubKey == "" {
		return nil
	}
	content, err := os.ReadFile(c.Global.Auth.JWTPubKey)
	if err != nil {
		return err
	}
	c.Global.Auth.JWTPubKeyObj, err = jwtGo.ParseRSAPublicKeyFromPEM(content)
	return err
}

// initIAM 初始化 iam
func (c *ClusterResourcesConf) initIAM() error {
	systemID, appCode, appSecret := c.Global.IAM.SystemID, c.Global.Basic.AppCode, c.Global.Basic.AppSecret
	if systemID == "" || appCode == "" || appSecret == "" {
		return errorx.New(errcode.ValidateErr, "SystemID/AppCode/AppSecret required")
	}
	// 支持蓝鲸 APIGW / 直连 IAMHost
	if c.Global.IAM.UseBKAPIGW {
		bkAPIGWHost := c.Global.Basic.BKAPIGWHost
		if bkAPIGWHost == "" {
			return errorx.New(errcode.ValidateErr, "BKAPIGWHost required")
		}
		c.Global.IAM.Cli = bkiam.NewAPIGatewayIAM(systemID, appCode, appSecret, bkAPIGWHost)
	} else {
		bkIAMHost, bkPaaSHost := c.Global.IAM.Host, c.Global.Basic.BKPaaSHost
		if bkIAMHost == "" || bkPaaSHost == "" {
			return errorx.New(errcode.ValidateErr, "BKIAMHost/BKPaaSHost required")
		}
		c.Global.IAM.Cli = bkiam.NewIAM(systemID, appCode, appSecret, bkIAMHost, bkPaaSHost)
	}
	// 指标相关
	if c.Global.IAM.Metric {
		metric.RegisterMetrics()
	}
	// 调试模式
	defaultLogLevel := logrus.ErrorLevel
	if c.Global.IAM.Debug {
		defaultLogLevel = logrus.DebugLevel
	}
	log := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        defaultLogLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	logger.SetLogger(log)
	return nil
}

// initCompoment 初始化 compoment
func (c *ClusterResourcesConf) initCompoment() error {
	if c.Global.Component.BCSStorageHost == "" {
		c.Global.Component.BCSStorageHost = "https://bcs-storage:50024"
	}
	return nil
}

// EtcdConf Etcd 相关配置
type EtcdConf struct {
	EtcdEndpoints string `yaml:"endpoints" usage:"Etcd Endpoints"`
	EtcdCert      string `yaml:"cert" usage:"Etcd Cert"`
	EtcdKey       string `yaml:"key" usage:"Etcd Key"`
	EtcdCa        string `yaml:"ca" usage:"Etcd CA"`
}

// ServerConf Server 配置
type ServerConf struct {
	UseLocalIP       bool   `yaml:"useLocalIP" usage:"是否使用 Local IP"`
	Address          string `yaml:"address" usage:"服务启动地址"`
	AddressIPv6      string `yaml:"addressIPv6" usage:"服务启动ipv6地址"`
	InsecureAddress  string `yaml:"insecureAddress" usage:"服务启动地址（非安全）"`
	Port             int    `yaml:"port" usage:"GRPC 服务端口"`
	HTTPPort         int    `yaml:"httpPort" usage:"HTTP 服务端口"`
	MetricPort       int    `yaml:"metricPort" usage:"Metric 服务端口"`
	RegisterTTL      int    `yaml:"registerTTL" usage:"注册TTL"` // nolint:tagliatelle
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
	Enabled bool `yaml:"enabled" usage:"是否启用 swagger 服务"`
}

// LogConf 日志配置
type LogConf struct {
	Level         string `yaml:"level" usage:"日志级别"`
	FlushInterval int    `yaml:"flushInterval" usage:"刷新数据的间隔"`
	AutoCreateDir bool   `yaml:"autoCreateDir" usage:"是否自动创建日志目录"`
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
	RedisMode    string `yaml:"redisMode" usage:"Redis Mode"`
	MasterName   string `yaml:"masterName" usage:"Redis MasterName for Sentinel Mode"`
	DialTimeout  int    `yaml:"dialTimeout" usage:"Redis Dial Timeout"`
	ReadTimeout  int    `yaml:"readTimeout" usage:"Redis Read Timeout(s)"`
	WriteTimeout int    `yaml:"writeTimeout" usage:"Redis Write Timeout(s)"`
	PoolSize     int    `yaml:"poolSize" usage:"Redis Pool Size"`
	MinIdleConns int    `yaml:"minIdleConns" usage:"Redis Min Idle Conns"`
	IdleTimeout  int    `yaml:"idleTimeout" usage:"Redis Idle Timeout(min)"`
}

// MongoConfig option for mongo
type MongoConfig struct {
	Address        string `json:"address" yaml:"address"`
	Replicaset     string `json:"replicaset" yaml:"replicaset"`
	ConnectTimeout uint   `json:"connectTimeout" yaml:"connectTimeout"`
	AuthDatabase   string `json:"authDatabase" yaml:"authDatabase"`
	Database       string `json:"database" yaml:"database"`
	Username       string `json:"username" yaml:"username"`
	Password       string `json:"password" yaml:"password"`
	MaxPoolSize    uint   `json:"maxPoolSize" yaml:"maxPoolSize"`
	MinPoolSize    uint   `json:"minPoolSize" yaml:"minPoolSize"`
	Encrypted      bool   `json:"encrypted" yaml:"encrypted"`
}

// GlobalConf 全局配置，可在业务逻辑中使用
type GlobalConf struct {
	Auth          AuthConf          `yaml:"auth"`
	Basic         BasicConf         `yaml:"basic"`
	BCSAPIGW      BCSAPIGatewayConf `yaml:"bcsApiGW"` // nolint:tagliatelle
	Component     ComponentConf     `yaml:"component"`
	IAM           IAMConf           `yaml:"iam"`
	SharedCluster SharedClusterConf `yaml:"sharedCluster"`
	MultiCluster  MultiClusterConf  `yaml:"multiCluster"`
}

// ComponentConf 组件配置
type ComponentConf struct {
	BCSStorageHost string `yaml:"bcsStorageHost" usage:"BCS Storage Host"`
}

// AuthConf 认证相关配置
type AuthConf struct {
	Disabled     bool           `yaml:"disabled" usage:"是否禁用身份认证"`
	JWTPubKey    string         `yaml:"jwtPublicKey" usage:"jwt 公钥（文件路径）"`
	JWTPubKeyObj *rsa.PublicKey `yaml:"-" usage:"jwt 公钥对象（自动生成）"`
}

// BasicConf 项目基础配置
type BasicConf struct {
	AppCode      string `yaml:"appCode" usage:"应用 ID"`
	AppSecret    string `yaml:"appSecret" usage:"应用 Secret"`
	BKAPIGWHost  string `yaml:"bkApiGWHost" usage:"蓝鲸 API 网关 Host"` // nolint:tagliatelle
	BKPaaSHost   string `yaml:"bkPaaSHost" usage:"蓝鲸 PaaS（esb）Host"`
	HealthzToken string `yaml:"healthzToken" usage:"Healthz API Token"`
	CacheToken   string `yaml:"cacheToken" usage:"刷新缓存用 Token"`
}

// BCSAPIGatewayConf 容器服务网关配置
type BCSAPIGatewayConf struct {
	Host                  string `yaml:"host" usage:"容器服务网关 Host"`
	AuthToken             string `yaml:"authToken" usage:"网关 AuthToken"`
	EnableMultiTenantMode bool   `yaml:"enableMultiTenantMode" usage:"多租户模式"`
	ReadAuthTokenFromEnv  bool   `yaml:"readAuthTokenFromEnv" usage:"是否从环境变量获取 AuthToken（适用于同集群部署情况）"`
}

// IAMConf 权限中心相关配置
type IAMConf struct {
	Host       string     `yaml:"host" usage:"权限中心 V3 Host"`
	SystemID   string     `yaml:"systemID" usage:"接入系统的 ID"`                                  // nolint:tagliatelle
	UseBKAPIGW bool       `yaml:"useBKApiGW" usage:"为真则使用蓝鲸 apigw，否则使用 iamHost + bkPaaSHost"` // nolint:tagliatelle
	Metric     bool       `yaml:"metric" usage:"支持 prometheus metrics"`
	Debug      bool       `yaml:"debug" usage:"启用 iam 调试模式"`
	Cli        *bkiam.IAM `yaml:"-" usage:"iam Client 对象（自动生成）"`
}

// SharedClusterConf 共享集群相关配置
type SharedClusterConf struct {
	EnabledCObjKinds         []string `yaml:"enabledCObjKinds" usage:"共享集群中支持的自定义对象 Kind"`
	EnabledCRDs              []string `yaml:"enabledCRDs" usage:"共享集群中支持的 CRD"` // nolint:tagliatelle
	AnnotationKeyProjectCode string   `yaml:"annotationKeyProjectCode" usage:"共享集群ProjectCode注解的key"`
}

// MultiClusterConf 多集群相关配置
type MultiClusterConf struct {
	EnabledQueryFromStorageKinds []string             `yaml:"enabledQueryFromStorageKinds"`
	EnabledCountKinds            []string             `yaml:"enabledCountKinds"`
	BlacklistForAPIServerQuery   []ProjectClusterConf `yaml:"blacklistForAPIServerQuery"`
}

// ProjectClusterConf 项目集群配置
type ProjectClusterConf struct {
	ProjectCode  string `yaml:"projectCode"`
	ClusterIDReg string `yaml:"clusterIDReg"`
}
