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

package kubectl_agg

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/apis/aggregation/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-federated-apiserver/pkg/client/clientset_generated/clientset"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
)

// AggPodOptions struct is the basic commandline options.
type AggPodOptions struct {
	ResourceName  string
	Namespace     string
	Selector      string
	AllNamespaces bool
	WideMessage   string
	LabelsMessage bool
	HelpMessage   bool
}

// GetPodRestartCount function return the pod's starts counts.
func GetPodRestartCount(pod v1alpha1.PodAggregation) int32 {
	var restartCount int32 = 0
	for _, v := range pod.Status.ContainerStatuses {
		restartCount = restartCount + v.RestartCount
	}
	return restartCount
}

// GetContainerReadyStatus return the pod's Ready status.
func GetContainerReadyStatus(pod v1alpha1.PodAggregation) string {
	var containerCount int32 = 0
	var containerReadyCount int32 = 0

	for _, v := range pod.Status.ContainerStatuses {
		containerCount++
		if v.Ready {
			containerReadyCount++
		}
	}
	return fmt.Sprintf("%d/%d", containerReadyCount, containerCount)
}

// GetPodAge function return the Pod's age, if < 24h, use H; otherwise use D respect for day,
// and Y respect for year.
func GetPodAge(pod v1alpha1.PodAggregation) string {
	var createAgeHour float64

	createAgeHour = time.Since(pod.CreationTimestamp.Time).Hours()
	if createAgeHour < 24.0 {
		return fmt.Sprintf("%dh", int64(math.Ceil(createAgeHour)))
	} else {
		createAgeDay := createAgeHour / 24.0
		if createAgeDay < 365.0 {
			return fmt.Sprintf("%dd", int64(math.Ceil(createAgeDay)))
		} else {
			createAgeYear := createAgeDay / 365.0
			return fmt.Sprintf("%dy", int64(math.Ceil(createAgeYear)))
		}
	}
}

// GetPodLabel function return the Pod's Label.
func GetPodLabel(pod v1alpha1.PodAggregation) string {
	if len(pod.Labels) != 0 {
		var labels, labelsTmp string
		for k, v := range pod.Labels {
			labelsTmp = fmt.Sprintf("%s=%s", k, v) + ","
			labels += labelsTmp
		}
		labels = strings.TrimRight(labels, ",")
		return labels
	} else {
		return "<none>"
	}
}

// GetReadinessGateStatus function return the Pod's ReadinessGateStatus.
func GetReadinessGateStatus(pod v1alpha1.PodAggregation) string {
	var readinessGateCount int32
	var readinessGateReadyCount int32
	for _, v := range pod.Status.Conditions {
		if v.Type != "Initialized" && v.Type != "Ready" && v.Type != "ContainersReady" && v.Type != "PodScheduled" {
			readinessGateCount++
			if v.Status == "true" {
				readinessGateReadyCount++
			}
		}
	}

	if readinessGateCount == 0 {
		return "<none>"
	} else {
		return fmt.Sprintf("%d/%d", readinessGateReadyCount, readinessGateCount)
	}
}


// GetNominatedNode function return the Pod's NominatedNode info.
func GetNominatedNode(pod v1alpha1.PodAggregation) string {
	if pod.Status.NominatedNodeName == "" {
		return "<none>"
	} else {
		return pod.Status.NominatedNodeName
	}
}


// GetPodAggregationList function implement the pod info from command line,
// output the Pods from the backend storage. If the name is offered, using the Get function,
// otherwise using the List function.
func GetPodAggregationList(clientSet *clientset.Clientset, o *AggPodOptions) (pods *v1alpha1.PodAggregationList, err error) {
	if o.ResourceName != "" {
		pods, err = clientSet.AggregationV1alpha1().PodAggregations(o.Namespace).Get(context.TODO(),
			o.ResourceName, v1.GetOptions{})
		if err != nil {
			klog.Errorf("Error: failed to get pod: %s/%s: %s\n", o.Namespace, o.ResourceName, err)
			return &v1alpha1.PodAggregationList{}, nil
		}
	} else {
		if o.AllNamespaces {
			o.Namespace = ""
		}

		selector := labels.Everything()
		if len(o.Selector) > 0 {
			selector, err = labels.Parse(o.Selector)
			if err != nil {
				return &v1alpha1.PodAggregationList{}, nil
			}
		}

		pods, err = clientSet.AggregationV1alpha1().PodAggregations(o.Namespace).List(context.TODO(),
			v1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			klog.Errorf("Error: failed to list pods: %s\n", err)
			return &v1alpha1.PodAggregationList{}, nil
		}
	}
	return pods, nil
}

// PrintPodAggregation prints the output message of the specified pods.
func PrintPodAggregation(o *AggPodOptions, pods *v1alpha1.PodAggregationList) {
	if len(pods.Items) == 0 {
		fmt.Println("No resources found")
		return
	}
	if o.WideMessage != "wide" {
		var headerMessage string
		if !o.LabelsMessage {
			headerMessage = fmt.Sprintf("%-16s%-64s%-8s%-16s%-10s%-8s\n", "NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE")
		} else {
			headerMessage = fmt.Sprintf("%-16s%-64s%-8s%-16s%-10s%-8s%-20s\n", "NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE", "LABELS")
		}
		fmt.Printf(headerMessage)
	} else {
		var headerMessage string
		if !o.LabelsMessage {
			headerMessage = fmt.Sprintf("%-16s%-64s%-8s%-16s%-10s%-10s%-20s%-40s%-20s%-16s\n", "NAMESPACE", "NAME", "READY",
				"STATUS", "RESTARTS", "AGE", "IP", "NODE", "NOMINATED NODE", "READINESS GATES")
		} else {
			headerMessage = fmt.Sprintf("%-16s%-64s%-8s%-16s%-10s%-10s%-20s%-40s%-20s%-16s%-20s\n", "NAMESPACE", "NAME", "READY",
				"STATUS", "RESTARTS", "AGE", "IP", "NODE", "NOMINATED NODE", "READINESS GATES", "LABELS")
		}
		fmt.Printf(headerMessage)
	}

	for _, v := range pods.Items {
		if o.WideMessage != "wide" {
			if !o.LabelsMessage {
				fmt.Printf("%-16s%-64s%-8s%-16s%-10d%-8s\n", v.Namespace, v.Name,
					GetContainerReadyStatus(v), string(v.Status.Phase),
					GetPodRestartCount(v),
					GetPodAge(v))
			} else {
				fmt.Printf("%-16s%-64s%-8s%-16s%-10d%-8s%-20s\n", v.Namespace, v.Name,
					GetContainerReadyStatus(v), string(v.Status.Phase),
					GetPodRestartCount(v),
					GetPodAge(v),
					GetPodLabel(v))
			}
		} else {
			if !o.LabelsMessage {
				fmt.Printf("%-16s%-64s%-8s%-16s%-10d%-10s%-20s%-40s%-20s%-16s\n", v.Namespace, v.Name,
					GetContainerReadyStatus(v), string(v.Status.Phase),
					GetPodRestartCount(v),
					GetPodAge(v), v.Status.PodIP, v.Spec.NodeName,
					GetNominatedNode(v),
					GetReadinessGateStatus(v))
			} else {
				fmt.Printf("%-16s%-64s%-8s%-16s%-10d%-10s%-20s%-40s%-20s%-16s%-20s\n", v.Namespace, v.Name,
					GetContainerReadyStatus(v), string(v.Status.Phase),
					GetPodRestartCount(v),
					GetPodAge(v), v.Status.PodIP, v.Spec.NodeName,
					GetNominatedNode(v),
					GetReadinessGateStatus(v),
					GetPodLabel(v))
			}
		}
	}
}
