/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nodenetwork

import (
	"fmt"
	"time"

	"k8s.io/client-go/tools/clientcmd"

	"bk-bcs/bcs-common/common/blog"
	cloud "bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/apis/cloud/v1"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/informers"
	cloudinformer "bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/informers/cloud/v1"
	clientset "bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/internalclientset"
	cloudclient "bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/internalclientset/typed/cloud/v1"
	cloudlister "bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/client/lister/cloud/v1"
)

// Interface operation for node network struct
type Interface interface {
	Create(nodeNetwork *cloud.NodeNetwork) error
	Get(ns, name string) (*cloud.NodeNetwork, error)
}

// Client client for operation node network
type Client struct {
	kubeconfig           string
	kubeResyncPeriod     int
	kubeCacheSyncTimeout int
	client               cloudclient.CloudV1Interface
	lister               cloudlister.NodeNetworkLister
	informer             cloudinformer.NodeNetworkInformer
	stopCh               chan struct{}
}

// New create node network client
func New(kubeconfig string, kubeResyncPeriod, kubeCacheSyncTimeout int) *Client {
	return &Client{
		kubeconfig:           kubeconfig,
		kubeResyncPeriod:     kubeResyncPeriod,
		kubeCacheSyncTimeout: kubeCacheSyncTimeout,
		stopCh:               make(chan struct{}),
	}
}

// Init init etcd config
func (c *Client) Init() error {
	config, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		blog.Fatalf("build config from kubeconfig %s failed, err %s", c.kubeconfig, err.Error())
	}
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		blog.Fatalf("build clientset failed, err %s", err.Error())
	}
	factory := informers.NewSharedInformerFactory(clientset, time.Duration(c.kubeResyncPeriod)*time.Second)
	c.informer = factory.Cloud().V1().NodeNetworks()
	c.lister = factory.Cloud().V1().NodeNetworks().Lister()
	c.client = clientset.CloudV1()
	factory.Start(c.stopCh)

	// start informer factory and wait for cache sync, when timeout, return error
	syncFlag := make(chan struct{})
	go func() {
		blog.Infof("wait for informer factory cache sync")
		factory.WaitForCacheSync(c.stopCh)
		close(syncFlag)
	}()
	select {
	case <-time.After(time.Duration(c.kubeCacheSyncTimeout) * time.Second):
		return fmt.Errorf("wait for cache sync timeout after %d seconds", c.kubeCacheSyncTimeout)
	case <-syncFlag:
		break
	}
	blog.Infof("wait informer factory cache sync done")

	return nil
}

// Create create node network
func (c *Client) Create(n *cloud.NodeNetwork) error {
	_, err := c.client.NodeNetworks(n.GetNamespace()).Create(n)
	if err != nil {
		blog.Errorf("create node network %+v failed, err %s", err.Error())
		return err
	}
	return nil
}

// Get get node network
func (c *Client) Get(ns, name string) (*cloud.NodeNetwork, error) {
	return c.lister.NodeNetworks(ns).Get(name)
}
