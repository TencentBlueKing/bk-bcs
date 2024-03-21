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
	"os"
	"path"
	"sync"
	"sync/atomic"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/cachemanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator/local"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator/remote"
)

// DescheduleMigratorInterface defines the interface of migrator,
type DescheduleMigratorInterface interface {
	// SendCalculateJob 发送 Policy 计算任务
	SendCalculateJob(policy *v1alpha1.DeschedulePolicy)

	// FilterNodes 用于计算重调度的 Pod 需要落在哪个节点
	FilterNodes(ctx context.Context, pod *corev1.Pod) (*corev1.NodeList, error)

	CreateMigrateJob(policy *v1alpha1.DeschedulePolicy) error

	DeleteMigrateJob(policy *v1alpha1.DeschedulePolicy)
}

// descheduleMigratorManager the manager of migrate. It will evict
// pods with MigrateMessage and control the process of it.
type descheduleMigratorManager struct {
	sync.RWMutex

	op *options.DeSchedulerOption

	// plans 存储计算到的迁移计划
	plans map[string][]*calculator.ResponseMigratePlan

	// workloadPlansMap 存储 Workload 级别的重调度结果
	workloadPlansMap *sync.Map

	migrateJob *cronJobInstance
	migrating  atomic.Bool

	cacheManager     cachemanager.CacheInterface
	calculatorRemote calculator.CalculateInterface
	calculatorLocal  calculator.CalculateInterface
}

var (
	once sync.Once

	migratorManager *descheduleMigratorManager
)

// GlobalMigratorManager create the instance of migrator manager
func GlobalMigratorManager() DescheduleMigratorInterface {
	once.Do(func() {
		migratorManager = &descheduleMigratorManager{
			op:               options.GlobalConfigHandler().GetOptions(),
			workloadPlansMap: &sync.Map{},
			cacheManager:     cachemanager.NewCacheManager(),
			calculatorLocal:  local.NewCalculatorLocal(),
			calculatorRemote: remote.NewCalculatorRemote(options.GlobalConfigHandler().GetOptions()),
		}
		migratorManager.migrating.Store(false)
		if err := os.MkdirAll(path.Join(migratorManager.op.LogDir, "calculator"), 0655); err != nil {
			blog.Errorf("create calculator result dir failed: %s", err.Error())
		}
	})
	return migratorManager
}

func (m *descheduleMigratorManager) getPodOwnerName(ctx context.Context, namespace,
	podName string) (*corev1.Pod, string, error) {
	podCtx, podCancel := context.WithTimeout(ctx, apis.DefaultQueryTimeout)
	defer podCancel()
	return m.cacheManager.GetPodOwnerName(podCtx, namespace, podName)
}
