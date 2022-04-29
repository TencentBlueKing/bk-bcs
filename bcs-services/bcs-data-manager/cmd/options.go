/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
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

// BcsMonitorConfig options for bcs monitor config
type BcsMonitorConfig struct {
	Schema              string `json:"schema"`
	BcsMonitorEndpoints string `json:"endpoints"`
	Password            string `json:"password"`
	User                string `json:"user"`
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

// DataManagerOptions options of data manager
type DataManagerOptions struct {
	conf.FileConfig
	conf.LogConfig
	ClientConfig
	ServerConfig
	Mongo          MongoOption        `json:"mongoConf"`
	BcsMonitorConf BcsMonitorConfig   `json:"bcsMonitorConf"`
	QueueConfig    QueueConfig        `json:"queueConfig"`
	HandleConfig   HandleConfig       `json:"handleConfig"`
	Etcd           EtcdOption         `json:"etcd"`
	BcsAPIConf     BcsAPIConfig       `json:"bcsApiConf"`
	Debug          bool               `json:"debug"`
	FilterRules    ClusterFilterRules `json:"filterRules"`
	AppCode        string             `json:"appCode"`
	AppSecret      string             `json:"appSecret"`
}

// ClusterFilterRules rules for cluster filter
type ClusterFilterRules struct {
	NeedFilter bool   `json:"needFilter"`
	ClusterIDs string `json:"clusterIDs"`
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
			Resource:        common.DataJobQueue,
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
		Debug: true,
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
