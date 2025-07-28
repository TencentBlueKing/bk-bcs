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

package endpoint

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corev1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// ErrK8sConfigNotInited show K8sConfig not inited
	ErrK8sConfigNotInited = errors.New("k8sConfig not inited")
)

// K8sConfig xxx
type K8sConfig struct {
	Mater      string `json:"master"`
	KubeConfig string `json:"kubeConfig"`
}

func (c *K8sConfig) getRestConfig() (*rest.Config, error) {
	if c == nil {
		return nil, ErrK8sConfigNotInited
	}

	config, err := clientcmd.BuildConfigFromFlags(c.Mater, c.KubeConfig)
	if err != nil {
		blog.Errorf("BuildConfigFromFlags failed: %v", err)
		return nil, err
	}

	// config client qps/burst
	config.QPS = 1e6
	config.Burst = 1e6

	return config, nil
}

// GetKubernetesClient init kubernetes clientSet by k8sConfig
func (c *K8sConfig) GetKubernetesClient() (kubernetes.Interface, error) {
	if c == nil {
		return nil, ErrK8sConfigNotInited
	}

	config, err := c.getRestConfig()
	if err != nil {
		blog.Errorf("GetKubernetesClient call getRestConfig return err: %v", err)
		return nil, err
	}
	blog.Infof("GetKubernetesClient call getRestConfig successful")

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		blog.Errorf("GetKubernetesClient call NewForConfig failed: %v", err)
		return nil, err
	}

	return clientset, nil
}

// GetNodeLister init kubernetes node lister by k8sConfig
func (c *K8sConfig) GetNodeLister() (corev1lister.NodeLister, corev1lister.NodeLister, error) {

	// 因为selector不支持“或”逻辑，所以这里分别初始化master和controlPlane的NodeLister
	masterNodeLister, err := c.getNodeLister(masterLabel)
	if err != nil {
		return nil, nil, err
	}

	cpNodeLister, err := c.getNodeLister(controlPlaneLabel)
	if err != nil {
		return nil, nil, err
	}

	return masterNodeLister, cpNodeLister, nil
}

func (c *K8sConfig) getNodeLister(label string) (corev1lister.NodeLister, error) {
	clientset, err := c.GetKubernetesClient()
	if err != nil {
		blog.Errorf("GetNodeLister call GetKubernetesClient failed: %v", err)
		return nil, err
	}

	tweakFunc := func(opts *metav1.ListOptions) {
		opts.LabelSelector = label // 只监听带此标签的 Node
	}

	factory := informers.NewFilteredSharedInformerFactory(clientset, 10*time.Hour, "", tweakFunc)
	nodeInformer := factory.Core().V1().Nodes()
	nodeLister := nodeInformer.Lister()

	// 5. 启动 Informer 并等待缓存同步
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), nodeInformer.Informer().HasSynced) {
		blog.Errorf("wait for cache sync failed for node selector %s", label)
		return nil, fmt.Errorf("wait for cache sync failed for node selector %s", label)
	}
	blog.Infof("wait for cache sync successful for node selector %s", label)

	return nodeLister, nil
}
