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
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common"
)

// 检查配置加载情况，若默认配置修改，需要同步调整该单元测试
func TestLoadConf(t *testing.T) {
	crOpts, err := LoadConf("../../" + common.DefaultConfPath)
	if err != nil {
		t.Errorf("Load default conf error: %v", err)
	}
	// 检查 debug 配置
	debug := false
	if crOpts.Debug != debug {
		t.Errorf("Conf debug, Excepted: %v, Result: %v", debug, crOpts.Debug)
	}
	// 检查 etcd 配置
	etcdEndpoints := "127.0.0.1:2379"
	if crOpts.Etcd.EtcdEndpoints != etcdEndpoints {
		t.Errorf("Conf etcd.endpoints, Excepted: %v, Result: %v", etcdEndpoints, crOpts.Etcd.EtcdEndpoints)
	}
	// 检查 server 配置
	address, httpPort := "127.0.0.1", 9091
	if crOpts.Server.Address != address {
		t.Errorf("Conf server.address, Excepted: %v, Result: %v", address, crOpts.Server.Address)
	}
	if crOpts.Server.HTTPPort != httpPort {
		t.Errorf("Conf server.httpPort, Excepted: %v, Result: %v", httpPort, crOpts.Server.HTTPPort)
	}
	// 检查 client 配置
	clientCert := ""
	if crOpts.Client.Cert != clientCert {
		t.Errorf("Conf client.cert, Excepted: %v, Result: %v", clientCert, crOpts.Client.Cert)
	}
	// 检查 swagger 配置
	swaggerDir := ""
	if crOpts.Swagger.Dir != swaggerDir {
		t.Errorf("Conf swagger.dir, Excepted: %v, Result: %v", swaggerDir, crOpts.Swagger.Dir)
	}
	// 检查 log 配置
	logDir, logMaxSize := "logs", uint64(500)
	if crOpts.Log.LogDir != logDir {
		t.Errorf("Conf log.logdir, Excepted: %v, Result: %v", logDir, crOpts.Log.LogDir)
	}
	if crOpts.Log.LogMaxSize != logMaxSize {
		t.Errorf("Conf log.logMaxSize, Excepted: %v, Result: %v", logMaxSize, crOpts.Log.LogMaxSize)
	}
	// 检查 redis 配置
	redisAddress, redisPwd := "127.0.0.1:6379", ""
	if crOpts.Redis.Address != redisAddress {
		t.Errorf("Conf redis.host, Excepted: %v, Result: %v", redisAddress, crOpts.Redis.Address)
	}
	if crOpts.Redis.Password != redisPwd {
		t.Errorf("Conf redis.password, Excepted: %v, Result: %v", redisPwd, crOpts.Redis.Password)
	}

}
