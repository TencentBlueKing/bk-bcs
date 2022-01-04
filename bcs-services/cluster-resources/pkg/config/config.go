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
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// EtcdConf Etcd 相关配置
type EtcdConf struct {
	EtcdEndpoints string `yaml:"endpoints" value:"" usage:"Etcd Endpoints"`
	EtcdCert      string `yaml:"cert" value:"" usage:"Etcd Cert"`
	EtcdKey       string `yaml:"key" value:"" usage:"Etcd Key"`
	EtcdCa        string `yaml:"ca" value:"" usage:"Etcd CA"`
}

// ServerConf Server 配置
type ServerConf struct {
	Address          string `yaml:"address" value:"127.0.0.1" usage:"服务启动地址"`
	InsecureAddress  string `yaml:"insecureAddress" value:"127.0.0.1" usage:"服务启动地址（非安全）"`
	Port             int    `yaml:"port" value:"9090" usage:"GRPC 服务端口"`
	HTTPPort         int    `yaml:"httpPort" value:"9091" usage:"HTTP 服务端口"`
	MetricPort       int    `yaml:"metricPort" value:"9092" usage:"Metric 服务端口"`
	RegisterTTL      int    `yaml:"registerTTL" value:"30" usage:"注册TTL"` //nolint:tagliatelle
	RegisterInterval int    `yaml:"registerInterval" value:"25" usage:"注册间隔"`
	Cert             string `yaml:"cert" value:"" usage:"Server Cert"`
	Key              string `yaml:"key" value:"" usage:"Server Key"`
	Ca               string `yaml:"ca" value:"" usage:"Server CA"`
}

// ClientConf Client 配置
type ClientConf struct {
	Cert string `yaml:"cert" value:"" usage:"Client Cert"`
	Key  string `yaml:"key" value:"" usage:"Client Key"`
	Ca   string `yaml:"ca" value:"" usage:"Client CA"`
}

// SwaggerConf Swagger 配置
type SwaggerConf struct {
	Enabled bool   `yaml:"enabled" value:"false" usage:"是否启用 swagger 服务"`
	Dir     string `yaml:"dir" value:"./swagger/data" usage:"swagger.json 存放目录"`
}

// LogConf 日志配置，字段同 bcs-common.conf.LogConfig，来源调整为 yaml
type LogConf struct {
	LogDir          string `yaml:"logDir" value:"./logs" usage:"日志文件存储路径"`
	LogMaxSize      uint64 `yaml:"logMaxSize" value:"500" usage:"单个文件最大 size (MB)"`
	LogMaxNum       int    `yaml:"logMaxNum" value:"10" usage:"最大日志文件数量，若超过则移除最先生成的文件"`
	ToStdErr        bool   `yaml:"logToStderr" value:"false" usage:"输出日志到 stderr 而不是文件"`
	AlsoToStdErr    bool   `yaml:"alsoLogToStderr" value:"false" usage:"输出日志到文件同时输出到 stderr"`
	Verbosity       int32  `yaml:"v" value:"0" usage:"显示所有 VLOG(m) 的日志， m 小于等于该 flag 的值，会被 VModule 覆盖"`
	StdErrThreshold string `yaml:"stderrThreshold" value:"2" usage:"将大于等于该级别的日志同时输出到 stderr"`
	VModule         string `yaml:"VModule" value:"" usage:"每个模块的详细日志的级别"` //nolint:tagliatelle
	TraceLocation   string `yaml:"logBacktraceAt" value:"" usage:"当日志记录命中 line file:N 时，发出堆栈跟踪"`
}

// RedisConf Redis 配置
type RedisConf struct {
	Address      string `yaml:"address" value:"127.0.0.1:6379" usage:"Redis Server Address"`
	DB           int    `yaml:"db" value:"0" usage:"Redis DB"`
	Password     string `yaml:"password" value:"" usage:"Redis Password"`
	URL          string `yaml:"url" value:"redis://:@127.0.0.1:6379/0" usage:"Redis URL"`
	DialTimeout  int    `yaml:"dialTimeout" value:"" usage:"Redis Dial Timeout"`
	ReadTimeout  int    `yaml:"readTimeout" value:"" usage:"Redis Read Timeout(s)"`
	WriteTimeout int    `yaml:"writeTimeout" value:"" usage:"Redis Write Timeout(s)"`
	PoolSize     int    `yaml:"poolSize" value:"" usage:"Redis Pool Size"`
	MinIdleConns int    `yaml:"minIdleConns" value:"" usage:"Redis Min Idle Conns"`
	IdleTimeout  int    `yaml:"idleTimeout" value:"" usage:"Redis Idle Timeout(min)"`
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
}

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
	// 加载后处理
	if err = postLoadConf(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// postLoadConf 加载配置之后处理逻辑
func postLoadConf(conf *ClusterResourcesConf) error {
	// 如果配置中没有指定 Redis.URL，则根据规则和其他配置生成
	if len(conf.Redis.URL) == 0 {
		conf.Redis.URL = fmt.Sprintf("redis://:%s@%s/%d", conf.Redis.Password, conf.Redis.Address, conf.Redis.DB)
	}
	return nil
}
