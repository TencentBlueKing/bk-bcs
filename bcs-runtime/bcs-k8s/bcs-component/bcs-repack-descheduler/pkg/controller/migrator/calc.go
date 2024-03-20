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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
)

var (
	calcJobCh = make(chan struct{}, 1)
	calcJobs  = &sync.Map{}
)

// SendCalculateJob 发送 Policy 计算任务
func (m *descheduleMigratorManager) SendCalculateJob(policy *v1alpha1.DeschedulePolicy) {
	m.Lock()
	defer m.Unlock()
	if _, ok := calcJobs.Load(apis.NamespacedNamePolicy(policy)); !ok {
		calcJobs.Store(apis.NamespacedNamePolicy(policy), struct{}{})
		go m.calculateJobRun(policy.Namespace, policy.Name)
	} else {
		// 当前 channel 中仍然有任务，忽略本次计算
		if len(calcJobCh) == 1 {
			blog.Warnf("calculate job for '%s/%s' ignored, because still have job wait to be executed.",
				policy.Namespace, policy.Name)
			return
		}
	}
	calcJobCh <- struct{}{}
}

var (
	defaultCalculateDuration = 300
)

func (m *descheduleMigratorManager) calculateJobRun(namespace, name string) {
	defer func() {
		m.Lock()
		calcJobs.Delete(apis.NamespacedNameString(namespace, name))
		m.Unlock()
	}()
	blog.Infof("calculate job for '%s/%s' started", namespace, name)
	m.calculate(namespace, name)
	ticker := time.NewTicker(time.Duration(defaultCalculateDuration) * time.Second)
	defer ticker.Stop()
	// 根据收敛时间执行计算任务，避免大量计算任务产生堆积。将 600s 内的相同计算任务收敛
	for range ticker.C {
		<-calcJobCh
		m.calculate(namespace, name)
	}
}

func (m *descheduleMigratorManager) calculate(namespace, name string) {
	// 如果迁移任务正在执行，则不进行计算. 避免数据覆盖
	if m.migrating.Load() {
		blog.Warnf("calculate job for '%s/%s' cannot start, because migrate job is running", namespace, name)
		return
	} else {
		blog.Infof("calculate job for '%s/%s' is running", namespace, name)
	}
	resultPlan, err := m.calculateFromRemoteOrLocal()
	if err != nil {
		blog.Errorf("calculate job for '%s/%s' is run failed: %s", namespace, name, err.Error())
		return
	} else {
		blog.Infof("calculate job for '%s/%s' is run completed", namespace, name)
	}
	if len(resultPlan.Plans) == 0 {
		blog.Errorf("calculate job have not result plans")
		return
	}
	// DOTO: 暂时取第一个 plan
	migratePlans := resultPlan.Plans[0].MigratePlan
	bs, _ := json.Marshal(migratePlans)
	m.saveCalculateResult(bs)

	m.plans = make(map[string][]*calculator.ResponseMigratePlan)
	for i := range migratePlans {
		migratePlan := &migratePlans[i]
		migrateNS, _, splitErr := apis.PodNameSplit(migratePlan.Item)
		if splitErr != nil {
			blog.Errorf("pod name '%s' of plan split failed: %s", migratePlan.Item, splitErr.Error())
			continue
		}
		m.plans[migrateNS] = append(m.plans[migrateNS], migratePlan)
	}
}

func (m *descheduleMigratorManager) saveCalculateResult(bs []byte) {
	t := time.Now()
	name := fmt.Sprintf("%s-%d.txt", t.Format("2006-01-02T15:04:05"), t.UnixNano())
	saveFile := path.Join(migratorManager.op.LogDir, "calculator", name)
	if err := os.WriteFile(saveFile, bs, 0644); err != nil {
		blog.Errorf("calculate job save calculate result failed: %s", err.Error())
	} else {
		blog.Infof("calculate job save calculate result to '%s'", saveFile)
	}
}

var (
	defaultCalculateTimeout = 300
)

func (m *descheduleMigratorManager) calculateFromRemoteOrLocal() (
	plan *calculator.ResultPlan, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(defaultCalculateTimeout)*time.Second)
	defer cancel()
	req, err := m.cacheManager.BuildCalculatorRequest(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "build calculator request failed")
	}
	result, err := m.calculatorRemote.Calculate(ctx, req)
	if err == nil {
		return &result, nil
	}
	blog.Errorf("calculate from remote failed: %s", err.Error())
	result, err = m.calculatorLocal.Calculate(ctx, req)
	if err == nil {
		return &result, nil
	}
	return nil, errors.Wrapf(err, "calculate from local failed")
}
