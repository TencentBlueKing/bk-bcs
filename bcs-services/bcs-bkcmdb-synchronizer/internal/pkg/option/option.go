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

// Package option define options for synchronizer
package option

import (
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
)

var globalConfig *BkcmdbSynchronizerOption

// GetGlobalConfig get global config
func GetGlobalConfig() *BkcmdbSynchronizerOption {
	return globalConfig
}

// SetGlobalConfig set global config
func SetGlobalConfig(config *BkcmdbSynchronizerOption) {
	globalConfig = config
}

// BkcmdbSynchronizerOption options for CostManager
type BkcmdbSynchronizerOption struct {
	Synchronizer  SynchronizerConfig  `json:"synchronizer" value:"synchronizer"`
	Client        ClientConfig        `json:"client"`
	Bcslog        conf.LogConfig      `json:"bcslog"`
	Bcsapi        BcsapiConfig        `json:"bcsapi"`
	RabbitMQ      RabbitMQConfig      `json:"rabbitmq"`
	CMDB          CMDBConfig          `json:"cmdb"`
	SharedCluster SharedClusterConfig `json:"shared_cluster"`
	BkUser        BkUserConfig        `json:"bkUser"`
}

// SynchronizerConfig synchronizer config
type SynchronizerConfig struct {
	Env                   string `json:"env"`
	Replicas              int    `json:"replicas"`
	BkBizID               int64  `json:"bkBizID"`
	HostID                int64  `json:"hostID"`
	WhiteList             string `json:"whiteList"`
	BlackList             string `json:"blackList"`
	EnableMultiTenantMode bool   `json:"enableMultiTenantMode"`
}

// ClientConfig client config
type ClientConfig struct {
	ClientCrtPwd string `json:"client_crt_pwd"`
	ClientCa     string `json:"client_ca"`
	ClientCrt    string `json:"client_crt"`
	ClientKey    string `json:"client_key"`
}

// BcsapiConfig bcsapi config
type BcsapiConfig struct {
	HttpAddr        string `json:"http_addr"`
	GrpcAddr        string `json:"grpc_addr"`
	BearerToken     string `json:"bearer_token"`
	ProjectToken    string `json:"project_token"`
	ProjectUsername string `json:"project_username"`
}

// RabbitMQConfig rabbitmq config
type RabbitMQConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Vhost          string `json:"vhost"`
	SourceExchange string `json:"source_exchange"`
}

// CMDBConfig cmdb config
type CMDBConfig struct {
	AppCode    string `json:"app_code"`
	AppSecret  string `json:"app_secret"`
	BKUserName string `json:"bk_username"`
	Server     string `json:"server"`
	Debug      bool   `json:"debug"`
}

// SharedClusterConfig shared cluster config
type SharedClusterConfig struct {
	AnnotationKeyProjCode string `json:"annotation_key_proj_code"`
}

// BkUserConfig bkuser config
type BkUserConfig struct {
	AppCode   string `json:"appCode"`
	AppSecret string `json:"appSecret"`
	Server    string `json:"server"`
	Debug     bool   `json:"debug"`
}

// InitTClientTlsConfig init client tls config
func InitTClientTlsConfig() (*tls.Config, error) {
	var (
		err       error
		tlsConfig *tls.Config
	)

	if len(globalConfig.Client.ClientCrt) != 0 &&
		len(globalConfig.Client.ClientKey) != 0 &&
		len(globalConfig.Client.ClientCa) != 0 {
		tlsConfig, err = ssl.ClientTslConfVerity(
			globalConfig.Client.ClientCa,
			globalConfig.Client.ClientCrt,
			globalConfig.Client.ClientKey,
			globalConfig.Client.ClientCrtPwd,
		)
		// static.ClientCertPwd)
		if err != nil {
			blog.Errorf("init tls config failed, err: %s", err.Error())
			return nil, err
		}
		blog.Infof("init tls config success")

	}
	return tlsConfig, nil
}
