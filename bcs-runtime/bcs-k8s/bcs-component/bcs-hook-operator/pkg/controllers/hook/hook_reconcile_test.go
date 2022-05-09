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

package hook

import (
	"errors"
	"fmt"
	hookmock "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/mock"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/providers"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/testutil"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"reflect"
	"testing"
	"time"
)

func TestGarbageCollectMeasurements(t *testing.T) {
	tests := []struct {
		name          string
		hr            *hookv1alpha1.HookRun
		limit         int
		expectedError error
	}{
		{
			name: "two metrics",
			hr: func() *hookv1alpha1.HookRun {
				hr := testutil.NewHookRun("hr1")
				hr.Spec.Metrics = append(hr.Spec.Metrics, hookv1alpha1.Metric{Name: "m1"})
				hr.Spec.Metrics = append(hr.Spec.Metrics, hookv1alpha1.Metric{Name: "m2"})
				hr.Status.MetricResults = append(hr.Status.MetricResults, hookv1alpha1.MetricResult{Name: "m1",
					Measurements: []hookv1alpha1.Measurement{{Value: "1"}, {Value: "2"}}})
				hr.Status.MetricResults = append(hr.Status.MetricResults, hookv1alpha1.MetricResult{Name: "m2",
					Measurements: []hookv1alpha1.Measurement{{Value: "1"}, {Value: "2"}}})
				return hr
			}(),
			limit: 1,
		},
		{
			name: "spec metrics less than result metrics",
			hr: func() *hookv1alpha1.HookRun {
				hr := testutil.NewHookRun("hr")
				hr.Spec.Metrics = append(hr.Spec.Metrics, hookv1alpha1.Metric{Name: "m1"})
				hr.Status.MetricResults = append(hr.Status.MetricResults, hookv1alpha1.MetricResult{Name: "m1",
					Measurements: []hookv1alpha1.Measurement{{Value: "1"}, {Value: "2"}}})
				hr.Status.MetricResults = append(hr.Status.MetricResults, hookv1alpha1.MetricResult{Name: "m2",
					Measurements: []hookv1alpha1.Measurement{{Value: "1"}, {Value: "2"}}})
				return hr
			}(),
			limit: 0,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			f := newFixture(t)
			f.newController()
			mp := new(hookmock.MockProvider)
			metricsByName := make(map[string]hookv1alpha1.Metric)
			for _, metric := range s.hr.Spec.Metrics {
				metricsByName[metric.Name] = metric
			}

			for _, result := range s.hr.Status.MetricResults {
				length := len(result.Measurements)
				if length > s.limit {
					if metric, ok := metricsByName[result.Name]; ok {
						mp.On("GarbageCollect", s.hr, metric, s.limit).
							Once().Return(nil)
					}
				}
			}
			f.c.newProvider = func(metric hookv1alpha1.Metric) (providers.Provider, error) { return mp, nil }
			err := f.c.garbageCollectMeasurements(s.hr, s.limit)
			if s.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, s.expectedError.Error())
			}
			mp.AssertExpectations(t)
		})
	}
}

func newHr(metrics, failureLimit, successfulLimit int32) *hookv1alpha1.HookRun {
	hr := testutil.NewHookRun("hr1")
	for i := int32(0); i < metrics; i++ {
		hr.Spec.Metrics = append(hr.Spec.Metrics, hookv1alpha1.Metric{
			Name:            fmt.Sprintf("m%d", i),
			FailureLimit:    failureLimit,
			SuccessfulLimit: successfulLimit,
		})
	}
	return hr
}

func newMetricResult(name string, phase hookv1alpha1.HookPhase, successful, failed int32, measurements []hookv1alpha1.Measurement) hookv1alpha1.MetricResult {
	mr := hookv1alpha1.MetricResult{
		Name:         name,
		Phase:        phase,
		Measurements: measurements,
		Successful:   successful,
		Failed:       failed,
	}
	return mr
}

func newMeasurement(phase hookv1alpha1.HookPhase) hookv1alpha1.Measurement {
	return hookv1alpha1.Measurement{
		Phase: phase,
	}
}

func TestAssessRunStatus(t *testing.T) {
	tests := []struct {
		name              string
		hr                *hookv1alpha1.HookRun
		expectedHookPhase hookv1alpha1.HookPhase
	}{
		{
			name:              "not started",
			hr:                newHr(1, 1, 2),
			expectedHookPhase: hookv1alpha1.HookPhaseRunning,
		},
		{
			name: "all succeeded",
			hr: func() *hookv1alpha1.HookRun {
				hr := newHr(2, 1, 2)
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m0", hookv1alpha1.HookPhaseSuccessful, 1, 1, nil))
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m1", hookv1alpha1.HookPhaseSuccessful, 1, 1, nil))
				return hr
			}(),
			expectedHookPhase: hookv1alpha1.HookPhaseSuccessful,
		},
		{
			name: "one failed, other succeeded",
			hr: func() *hookv1alpha1.HookRun {
				hr := newHr(2, 1, 2)
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m0", hookv1alpha1.HookPhaseSuccessful, 1, 1, nil))
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m1", hookv1alpha1.HookPhaseFailed, 1, 1, nil))
				return hr
			}(),
			expectedHookPhase: hookv1alpha1.HookPhaseFailed,
		},
		{
			name: "one running, other completed",
			hr: func() *hookv1alpha1.HookRun {
				hr := newHr(2, 1, 2)
				m := newMeasurement(hookv1alpha1.HookPhaseRunning)
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m0", hookv1alpha1.HookPhaseRunning, 1, 1, []hookv1alpha1.Measurement{m}))
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m1", hookv1alpha1.HookPhaseFailed, 1, 1, nil))
				return hr
			}(),
			expectedHookPhase: hookv1alpha1.HookPhaseRunning,
		},
		{
			name: "result running, but measurement is already succeeded",
			hr: func() *hookv1alpha1.HookRun {
				hr := newHr(2, 1, 2)
				m := newMeasurement(hookv1alpha1.HookPhaseSuccessful)
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m0", hookv1alpha1.HookPhaseRunning, 2, 1, []hookv1alpha1.Measurement{m}))
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m1", hookv1alpha1.HookPhaseSuccessful, 2, 1, nil))
				return hr
			}(),
			expectedHookPhase: hookv1alpha1.HookPhaseSuccessful,
		},
		{
			name: "result running, but measurement is already failed",
			hr: func() *hookv1alpha1.HookRun {
				hr := newHr(2, 1, 2)
				m := newMeasurement(hookv1alpha1.HookPhaseFailed)
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m0", hookv1alpha1.HookPhaseRunning, 0, 2, []hookv1alpha1.Measurement{m}))
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m1", hookv1alpha1.HookPhaseSuccessful, 0, 2, nil))
				return hr
			}(),
			expectedHookPhase: hookv1alpha1.HookPhaseFailed,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			hc := &HookController{recorder: &record.FakeRecorder{}}
			if got := hc.assessRunStatus(s.hr); got != s.expectedHookPhase {
				t.Errorf("assessRunStatus() = %v, want %v", got, s.expectedHookPhase)
			}
		})
	}
}

func TestAssessMetricStatus(t *testing.T) {
	tests := []struct {
		name          string
		metric        hookv1alpha1.Metric
		result        hookv1alpha1.MetricResult
		terminating   bool
		expectedPhase hookv1alpha1.HookPhase
	}{
		{
			name:   "test result complete",
			metric: hookv1alpha1.Metric{Name: "m1"},
			result: hookv1alpha1.MetricResult{
				Phase: hookv1alpha1.HookPhaseSuccessful,
			},
			expectedPhase: hookv1alpha1.HookPhaseSuccessful,
		},
		{
			name:   "measurements is empty",
			metric: hookv1alpha1.Metric{Name: "m1"},
			result: hookv1alpha1.MetricResult{
				Phase: hookv1alpha1.HookPhaseRunning,
			},
			expectedPhase: hookv1alpha1.HookPhasePending,
		},
		{
			name:   "measurements is empty and terminating",
			metric: hookv1alpha1.Metric{Name: "m1"},
			result: hookv1alpha1.MetricResult{
				Phase: hookv1alpha1.HookPhaseRunning,
			},
			terminating:   true,
			expectedPhase: hookv1alpha1.HookPhaseSuccessful,
		},
		{
			name:   "lastMeasurement is not completed",
			metric: hookv1alpha1.Metric{Name: "m1"},
			result: hookv1alpha1.MetricResult{
				Phase:        hookv1alpha1.HookPhaseRunning,
				Measurements: []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseRunning)},
			},
			expectedPhase: hookv1alpha1.HookPhaseRunning,
		},
		{
			name:   "failed greater than failureLimit",
			metric: hookv1alpha1.Metric{Name: "m1", FailureLimit: 2},
			result: hookv1alpha1.MetricResult{
				Phase:        hookv1alpha1.HookPhaseRunning,
				Failed:       3,
				Measurements: []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseFailed)},
			},
			expectedPhase: hookv1alpha1.HookPhaseFailed,
		},
		{
			name:   "successful greater than successfulLimit",
			metric: hookv1alpha1.Metric{Name: "m1", SuccessfulLimit: 2},
			result: hookv1alpha1.MetricResult{
				Phase:        hookv1alpha1.HookPhaseRunning,
				Successful:   3,
				Measurements: []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseFailed)},
			},
			expectedPhase: hookv1alpha1.HookPhaseSuccessful,
		},
		{
			name:   "inconclusive greater than inconclusiveLimit",
			metric: hookv1alpha1.Metric{Name: "m1", InconclusiveLimit: 2},
			result: hookv1alpha1.MetricResult{
				Phase:        hookv1alpha1.HookPhaseRunning,
				Inconclusive: 3,
				Measurements: []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseFailed)},
			},
			expectedPhase: hookv1alpha1.HookPhaseInconclusive,
		},
		{
			name:   "got ConsecutiveError",
			metric: hookv1alpha1.Metric{Name: "m1", ConsecutiveErrorLimit: func() *int32 { i := int32(1); return &i }()},
			result: hookv1alpha1.MetricResult{
				Phase:            hookv1alpha1.HookPhaseRunning,
				ConsecutiveError: DefaultConsecutiveErrorLimit,
				Measurements:     []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseFailed)},
			},
			expectedPhase: hookv1alpha1.HookPhaseError,
		},
		{
			name:   "got ConsecutiveSuccessful",
			metric: hookv1alpha1.Metric{Name: "m1", ConsecutiveSuccessfulLimit: func() *int32 { i := int32(1); return &i }()},
			result: hookv1alpha1.MetricResult{
				Phase:                 hookv1alpha1.HookPhaseRunning,
				ConsecutiveSuccessful: DefaultConsecutiveErrorLimit,
				Measurements:          []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseFailed)},
			},
			expectedPhase: hookv1alpha1.HookPhaseSuccessful,
		},
		{
			name:   "metric effective count reached",
			metric: hookv1alpha1.Metric{Name: "m1", Count: 3},
			result: hookv1alpha1.MetricResult{
				Phase:        hookv1alpha1.HookPhaseRunning,
				Count:        3,
				Measurements: []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseFailed)},
			},
			expectedPhase: hookv1alpha1.HookPhaseSuccessful,
		},
		{
			name:   "metric effective count not reached and terminating",
			metric: hookv1alpha1.Metric{Name: "m1"},
			result: hookv1alpha1.MetricResult{
				Phase:        hookv1alpha1.HookPhaseRunning,
				Measurements: []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseFailed)},
			},
			terminating:   true,
			expectedPhase: hookv1alpha1.HookPhaseSuccessful,
		},
		{
			name:   "metric effective count not reached and not terminating",
			metric: hookv1alpha1.Metric{Name: "m1"},
			result: hookv1alpha1.MetricResult{
				Phase:        hookv1alpha1.HookPhaseRunning,
				Measurements: []hookv1alpha1.Measurement{newMeasurement(hookv1alpha1.HookPhaseFailed)},
			},
			expectedPhase: hookv1alpha1.HookPhaseRunning,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			if phase := assessMetricStatus(s.metric, s.result, s.terminating); phase != s.expectedPhase {
				t.Errorf("got: %v, want: %v", phase, s.expectedPhase)
			}
		})
	}
}

func newRunWithMetricsAndResult(metrics []hookv1alpha1.Metric, metricResult []hookv1alpha1.MetricResult,
	startedAt metav1.Time) *hookv1alpha1.HookRun {
	return &hookv1alpha1.HookRun{
		Spec: hookv1alpha1.HookRunSpec{
			Metrics: metrics,
		},
		Status: hookv1alpha1.HookRunStatus{
			StartedAt:     &startedAt,
			MetricResults: metricResult,
		},
	}
}

func TestCalculateNextReconcileTime(t *testing.T) {
	tests := []struct {
		name         string
		metrics      []hookv1alpha1.Metric
		metricResult []hookv1alpha1.MetricResult
		startedAt    metav1.Time
		expectedTime *time.Time
	}{
		{
			name:         "all completed",
			metrics:      []hookv1alpha1.Metric{{Name: "m1"}},
			metricResult: []hookv1alpha1.MetricResult{{Name: "m1", Phase: hookv1alpha1.HookPhaseSuccessful}},
		},
		{
			name:         "metric not started",
			metrics:      []hookv1alpha1.Metric{{Name: "m1"}},
			metricResult: []hookv1alpha1.MetricResult{{Name: "m1", Phase: hookv1alpha1.HookPhaseRunning}},
		},
		{
			name:         "metric with InitialDelay, but failed to parse interval",
			metrics:      []hookv1alpha1.Metric{{Name: "m1", InitialDelay: "1"}},
			metricResult: []hookv1alpha1.MetricResult{{Name: "m1", Phase: hookv1alpha1.HookPhaseRunning}},
		},
		{
			name:         "metric with InitialDelay",
			metrics:      []hookv1alpha1.Metric{{Name: "m1", InitialDelay: "1s"}},
			metricResult: []hookv1alpha1.MetricResult{{Name: "m1", Phase: hookv1alpha1.HookPhaseRunning}},
			expectedTime: func() *time.Time { tt := time.Time{}.Add(time.Second); return &tt }(),
		},
		{
			name:    "measurement is not finished",
			metrics: []hookv1alpha1.Metric{{Name: "m1"}},
			metricResult: []hookv1alpha1.MetricResult{
				{Name: "m1", Measurements: []hookv1alpha1.Measurement{{FinishedAt: nil, ResumeAt: &metav1.Time{Time: time.Time{}.Add(time.Second)}}}},
			},
			expectedTime: func() *time.Time { tt := time.Time{}.Add(time.Second); return &tt }(),
		},
		{
			name:    "metrics count is reached, not next reconcile time",
			metrics: []hookv1alpha1.Metric{{Name: "m1", Count: 1}},
			metricResult: []hookv1alpha1.MetricResult{
				{Name: "m1", Measurements: []hookv1alpha1.Measurement{{FinishedAt: &metav1.Time{
					Time: time.Time{}.Add(time.Second)}}},
					Count: 1,
				},
			},
		},
		{
			name:    "metrics with interval and failed to parse interval",
			metrics: []hookv1alpha1.Metric{{Name: "m1", Interval: "30"}},
			metricResult: []hookv1alpha1.MetricResult{
				{Name: "m1", Measurements: []hookv1alpha1.Measurement{{FinishedAt: &metav1.Time{
					Time: time.Time{}.Add(time.Second)}}},
					Count: 1,
				},
			},
		},
		{
			name:    "metrics with interval",
			metrics: []hookv1alpha1.Metric{{Name: "m1", Interval: "30s"}},
			metricResult: []hookv1alpha1.MetricResult{
				{Name: "m1", Measurements: []hookv1alpha1.Measurement{{FinishedAt: &metav1.Time{
					Time: time.Time{}.Add(time.Second)}}},
					Count: 1,
				},
			},
			expectedTime: func() *time.Time { tt := time.Time{}.Add(31 * time.Second); return &tt }(),
		},
		{
			name:    "not setting interval",
			metrics: []hookv1alpha1.Metric{{Name: "m1"}},
			metricResult: []hookv1alpha1.MetricResult{
				{Name: "m1", Measurements: []hookv1alpha1.Measurement{{FinishedAt: &metav1.Time{
					Time: time.Time{}.Add(time.Second)}}},
				},
			},
		},
		{
			name:    "lastMeasurement is error, need to retry",
			metrics: []hookv1alpha1.Metric{{Name: "m1"}},
			metricResult: []hookv1alpha1.MetricResult{
				{Name: "m1", Measurements: []hookv1alpha1.Measurement{{FinishedAt: &metav1.Time{
					Time: time.Time{}.Add(time.Second)}, Phase: hookv1alpha1.HookPhaseError}},
				},
			},
			expectedTime: func() *time.Time { tt := time.Time{}.Add(11 * time.Second); return &tt }(),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			if got := calculateNextReconcileTime(
				newRunWithMetricsAndResult(s.metrics, s.metricResult, s.startedAt)); !reflect.DeepEqual(got, s.expectedTime) {
				t.Errorf("got: %v, want: %v", got, s.expectedTime)
			}
		})
	}
}

func TestRunMeasurements(t *testing.T) {
	tests := []struct {
		name                  string
		run                   *hookv1alpha1.HookRun
		tasks                 []metricTask
		expectedRun           bool
		expectedTerminate     bool
		expectedResume        bool
		expectedMetricResults []hookv1alpha1.MetricResult
	}{
		{
			name: "metric result is empty",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				return r
			}(),
			tasks: []metricTask{
				{metric: hookv1alpha1.Metric{Name: "m0"}},
			},
			expectedRun: true,
			expectedMetricResults: []hookv1alpha1.MetricResult{
				{
					Name:         "m0",
					Phase:        hookv1alpha1.HookPhaseRunning,
					Measurements: []hookv1alpha1.Measurement{{}},
				},
			},
		},
		{
			name: "incompleteMeasurement is empty with terminate",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(2, 0, 0)
				r.Spec.Terminate = true
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{Name: "m0"})
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name: "m1", Measurements: []hookv1alpha1.Measurement{
						{},
					}})
				return r
			}(),
			tasks: []metricTask{
				{metric: hookv1alpha1.Metric{Name: "m0"}},
				{metric: hookv1alpha1.Metric{Name: "m1"}, incompleteMeasurement: &hookv1alpha1.Measurement{}},
			},
			expectedRun:       true,
			expectedTerminate: true,
			expectedMetricResults: []hookv1alpha1.MetricResult{
				{Name: "m0", Measurements: []hookv1alpha1.Measurement{{}}},
				{Name: "m1", Measurements: []hookv1alpha1.Measurement{{
					Phase:   hookv1alpha1.HookPhaseSuccessful,
					Message: "metric terminated"},
				}, Count: 1, Successful: 1, ConsecutiveSuccessful: 1},
			},
		},
		{
			name: "incompleteMeasurement is not empty",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(2, 0, 0)
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{Name: "m0"})
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name: "m1", Measurements: []hookv1alpha1.Measurement{
						{},
					}})
				return r
			}(),
			tasks: []metricTask{
				{metric: hookv1alpha1.Metric{Name: "m0"}},
				{metric: hookv1alpha1.Metric{Name: "m1"}, incompleteMeasurement: &hookv1alpha1.Measurement{}},
			},
			expectedRun:    true,
			expectedResume: true,
			expectedMetricResults: []hookv1alpha1.MetricResult{
				{Name: "m0", Measurements: []hookv1alpha1.Measurement{{}}},
				{Name: "m1", Measurements: []hookv1alpha1.Measurement{{}}},
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			f := newFixture(t)
			f.newController()
			mp := new(hookmock.MockProvider)
			if s.expectedRun {
				mp.On("Run", s.run, mock.Anything).Once().Return(hookv1alpha1.Measurement{})
			}
			if s.expectedResume {
				mp.On("Resume", s.run, mock.Anything, mock.Anything).Once().Return(hookv1alpha1.Measurement{})
			}
			if s.expectedTerminate {
				mp.On("Terminate", s.run, mock.Anything, mock.Anything).Once().Return(hookv1alpha1.Measurement{
					Phase: hookv1alpha1.HookPhaseSuccessful})
			}
			f.c.newProvider = func(metric hookv1alpha1.Metric) (providers.Provider, error) { return mp, nil }
			f.c.runMeasurements(s.run, s.tasks)
			mp.AssertExpectations(t)
			// remove time, because time.Now() is always different
			for i := range s.run.Status.MetricResults {
				for j := range s.run.Status.MetricResults[i].Measurements {
					s.run.Status.MetricResults[i].Measurements[j].FinishedAt = nil
				}
			}
			assert.Equal(t, s.expectedMetricResults, s.run.Status.MetricResults)
		})
	}
}

func TestGenerateMetricTasks(t *testing.T) {
	tests := []struct {
		name          string
		run           *hookv1alpha1.HookRun
		expectedTasks []metricTask
	}{
		{
			name: "only one complete",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(3, 0, 0)
				// run in progress
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name:         "m0",
					Measurements: []hookv1alpha1.Measurement{{}},
				})
				// complete
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name:  "m1",
					Phase: hookv1alpha1.HookPhaseSuccessful})
				// resume not yet
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name: "m2",
					Measurements: []hookv1alpha1.Measurement{
						{
							ResumeAt: &metav1.Time{Time: time.Unix(10000000000, 0)},
						},
					},
				})
				// run is terminating, skip
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name:  "m0",
					Phase: hookv1alpha1.HookPhaseSuccessful})
				return r
			}(),
			expectedTasks: []metricTask{
				{
					metric:                hookv1alpha1.Metric{Name: "m0"},
					incompleteMeasurement: &hookv1alpha1.Measurement{},
				},
			},
		},
		{
			name: "terminating",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Terminate = true
				return r
			}(),
		},
		{
			name: "measurement never taken",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				return r
			}(),
			expectedTasks: []metricTask{{metric: hookv1alpha1.Metric{Name: "m0"}}},
		},
		{
			name: "metric with InitialDelay, and not started",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Metrics[0].InitialDelay = "1s"
				r.Status.StartedAt = nil
				return r
			}(),
		},
		{
			name: "metric with InitialDelay, and failed to parse duration",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Metrics[0].InitialDelay = "1"
				r.Status.StartedAt = &metav1.Time{Time: time.Now()}
				return r
			}(),
		},
		{
			name: "metric with InitialDelay, and duration not reached",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Metrics[0].InitialDelay = "1h"
				r.Status.StartedAt = &metav1.Time{Time: time.Now()}
				return r
			}(),
		},
		{
			name: "has measurements",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(2, 0, 0)
				// reached desired count
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name: "m0",
					Measurements: []hookv1alpha1.Measurement{
						{Phase: hookv1alpha1.HookPhaseRunning, FinishedAt: &metav1.Time{Time: time.Now()}},
					},
					Count: 1,
				})
				// overdue measurement
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name: "m1",
					Measurements: []hookv1alpha1.Measurement{
						{Phase: hookv1alpha1.HookPhaseRunning, FinishedAt: &metav1.Time{Time: time.Now().Add(-time.Hour)}},
					},
				})
				return r
			}(),
			expectedTasks: []metricTask{{metric: hookv1alpha1.Metric{Name: "m1"}}},
		},
		{
			name: "metric with Interval",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Metrics[0].Interval = "30s"
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name: "m0",
					Measurements: []hookv1alpha1.Measurement{
						{Phase: hookv1alpha1.HookPhaseRunning, FinishedAt: &metav1.Time{Time: time.Now()}},
					},
				})
				return r
			}(),
		},
		{
			name: "metric with Interval and failed to parse duration",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Metrics[0].Interval = "30"
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name: "m0",
					Measurements: []hookv1alpha1.Measurement{
						{Phase: hookv1alpha1.HookPhaseRunning, FinishedAt: &metav1.Time{Time: time.Now()}},
					},
				})
				return r
			}(),
		},
		{
			name: "Ordered polocy, only generate one task for the first metric",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(2, 0, 0)
				r.Spec.Policy = hookv1alpha1.OrderedPolicy
				return r
			}(),
			expectedTasks: []metricTask{{metric: hookv1alpha1.Metric{Name: "m0"}}},
		},
		{
			name: "Ordered polocy, waitting for the first metric to be completed",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(2, 0, 0)
				// reached desired count
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name: "m0",
					Measurements: []hookv1alpha1.Measurement{
						{Phase: hookv1alpha1.HookPhaseRunning, FinishedAt: &metav1.Time{Time: time.Now()}},
					},
					Count: 1,
				})
				r.Spec.Policy = hookv1alpha1.OrderedPolicy
				return r
			}(),
		},
		{
			name: "Ordered polocy, only generate one task",
			run: func() *hookv1alpha1.HookRun {
				r := newHr(3, 0, 0)
				r.Status.MetricResults = append(r.Status.MetricResults, hookv1alpha1.MetricResult{
					Name:  "m0",
					Phase: hookv1alpha1.HookPhaseSuccessful,
					Measurements: []hookv1alpha1.Measurement{
						{Phase: hookv1alpha1.HookPhaseSuccessful, FinishedAt: &metav1.Time{Time: time.Now()}},
					},
					Count:      1,
					Successful: 1,
				})
				r.Spec.Policy = hookv1alpha1.OrderedPolicy
				return r
			}(),
			expectedTasks: []metricTask{{metric: hookv1alpha1.Metric{Name: "m1"}}},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			got := generateMetricTasks(s.run)
			assert.Equal(t, s.expectedTasks, got)
		})
	}
}

func TestReconcileHookRun(t *testing.T) {
	tests := []struct {
		name              string
		originRun         *hookv1alpha1.HookRun
		expectedHookRun   *hookv1alpha1.HookRun
		expectedRun       bool
		expectedTerminate bool
		expectedResume    bool
	}{
		{
			name: "hook run complete",
			originRun: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Status.Phase = hookv1alpha1.HookPhaseSuccessful
				return r
			}(),
			expectedHookRun: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Status.Phase = hookv1alpha1.HookPhaseSuccessful
				return r
			}(),
		},
		{
			name: "no provider",
			originRun: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				return r
			}(),
			expectedHookRun: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Status.MetricResults = make([]hookv1alpha1.MetricResult, 0)
				message := fmt.Sprintf("HookRun: %s/%s, hook spec invalid: %v", r.Namespace, r.Name,
					errors.New("metrics[0]: no provider specified"))
				r.Status.Phase = hookv1alpha1.HookPhaseError
				r.Status.Message = message
				return r
			}(),
		},
		{
			name: "start hook run",
			originRun: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Metrics[0].Provider = hookv1alpha1.MetricProvider{
					Web: &hookv1alpha1.WebMetric{},
				}
				return r
			}(),
			expectedHookRun: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Metrics[0].Provider = hookv1alpha1.MetricProvider{
					Web: &hookv1alpha1.WebMetric{},
				}
				r.Status = hookv1alpha1.HookRunStatus{
					Phase: hookv1alpha1.HookPhaseRunning,
					MetricResults: []hookv1alpha1.MetricResult{
						{
							Name:         "m0",
							Phase:        hookv1alpha1.HookPhaseRunning,
							Measurements: []hookv1alpha1.Measurement{{}},
						},
					},
				}
				return r
			}(),
			expectedRun: true,
		},
		{
			name: "new status complete",
			originRun: func() *hookv1alpha1.HookRun {
				hr := newHr(2, 0, 0)
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m0", hookv1alpha1.HookPhaseSuccessful, 1, 1, nil))
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m1", hookv1alpha1.HookPhaseSuccessful, 1, 1, nil))
				return hr
			}(),
			expectedHookRun: func() *hookv1alpha1.HookRun {
				r := newHr(2, 0, 0)
				r.Status = hookv1alpha1.HookRunStatus{
					Phase: hookv1alpha1.HookPhaseSuccessful,
					MetricResults: []hookv1alpha1.MetricResult{
						{
							Name:       "m0",
							Phase:      hookv1alpha1.HookPhaseSuccessful,
							Successful: 1,
							Failed:     1,
						},
						{
							Name:       "m1",
							Phase:      hookv1alpha1.HookPhaseSuccessful,
							Successful: 1,
							Failed:     1,
						},
					},
				}
				return r
			}(),
		},
		{
			name: "new status not complete",
			originRun: func() *hookv1alpha1.HookRun {
				hr := newHr(1, 0, 0)
				hr.Spec.Metrics[0].InitialDelay = "10s"
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m0", hookv1alpha1.HookPhaseRunning, 1, 1, nil))
				return hr
			}(),
			expectedHookRun: func() *hookv1alpha1.HookRun {
				r := newHr(1, 0, 0)
				r.Spec.Metrics[0].InitialDelay = "10s"
				r.Status = hookv1alpha1.HookRunStatus{
					Phase: hookv1alpha1.HookPhaseRunning,
					MetricResults: []hookv1alpha1.MetricResult{
						newMetricResult("m0", hookv1alpha1.HookPhasePending, 1, 1, nil),
					},
				}
				return r
			}(),
		},
		{
			name: "new status failed",
			originRun: func() *hookv1alpha1.HookRun {
				hr := newHr(2, 0, 0)
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m0", hookv1alpha1.HookPhaseSuccessful, 1, 1, nil))
				hr.Status.MetricResults = append(hr.Status.MetricResults,
					newMetricResult("m1", hookv1alpha1.HookPhaseFailed, 1, 1, nil))
				return hr
			}(),
			expectedHookRun: func() *hookv1alpha1.HookRun {
				r := newHr(2, 0, 0)
				r.Status = hookv1alpha1.HookRunStatus{
					Phase: hookv1alpha1.HookPhaseFailed,
					MetricResults: []hookv1alpha1.MetricResult{
						{
							Name:       "m0",
							Phase:      hookv1alpha1.HookPhaseSuccessful,
							Successful: 1,
							Failed:     1,
						},
						{
							Name:       "m1",
							Phase:      hookv1alpha1.HookPhaseFailed,
							Successful: 1,
							Failed:     1,
						},
					},
				}
				return r
			}(),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			f := newFixture(t)
			f.newController()
			mp := new(hookmock.MockProvider)
			f.c.newProvider = func(metric hookv1alpha1.Metric) (providers.Provider, error) { return mp, nil }
			if s.expectedRun {
				mp.On("Run", mock.Anything, mock.Anything).Once().Return(hookv1alpha1.Measurement{})
			}
			if s.expectedResume {
				mp.On("Resume", s.originRun, mock.Anything, mock.Anything).Once().Return(hookv1alpha1.Measurement{})
			}
			if s.expectedTerminate {
				mp.On("Terminate", s.originRun, mock.Anything, mock.Anything).Once().Return(hookv1alpha1.Measurement{
					Phase: hookv1alpha1.HookPhaseSuccessful})
			}
			hr := f.c.reconcileHookRun(s.originRun)
			mp.AssertExpectations(t)
			// remove time, because time.Now() is always different
			hr.Status.StartedAt = nil
			assert.Equal(t, s.expectedHookRun, hr)
		})
	}
}
