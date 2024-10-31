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

package task

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

var (
	// 当前step执行数量
	stepRunningCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "step_running_count",
		Help: "The number of running step.",
	}, []string{"task_type", "task_name", "executor"})

	// step执行总数
	stepExecuteTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "step_execute_total",
		Help: "Counter of step execute count.",
	}, []string{"task_type", "task_name", "executor", "status"})

	// step执行耗时
	stepExecuteDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "step_execute_duration_seconds",
		Help:    "Histogram of duration for step execute.",
		Buckets: []float64{1, 10, 30, 60, 60 * 5, 60 * 10, 60 * 30, 3600, 3600 * 2},
	}, []string{"task_type", "task_name", "executor", "status"})

	// task执行总数
	taskExecuteTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "task_execute_total",
		Help: "Counter of task execute.",
	}, []string{"task_type", "task_name", "status"})

	// task执行耗时
	taskExecuteDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "task_execute_duration_seconds",
		Help:    "Histogram of duration for task execute.",
		Buckets: []float64{1, 10, 30, 60, 60 * 5, 60 * 10, 60 * 30, 3600, 3600 * 2, 3600 * 4, 3600 * 8},
	}, []string{"task_type", "task_name", "status"})
)

func init() {
	prometheus.MustRegister(stepRunningCount)
	prometheus.MustRegister(stepExecuteTotal)
	prometheus.MustRegister(stepExecuteDuration)
	prometheus.MustRegister(taskExecuteTotal)
	prometheus.MustRegister(taskExecuteDuration)
}

// collectMetricStart metrics for task start
func collectMetricStart(state *State) {
	stepRunningCount.WithLabelValues(
		state.task.GetTaskType(),
		state.task.GetTaskName(),
		state.step.Executor).Inc()
}

// collectMetricEnd metrics for task end
func collectMetricEnd(state *State) {
	// 任务状态完成时, 记录执行结果
	if state.task.GetStatus() != types.TaskStatusInit && state.task.GetStatus() != types.TaskStatusRunning {
		taskExecuteTotal.WithLabelValues(
			state.task.GetTaskType(),
			state.task.GetTaskName(),
			state.task.GetStatus()).Inc()

		taskExecuteDuration.WithLabelValues(
			state.task.GetTaskType(),
			state.task.GetTaskName(),
			state.task.GetStatus()).Observe(state.task.GetExecutionTime().Seconds())
	}

	// 任务步骤完成时, 记录执行结果
	stepRunningCount.WithLabelValues(
		state.task.GetTaskType(),
		state.task.GetTaskName(),
		state.step.Executor).Dec()

	stepExecuteTotal.WithLabelValues(
		state.task.GetTaskType(),
		state.task.GetTaskName(),
		state.step.Executor,
		state.step.GetStatus()).Inc()

	stepExecuteDuration.WithLabelValues(
		state.task.GetTaskType(),
		state.task.GetTaskName(),
		state.step.Executor,
		state.step.GetStatus()).Observe(state.step.GetExecutionTime().Seconds())
}
