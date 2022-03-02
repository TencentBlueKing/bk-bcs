/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8sclient

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	bcsAPIGWK8SBaseURI = "%s/clusters/%s/"
)

// Config describe the configuration for k8s client group
type Config struct {
	BcsAPI string
	Token  string
}

// NewGroup return a new Group instance
func NewGroup(c Config) Group {
	return &k8sClientGroup{
		config: &c,
		groups: make(map[string]*k8sClient),
		locks:  make(map[string]*sync.Mutex),
	}
}

// Group 定义了一组k8s client, 用来操作不同的集群
type Group interface {

	// Cluster 通过指定集群 clusterID 可以拿到一个 kubernetes.Interface 对象
	Cluster(clusterID string) (kubernetes.Interface, error)
}

type k8sClientGroup struct {
	config *Config

	sync.RWMutex
	groups map[string]*k8sClient
	locks  map[string]*sync.Mutex
}

// Cluster 从缓存中拿对应 clusterID 的client set, 如果不存在则新建一个
func (kg *k8sClientGroup) Cluster(clusterID string) (kubernetes.Interface, error) {
	if cs := kg.getClientSet(clusterID); cs != nil {
		return cs.clientSet, nil
	}

	clusterLock := kg.getClusterLock(clusterID)
	clusterLock.Lock()
	defer clusterLock.Unlock()

	if cs := kg.getClientSet(clusterID); cs != nil {
		return cs.clientSet, nil
	}

	cs, err := kg.generateClientSet(clusterID)
	if err != nil {
		return nil, err
	}
	return cs.clientSet, nil
}

func (kg *k8sClientGroup) getClientSet(clusterID string) *k8sClient {
	kg.RLock()
	defer kg.RUnlock()

	cs, ok := kg.groups[clusterID]
	if ok {
		return cs
	}

	return nil
}

func (kg *k8sClientGroup) getClusterLock(clusterID string) *sync.Mutex {
	kg.Lock()
	defer kg.Unlock()

	if _, ok := kg.locks[clusterID]; !ok {
		kg.locks[clusterID] = new(sync.Mutex)
	}

	return kg.locks[clusterID]
}

func (kg *k8sClientGroup) generateClientSet(clusterID string) (*k8sClient, error) {
	host := kg.getHost(clusterID)
	blog.Infof("generate new k8s client for cluster %s to %s", clusterID, host)

	c := &rest.Config{
		Host:        host,
		BearerToken: kg.config.Token,
		QPS:         1e6,
		Burst:       1e6,
		Transport: &http.Transport{
			TLSHandshakeTimeout: 5 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ResponseHeaderTimeout: 30 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}

	clientSet, err := kubernetes.NewForConfig(c)
	if err != nil {
		return nil, err
	}

	cs := &k8sClient{
		clientSet: clientSet,
	}

	kg.Lock()
	kg.groups[clusterID] = cs
	kg.Unlock()
	return cs, nil
}

func (kg *k8sClientGroup) getHost(clusterID string) string {
	return fmt.Sprintf(bcsAPIGWK8SBaseURI, kg.config.BcsAPI, clusterID)
}

type k8sClient struct {
	clientSet *kubernetes.Clientset
}
