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

// Package cmd xxx
package cmd

import (
	"crypto/tls"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	jsoniter "github.com/json-iterator/go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

// MongoOption option for mongo db
type MongoOption struct {
	// MongoEndpoints addr of mongodb
	MongoEndpoints string `json:"endpoints"`
	// MongoConnectTimeout connect timeout of mongodb
	MongoConnectTimeout int `json:"connecttimeout"`
	// MongoDatabaseName database of mongodb
	MongoDatabaseName string `json:"database"`
	// MongoUsername username of mongodb
	MongoUsername string `json:"username"`
	// MongoPassword password of mongodb
	MongoPassword string `json:"password"`
}

// TspiderOption option for mysql db
type TspiderOption struct {
	StoreName  string `json:"storename"`
	Connection string `json:"connection"`
}

// BcsMonitorConfig options for bcs monitor config
type BcsMonitorConfig struct {
	BcsMonitorEndpoints string `json:"endpoints"`
}

// EtcdOption option for etcd
type EtcdOption struct {
	EtcdEndpoints string `json:"endpoints" value:"" usage:"endpoints of etcd"`
	EtcdCert      string `json:"cert" value:"" usage:"cert file of etcd"`
	EtcdKey       string `json:"key" value:"" usage:"key file for etcd"`
	EtcdCa        string `json:"ca" value:"" usage:"ca file for etcd"`
}

// BcsAPIConfig contains several bcs module endpoint
type BcsAPIConfig struct {
	BcsAPIGwURL    string `json:"bcsApiGatewayUrl"`
	OldBcsAPIGwURL string `json:"oldBcsApiGwUrl"`
	AdminToken     string `json:"adminToken"`
	GrpcGWAddress  string `json:"grpcGwAddress"`
	UserName       string `json:"userName"`
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

// ClientConfig option for as client
type ClientConfig struct {
	ClientCert string `json:"clientcert"`
	ClientKey  string `json:"clientkey"`
	ClientCa   string `json:"clientca"`
}

// QueueConfig queue config
type QueueConfig struct {
	Resource     string `json:"resource"`
	QueueFlag    bool   `json:"queue_flag"`
	QueueAddress string `json:"queue_address"`
	ExchangeName string `json:"exchange_name"`
	// nats-streaming
	ClusterID      string `json:"clusterID"`
	ConnectTimeout int    `json:"connectTimeout"`
	ConnectRetry   bool   `json:"connectRetry"`

	// publishOpts
	PublishDelivery int    `json:"publishDelivery"`
	QueueArguments  string `json:"queueArguments"`
}

// HandleConfig handle config
type HandleConfig struct {
	Concurrency  int64 `json:"concurrency"`
	ChanQueueLen int64 `json:"chanQueueLen"`
}

// KafkaConfig kafka config
type KafkaConfig struct {
	Address   string `json:"address"`
	Network   string `json:"network"`
	Topic     string `json:"topic"`
	Partition int    `json:"partition"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

// SharedClusterConfig options of shared cluster
type SharedClusterConfig struct {
	AnnoKeyProjCode string `json:"annoKeyProjCode"`
}

// DataManagerOptions options of data manager
type DataManagerOptions struct {
	conf.FileConfig
	conf.LogConfig
	ClientConfig
	ServerConfig
	Mongo                  MongoOption         `json:"mongoConf"`
	BcsMonitorConf         BcsMonitorConfig    `json:"bcsMonitorConf"`
	QueueConfig            QueueConfig         `json:"queueConfig"`
	HandleConfig           HandleConfig        `json:"handleConfig"`
	Etcd                   EtcdOption          `json:"etcd"`
	BcsAPIConf             BcsAPIConfig        `json:"bcsApiConf"`
	Debug                  bool                `json:"debug"`
	FilterRules            ClusterFilterRules  `json:"filterRules"`
	AppCode                string              `json:"appCode"`
	AppSecret              string              `json:"appSecret"`
	ProducerConfig         ProducerConfig      `json:"producerConfig"`
	KafkaConfig            KafkaConfig         `json:"kafkaConfig"`
	SharedClusterConfig    SharedClusterConfig `json:"sharedClusterConfig"`
	NeedSendKafka          bool                `json:"needSendKafka"`
	IgnoreBkMonitorCluster bool                `json:"ignoreBkMonitorCluster"`
	QueryFromBkMonitor     bool                `json:"queryFromBkMonitor"`
	BkbaseConfigPath       string              `json:"bkbaseConfigPath"`
	TspiderConfigPath      string              `json:"tspiderConfigPath"`
}

// ClusterFilterRules rules for cluster filter
type ClusterFilterRules struct {
	NeedFilter bool   `json:"needFilter"`
	ClusterIDs string `json:"clusterIDs"`
	Env        string `json:"env"`
}

// ProducerConfig config for producer
type ProducerConfig struct {
	Concurrency int `json:"concurrency"`
}

// NewDataManagerOptions new dataManagerOptions
func NewDataManagerOptions() *DataManagerOptions {
	return &DataManagerOptions{
		LogConfig: conf.LogConfig{
			LogDir:       "./logs",
			AlsoToStdErr: true,
			Verbosity:    2,
		},
		ServerConfig: ServerConfig{
			Address:    "127.0.0.1",
			Port:       8081,
			HTTPPort:   8080,
			MetricPort: 8082,
		},
		Mongo: MongoOption{
			MongoEndpoints:      "127.0.0.1:27017",
			MongoConnectTimeout: 3,
		},
		QueueConfig: QueueConfig{
			Resource:        types.DataJobQueue,
			QueueFlag:       true,
			ExchangeName:    "bcs-data-manager",
			ConnectTimeout:  300,
			ConnectRetry:    true,
			PublishDelivery: 2,
		},
		HandleConfig: HandleConfig{
			Concurrency:  100,
			ChanQueueLen: 1024,
		},
		Etcd: EtcdOption{
			EtcdEndpoints: "127.0.0.1:2379",
		},
		Debug:            true,
		BkbaseConfigPath: "/data/bcs/bkbase/bkbaseconfig.json",
	}
}

// GetEtcdRegistryTLS get specified etcd registry tls config
func (opt *DataManagerOptions) GetEtcdRegistryTLS() (*tls.Config, error) {
	config, err := ssl.ClientTslConfVerity(opt.Etcd.EtcdCa, opt.Etcd.EtcdCert,
		opt.Etcd.EtcdKey, "")
	if err != nil {
		blog.Errorf("gateway-discovery etcd TLSConfig with CA/Cert/Key failed, %s", err.Error())
		return nil, err
	}
	return config, nil
}

// ParseTspiderConfig parse config for tspider db
func (opt *DataManagerOptions) ParseTspiderConfig() ([]*TspiderOption, error) {
	config := []*TspiderOption{}
	bytes, err := os.ReadFile(opt.TspiderConfigPath)
	if err != nil {
		blog.Errorf("open tspider config file(%s) failed: %s", opt.TspiderConfigPath, err.Error())
		return nil, err
	}
	if err := jsoniter.Unmarshal(bytes, &config); err != nil {
		blog.Errorf("unmarshal config file(%s) failed: %s", opt.BkbaseConfigPath, err.Error())
		return nil, err
	}
	for _, item := range config {
		blog.Infof("tspider config: %+v", *item)
	}

	return config, nil
}

// ParseBkbaseConfig parse config for bkbase data
func (opt *DataManagerOptions) ParseBkbaseConfig() (*types.BkbaseConfig, error) {
	config := &types.BkbaseConfig{}
	bytes, err := os.ReadFile(opt.BkbaseConfigPath)
	if err != nil {
		blog.Errorf("open bkbase config file(%s) failed: %s", opt.BkbaseConfigPath, err.Error())
		return nil, err
	}
	if err := jsoniter.Unmarshal(bytes, config); err != nil {
		blog.Errorf("unmarshal config file(%s) failed: %s", opt.BkbaseConfigPath, err.Error())
		return nil, err
	}
	blog.Infof("bkbase config: %+v", config)
	return config, nil
}
