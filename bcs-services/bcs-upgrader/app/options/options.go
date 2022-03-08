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
 *
 */

package options

import (
	"flag"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
)

//CertConfig is configuration of Cert
type CertConfig struct {
	CAFile   string
	CertFile string
	KeyFile  string
	CertPwd  string
	IsSSL    bool
}

//UpgraderOptions is options in flags
type UpgraderOptions struct {
	conf.FileConfig
	conf.ServiceConfig
	conf.ServerOnlyCertConfig
	conf.LogConfig
	conf.ProcessConfig
	MongoConfig
	HttpCliConfig
	ServerCert *CertConfig

	DebugMode bool `json:"debug_mode" value:"false" usage:"Debug mode, use pprof."`
}

// MongoConfig option for mongo
type MongoConfig struct {
	MongoAuthMechanism  string `json:"mongo_authmechanism" value:"" usage:"the mechanism to use for authentication"`
	MongoAddress        string `json:"mongo_address" value:"127.0.0.1:27017" usage:"mongo server address"`
	MongoConnectTimeout uint   `json:"mongo_connecttimeout"  value:"3" usage:"mongo server connnect timeout"`
	MongoDatabase       string `json:"mongo_database" value:"bcs" usage:"database in mongo for cluster manager"`
	MongoUsername       string `json:"mongo_username"  value:"" usage:"mongo username for cluster manager"`
	MongoPassword       string `json:"mongo_password" value:"" usage:"mongo passsword for cluster manager"`
	MongoMaxPoolSize    uint   `json:"mongo_maxpoolsize" value:"0" usage:"mongo client connection pool max size"`
	MongoMinPoolSize    uint   `json:"mongo_minpoolsize" value:"0" usage:"mongo client connection pool min size"`
}

// HttpCliConfig option for HttpCliConfig
type HttpCliConfig struct {
	CcHOST                   string `json:"cc_host" value:"" usage:"request bcs saas cc host"`
	BkAppSecret              string `json:"bk_app_secret" value:"" usage:"request ssm for http header"`
	SsmHost                  string `json:"ssm_host" value:"" usage:"request ssm host"`
	SsmAccessToken           string `json:"ssm_access_token" value:"" usage:"ssm access token"`
	CmHost                   string `json:"cm_host"  value:"" usage:"request cluster manager host"`
	GatewayToken             string `json:"gateway_token" value:"" usage:"bcs api gateway token"`
	HttpCliCertConfig        *commtypes.CertConfig
	ClusterManagerCertConfig *commtypes.CertConfig
}

// AddFlags add cmdline flags
func AddFlags() {
	// mongo config
	flag.String("mongo_address", "127.0.0.1:27017", "mongo server address")
	flag.Uint("mongo_connecttimeout", 3, "mongo server connnect timeout")
	flag.String("mongo_database", "", "database in mongo for cluster manager")
	flag.String("mongo_username", "", "mongo username for cluster manager")
	flag.String("mongo_password", "", "mongo passsword for cluster manager")
	flag.Uint("mongo_maxpoolsize", 0, "mongo client connection pool max size, 0 means not set")
	flag.Uint("mongo_minpoolsize", 0, "mongo client connection pool min size, 0 means not set")
}

//NewUpgraderOptions create UpgraderOptions object
func NewUpgraderOptions() *UpgraderOptions {
	return &UpgraderOptions{
		ServerCert: &CertConfig{
			CertPwd: static.ServerCertPwd,
			IsSSL:   false,
		},
	}
}
