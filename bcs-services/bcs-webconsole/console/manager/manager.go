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

package manager

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	"github.com/go-redis/redis/v7"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type manager struct {
	sync.RWMutex
	conf                *config.ConsoleConfig
	k8sClient           *kubernetes.Clientset
	k8sConfig           *rest.Config
	redisClient         *redis.Client // redis 客户端
	connectedContainers map[string]bool
	PodMap              map[string]types.UserPodData
}

// NewManager create a Manager object
func NewManager(conf *config.ConsoleConfig) Manager {
	return &manager{
		conf:                conf,
		connectedContainers: make(map[string]bool),
	}
}

func (m *manager) Init() error {
	var err error

	// 初始化 k8s client
	err = m.newK8sClient()
	if err != nil {
		return err
	}

	err = m.newRedisCli()
	if err != nil {
		return err
	}

	m.PodMap = make(map[string]types.UserPodData)

	return nil

}

func (m *manager) newK8sClient() error {
	// 配置 k8s 集群外 kubeconfig 配置文件
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	//在 kubeconfig 中使用当前上下文环境，config 获取支持 url 和 path 方式
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return err
		}
	}

	//在 kubeconfig 中使用当前上下文环境，config 获取支持 url 和 path 方式
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	m.k8sConfig = k8sConfig
	m.k8sClient = k8sClient

	stopChan := make(chan struct{})
	go m.startCleanUserPod(stopChan)

	return nil
}

// newRedisCli returns a client to the Redis Server specified by Options
func (m *manager) newRedisCli() error {
	dbNum, err := strconv.Atoi(m.conf.RedisConfig.Database)
	if nil != err {
		return err
	}
	if m.conf.RedisConfig.MaxOpenConns == 0 {
		m.conf.RedisConfig.MaxOpenConns = 3000
	}

	client := new(redis.Client)

	if m.conf.RedisConfig.MasterName == "" {
		option := &redis.Options{
			Addr:     m.conf.RedisConfig.Address,
			Password: m.conf.RedisConfig.Password,
			DB:       dbNum,
			PoolSize: m.conf.RedisConfig.MaxOpenConns,
		}
		client = redis.NewClient(option)

	} else {
		hosts := strings.Split(m.conf.RedisConfig.Address, ",")
		option := &redis.FailoverOptions{
			MasterName:       m.conf.RedisConfig.MasterName,
			SentinelAddrs:    hosts,
			Password:         m.conf.RedisConfig.Password,
			DB:               dbNum,
			PoolSize:         m.conf.RedisConfig.MaxOpenConns,
			SentinelPassword: m.conf.RedisConfig.SentinelPassword,
		}
		client = redis.NewFailoverClient(option)
	}

	err = client.Ping().Err()
	if err != nil {
		return err
	}

	m.redisClient = client
	return nil
}

// startCleanUserPod 清理pod
func (m *manager) startCleanUserPod(stopCh <-chan struct{}) {
	go wait.NonSlidingUntil(m.cleanUserPod, CleanUserPodInterval, stopCh)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
