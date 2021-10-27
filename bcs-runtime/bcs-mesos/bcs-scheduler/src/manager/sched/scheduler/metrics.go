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

package scheduler

import (
	"time"

	types "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	/*schedule taskgroup type*/
	LaunchTaskgroupType     = "launch"
	RescheduleTaskgroupType = "reschedule"
	ScaleTaskgroupType      = "scale"
	UpdateTaskgroupType     = "update"

	/*operate application type*/
	LaunchApplicationType        = "launch"
	DeleteApplicationType        = "delete"
	ScaleApplicationType         = "scale"
	UpdateApplicationType        = "update"
	RollingupdateApplicationType = "rollingupdate"
)

// Metrics the scheduler info
var (
	ScheduleTaskgroupTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "schedule_taskgroup_total",
		Help:      "Total counter of schedule taskgroup",
	}, []string{"namespace", "application", "taskgroup", "type"})

	ScheduleTaskgroupLatencyMs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "schedule_taskgroup_latency_ms",
		Help:      "Schedule taskgroup latency milliseconds",
	}, []string{"namespace", "application", "taskgroup", "type"})

	OperateAppTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "operate_application_total",
		Help:      "Total counter of operate application",
	}, []string{"namespace", "application", "type"})

	OperateAppLatencySecond = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "operate_application_latency_second",
		Help:      "Operate application latency seconds",
	}, []string{"namespace", "application", "type"})

	TaskgroupReportTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "taskgroup_report_total",
		Help:      "Total counter of report taskgroup status",
	}, []string{"namespace", "application", "taskgroup", "status"})

	TaskgroupOomTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: types.MetricsNamespaceScheduler,
		Subsystem: types.MetricsSubsystemScheduler,
		Name:      "application_oom_total",
		Help:      "Total counter of application oom killed",
	}, []string{"namespace", "application"})
)

func init() {
	prometheus.MustRegister(ScheduleTaskgroupTotal, ScheduleTaskgroupLatencyMs, OperateAppTotal, OperateAppLatencySecond,
		TaskgroupReportTotal, TaskgroupOomTotal)
}

func reportScheduleTaskgroupMetrics(ns, name, taskgroup, scheduleType string, started time.Time) {
	ScheduleTaskgroupTotal.WithLabelValues(ns, name, taskgroup, scheduleType).Inc()
	d := time.Duration(time.Since(started).Nanoseconds())
	sec := d / time.Millisecond
	nsec := d % time.Millisecond
	ms := float64(sec) + float64(nsec)/1e6
	ScheduleTaskgroupLatencyMs.WithLabelValues(ns, name, taskgroup, scheduleType).Observe(ms)
}

func reportOperateAppMetrics(ns, name, operateType string, started time.Time) {
	OperateAppTotal.WithLabelValues(ns, name, operateType).Inc()
	OperateAppLatencySecond.WithLabelValues(ns, name, operateType).Observe(time.Since(started).Seconds())
}

func reportTaskgroupReportMetrics(ns, name, taskgroup, status string) {
	TaskgroupReportTotal.WithLabelValues(ns, name, taskgroup, status).Inc()
}

func reportTaskgroupOomMetrics(ns, name, taskgroupId string) {
	TaskgroupOomTotal.WithLabelValues(ns, name, taskgroupId).Inc()
}
