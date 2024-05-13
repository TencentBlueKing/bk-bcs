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

package leaderelection

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// Client client for leader election
type Client struct {
	ctx context.Context
	// lock type in kubernetes, available [resourcelock.EndpointsResourceLock, resourcelock.LeasesResourceLock ..... ]
	lockType      string
	name          string
	namespace     string
	leaseDuration time.Duration
	renewDuration time.Duration
	retryPeriod   time.Duration

	lock resourcelock.Interface
	el   *leaderelection.LeaderElector

	isMaster bool
}

// New create client
func New(lockType, name, ns, kubeconfig string,
	leaseDuration, renewDuration, retryPeriod time.Duration) (*Client, error) {

	var restConfig *rest.Config
	var err error

	cl := new(Client)
	cl.lockType = lockType
	cl.name = name
	cl.namespace = ns
	cl.leaseDuration = leaseDuration
	cl.renewDuration = renewDuration
	cl.retryPeriod = retryPeriod

	id, err := os.Hostname()
	if err != nil {
		blog.Errorf("get hostname failed, err %s", err.Error())
		return nil, err
	}

	// create kubernetes client for leader election
	if len(kubeconfig) != 0 {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			blog.Errorf("create internal client with kubeconfig %s failed, err %s", kubeconfig, err.Error())
			return nil, fmt.Errorf("create internal client with kubeconfig %s failed, err %s", kubeconfig, err.Error())
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			blog.Errorf("build incluster config failed, err %s", err.Error())
			return nil, fmt.Errorf("buidl incluster config failed, err %s", err.Error())
		}
	}
	k8sClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("create client set from config failed, err %s", err.Error())
		return nil, fmt.Errorf("create client set from config failed, err %s", err.Error())
	}

	id = id + "_" + string(uuid.NewUUID())

	rl, err := resourcelock.New(cl.lockType, cl.namespace, cl.name,
		k8sClientSet.CoreV1(), k8sClientSet.CoordinationV1(),
		resourcelock.ResourceLockConfig{
			Identity: id,
		})
	if err != nil {
		blog.Errorf("create resource lock failed, err %s", err.Error())
		return nil, err
	}
	cl.lock = rl

	el, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          rl,
		LeaseDuration: leaseDuration,
		RenewDeadline: renewDuration,
		RetryPeriod:   retryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: cl.onStartedLeading,
			OnStoppedLeading: cl.onReaquireLeading,
		},
	})
	if err != nil {
		blog.Errorf("create client-go leader elector failed, err %s", err.Error())
		return nil, fmt.Errorf("create client-go leader elector failed, err %s", err.Error())
	}
	cl.el = el

	return cl, nil
}

// Run run election
func (c *Client) Run(ctx context.Context) {
	c.ctx = ctx
	c.el.Run(ctx)
}

func (c *Client) onStartedLeading(ctx context.Context) {
	blog.Infof("become leader")
	c.isMaster = true
}

func (c *Client) onReaquireLeading() {
	blog.Infof("become follower")
	c.isMaster = false
	go c.el.Run(c.ctx)
}

// IsMaster to see if it is master
func (c *Client) IsMaster() bool {
	return c.isMaster
}
