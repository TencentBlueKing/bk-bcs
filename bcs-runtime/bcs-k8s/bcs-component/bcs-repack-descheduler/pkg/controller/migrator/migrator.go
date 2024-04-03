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

package migrator

import (
	"context"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/uuid"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/migrator/eviction"
)

var (
	defaultMigrateTimeout = 30
)

func (m *descheduleMigratorManager) Migrate() {
	m.Lock()
	if m.migrateJob == nil || !m.migrateJob.prevJobStopped.Load() || m.migrateJob.stopped.Load() {
		return
	}
	m.migrating.Store(true)
	m.workloadPlansMap = &sync.Map{}
	m.Unlock()
	defer m.migrating.Store(false)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultMigrateTimeout)*time.Minute)
	defer cancel()
	for ns := range m.plans {
		allWorkloadPlans := m.filterWorkloadPlans(ctx, ns)
		for wkName, wkPlans := range allWorkloadPlans {
			m.workloadPlansMap.Store(wkName, wkPlans)
		}
	}
	runner := &migratorRunner{
		op:           options.GlobalConfigHandler().GetOptions(),
		plans:        m.workloadPlansMap,
		evictManager: eviction.NewEvictManager(),
	}
	runner.Run(ctx)
}

type WorkloadPlans []*calculator.ResponseMigratePlan

func (m *descheduleMigratorManager) filterWorkloadPlans(ctx context.Context, ns string) map[string]WorkloadPlans {
	nsPlans := m.plans[ns]
	wkPlans := make(map[string]WorkloadPlans)
	for _, plan := range nsPlans {
		namespace, name, err := apis.PodNameSplit(plan.Item)
		if err != nil {
			blog.Warnf("migrate plan '%s' split failed: %s", plan.Item, err.Error())
			continue
		}
		if namespace != ns {
			blog.Warnf("migrate plan '%s' not belong to namespace '%s'", plan.Item, ns)
			continue
		}
		// 如果 Pod 没有 Owner 则不能驱逐，防止驱逐造成无父 Pod 被删除
		pod, podOwnerName, err := m.getPodOwnerName(ctx, namespace, name)
		if err != nil {
			blog.Warnf("Migrator namespace '%s' pod name '%s' get owner name failed", ns, plan.Item)
			continue
		}
		if pod.Status.HostIP != plan.From {
			blog.Warnf("migrator namespace '%s' pod '%s' hostIP '%s' not same as plan's hostIP '%s'",
				ns, plan.Item, pod.Status.HostIP, plan.From)
			continue
		}
		wkPlans[podOwnerName] = append(wkPlans[podOwnerName], plan)
	}
	return wkPlans
}

type migratorRunner struct {
	op           *options.DeSchedulerOption
	plans        *sync.Map
	nodePlans    *sync.Map
	evictManager eviction.EvictInterface
}

func (r *migratorRunner) Run(ctx context.Context) {
	// DOTO: 需要划分下迁移节点的优先级
	r.nodePlans = r.filterMigrateNode()
	wg := &sync.WaitGroup{}
	wg.Add(int(r.op.MaxEvictionNodes))
	nodeCh := make(chan string)
	for i := 0; i < int(r.op.MaxEvictionNodes); i++ {
		go r.run(ctx, wg, nodeCh)
	}
	r.nodePlans.Range(func(key, value any) bool {
		node := key.(string)
		nodeCh <- node
		return true
	})
	close(nodeCh)
	wg.Wait()
}

func (r *migratorRunner) run(ctx context.Context, wg *sync.WaitGroup, nodeCh chan string) {
	defer wg.Done()
	for {
		select {
		case node, ok := <-nodeCh:
			if !ok {
				return
			}
			v, ok := r.nodePlans.Load(node)
			if !ok {
				continue
			}
			plans := v.([]*calculator.ResponseMigratePlan)
			r.evictManager.EvictNode(ctx, &eviction.WorkloadEvictionMessage{
				TraceID: uuid.New().String(),
				Node:    node,
				Plans:   plans,
			})
		case <-ctx.Done():
			return
		}
	}
}

func (r *migratorRunner) filterMigrateNode() *sync.Map {
	result := &sync.Map{}
	r.plans.Range(func(key, value any) bool {
		wkPlans := value.(WorkloadPlans)
		for _, plan := range wkPlans {
			v, ok := result.Load(plan.From)
			if ok {
				v = append(v.([]*calculator.ResponseMigratePlan), plan)
				result.Store(plan.From, v)
			} else {
				result.Store(plan.From, []*calculator.ResponseMigratePlan{plan})
			}
		}
		return true
	})
	return result
}
