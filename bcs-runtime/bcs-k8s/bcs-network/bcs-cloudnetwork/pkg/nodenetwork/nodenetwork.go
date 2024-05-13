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

package nodenetwork

import (
	"context"
	"fmt"
	"time"

	coreapiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8score "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cloud "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
	clientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned"
	cloudclient "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	informers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/informers/externalversions"
	cloudinformer "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/informers/externalversions/cloud/v1"
	cloudlister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/listers/cloud/v1"
	"k8s.io/client-go/tools/cache"
)

// Interface operation for node network struct
type Interface interface {
	Register(handler cache.ResourceEventHandler)
	Run() error
	Create(nodeNetwork *cloud.NodeNetwork) error
	Get(ns, name string) (*cloud.NodeNetwork, error)
}

// Client client for operation node network
type Client struct {
	kubeconfig           string
	kubeResyncPeriod     int
	kubeCacheSyncTimeout int
	factory              informers.SharedInformerFactory
	client               cloudclient.CloudV1Interface
	lister               cloudlister.NodeNetworkLister
	informer             cloudinformer.NodeNetworkInformer
	nsClient             k8score.NamespaceInterface
	handler              cache.ResourceEventHandler
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

// Register register event handler
func (c *Client) Register(handler cache.ResourceEventHandler) {
	c.handler = handler
	c.informer.Informer().AddEventHandlerWithResyncPeriod(c.handler, time.Duration(c.kubeCacheSyncTimeout)*time.Second)
}

// Init init etcd config
func (c *Client) Init() error {
	config, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		blog.Errorf("build config from kubeconfig %s failed, err %s", c.kubeconfig, err.Error())
		return err
	}
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		blog.Errorf("build clientset failed, err %s", err.Error())
		return err
	}
	c.factory = informers.NewSharedInformerFactory(clientset, time.Duration(c.kubeResyncPeriod)*time.Second)
	c.informer = c.factory.Cloud().V1().NodeNetworks()
	c.lister = c.factory.Cloud().V1().NodeNetworks().Lister()
	c.client = clientset.CloudV1()

	k8scorecliset, err := kubernetes.NewForConfig(config)
	if err != nil {
		blog.Errorf("build k8s core clientset failed, err %s", err.Error())
		return err
	}
	c.nsClient = k8scorecliset.CoreV1().Namespaces()

	return nil
}

// Run run the client
func (c *Client) Run() error {

	c.factory.Start(c.stopCh)

	// start informer factory and wait for cache sync, when timeout, return error
	syncFlag := make(chan struct{})
	go func() {
		blog.Infof("wait for informer factory cache sync")
		c.factory.WaitForCacheSync(c.stopCh)
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

func (c *Client) ensureNamespace(ns string) error {
	_, err := c.nsClient.Get(context.TODO(), ns, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			blog.Errorf("get ns %s failed, err %s", err.Error())
			return err
		}
		newNs := &coreapiv1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Namespace",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}
		_, err := c.nsClient.Create(context.TODO(), newNs, metav1.CreateOptions{})
		if err != nil {
			blog.Errorf("create namespace %+v failed, err %s", newNs, err.Error())
			return err
		}
	}
	return nil
}

// Create create node network
func (c *Client) Create(n *cloud.NodeNetwork) error {
	if err := c.ensureNamespace(n.GetNamespace()); err != nil {
		return err
	}
	// clean resource version when create
	n.ResourceVersion = ""
	_, err := c.client.NodeNetworks(n.GetNamespace()).Create(context.TODO(), n, metav1.CreateOptions{})
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
