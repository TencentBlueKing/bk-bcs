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

package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	comconf "github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"

	"github.com/go-redis/redis/v8"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ConsoleManager is an console struct
type ConsoleManager struct {
	// options for console manager
	opt *options.ConsoleManagerOption

	// tls config for cluster manager server
	tlsConfig *tls.Config

	k8sClient   *kubernetes.Clientset
	k8sConfig   *rest.Config
	redisClient *redis.Client // redis 客户端

	backend manager.Manager
	route   *api.Router
	conf    *config.ConsoleConfig
}

// NewConsoleManager create an ConsoleProxy object
func NewConsoleManager(opt *options.ConsoleManagerOption) *ConsoleManager {

	return &ConsoleManager{
		opt:     opt,
		backend: nil,
		route:   nil,
		conf:    nil,
	}
}

func (c *ConsoleManager) Init() error {

	err := setConfig(c.opt)
	if err != nil {
		return err
	}
	// init server and client tls config
	c.initTLSConfig()
	// init redis
	if err := c.initRedisCli(); err != nil {
		return err
	}
	// init k8s client
	if err := c.initK8sClient(); err != nil {
		return err
	}

	return nil
}

// Run create a pid
func (c *ConsoleManager) Run() error {

	//pid
	if err := common.SavePid(comconf.ProcessConfig{}); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	c.backend = manager.NewManager(&c.opt.Conf, c.k8sClient, c.k8sConfig, c.redisClient, nil)
	c.route = api.NewRouter(c.backend, &c.opt.Conf)
	stopCh := make(chan struct{})
	// 定期清理pod
	go wait.NonSlidingUntil(c.backend.CleanUserPod, manager.CleanUserPodInterval*time.Second, stopCh)

	return nil
}

func (c *ConsoleManager) initTLSConfig() {

	// server cert directoty
	if c.opt.CertConfig.ServerCertFile != "" && c.opt.CertConfig.CAFile != "" && c.opt.CertConfig.ServerKeyFile != "" {
		c.opt.Conf.ServCert.IsSSL = true
	}
}

func (c *ConsoleManager) initRedisCli() error {

	dbNum, err := strconv.Atoi(c.opt.Redis.Database)
	if nil != err {
		return err
	}
	if c.opt.Redis.PoolSize == 0 {
		c.opt.Redis.PoolSize = 3000
	}

	var client *redis.Client

	if c.opt.Redis.MasterName == "" {
		option := &redis.Options{
			Addr:     c.opt.Redis.Address,
			Password: c.opt.Redis.Password,
			DB:       dbNum,
			PoolSize: c.opt.Redis.PoolSize,
		}
		client = redis.NewClient(option)

	} else {
		hosts := strings.Split(c.opt.Redis.Address, ",")
		option := &redis.FailoverOptions{
			MasterName:       c.opt.Redis.MasterName,
			SentinelAddrs:    hosts,
			Password:         c.opt.Redis.Password,
			DB:               dbNum,
			PoolSize:         c.opt.Redis.PoolSize,
			SentinelPassword: c.opt.Redis.SentinelPassword,
		}
		client = redis.NewFailoverClient(option)
	}

	err = client.Ping(context.Background()).Err()
	if err != nil {
		return err
	}

	c.redisClient = client

	return nil
}

func (c *ConsoleManager) initK8sClient() error {
	// 配置 k8s 集群外 kubeconfig 配置文件
	if home := homeDir(); home != "" {
		c.opt.KubeConfigFile = filepath.Join(home, ".kube", "config")
	}

	//在 kubeconfig 中使用当前上下文环境，config 获取支持 url 和 path 方式
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", c.opt.KubeConfigFile)
		if err != nil {
			return err
		}
	}

	//在 kubeconfig 中使用当前上下文环境，config 获取支持 url 和 path 方式
	config, err := clientcmd.BuildConfigFromFlags("", c.opt.KubeConfigFile)
	if err != nil {
		return err
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.k8sConfig = k8sConfig
	c.k8sClient = k8sClient

	return nil
}

func setConfig(op *options.ConsoleManagerOption) error {
	op.Redis.Address = op.RedisAddress
	op.Redis.Password = op.RedisPassword
	op.Redis.Database = op.RedisDatabase
	op.Redis.MasterName = op.RedisMasterName
	op.Redis.SentinelPassword = op.RedisSentinelPassword
	op.Redis.PoolSize = op.RedisPoolSize

	op.Conf.Address = op.Address
	op.Conf.Port = int(op.Port)
	if op.WebConsoleImage == "" {
		return fmt.Errorf("web-console-image required")
	}
	op.Conf.WebConsoleImage = op.WebConsoleImage
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
