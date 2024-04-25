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

// Package eviction xx
package eviction

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/cachemanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
)

// WorkloadEvictionMessage defines the message that should be evicted. One message to
// one workload eviction.
type WorkloadEvictionMessage struct {
	TraceID string
	Node    string
	Plans   []*calculator.ResponseMigratePlan
}

// EvictInterface defines the interface of evict.
type EvictInterface interface {
	EvictNode(ctx context.Context, message *WorkloadEvictionMessage)
}

// EvictManager manages all the tasks of eviction. EvictManager will evict workload with WorkloadEvictionMessage.
// Every message contains migrate plans, manager evict workload with them.
type EvictManager struct {
	sync.Mutex

	op             *options.DeSchedulerOption
	cacheManager   cachemanager.CacheInterface
	currentWorkNum int32
}

// NewEvictManager create the instance of EvictManager.
func NewEvictManager() EvictInterface {
	return &EvictManager{
		op:             options.GlobalConfigHandler().GetOptions(),
		cacheManager:   cachemanager.NewCacheManager(),
		currentWorkNum: 0,
	}
}

// EvictNode 驱逐节点上的 Pod 实例
func (m *EvictManager) EvictNode(ctx context.Context, message *WorkloadEvictionMessage) {
	blog.Infof("eviction node '%s' is started, trace=%s", message.Node, message.TraceID)
	if err := m.cacheManager.CordonNode(ctx, message.Node); err != nil {
		blog.Errorf("eviction node '%s' cordon failed: %s, trace=%s", message.Node, err.Error(), message.TraceID)
		return
	}
	wg := &sync.WaitGroup{}
	wg.Add(len(message.Plans))
	for _, plan := range message.Plans {
		go func(plan *calculator.ResponseMigratePlan) {
			defer wg.Done()
			if err := m.EvictPlan(ctx, message.TraceID, plan); err != nil {
				blog.Errorf("eviction migrate plan failed: %s, trace=%s", err.Error(), message.TraceID)
			}
			blog.Infof("eviction plan '%s' completed, trace=%s", plan.Item, message.TraceID)
		}(plan)
	}
	wg.Wait()
}

var (
	evictRetryDuration = time.Duration(5) * time.Second
)

// EvictPlan 根据计算出的迁移计划，将对应的 Pod 进行驱逐，并同步等待 Pod 删除成功
func (m *EvictManager) EvictPlan(ctx context.Context, traceID string, plan *calculator.ResponseMigratePlan) error {
	m.lock()
	defer m.unlock()
	blog.Infof("eviction for item '%s' is handling, trace=%s", plan.Item, traceID)
	podNamespace, podName, err := apis.PodNameSplit(plan.Item)
	if err != nil {
		return errors.Wrapf(err, "item '%s' split failed", plan.Item)
	}

	deleted := make(chan struct{})
	registerId := traceID + plan.Item
	m.cacheManager.RegisterPodDeleteEvents(registerId, func(pod *corev1.Pod) {
		if plan.Item == apis.PodName(pod.Namespace, pod.Name) {
			blog.Infof("Eviction item '%s' delete success, trace=%s", plan.Item, traceID)
			deleted <- struct{}{}
		}
	})
	defer m.cacheManager.UnRegisterPodDeleteEvents(registerId)

	ticker := time.NewTicker(evictRetryDuration)
	defer ticker.Stop()
L:
	for {
		select {
		case <-ticker.C:
			if err = m.cacheManager.EvictionPod(ctx, podNamespace, podName); err == nil {
				break L
			}
			if !apis.IsPDBError(err) {
				return errors.Wrapf(err, "item '%s' evict failed", plan.Item)
			}
			blog.Infof("eviction item '%s' triggerred pod's disruption budget, will try after %s",
				plan.Item, evictRetryDuration.String())
		case <-ctx.Done():
			blog.Warn("eviction item '%s' is canceled when evict, trace=%s", plan.Item, traceID)
			return nil
		}
	}
	blog.Infof("eviction item '%s' call success, wait it deleted, trace=%s", plan.Item, traceID)
	select {
	case <-deleted:
		blog.Infof("eviction item '%s' is deleted, trace=%s", plan.Item, traceID)
	case <-ctx.Done():
		blog.Warnf("eviction item '%s' is canceled when wait delete, trace=%s", plan.Item, traceID)
	}
	return nil
}

// lock will lock the eviction process when currentWorkNum < MaxEvictionParallel, there will
// set the max parallel num of eviction. If the working eviction reach the threshold, new eviction
// should wait for others completed.
func (m *EvictManager) lock() {
	m.Lock()
	defer m.Unlock()

	if m.currentWorkNum < m.op.MaxEvictionParallel {
		atomic.AddInt32(&m.currentWorkNum, 1)
		return
	}
	for m.currentWorkNum >= m.op.MaxEvictionParallel {
		time.Sleep(1 * time.Second)
	}
}

func (m *EvictManager) unlock() {
	atomic.AddInt32(&m.currentWorkNum, -1)
}
