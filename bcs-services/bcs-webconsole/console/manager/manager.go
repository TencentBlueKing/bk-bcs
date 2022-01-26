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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	"github.com/go-redis/redis/v8"
	microconf "go-micro.dev/v4/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type manager struct {
	sync.RWMutex
	conf                *config.ConsoleConfig
	k8sClient           *kubernetes.Clientset
	k8sConfig           *rest.Config
	redisClient         *redis.Client // redis 客户端
	connectedContainers map[string]bool
	PodMap              map[string]types.UserPodData
	Config              microconf.Config
}

// NewManager create a Manager object
func NewManager(conf *config.ConsoleConfig, k8sClient *kubernetes.Clientset, k8sConfig *rest.Config,
	redisClient *redis.Client, mc microconf.Config) Manager {
	return &manager{
		conf:                conf,
		k8sClient:           k8sClient,
		k8sConfig:           k8sConfig,
		redisClient:         redisClient,
		connectedContainers: make(map[string]bool),
		PodMap:              make(map[string]types.UserPodData),
		Config:              mc,
	}
}
