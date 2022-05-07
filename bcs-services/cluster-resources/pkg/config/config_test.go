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

package config_test

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

// 检查配置加载情况，若默认配置修改，需要同步调整该单元测试
func TestLoadConf(t *testing.T) { //nolint:cyclop
	c, err := config.LoadConf("../../" + conf.DefaultConfPath)
	if err != nil {
		t.Errorf("Load default conf error: %v", err)
	}
	// 检查 debug 配置
	if c.Debug != false {
		t.Errorf("Conf debug, Excepted: false, Result: %v", c.Debug)
	}
	// 检查 etcd 配置
	etcdEndpoints := "127.0.0.1:2379"
	if c.Etcd.EtcdEndpoints != etcdEndpoints {
		t.Errorf("Conf etcd.endpoints, Excepted: %v, Result: %v", etcdEndpoints, c.Etcd.EtcdEndpoints)
	}
	// 检查 server 配置
	address, httpPort := "", 9091
	if c.Server.Address != address {
		t.Errorf("Conf server.address, Excepted: %v, Result: %v", address, c.Server.Address)
	}
	if c.Server.HTTPPort != httpPort {
		t.Errorf("Conf server.httpPort, Excepted: %v, Result: %v", httpPort, c.Server.HTTPPort)
	}
	// 检查 client 配置
	if clientCert := ""; c.Client.Cert != clientCert {
		t.Errorf("Conf client.cert, Excepted: %v, Result: %v", clientCert, c.Client.Cert)
	}
	// 检查 swagger 配置
	if swaggerDir := ""; c.Swagger.Dir != swaggerDir {
		t.Errorf("Conf swagger.dir, Excepted: %v, Result: %v", swaggerDir, c.Swagger.Dir)
	}
	// 检查 log 配置
	level, fileName := "info", "cr.log"
	if c.Log.Level != level {
		t.Errorf("Conf log.level, Excepted: %v, Result: %v", level, c.Log.Level)
	}
	if c.Log.Name != fileName {
		t.Errorf("Conf log.name, Excepted: %v, Result: %v", fileName, c.Log.Name)
	}
	// 检查 redis 配置
	redisAddress, redisPwd := "127.0.0.1:6379", ""
	if c.Redis.Address != redisAddress {
		t.Errorf("Conf redis.host, Excepted: %v, Result: %v", redisAddress, c.Redis.Address)
	}
	if c.Redis.Password != redisPwd {
		t.Errorf("Conf redis.password, Excepted: %v, Result: %v", redisPwd, c.Redis.Password)
	}
}
