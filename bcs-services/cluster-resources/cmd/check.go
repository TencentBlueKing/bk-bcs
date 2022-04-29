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

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	microEtcd "github.com/asim/go-micro/plugins/registry/etcd/v4"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache/redis"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

const (
	// 重试次数
	RetryTimes = 10
	// 检查间隔，单位：秒
	CheckInterval = 30
	// 依赖服务检查失败退出码
	ExitCode = 1
)

// DependencyServiceChecker ...
type DependencyServiceChecker struct {
	conf       *config.ClusterResourcesConf
	microRtr   registry.Registry
	cliTLSConf *tls.Config
}

// NewDependencyServiceChecker ...
func NewDependencyServiceChecker(conf *config.ClusterResourcesConf) *DependencyServiceChecker {
	return &DependencyServiceChecker{
		conf:       conf,
		microRtr:   initMicroRtr(conf.Etcd),
		cliTLSConf: initCliTLSConf(conf.Client),
	}
}

func initMicroRtr(conf config.EtcdConf) registry.Registry {
	etcdEndpoints := stringx.Split(conf.EtcdEndpoints)
	etcdSecure := false

	var etcdTLS *tls.Config
	var err error
	if conf.EtcdCa != "" && conf.EtcdCert != "" && conf.EtcdKey != "" {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(
			conf.EtcdCa, conf.EtcdCert, conf.EtcdKey, "",
		)
		if err != nil {
			panic(fmt.Sprintf("init etcd tls config failed: %v", err))
		}
	}

	fmt.Printf("registry: etcd endpoints: %v, secure: %t\n", etcdEndpoints, etcdSecure) // nolint:forbidigo

	microRtr := microEtcd.NewRegistry(
		registry.Addrs(etcdEndpoints...),
		registry.Secure(etcdSecure),
		registry.TLSConfig(etcdTLS),
	)
	if err = microRtr.Init(); err != nil {
		panic(fmt.Sprintf("init microRtr failed: %v", err))
	}
	return microRtr
}

func initCliTLSConf(conf config.ClientConf) *tls.Config {
	if conf.Cert != "" && conf.Key != "" && conf.Ca != "" {
		tlsConf, err := ssl.ClientTslConfVerity(conf.Ca, conf.Cert, conf.Key, conf.CertPwd)
		if err != nil {
			panic(fmt.Sprintf("load cluster resources client tls config failed: %v", err))
		}
		fmt.Println("load client tls config successfully") // nolint:forbidigo
		return tlsConf
	}
	return nil
}

func (c *DependencyServiceChecker) DoAndExit() {
	for i := 0; i < RetryTimes; i++ {
		fmt.Printf("try %d times\n", i) // nolint:forbidigo

		if err := c.doOnce(); err == nil {
			fmt.Println("success and exit") // nolint:forbidigo
			os.Exit(0)
		} else {
			fmt.Printf("error: %v\n", err) // nolint:forbidigo
		}

		time.Sleep(CheckInterval * time.Second)
	}
	fmt.Printf("failed after retry %d times\n", RetryTimes) // nolint:forbidigo
	os.Exit(ExitCode)
}

// 对依赖服务进行一次检查，任意服务不可用，都返回错误
func (c *DependencyServiceChecker) doOnce() error {
	// 检查 Redis 服务，若服务异常，则返回错误
	rds := redis.NewStandaloneClient(&c.conf.Redis)
	if _, err := rds.Ping(context.TODO()).Result(); err != nil {
		return err
	}

	// 检查 ClusterManager 服务，若服务未注册，则返回错误
	if _, err := cluster.NewCMClient(c.microRtr, c.cliTLSConf); err != nil {
		return err
	}

	// 检查 BcsProject 服务，若服务未注册，则返回错误
	if _, err := project.NewProjClient(c.microRtr, c.cliTLSConf); err != nil {
		return err
	}
	return nil
}
