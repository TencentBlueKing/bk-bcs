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

package cachemanager

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
)

// BuildCalculatorRequest will build the request of calculator
func (m *CacheManager) BuildCalculatorRequest(ctx context.Context) (*calculator.CalculateConvergeRequest, error) {
	podsItems, err := m.filterPods(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "filter all pods failed")
	}
	requestPods, workloadPods, nodeDecreasePods, err := m.buildPods(ctx, podsItems)
	if err != nil {
		return nil, errors.Wrapf(err, "get all pods failed")
	}
	gzipPodBS, err := m.gzipPods(requestPods)
	if err != nil {
		return nil, errors.Wrapf(err, "gzip pods failed")
	}

	nodeItems, err := m.filterNodes(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "filter all nodes failed")
	}
	requestNodes, err := m.buildNodes(ctx, nodeItems, nodeDecreasePods)
	if err != nil {
		return nil, errors.Wrapf(err, "get all nodes failed")
	}
	gzipNodeBS, err := m.gzipNodes(requestNodes)
	if err != nil {
		return nil, errors.Wrapf(err, "gzip nodes failed")
	}

	timeNow, _ := strconv.ParseInt(time.Now().Format("20060102"), 10, 64)
	req := &calculator.CalculateConvergeRequest{
		AuthenticationMethod: "token",
		Token:                m.op.BKDataToken,
		AppCode:              m.op.BKDataAppCode,
		AppSecret:            m.op.BKDataAppSecret,
		Data: &calculator.CalculateData{
			Inputs: []calculator.RequestInputs{
				{
					Pod:  string(gzipPodBS),
					Node: string(gzipNodeBS),
					Time: timeNow,
				},
			},
		},
		Config: m.buildConfig(workloadPods),
		Original: &calculator.CalculateOriginalData{
			Pods:  podsItems,
			Nodes: nodeItems,
		},
	}
	blog.V(4).Infof("Calculator request built: %s", req.String())
	return req, nil
}

var (
	defaultListAllPods = time.Duration(30) * time.Second
)

func (m *CacheManager) gzipPods(requestPods []map[string]interface{}) ([]byte, error) {
	podBS, err := json.Marshal(requestPods)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal request pods failed")
	}
	podBSBS, err := json.Marshal(string(podBS))
	if err != nil {
		return nil, errors.Wrapf(err, "marshal pods bytes failed")
	}
	gzipPodBS, err := utils.GzipAndBase64Bytes(podBSBS)
	if err != nil {
		return nil, errors.Wrapf(err, "gzip rquest pods failed")
	}
	return gzipPodBS, nil
}

func (m *CacheManager) gzipNodes(requestNodes []map[string]interface{}) ([]byte, error) {
	nodeBS, err := json.Marshal(requestNodes)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal request nodes failed")
	}
	nodeBSBS, err := json.Marshal(string(nodeBS))
	if err != nil {
		return nil, errors.Wrapf(err, "marshal failed")
	}
	gzipNodeBS, err := utils.GzipAndBase64Bytes(nodeBSBS)
	if err != nil {
		return nil, errors.Wrapf(err, "gzip request nodes failed")
	}
	return gzipNodeBS, nil
}

func (m *CacheManager) filterPods(ctx context.Context) ([]*calculator.PodItem, error) {
	queryCtx, queryCancel := context.WithTimeout(ctx, defaultListAllPods)
	defer queryCancel()
	pods, err := m.ListPods(queryCtx, "", labels.NewSelector())
	if err != nil {
		return nil, errors.Wrapf(err, "list all pods failed")
	}
	pdbPodsMap, err := m.ListPDBPods(ctx, "")
	if err != nil {
		return nil, errors.Wrapf(err, "list pdb pods failed")
	}

	items := make([]*calculator.PodItem, 0, len(pods))
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded ||
			pod.Status.Phase == corev1.PodUnknown || pod.Status.HostIP == "" || pod.Spec.NodeName == "" {
			continue
		}
		items = append(items, &calculator.PodItem{
			Item:              apis.PodName(pod.Namespace, pod.Name),
			Index1:            getPodMemory(pod),
			Index2:            getPodCPU(pod),
			Container:         pod.Spec.NodeName,
			IsAllowMigrate:    m.checkPodAllowMigrate(pod, pdbPodsMap),
			MigrationPriority: 0,
			OriginalPod:       pod,
		})
	}
	return items, nil
}

func (m *CacheManager) buildPods(ctx context.Context, podItems []*calculator.PodItem) ([]map[string]interface{},
	map[string][]*corev1.Pod, map[string][]*corev1.Pod, error) {
	nodeDecreasePods := make(map[string][]*corev1.Pod)
	workloadPods := make(map[string][]*corev1.Pod)
	requestPods := make([]map[string]interface{}, 0, len(podItems))
	for _, podItem := range podItems {
		originalPod := podItem.OriginalPod
		// we should record the pod's cpu/mem, if pod not allow migrated
		if podItem.IsAllowMigrate == 0 {
			nodeName := originalPod.Spec.NodeName
			nodeDecreasePods[nodeName] = append(nodeDecreasePods[nodeName], podItem.OriginalPod)
			continue
		}
		bs, err := json.Marshal(podItem)
		if err != nil {
			blog.Warnf("build pods marshal pod '%s/%s' failed: %s",
				originalPod.Namespace, originalPod.Name, err.Error())
			continue
		}
		podMap := make(map[string]interface{})
		if err = json.Unmarshal(bs, &podMap); err != nil {
			blog.Warnf("build pods unmarshal failed: %s", err.Error())
			continue
		}
		for k, v := range originalPod.Labels {
			podMap[k] = v
		}
		ownerName, err := m.getPodOwnerName(ctx, originalPod)
		if err != nil {
			blog.Warnf("build pods get pod '%s/%s' owner name failed: %s",
				originalPod.Namespace, originalPod.Name, err.Error())
		} else {
			podMap[apis.WorkloadName] = ownerName
			workloadPods[ownerName] = append(workloadPods[ownerName], originalPod)
		}
		podMap[apis.PodNameLabel] = originalPod.Name
		podMap[apis.NodeNameLabel] = originalPod.Spec.Hostname
		podMap[apis.PodNamespaceLabel] = originalPod.Namespace
		requestPods = append(requestPods, podMap)
	}
	blog.V(4).Infof("BuildRequest build pods success.")
	return requestPods, workloadPods, nodeDecreasePods, nil
}

func (m *CacheManager) getPodOwnerName(ctx context.Context, pod *corev1.Pod) (string, error) {
	podCtx, podCancel := context.WithTimeout(ctx, apis.DefaultQueryTimeout)
	defer podCancel()
	_, ownerName, err := m.GetPodOwnerName(podCtx, pod.Namespace, pod.Name)
	if err != nil {
		return ownerName, errors.Wrapf(err, "get pod's ownerName failed")
	}
	return ownerName, err
}

func (m *CacheManager) filterNodes(ctx context.Context) ([]*calculator.NodeItem, error) {
	queryCtx, queryCancel := context.WithTimeout(ctx, apis.DefaultQueryTimeout)
	defer queryCancel()
	nodes, err := m.ListNodes(queryCtx, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "list all nodes failed")
	}
	items := make([]*calculator.NodeItem, 0, len(nodes))
	for _, node := range nodes {
		// filter node if type is master
		if v, ok := node.Labels[apis.NodeMasterLabel]; ok && v == "true" {
			continue
		}
		items = append(items, &calculator.NodeItem{
			Container:      node.Name,
			Index1:         getNodeMemory(node),
			Index2:         getNodeCPU(node),
			ItemNums:       64,
			IsAllowMigrate: m.checkNodeAllowMigrate(node),
			OriginalNode:   node,
		})
	}
	return items, nil
}

func (m *CacheManager) buildNodes(ctx context.Context, nodeItems []*calculator.NodeItem,
	nodeDecreasePods map[string][]*corev1.Pod) ([]map[string]interface{}, error) {
	requestNodes := make([]map[string]interface{}, 0, len(nodeItems))
	for _, nodeItem := range nodeItems {
		originalNode := nodeItem.OriginalNode
		if pods, ok := nodeDecreasePods[originalNode.Name]; ok {
			for _, pod := range pods {
				nodeItem.Index1 -= getPodMemory(pod)
				nodeItem.Index2 -= getPodCPU(pod)
			}
		}
		bs, err := json.Marshal(nodeItem)
		if err != nil {
			blog.Warnf("build nodes marshal node '%s/%s' failed: %s", originalNode.Name, err.Error())
			continue
		}
		nodeMap := make(map[string]interface{})
		if err = json.Unmarshal(bs, &nodeMap); err != nil {
			blog.Warnf("build nodes unmarshal failed: %s", err.Error())
			continue
		}
		for k, v := range originalNode.Labels {
			nodeMap[k] = v
		}
		requestNodes = append(requestNodes, nodeMap)
	}
	blog.V(4).Infof("BuildRequest build nodes success.")
	return requestNodes, nil
}

func (m *CacheManager) buildConfig(workloadPods map[string][]*corev1.Pod) *calculator.CalculateConfig {
	scope := calculator.PredictScope{}
	for workload, pods := range workloadPods {
		pod := pods[0]
		if len(pod.Spec.NodeSelector) != 0 {
			scope.ContainerAffinity = append(scope.ContainerAffinity, buildNodeSelector(workload, pod))
		}
		affinity := pod.Spec.Affinity
		if affinity == nil {
			continue
		}
		if affinity.NodeAffinity != nil {
			scope.ContainerAffinity = append(scope.ContainerAffinity, buildNodeAffinity(workload, pod)...)
		}
		if affinity.PodAffinity != nil {
			scope.ItemAffinity = append(scope.ItemAffinity, buildPodAffinity(workload, pod)...)
		}
		if affinity.PodAntiAffinity != nil {
			scope.ItemAntiAffinity = append(scope.ItemAffinity, buildPodAntiAffinity(workload, pod)...)
		}
	}
	blog.V(4).Infof("BuildRequest build config success.")
	return &calculator.CalculateConfig{
		PredictArgs: calculator.PredictArgs{
			Scope: scope,
			OptimizeTarget: []calculator.OptimizeTarget{
				{
					Name:              "index1",
					OptimizeDirection: "max",
					FieldType:         "double",
				},
				{
					Name:              "index2",
					OptimizeDirection: "max",
					FieldType:         "double",
				},
				{
					Name:              "cost",
					OptimizeDirection: "min",
					FieldType:         "double",
				},
				{
					Name:              "container_released",
					OptimizeDirection: "max",
					FieldType:         "int",
				},
			},
			IterationLimit:     10,
			PopulationSize:     5,
			MigrationCostLimit: 1,
			MigrationWaterline: 0.8,
			MigrationDegree:    "all",
			IsCompressed:       1,
		},
	}
}

func buildNodeSelector(workload string, pod *corev1.Pod) calculator.Affinity {
	itemConditions := []calculator.Condition{
		{
			Table:         "item",
			Col:           apis.WorkloadName,
			ConditionType: "=",
			Value:         workload,
		},
	}
	containerConditions := make([]calculator.Condition, 0, len(pod.Spec.NodeSelector))
	for k, v := range pod.Spec.NodeSelector {
		containerConditions = append(containerConditions, calculator.Condition{
			Table:         "container",
			Col:           k,
			ConditionType: "=",
			Value:         v,
		})
	}
	return calculator.Affinity{
		ItemCondition:      itemConditions,
		ContainerCondition: containerConditions,
		IsForced:           true,
	}
}

func buildNodeAffinity(workload string, pod *corev1.Pod) []calculator.Affinity {
	result := make([]calculator.Affinity, 0)
	itemConditions := []calculator.Condition{
		{
			Table:         "item",
			Col:           apis.WorkloadName,
			ConditionType: "=",
			Value:         workload,
		},
	}
	nodeAffinity := pod.Spec.Affinity.NodeAffinity
	if required := nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution; required != nil {
		containerConditions := buildContainerConditions(required.NodeSelectorTerms)
		if len(containerConditions) != 0 {
			result = append(result, calculator.Affinity{
				ItemCondition:      itemConditions,
				ContainerCondition: containerConditions,
				IsForced:           true,
			})
		}
	}

	preferred := nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	preferredTerms := make([]corev1.NodeSelectorTerm, 0, len(preferred))
	for _, term := range preferred {
		preferredTerms = append(preferredTerms, term.Preference)
	}
	if len(preferredTerms) != 0 {
		containerConditions := buildContainerConditions(preferredTerms)
		if len(containerConditions) != 0 {
			result = append(result, calculator.Affinity{
				ItemCondition:      itemConditions,
				ContainerCondition: buildContainerConditions(preferredTerms),
				IsForced:           false,
			})
		}
	}
	return result
}

func buildContainerConditions(terms []corev1.NodeSelectorTerm) []calculator.Condition {
	containerConditions := make([]calculator.Condition, 0)
	for _, term := range terms {
		for _, label := range term.MatchExpressions {
			conditionType := getConditionTypeWithNodeSelector(label.Operator)
			if conditionType == "" {
				continue
			}
			containerConditions = append(containerConditions, calculator.Condition{
				Table:         "container",
				Col:           label.Key,
				ConditionType: conditionType,
				Value:         label.Values,
			})
		}
		// TODO https://github.com/kubernetes/kubernetes/blob/master/pkg/apis/core/v1/conversion.go#L33
		// for _, field := range term.MatchFields {}
	}
	return containerConditions
}

func getConditionTypeWithNodeSelector(operator corev1.NodeSelectorOperator) string {
	condition := ""
	switch operator {
	case corev1.NodeSelectorOpIn:
		condition = "IN"
	case corev1.NodeSelectorOpNotIn:
		condition = "NOT IN"
	case corev1.NodeSelectorOpExists:
		condition = "="
	case corev1.NodeSelectorOpDoesNotExist:
		condition = "!="
	case corev1.NodeSelectorOpGt:
		// TODO
	case corev1.NodeSelectorOpLt:
		// TODO
	}
	return condition
}

func buildPodAffinity(workload string, pod *corev1.Pod) []calculator.Affinity {
	result := make([]calculator.Affinity, 0)
	originalItemConditions := []calculator.Condition{
		{
			Table:         "item",
			Col:           apis.WorkloadName,
			ConditionType: "=",
			Value:         workload,
		},
	}

	podAffinity := pod.Spec.Affinity.PodAffinity
	if required := podAffinity.RequiredDuringSchedulingIgnoredDuringExecution; len(required) != 0 {
		itemConditions := buildItemConditions(required)
		if len(itemConditions) != 0 {
			result = append(result, calculator.Affinity{
				ItemCondition: append(originalItemConditions, itemConditions...),
				IsForced:      true,
			})
		}
	}

	preferred := podAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	preferredTerms := make([]corev1.PodAffinityTerm, 0)
	for _, term := range preferred {
		preferredTerms = append(preferredTerms, term.PodAffinityTerm)
	}
	if len(preferredTerms) != 0 {
		result = append(result, calculator.Affinity{
			ItemCondition: append(originalItemConditions, buildItemConditions(preferredTerms)...),
			IsForced:      false,
		})
	}
	return result
}

func buildPodAntiAffinity(workload string, pod *corev1.Pod) []calculator.Affinity {
	result := make([]calculator.Affinity, 0)
	originalItemConditions := []calculator.Condition{
		{
			Table:         "item",
			Col:           apis.WorkloadName,
			ConditionType: "=",
			Value:         workload,
		},
	}

	podAffinity := pod.Spec.Affinity.PodAntiAffinity
	if required := podAffinity.RequiredDuringSchedulingIgnoredDuringExecution; len(required) != 0 {
		itemConditions := buildItemConditions(required)
		if len(itemConditions) != 0 {
			result = append(result, calculator.Affinity{
				ItemCondition: append(originalItemConditions, itemConditions...),
				IsForced:      true,
			})
		}
	}

	preferred := podAffinity.PreferredDuringSchedulingIgnoredDuringExecution
	preferredTerms := make([]corev1.PodAffinityTerm, 0)
	for _, term := range preferred {
		preferredTerms = append(preferredTerms, term.PodAffinityTerm)
	}
	if len(preferredTerms) != 0 {
		result = append(result, calculator.Affinity{
			ItemCondition: append(originalItemConditions, buildItemConditions(preferredTerms)...),
			IsForced:      false,
		})
	}
	return result
}

func buildItemConditions(terms []corev1.PodAffinityTerm) []calculator.Condition {
	itemConditions := make([]calculator.Condition, 0)
	for _, term := range terms {
		// TODO ignore namespaces/namespaceSelector
		// TODO ignore topologyKey
		if term.LabelSelector == nil {
			continue
		}
		for k, v := range term.LabelSelector.MatchLabels {
			itemConditions = append(itemConditions, calculator.Condition{
				Table:         "item",
				Col:           k,
				ConditionType: "=",
				Value:         v,
			})
		}
		for _, label := range term.LabelSelector.MatchExpressions {
			conditionType := getConditionTypeWithLabelSelector(label.Operator)
			if conditionType == "" {
				continue
			}
			itemConditions = append(itemConditions, calculator.Condition{
				Table:         "item",
				Col:           label.Key,
				ConditionType: conditionType,
				Value:         label.Values,
			})
		}
	}
	return itemConditions
}

func (m *CacheManager) checkPodAllowMigrate(pod *corev1.Pod, podsMap map[string]*corev1.Pod) int32 {
	// check pod is managed by pdb
	// TODO: 临时禁用
	//if _, ok := podsMap[apis.PodName(pod.Namespace, pod.Name)]; !ok {
	//	return 0
	//}
	// check pod namespace
	if _, ok := apis.NotAllowMigrateNamespace[pod.Namespace]; ok {
		return 0
	}
	// check pod owner
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == apis.DaemonSetKind {
			return 0
		}
	}
	return 1
}

func (m *CacheManager) checkNodeAllowMigrate(node *corev1.Node) int32 {
	return 1
}

func getConditionTypeWithLabelSelector(operator metav1.LabelSelectorOperator) string {
	condition := ""
	switch operator {
	case metav1.LabelSelectorOpIn:
		condition = "IN"
	case metav1.LabelSelectorOpNotIn:
		condition = "NOTIN"
	case metav1.LabelSelectorOpExists:
		condition = "="
	case metav1.LabelSelectorOpDoesNotExist:
		condition = "!="
	}
	return condition
}

func getPodMemory(pod *corev1.Pod) (mem float64) {
	cs := pod.Spec.Containers
	for _, c := range cs {
		mem += c.Resources.Requests.Memory().AsApproximateFloat64()
	}
	return mem
}

func getPodCPU(pod *corev1.Pod) (cpu float64) {
	cs := pod.Spec.Containers
	var cpus float64
	for _, c := range cs {
		cpus += c.Resources.Requests.Cpu().AsApproximateFloat64()
	}
	return cpus
}

func getNodeMemory(node *corev1.Node) float64 {
	return node.Status.Allocatable.Memory().AsApproximateFloat64()
}

func getNodeCPU(node *corev1.Node) float64 {
	return node.Status.Allocatable.Cpu().AsApproximateFloat64()
}
