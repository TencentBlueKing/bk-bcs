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

package hook

import (
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/metrics"
	hooksutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/hook"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	// DefaultMeasurementHistoryLimit is the default maximum number of measurements to retain per metric,
	// before trimming the list.
	DefaultMeasurementHistoryLimit = 10
	// DefaultErrorRetryInterval is the default interval to retry a measurement upon error, in the
	// event an interval was not specified
	DefaultErrorRetryInterval time.Duration = 10 * time.Second
	// DefaultConsecutiveErrorLimit is the default number times a metric can error in sequence before
	// erroring the entire metric.
	DefaultConsecutiveErrorLimit int32 = 4
)

// Event reasons for hook events
const (
	EventReasonStatusFailed    = "Failed"
	EventReasonStatusCompleted = "Complete"
)

// metricTask holds the metric which need to be measured during this reconciliation along with
// an in-progress measurement
type metricTask struct {
	metric                v1alpha1.Metric
	incompleteMeasurement *v1alpha1.Measurement
}

// reconcileHookRun main logic for reconcile
func (hc *HookController) reconcileHookRun(origRun *v1alpha1.HookRun) *v1alpha1.HookRun {
	if origRun.Status.Phase.Completed() {
		return origRun
	}

	run := origRun.DeepCopy()

	if run.Status.MetricResults == nil {
		run.Status.MetricResults = make([]v1alpha1.MetricResult, 0)
		startTime := time.Now()
		err := hooksutil.ValidateMetrics(run.Spec.Metrics)
		if err != nil {
			message := fmt.Sprintf("HookRun: %s/%s, hook spec invalid: %v", run.Namespace, run.Name, err)
			klog.Warning(message)
			run.Status.Phase = v1alpha1.HookPhaseError
			run.Status.Message = message
			hc.recorder.Eventf(run, corev1.EventTypeWarning, EventReasonStatusFailed, "hook completed %s",
				run.Status.Phase)
			hc.metrics.CollectHookrunExecDurations(run.Namespace, string(run.UID), hooksutil.GetOwnerRef(run),
				"validateMetrics", "failure", time.Since(startTime))
			return run
		}
		hc.metrics.CollectHookrunExecDurations(run.Namespace, string(run.UID), hooksutil.GetOwnerRef(run),
			"validateMetrics", "success", time.Since(startTime))
	}
	tasks := generateMetricTasks(run)
	klog.Infof("HookRun: %s/%s, taking %d measurements", run.Namespace, run.Name, len(tasks))
	startime := time.Now()
	hc.runMeasurements(run, tasks, hc.metrics)

	newStatus := hc.assessRunStatus(run)
	if newStatus != run.Status.Phase {
		message := fmt.Sprintf("HookRun: %s/%s, hook transitioned from %s -> %s", run.Namespace, run.Name,
			run.Status.Phase, newStatus)
		if newStatus.Completed() {
			switch newStatus {
			case v1alpha1.HookPhaseError, v1alpha1.HookPhaseFailed:
				hc.recorder.Eventf(run, corev1.EventTypeWarning, EventReasonStatusFailed,
					"hook completed %s", newStatus)
				hc.metrics.CollectHookrunExecDurations(run.Namespace, string(run.UID), hooksutil.GetOwnerRef(run),
					"executeHookRun", "failure", time.Since(startime))
			default:
				hc.recorder.Eventf(run, corev1.EventTypeNormal, EventReasonStatusCompleted,
					"hook completed %s", newStatus)
				hc.metrics.CollectHookrunExecDurations(run.Namespace, string(run.UID), hooksutil.GetOwnerRef(run),
					"executeHookRun", "success", time.Since(startime))
			}
		}
		klog.Info(message)
		run.Status.Phase = newStatus
	}

	err := hc.garbageCollectMeasurements(run, DefaultMeasurementHistoryLimit)
	if err != nil {
		klog.Warningf("HookRun: %s/%s, Failed to garbage collect measurements: %v", run.Namespace, run.Name, err)
	}

	nextReconcileTime := calculateNextReconcileTime(run)
	if nextReconcileTime != nil {
		enqueueSeconds := nextReconcileTime.Sub(time.Now())
		if enqueueSeconds < 0 {
			enqueueSeconds = 0
		}
		klog.Infof("HookRun: %s/%s, enqueuing hook after %v", run.Namespace, run.Name, enqueueSeconds)
		hc.enqueueHookRunAfter(run, enqueueSeconds)
	}
	return run
}

func generateMetricTasks(run *v1alpha1.HookRun) []metricTask {
	var tasks []metricTask
	terminating := hooksutil.IsTerminating(run)
	for i, metric := range run.Spec.Metrics {
		// if previous metric has not been completed, wait until next reconcile
		if i > 0 && run.Spec.Policy == v1alpha1.OrderedPolicy && !hooksutil.MetricCompleted(run, run.Spec.Metrics[i-1].Name) {
			klog.Infof("With Ordered policy, waitting %s to be completed", run.Spec.Metrics[i-1].Name)
			break
		}
		if hooksutil.MetricCompleted(run, metric.Name) {
			continue
		}
		lastMeasurement := hooksutil.LastMeasurement(run, metric.Name)
		if lastMeasurement != nil && lastMeasurement.FinishedAt == nil {
			now := metav1.Now()
			if lastMeasurement.ResumeAt != nil && lastMeasurement.ResumeAt.After(now.Time) {
				continue
			}
			// last measurement is still in-progress. need to complete it
			tasks = append(tasks, metricTask{
				metric:                metric,
				incompleteMeasurement: lastMeasurement,
			})
			continue
		}
		if terminating {
			klog.Infof("HookRun: %s/%s, metric: %s. skipping measurement，run is terminating",
				run.Namespace, run.Name, metric.Name)
			continue
		}
		if lastMeasurement == nil {
			if metric.InitialDelay != "" {
				if run.Status.StartedAt == nil {
					continue
				}
				duration, err := metric.InitialDelay.Duration()
				if err != nil {
					klog.Warningf("HookRun: %s/%s, metric: %s. failed to parse duration: %s",
						run.Namespace, run.Name, metric.Name, err.Error())
					continue
				}
				if run.Status.StartedAt.Add(duration).After(time.Now()) {
					klog.Infof("HookRun: %s/%s, metric: %s. waiting until start delay duration passes",
						run.Namespace, run.Name, metric.Name)
					continue
				}
			}
			// measurement never taken
			tasks = append(tasks, metricTask{metric: metric})
			klog.Infof("HookRun: %s/%s, metric: %s. running initial measurement", run.Namespace, run.Name,
				metric.Name)
			continue
		}
		metricResult := hooksutil.GetResult(run, metric.Name)
		effectiveCount := metric.EffectiveCount()
		if effectiveCount != nil && metricResult.Count >= *effectiveCount {
			// we have reached desired count
			continue
		}
		// if we get here, we know we need to take a measurement (eventually). check last measurement
		// to decide if it should be taken now. metric.Interval can be null because we may be
		// retrying a metric due to error.
		interval := DefaultErrorRetryInterval
		if metric.Interval != "" {
			metricInterval, err := metric.Interval.Duration()
			if err != nil {
				klog.Warningf("HookRun: %s/%s, metric: %s. failed to parse internal: %s", run.Namespace,
					run.Name, metric.Name, err.Error())
				continue
			}
			interval = metricInterval
		}
		if time.Now().After(lastMeasurement.FinishedAt.Add(interval)) {
			tasks = append(tasks, metricTask{metric: metric})
			klog.Infof("HookRun: %s/%s, metric: %s. running overdue measurement", run.Namespace, run.Name,
				metric.Name)
			continue
		}
	}
	return tasks
}

// runMeasurements iterates a list of metric tasks, and runs, resumes, or terminates measurements
func (hc *HookController) runMeasurements(run *v1alpha1.HookRun, tasks []metricTask, metrics *metrics.Metrics) {
	var wg sync.WaitGroup
	var resultsLock sync.Mutex
	terminating := hooksutil.IsTerminating(run)

	for _, task := range tasks {
		wg.Add(1)

		go func(t metricTask) {
			defer wg.Done()

			resultsLock.Lock()
			metricResult := hooksutil.GetResult(run, t.metric.Name)
			resultsLock.Unlock()

			if metricResult == nil {
				metricResult = &v1alpha1.MetricResult{
					Name:  t.metric.Name,
					Phase: v1alpha1.HookPhaseRunning,
				}
			}

			var newMeasurement v1alpha1.Measurement
			startTime := time.Now()
			provider, err := hc.newProvider(t.metric)
			if err != nil {
				metrics.CollectHookrunExecDurations(run.Namespace, string(run.UID), hooksutil.GetOwnerRef(run),
					"executeHookRun", "error", time.Since(startTime))
				if t.incompleteMeasurement != nil {
					newMeasurement = *t.incompleteMeasurement
				} else {
					startedAt := metav1.Now()
					newMeasurement.StartedAt = &startedAt
				}
				newMeasurement.Phase = v1alpha1.HookPhaseError
				newMeasurement.Message = err.Error()
			} else {
				if t.incompleteMeasurement == nil {
					startTime := time.Now()
					newMeasurement = provider.Run(run, t.metric)
					hc.metrics.CollectMetricExecDurations(run.Namespace, hooksutil.GetOwnerRef(run), t.metric.Name,
						string(newMeasurement.Phase), time.Since(startTime))
					if newMeasurement.Phase == v1alpha1.HookPhaseError {
						metrics.CollectHookrunExecDurations(run.Namespace, string(run.UID), hooksutil.GetOwnerRef(run),
							"invalidHookConfig", "failure", time.Since(startTime))
					}
				} else {
					if terminating {
						klog.Infof("HookRun: %s/%s, metric: %s. terminating in-progress measurement",
							run.Namespace, run.Name, t.metric.Name)
						newMeasurement = provider.Terminate(run, t.metric, *t.incompleteMeasurement)
						if newMeasurement.Phase == v1alpha1.HookPhaseSuccessful {
							newMeasurement.Message = "metric terminated"
						}
					} else {
						newMeasurement = provider.Resume(run, t.metric, *t.incompleteMeasurement)
					}
				}
			}

			if newMeasurement.Phase.Completed() {
				klog.Infof("HookRun: %s/%s, metric: %s. measurement completed %s", run.Namespace, run.Name,
					t.metric.Name, newMeasurement.Phase)
				if newMeasurement.FinishedAt == nil {
					finishedAt := metav1.Now()
					newMeasurement.FinishedAt = &finishedAt
				}
				switch newMeasurement.Phase {
				case v1alpha1.HookPhaseSuccessful:
					metricResult.Successful++
					metricResult.Count++
					metricResult.ConsecutiveError = 0
					metricResult.ConsecutiveSuccessful++
				case v1alpha1.HookPhaseFailed:
					metricResult.Failed++
					metricResult.Count++
					metricResult.ConsecutiveError = 0
					metricResult.ConsecutiveSuccessful = 0
				case v1alpha1.HookPhaseInconclusive:
					metricResult.Inconclusive++
					metricResult.Count++
					metricResult.ConsecutiveError = 0
					metricResult.ConsecutiveSuccessful = 0
				case v1alpha1.HookPhaseError:
					metricResult.Error++
					metricResult.Count++
					metricResult.ConsecutiveError++
					metricResult.ConsecutiveSuccessful = 0
				}
			}
			if t.incompleteMeasurement == nil {
				metricResult.Measurements = append(metricResult.Measurements, newMeasurement)
			} else {
				metricResult.Measurements[len(metricResult.Measurements)-1] = newMeasurement
			}

			resultsLock.Lock()
			hooksutil.SetResult(run, *metricResult)
			resultsLock.Unlock()
		}(task)
	}

	wg.Wait()
}

// asssessRunStatus assesses the overall status of this HookRun
// If any metric is not yet completed, the HookRun is still considered Running
// Once all metrics are complete, the worst status is used as the overall HookRun status
func (hc *HookController) assessRunStatus(run *v1alpha1.HookRun) v1alpha1.HookPhase {
	var worstStatus v1alpha1.HookPhase
	terminating := hooksutil.IsTerminating(run)
	everythingCompleted := true

	if run.Status.StartedAt == nil {
		now := metav1.Now()
		run.Status.StartedAt = &now
	}

	// Iterate all metrics and update MetricResult.Phase fields based on lastest measurement(s)
	for _, metric := range run.Spec.Metrics {
		if result := hooksutil.GetResult(run, metric.Name); result != nil {
			metricStatus := assessMetricStatus(metric, *result, terminating)
			if result.Phase != metricStatus {
				if metricStatus.Completed() {
					switch metricStatus {
					case v1alpha1.HookPhaseError, v1alpha1.HookPhaseFailed:
						hc.recorder.Eventf(run, corev1.EventTypeWarning, EventReasonStatusFailed,
							"metric '%s' completed %s", metric.Name, metricStatus)
					default:
						hc.recorder.Eventf(run, corev1.EventTypeNormal, EventReasonStatusCompleted,
							"metric '%s' completed %s", metric.Name, metricStatus)
					}
				}
				if lastMeasurement := hooksutil.LastMeasurement(run, metric.Name); lastMeasurement != nil {
					result.Message = lastMeasurement.Message
				}
				result.Phase = metricStatus
				hooksutil.SetResult(run, *result)
			}
			if !metricStatus.Completed() {
				everythingCompleted = false
			} else {
				if worstStatus == "" {
					worstStatus = metricStatus
				} else {
					if hooksutil.IsWorse(worstStatus, metricStatus) {
						worstStatus = metricStatus
					}
				}
			}
		} else {
			everythingCompleted = false
		}
	}
	if !everythingCompleted || worstStatus == "" {
		return v1alpha1.HookPhaseRunning
	}
	return worstStatus
}

// assessMetricStatus assesses the status of a single metric based on:
// * current/latest measurement status
// * parameters given by the metric (failureLimit, count, etc...)
// * whether or not we are terminating (e.g. due to failing run, or termination request)
func assessMetricStatus(metric v1alpha1.Metric, result v1alpha1.MetricResult, terminating bool) v1alpha1.HookPhase {
	if result.Phase.Completed() {
		return result.Phase
	}

	if len(result.Measurements) == 0 {
		if terminating {
			klog.Infof("metric %s assessed %s: run terminated", metric.Name, v1alpha1.HookPhaseSuccessful)
			return v1alpha1.HookPhaseSuccessful
		}
		return v1alpha1.HookPhasePending
	}
	lastMeasurement := result.Measurements[len(result.Measurements)-1]
	if !lastMeasurement.Phase.Completed() {
		return v1alpha1.HookPhaseRunning
	}
	if metric.FailureLimit > 0 && result.Failed > metric.FailureLimit {
		klog.Infof("metric %s assessed %s: failed (%d) > failureLimit (%d)", metric.Name,
			v1alpha1.HookPhaseFailed, result.Failed, metric.FailureLimit)
		return v1alpha1.HookPhaseFailed
	}

	if metric.SuccessfulLimit > 0 && result.Successful >= metric.SuccessfulLimit {
		klog.Infof("metric %s assessed %s: successful (%d) > successfulLimit (%d)", metric.Name,
			v1alpha1.HookPhaseSuccessful, result.Successful, metric.SuccessfulLimit)
		return v1alpha1.HookPhaseSuccessful
	}

	if result.Inconclusive > metric.InconclusiveLimit {
		klog.Infof("metric %s assessed %s: inconclusive (%d) > inconclusiveLimit (%d)", metric.Name,
			v1alpha1.HookPhaseInconclusive, result.Inconclusive, metric.InconclusiveLimit)
		return v1alpha1.HookPhaseInconclusive
	}

	if metric.ConsecutiveErrorLimit != nil {
		consecutiveErrorLimit := *metric.ConsecutiveErrorLimit
		if result.ConsecutiveError > consecutiveErrorLimit {
			klog.Infof("metric %s assessed %s: consecutiveErrors (%d) > consecutiveErrorLimit (%d)",
				metric.Name, v1alpha1.HookPhaseError, result.ConsecutiveError, consecutiveErrorLimit)
			return v1alpha1.HookPhaseError
		}
	}

	if metric.ConsecutiveSuccessfulLimit != nil {
		if result.ConsecutiveSuccessful >= *metric.ConsecutiveSuccessfulLimit {
			klog.Infof("metric %s assessed %s: consecutiveSuccessful (%d) >= consecutiveSuccessfulLimit (%d)",
				metric.Name, v1alpha1.HookPhaseSuccessful, result.ConsecutiveSuccessful,
				*metric.ConsecutiveSuccessfulLimit)
			return v1alpha1.HookPhaseSuccessful
		}
	}

	// If a count was specified, and we reached that count, then metric is considered Successful.
	// The Error, Failed, Inconclusive counters are ignored because those checks have already been
	// taken into consideration above, and we do not want to fail if failures < failureLimit.
	effectiveCount := metric.EffectiveCount()
	if effectiveCount != nil && result.Count >= *effectiveCount {
		klog.Infof("metric %s assessed %s: count (%d) reached", metric.Name, v1alpha1.HookPhaseSuccessful,
			*effectiveCount)
		return v1alpha1.HookPhaseSuccessful
	}

	// if we get here, this metric runs indefinitely
	if terminating {
		klog.Infof("metric %s assessed %s: run terminated", metric.Name, v1alpha1.HookPhaseSuccessful)
		return v1alpha1.HookPhaseSuccessful
	}
	return v1alpha1.HookPhaseRunning
}

// garbageCollectMeasurements trims the measurement history to the specified limit and GCs old measurements
func (hc *HookController) garbageCollectMeasurements(run *v1alpha1.HookRun, limit int) error {
	var errors []error

	metricsByName := make(map[string]v1alpha1.Metric)
	for _, metric := range run.Spec.Metrics {
		metricsByName[metric.Name] = metric
	}

	for i, result := range run.Status.MetricResults {
		length := len(result.Measurements)
		if length > limit {
			metric, ok := metricsByName[result.Name]
			if !ok {
				continue
			}
			provider, err := hc.newProvider(metric)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			err = provider.GarbageCollect(run, metric, limit)
			if err != nil {
				return err
			}
			result.Measurements = result.Measurements[length-limit : length]
		}
		run.Status.MetricResults[i] = result
	}
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// calculateNextReconcileTime calculates the next time that this AnalysisRun should be reconciled,
// based on the earliest time of all metrics intervals, counts, and their finishedAt timestamps
func calculateNextReconcileTime(run *v1alpha1.HookRun) *time.Time {
	var reconcileTime *time.Time
	for _, metric := range run.Spec.Metrics {
		if hooksutil.MetricCompleted(run, metric.Name) {
			continue
		}
		lastMeasurement := hooksutil.LastMeasurement(run, metric.Name)
		if lastMeasurement == nil {
			if metric.InitialDelay != "" {
				startTime := metav1.Now()
				if run.Status.StartedAt != nil {
					startTime = *run.Status.StartedAt
				}
				duration, err := metric.InitialDelay.Duration()
				if err != nil {
					klog.Warningf("HookRun: %s/%s, metric: %s. failed to parse interval: %v",
						run.Namespace, run.Name, metric.Name, err)
					continue
				}
				endInitialDelay := startTime.Add(duration)
				if reconcileTime == nil || reconcileTime.After(endInitialDelay) {
					reconcileTime = &endInitialDelay
				}
				continue
			}
			klog.Warningf("HookRun: %s/%s, metric: %s. metric never started. "+
				"not factored into enqueue time", run.Namespace, run.Name, metric.Name)
			continue
		}
		if lastMeasurement.FinishedAt == nil {
			if lastMeasurement.ResumeAt != nil {
				if reconcileTime == nil || reconcileTime.After(lastMeasurement.ResumeAt.Time) {
					reconcileTime = &lastMeasurement.ResumeAt.Time
				}
			}
			continue
		}
		metricResult := hooksutil.GetResult(run, metric.Name)
		effectiveCount := metric.EffectiveCount()
		if effectiveCount != nil && metricResult.Count >= *effectiveCount {
			continue
		}
		var interval time.Duration
		if metric.Interval != "" {
			metricInterval, err := metric.Interval.Duration()
			if err != nil {
				klog.Warningf("HookRun: %s/%s, metric: %s. failed to parse interval: %v", run.Namespace,
					run.Name, metric.Name, err)
				continue
			}
			interval = metricInterval
		} else if lastMeasurement.Phase == v1alpha1.HookPhaseError {
			interval = DefaultErrorRetryInterval
		} else {
			// if we get here, an interval was not set (meaning reoccurrence was not desired), and
			// there was no error (meaning we don't need to retry). no need to requeue this metric.
			// NOTE: we shouldn't ever get here since it means we are not doing proper bookkeeping
			// of count.
			klog.Warningf("HookRun: %s/%s, metric: %s. skipping requeue. no interval or error (count: %d, "+
				"effectiveCount: %d)", run.Namespace, run.Name, metric.Name, metricResult.Count, metric.EffectiveCount())
			continue
		}

		metricReconcileTime := lastMeasurement.FinishedAt.Add(interval)
		if reconcileTime == nil || reconcileTime.After(metricReconcileTime) {
			reconcileTime = &metricReconcileTime
		}
	}
	return reconcileTime
}
