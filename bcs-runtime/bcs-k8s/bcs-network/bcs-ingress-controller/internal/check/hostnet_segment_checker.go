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

package check

import (
	"context"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/hostnetportpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// HostNetSegmentChecker periodically scans for leaked port segments whose pods no longer exist.
type HostNetSegmentChecker struct {
	cli   client.Client
	cache *hostnetportpoolcache.HostNetPortPoolCache
}

// NewHostNetSegmentChecker creates a new segment leak checker.
func NewHostNetSegmentChecker(cli client.Client,
	cache *hostnetportpoolcache.HostNetPortPoolCache) *HostNetSegmentChecker {
	return &HostNetSegmentChecker{cli: cli, cache: cache}
}

// Run implements Checker interface.
func (c *HostNetSegmentChecker) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	st := time.Now()
	defer func() { blog.Infof("hostnet segment leak check cost: %f s", time.Since(st).Seconds()) }()

	allocated := c.cache.GetAllocatedSegments()
	for _, seg := range allocated {
		parts := strings.SplitN(seg.PodKey, "/", 2)
		if len(parts) != 2 {
			continue
		}
		podNs, podName := parts[0], parts[1]

		pod := &k8scorev1.Pod{}
		err := c.cli.Get(ctx, types.NamespacedName{
			Namespace: podNs, Name: podName,
		}, pod)

		var reason string
		if err != nil {
			if k8serrors.IsNotFound(err) {
				reason = "pod_not_found"
			} else {
				blog.Warnf("hostnet segment checker: get pod %s failed: %v", seg.PodKey, err)
				continue
			}
		} else if pod.Status.Phase == k8scorev1.PodFailed || pod.Status.Phase == k8scorev1.PodSucceeded {
			reason = "pod_terminated"
		}

		if reason != "" {
			c.cache.ReleaseByPodKey(seg.PodKey)
			blog.Warnf("hostnet segment checker: released leaked segment node=%s ports=%d-%d pod=%s pool=%s reason=%s",
				seg.NodeName, seg.StartPort, seg.EndPort, seg.PodKey, seg.PoolKey, reason)
			metrics.IncreaseHostNetSegmentLeakReleased(seg.PoolKey, seg.NodeName, reason)
		}
	}
}
