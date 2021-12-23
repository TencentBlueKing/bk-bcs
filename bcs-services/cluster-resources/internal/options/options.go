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

package options

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Etcd 相关配置
type EtcdOption struct {
	EtcdEndpoints string `yaml:"endpoints" value:"" usage:"Etcd Endpoints"`
	EtcdCert      string `yaml:"cert" value:"" usage:"Etcd Cert"`
	EtcdKey       string `yaml:"key" value:"" usage:"Etcd Key"`
	EtcdCa        string `yaml:"ca" value:"" usage:"Etcd CA"`
}

// Server 配置
type ServerConfig struct {
	Address         string `yaml:"address" value:"127.0.0.1" usage:"服务启动地址"`
	InsecureAddress string `yaml:"insecureAddress" value:"127.0.0.1" usage:"服务启动地址（非安全）"`
	Port            uint   `yaml:"port" value:"9090" usage:"GRPC 服务端口"`
	HTTPPort        uint   `yaml:"httpPort" value:"9091" usage:"HTTP 服务端口"`
	MetricPort      uint   `yaml:"metricPort" value:"9092" usage:"Metric 服务端口"`
	Cert            string `yaml:"cert" value:"" usage:"Server Cert"`
	Key             string `yaml:"key" value:"" usage:"Server Key"`
	Ca              string `yaml:"ca" value:"" usage:"Server CA"`
}

// Client 配置
type ClientConfig struct {
	Cert string `yaml:"cert" value:"" usage:"Client Cert"`
	Key  string `yaml:"key" value:"" usage:"Client Key"`
	Ca   string `yaml:"ca" value:"" usage:"Client CA"`
}

// Swagger 配置
type SwaggerConfig struct {
	Enabled bool   `yaml:"enabled" value:"false" usage:"是否启用 swagger 服务"`
	Dir     string `yaml:"dir" value:"./swagger/data" usage:"swagger.json 存放目录"`
}

// 日志配置，字段同 bcs-common.conf.LogConfig，来源调整为 yaml
type LogConfig struct {
	LogDir          string `yaml:"logDir" value:"./logs" usage:"日志文件存储路径"`
	LogMaxSize      uint64 `yaml:"logMaxSize" value:"500" usage:"单个文件最大 size (MB)"`
	LogMaxNum       int    `yaml:"logMaxNum" value:"10" usage:"最大日志文件数量，若超过则移除最先生成的文件"`
	ToStdErr        bool   `yaml:"logToStderr" value:"false" usage:"输出日志到 stderr 而不是文件"`
	AlsoToStdErr    bool   `yaml:"alsoLogToStderr" value:"false" usage:"输出日志到文件同时输出到 stderr"`
	Verbosity       int32  `yaml:"v" value:"0" usage:"显示所有 VLOG(m) 的日志， m 小于等于该 flag 的值，会被 VModule 覆盖"`
	StdErrThreshold string `yaml:"stderrThreshold" value:"2" usage:"将大于等于该级别的日志同时输出到 stderr"`
	VModule         string `yaml:"VModule" value:"" usage:"每个模块的详细日志的级别"`
	TraceLocation   string `yaml:"logBacktraceAt" value:"" usage:"当日志记录命中 line file:N 时，发出堆栈跟踪"`
}

// ClusterResources 服务启动配置
type ClusterResourcesOptions struct {
	Debug   bool          `yaml:"debug"`
	Etcd    EtcdOption    `yaml:"etcd"`
	Server  ServerConfig  `yaml:"server"`
	Client  ClientConfig  `yaml:"client"`
	Swagger SwaggerConfig `yaml:"swagger"`
	Log     LogConfig     `yaml:"log"`
}

// 加载配置信息
func LoadConf(filePath string) (*ClusterResourcesOptions, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	opts := &ClusterResourcesOptions{}
	err = yaml.Unmarshal(yamlFile, opts)
	if err != nil {
		return nil, err
	}
	return opts, nil
}
