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

// EtcdConfig 依赖的 etcd 服务的配置
type EtcdConfig struct {
	EtcdEndpoints string `yaml:"endpoints" usage:"endpoints of etcd"`
	EtcdCert      string `yaml:"cert" usage:"cert file of etcd"`
	EtcdKey       string `yaml:"key" usage:"key file for etcd"`
	EtcdCa        string `yaml:"ca" usage:"ca file for etcd"`
}

// MongoConfig
type MongoConfig struct {
	Address        string `yaml:"address"`
	ConnectTimeout uint   `yaml:"connecttimeout"`
	Database       string `yaml:"database"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	MaxPoolSize    uint   `yaml:"maxpoolsize"`
	MinPoolSize    uint   `yaml:"minpoolsize"`
	Encrypted      bool   `yaml:"encrypted"`
}

// ServerConfig 服务的配置
type ServerConfig struct {
	Address         string `yaml:"address" usage:"server address"`
	InsecureAddress string `yaml:"insecureAddress" usage:"insecurue server address"`
	Port            int    `yaml:"port" usage:"grpc port"`
	HTTPPort        int    `yaml:"httpPort" usage:"http port"`
	MetricPort      int    `yaml:"metricPort" usage:"metric port"`
	Cert            string `yaml:"cert" usage:"server cert"`
	CertPwd         string `yaml:"certPwd" usage:"server cert password"`
	Key             string `yaml:"key" usage:"server key"`
	Ca              string `yaml:"ca" usage:"server ca"`
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Cert    string `yaml:"cert" usage:"client cert"`
	CertPwd string `yaml:"certPwd" usage:"client cert password"`
	Key     string `yaml:"key" usage:"client key"`
	Ca      string `yaml:"ca" usage:"client ca"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level         string `yaml:"level" usage:"log level"`
	FlushInterval int    `yaml:"flushInterval" usage:"interval of flush"`
	Path          string `yaml:"path" usage:"log file path, example: /tmp/logs"`
	Name          string `yaml:"name" usage:"log file name, example: cr.log"`
	Size          int    `yaml:"size" usage:"log file size, unit: MB"`
	Age           int    `yaml:"age" usage:"log reserved age, unit: day"`
	Backups       int    `yaml:"backups" usage:"the count of reserved log"`
}

// SwaggerConfig
type SwaggerConfig struct {
	Dir string `yaml:"dir" usage:"swagger dir"`
}

type ProjectConfig struct {
	Etcd    EtcdConfig    `yaml:"etcd"`
	Mongo   MongoConfig   `yaml:"mongo"`
	Log     LogConfig     `yaml:"log"`
	Swagger SwaggerConfig `yaml:"swagger"`
	Server  ServerConfig  `yaml:"server"`
	Client  ClientConfig  `yaml:"client"`
}

// LoadConfig 通过制定的path，加载对应的配置选项
func LoadConfig(filePath string) (*ProjectConfig, error) {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	conf := &ProjectConfig{}
	if err = yaml.Unmarshal(yamlFile, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
