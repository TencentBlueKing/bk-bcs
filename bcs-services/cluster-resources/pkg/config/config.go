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

// LogConf 日志配置
type LogConf struct {
	Level         string            `yaml:"level" usage:"日志级别"`
	WriterType    string            `yaml:"writerType" usage:"写到文件或者标准输出"`
	FlushInterval int               `yaml:"flushInterval" usage:"写入数据的间隔"`
	Settings      map[string]string `yaml:"settings" usage:"文件或者标准输出的配置信息"`
}

// ClusterResourcesConf ClusterResources 服务启动配置
type ClusterResourcesConf struct {
	Debug   bool        `yaml:"debug"`
	Etcd    EtcdConf    `yaml:"etcd"`
	Server  ServerConf  `yaml:"server"`
	Client  ClientConf  `yaml:"client"`
	Swagger SwaggerConf `yaml:"swagger"`
	Log     LogConf     `yaml:"log"`
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
	return conf, nil
}
