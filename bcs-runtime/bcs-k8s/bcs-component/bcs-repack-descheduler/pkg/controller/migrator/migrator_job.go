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
	"sync/atomic"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis/tkex/v1alpha1"
)

// CreateMigrateJob create the migrate job by descheduler policy
func (m *descheduleMigratorManager) CreateMigrateJob(policy *v1alpha1.DeschedulePolicy) error {
	m.Lock()
	defer m.Unlock()
	policyNN := apis.NamespacedNamePolicy(policy)
	if m.migrateJob != nil && m.migrateJob.stopped.Load() {
		m.migrateJob = nil
	}
	timeRange := policy.Spec.Converge.TimeRange
	if m.migrateJob == nil {
		m.migrateJob = &cronJobInstance{
			timeRange: timeRange,
		}
		m.migrateJob.prevJobStopped.Store(true)
		m.migrateJob.stopped.Store(false)
		if err := m.createCronJob(policyNN); err != nil {
			return errors.Wrapf(err, "create cron job failed")
		}
		blog.Infof("migrate job '%s' create success", policyNN)
		return nil
	}
	if m.migrateJob.timeRange == timeRange {
		blog.Infof("migrate job '%s' exist and not changed", policyNN)
		return nil
	}
	blog.Infof("migrate job '%s' timerange changed: '%s' -> '%s'", policyNN, m.migrateJob.timeRange, timeRange)
	m.migrateJob.prevJobStopped.Store(false)
	ctx := m.migrateJob.cronJob.Stop()
	go func(ctx context.Context) {
		<-ctx.Done()
		m.migrateJob.prevJobStopped.Store(true)
	}(ctx)
	m.migrateJob.timeRange = timeRange
	if err := m.createCronJob(policyNN); err != nil {
		return errors.Wrapf(err, "create cron job failed when update")
	}
	return nil
}

func (m *descheduleMigratorManager) createCronJob(policy string) error {
	m.migrateJob.cronJob = cron.New()
	schedule, err := cron.ParseStandard(m.migrateJob.timeRange)
	if err != nil {
		return errors.Wrapf(err, "cron check schedule time '%s' failed", m.migrateJob.timeRange)
	}
	time1 := schedule.Next(time.Now())
	time2 := schedule.Next(time1)
	time3 := schedule.Next(time2)
	blog.Infof("migrate cron job '%s' is started. next three times: [%v][%v][%v]",
		policy, time1, time2, time3)
	_, _ = m.migrateJob.cronJob.AddFunc(m.migrateJob.timeRange, func() {
		m.Migrate()
	})
	m.migrateJob.cronJob.Start()
	return nil
}

// DeleteMigrateJob delete the migrate job by descheduler policy
func (m *descheduleMigratorManager) DeleteMigrateJob(policy *v1alpha1.DeschedulePolicy) {
	m.Lock()
	defer m.Unlock()
	if m.migrateJob == nil {
		return
	}
	ctx := m.migrateJob.cronJob.Stop()
	go func(ctx context.Context) {
		<-ctx.Done()
		blog.Infof("migrate job '%s' stopped", apis.NamespacedNamePolicy(policy))
		m.migrateJob.stopped.Store(true)
	}(ctx)
}

type cronJobInstance struct {
	cronJob   *cron.Cron
	timeRange string

	prevJobStopped atomic.Bool
	stopped        atomic.Bool
}
