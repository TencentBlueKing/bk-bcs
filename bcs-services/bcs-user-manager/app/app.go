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

// Package app xxx
package app

import (
	"fmt"
	"net"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/esb/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/metrics"
	usermanager "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

// Run run app
func Run(op *options.UserManagerOptions) {
	conf, err := parseConfig(op)
	if err != nil {
		blog.Errorf("error parse config: %s", err.Error())
		os.Exit(1)
	}
	// set userManager global config
	config.SetGlobalConfig(conf)

	// init cluster cache
	component.CacheClusterList()

	if conf.ClientCert.IsSSL {
		// nolint
		cliTLS, err := ssl.ClientTslConfVerity(conf.ClientCert.CAFile, conf.ClientCert.CertFile, conf.ClientCert.KeyFile,
			conf.ClientCert.CertPasswd)
		if err != nil {
			blog.Errorf("set client tls config error %s", err.Error())
		} else {
			config.CliTls = cliTLS
			blog.Infof("set client tls config success")
		}
	}
	userManager := usermanager.NewUserManager(conf)

	// init jwt client
	err = jwt.InitJWTClient(op)
	if err != nil {
		blog.Errorf("init jwt client error: %s", err.Error())
		os.Exit(1)
	}

	// init cmdb client
	if err = cmdb.InitCMDBClient(op); err != nil {
		blog.Errorf("init cmdb client error: %s", err.Error())
		os.Exit(1)
	}

	// start userManager, and http service
	err = userManager.Start()
	if err != nil {
		blog.Errorf("start processor error %s, and exit", err.Error())
		os.Exit(1)
	}

	// pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	metrics.RunMetric(conf)
}

// parseConfig parse the option to config
// nolint:funlen
func parseConfig(op *options.UserManagerOptions) (*config.UserMgrConfig, error) {
	userMgrConfig := config.NewUserMgrConfig()

	userMgrConfig.Address = op.Address
	ipv6Address := util.InitIPv6Address(op.IPv6Address)
	// 如果没有主动配置IPv6，同时也没有从环境变量解析出可用IPv6，则不监听IPv6地址
	if op.IPv6Address == "" && ipv6Address == net.IPv6loopback.String() &&
		util.GetIPv6Address(os.Getenv(types.LOCALIPV6)) == "" {
		ipv6Address = ""
	}
	userMgrConfig.IPv6Address = ipv6Address
	userMgrConfig.Port = op.Port
	userMgrConfig.InsecureAddress = op.InsecureAddress
	userMgrConfig.InsecurePort = op.InsecurePort
	userMgrConfig.LocalIp = op.LocalIP
	userMgrConfig.MetricPort = op.MetricPort
	userMgrConfig.BootStrapUsers = op.BootStrapUsers
	userMgrConfig.TKE = op.TKE
	userMgrConfig.PeerToken = op.PeerToken
	userMgrConfig.PermissionSwitch = op.PermissionSwitch
	userMgrConfig.CommunityEdition = op.CommunityEdition
	userMgrConfig.BcsAPI = &op.BcsAPI
	userMgrConfig.Encrypt = op.Encrypt
	userMgrConfig.Activity = op.Activity
	userMgrConfig.EnableTokenSync = op.EnableTokenSync

	config.Tke = op.TKE
	secretID, err := encrypt.DesDecryptFromBase([]byte(config.Tke.SecretID))
	if err != nil {
		return nil, fmt.Errorf("error decrypting tke secretId and exit: %s", err.Error())
	}
	config.Tke.SecretID = string(secretID)
	secretKey, err := encrypt.DesDecryptFromBase([]byte(config.Tke.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("error decrypting tke secretKey and exit: %s", err.Error())
	}
	config.Tke.SecretKey = string(secretKey)

	dsn, err := encrypt.DesDecryptFromBase([]byte(op.DSN))
	if err != nil {
		return nil, fmt.Errorf("error decrypting db config and exit: %s", err.Error())
	}
	userMgrConfig.DSN = string(dsn)

	redisDSN, err := encrypt.DesDecryptFromBase([]byte(op.RedisDSN))
	if err != nil {
		return nil, fmt.Errorf("error decrypting redis config and exit: %s", err.Error())
	}
	userMgrConfig.RedisDSN = string(redisDSN)

	// RedisDSN 没有配置，检查 RedisConfig
	if userMgrConfig.RedisDSN == "" {
		redisConfig, aErr := parseRedisConfig(op.RedisConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing redis config and exit: %s", aErr.Error())
		}
		userMgrConfig.RedisConfig = redisConfig
	}

	userMgrConfig.VerifyClientTLS = op.VerifyClientTLS

	// server cert directory
	if op.CertConfig.ServerCertFile != "" && op.CertConfig.ServerKeyFile != "" {
		userMgrConfig.ServCert.CertFile = op.CertConfig.ServerCertFile
		userMgrConfig.ServCert.KeyFile = op.CertConfig.ServerKeyFile
		userMgrConfig.ServCert.CAFile = op.CertConfig.CAFile
		userMgrConfig.ServCert.IsSSL = true

		userMgrConfig.TlsServerConfig, err = ssl.ServerTslConfVerityClient(op.CertConfig.CAFile, op.CertConfig.ServerCertFile,
			op.CertConfig.ServerKeyFile, userMgrConfig.ServCert.CertPasswd)
		if err != nil {
			blog.Errorf("initServerTLSConfig failed: %v", err)
			return nil, err
		}
	}

	// client cert directory
	if op.CertConfig.ClientCertFile != "" && op.CertConfig.ClientKeyFile != "" {
		userMgrConfig.ClientCert.CertFile = op.CertConfig.ClientCertFile
		userMgrConfig.ClientCert.KeyFile = op.CertConfig.ClientKeyFile
		userMgrConfig.ClientCert.CAFile = op.CertConfig.CAFile
		userMgrConfig.ClientCert.IsSSL = true

		userMgrConfig.TlsClientConfig, err = ssl.ClientTslConfVerity(op.CertConfig.CAFile, op.CertConfig.ClientCertFile,
			op.CertConfig.ClientKeyFile, userMgrConfig.ClientCert.CertPasswd)
		if err != nil {
			blog.Errorf("initClientTLSConfig failed: %v", err)
			return nil, err
		}
	}

	userMgrConfig.EtcdConfig = op.Etcd
	userMgrConfig.IAMConfig = op.IAMConfig

	return userMgrConfig, nil
}

// parseRedisConfig parse redis option when redisDsn is empty
func parseRedisConfig(redisOp options.RedisConfig) (config.RedisConfig, error) {
	conf := config.RedisConfig{}
	redisPassword, err := encrypt.DesDecryptFromBase([]byte(redisOp.Password))
	if err != nil {
		return conf, fmt.Errorf("error decrypting redis config and exit: %s", err.Error())
	}
	conf.RedisMode = redisOp.RedisMode
	conf.MasterName = redisOp.MasterName
	conf.Addr = redisOp.Addr
	conf.Password = string(redisPassword)
	conf.DialTimeout = redisOp.DialTimeout
	conf.ReadTimeout = redisOp.ReadTimeout
	conf.WriteTimeout = redisOp.WriteTimeout
	conf.PoolSize = redisOp.PoolSize
	conf.MinIdleConns = redisOp.MinIdleConns
	conf.IdleTimeout = redisOp.IdleTimeout
	return conf, nil
}
