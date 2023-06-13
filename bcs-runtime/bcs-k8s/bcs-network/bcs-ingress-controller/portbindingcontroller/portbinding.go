/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package portbindingcontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	bcsnetcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type portBindingHandler struct {
	ctx         context.Context
	k8sClient   client.Client
	eventer     record.EventRecorder
	itemHandler *portBindingItemHandler
}

func newPortBindingHandler(ctx context.Context, k8sClient client.Client, eventer record.EventRecorder) *portBindingHandler {
	return &portBindingHandler{
		ctx:         ctx,
		k8sClient:   k8sClient,
		itemHandler: newPortBindingItemHandler(ctx, k8sClient),
		eventer:     eventer,
	}
}

// the returned bool value indicates whether you need to retry
func (pbh *portBindingHandler) ensurePortBinding(
	pod *k8scorev1.Pod, portBinding *networkextensionv1.PortBinding) (bool, error) {
	if portBinding == nil {
		blog.Warnf("port binding for pod '%s/%s' is empty", pod.GetNamespace(), pod.GetName())
		return false, nil
	}
	var newBindingStatusList []*networkextensionv1.PortBindingStatusItem
	for _, item := range portBinding.Spec.PortBindingList {
		var curStatus *networkextensionv1.PortBindingStatusItem
		// 找到和spec中item对应的status
		for _, tmpStatus := range portBinding.Status.PortBindingStatusList {
			if tmpStatus.PoolName == item.PoolName &&
				tmpStatus.PoolNamespace == item.PoolNamespace &&
				tmpStatus.PoolItemName == item.PoolItemName &&
				tmpStatus.StartPort == item.StartPort &&
				tmpStatus.EndPort == item.EndPort {
				curStatus = tmpStatus
			}
		}
		itemStatus := pbh.itemHandler.ensureItem(pod, item, curStatus)
		newBindingStatusList = append(newBindingStatusList, itemStatus)
	}
	portBinding.Status.PortBindingStatusList = newBindingStatusList
	retry := false
	unreadyNum := 0
	// 不断重试等待所有item对应的监听器就绪
	for _, status := range portBinding.Status.PortBindingStatusList {
		if status.Status != constant.PortBindingItemStatusReady {
			unreadyNum++
			retry = true
		}
	}

	rawStatus := portBinding.Status.Status
	if unreadyNum == 0 {
		portBinding.Status.Status = constant.PortBindingStatusReady
		pbh.recordEvent(portBinding, k8scorev1.EventTypeNormal, ReasonPortBindingReady, MsgPortBindingReady)
	} else {
		portBinding.Status.Status = constant.PortBindingStatusNotReady
		pbh.recordEvent(portBinding, k8scorev1.EventTypeNormal, ReasonPortBindingNotReady,
			fmt.Sprintf(MsgPortBindingNotReady, unreadyNum))
	}
	updateStatus := portBinding.Status.Status

	if err := pbh.k8sClient.Status().Update(context.Background(), portBinding, &client.UpdateOptions{}); err != nil {
		return true, fmt.Errorf("ensure port binding %s/%s failed, err %s",
			portBinding.GetName(), portBinding.GetNamespace(), err.Error())
	}

	// 根据portBinding status更新相关状态
	if err := pbh.postPortBindingUpdateStatus(rawStatus, updateStatus, portBinding); err != nil {
		return true, err
	}
	if err := pbh.updatePodCondition(pod, portBinding.Status.Status); err != nil {
		return true, err
	}
	if err := pbh.patchPodAnnotation(pod, portBinding.Status.Status); err != nil {
		return true, err
	}

	// 当portBinding的部分字段发生变化时，需要同步更新pod上的注解
	if err := pbh.ensurePod(pod, portBinding); err != nil {
		return true, errors.Wrapf(err, "ensurePod[%s/%s] failed", pod.GetNamespace(), pod.GetName())
	}
	pbh.recordEvent(portBinding, k8scorev1.EventTypeNormal, ReasonPortBindingUpdatePodSuccess,
		MsgPortBindingUpdatePodSuccess)

	return retry, nil
}

// updatePodCondition 在pod.condition上记录portBinding的绑定状态
func (pbh *portBindingHandler) updatePodCondition(pod *k8scorev1.Pod, status string) error {
	if _, ok := pod.Annotations[constant.AnnotationForPortPoolReadinessGate]; !ok {
		return nil
	}
	found := false
	for i, condition := range pod.Status.Conditions {
		if condition.Type == constant.ConditionTypeBcsIngressPortBinding {
			if condition.Status == k8scorev1.ConditionFalse {
				if status == constant.PortBindingStatusReady {
					pod.Status.Conditions[i].Status = k8scorev1.ConditionTrue
					pod.Status.Conditions[i].Reason = constant.ConditionReasonReadyBcsIngressPortBinding
					pod.Status.Conditions[i].Message = constant.ConditionMessageReadyBcsIngressPortBinding
				} else {
					pod.Status.Conditions[i].Status = k8scorev1.ConditionFalse
					pod.Status.Conditions[i].Reason = constant.ConditionReasonNotReadyBcsIngressPortBinding
					pod.Status.Conditions[i].Message = constant.ConditionMessageNotReadyBcsIngressPortBinding
				}
			}
			found = true
			break
		}
	}
	if !found && status == constant.PortBindingStatusReady {
		pod.Status.Conditions = append(pod.Status.Conditions, k8scorev1.PodCondition{
			Type:    constant.ConditionTypeBcsIngressPortBinding,
			Status:  k8scorev1.ConditionTrue,
			Reason:  constant.ConditionReasonReadyBcsIngressPortBinding,
			Message: constant.ConditionMessageReadyBcsIngressPortBinding,
		})
	}
	if err := pbh.k8sClient.Status().Update(context.Background(), pod, &client.UpdateOptions{}); err != nil {
		blog.Warnf("update pod %s/%s condition failed, err %s", pod.GetName(), pod.GetNamespace(), err.Error())
		return fmt.Errorf("update pod %s/%s condition failed, err %s", pod.GetName(), pod.GetNamespace(), err.Error())
	}
	return nil
}

func (pbh *portBindingHandler) patchPodAnnotation(pod *k8scorev1.Pod, status string) error {
	rawPatch := client.RawPatch(k8stypes.MergePatchType, []byte(
		"{\"metadata\":{\"annotations\":{\""+constant.AnnotationForPortPoolBindingStatus+
			"\":\""+status+"\"}}}"))
	updatePod := &k8scorev1.Pod{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		},
	}
	if err := pbh.k8sClient.Patch(context.Background(), updatePod, rawPatch, &client.PatchOptions{}); err != nil {
		blog.Errorf("patch pod %s/%s annotation status failed, err %s", pod.GetName(), pod.GetNamespace(), err.Error())
		return fmt.Errorf("patch pod %s/%s annotation status failed, err %s",
			pod.GetName(), pod.GetNamespace(), err.Error())
	}
	return nil
}

func (pbh *portBindingHandler) patchPodBindingAnnotation(
	pod *k8scorev1.Pod, bindingItemList []*networkextensionv1.PortBindingItem,
) error {
	bindingItemListBytes, err := json.Marshal(bindingItemList)
	if err != nil {
		return errors.Wrapf(err, "marshal bindingItemList for pod '%s/%s' failed",
			pod.GetNamespace(), pod.GetName())
	}
	patchStruct := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				constant.AnnotationForPortPoolBindings: string(bindingItemListBytes),
			},
		},
	}
	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		return errors.Wrapf(err, "marshal patchStruct for pod '%s/%s' failed", pod.GetNamespace(), pod.GetName())
	}
	blog.V(5).Infof("marshaled patchStruct of pod '%s/%s', patchStruct: %s", pod.GetNamespace(),
		pod.GetName(), string(patchBytes))
	rawPatch := client.RawPatch(k8stypes.MergePatchType, patchBytes)
	updatePod := &k8scorev1.Pod{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		},
	}
	if err := pbh.k8sClient.Patch(context.Background(), updatePod, rawPatch, &client.PatchOptions{}); err != nil {
		return errors.Wrapf(err, "patch pod %s/%s annotation status failed, patcheStruct: %s", pod.GetName(),
			pod.GetNamespace(), string(patchBytes))
	}
	return nil
}

// the returned bool value indicates whether you need to retry
func (pbh *portBindingHandler) cleanPortBinding(portBinding *networkextensionv1.PortBinding) (bool, error) {
	if portBinding == nil {
		blog.Warnf("port binding is empty")
		return false, nil
	}
	portBinding.Status.PortBindingStatusList = nil
	for _, item := range portBinding.Spec.PortBindingList {
		// 将item对应监听器的targetGroup重新设置为空
		itemStatus := pbh.itemHandler.deleteItem(item)
		portBinding.Status.PortBindingStatusList = append(portBinding.Status.PortBindingStatusList, itemStatus)
	}
	notCleanedNum := 0
	for _, status := range portBinding.Status.PortBindingStatusList {
		if status.Status != constant.PortBindingItemStatusCleaned {
			notCleanedNum++
		}
	}
	if notCleanedNum == 0 {
		portBinding.Status.Status = constant.PortBindingStatusCleaned
		pbh.recordEvent(portBinding, k8scorev1.EventTypeNormal, ReasonPortBindingCleanSuccess, MsgPortBindingCleanSuccess)
	} else {
		portBinding.Status.Status = constant.PortBindingStatusCleaning
		pbh.recordEvent(portBinding, k8scorev1.EventTypeWarning, ReasonPortBindingCleaning,
			fmt.Sprintf(MsgPortBindingCleaning, notCleanedNum))
	}
	portBinding.Status.UpdateTime = bcsnetcommon.FormatTime(time.Now())
	if err := pbh.k8sClient.Status().Update(
		context.Background(), portBinding, &client.UpdateOptions{}); err != nil {
		return true, fmt.Errorf("update port binding %s/%s when delete failed, err %s",
			portBinding.GetName(), portBinding.GetNamespace(), err.Error())
	}

	return notCleanedNum != 0, nil
}

// ensurePod update pod annotation if portBinding related field changed
func (pbh *portBindingHandler) ensurePod(pod *k8scorev1.Pod, portBinding *networkextensionv1.PortBinding) error {
	portBindingItemMap := make(map[string]*networkextensionv1.PortBindingItem)
	for _, portBindingItem := range portBinding.Spec.PortBindingList {
		portBindingItemMap[genUniqueIDOfPortBindingItem(portBindingItem)] = portBindingItem
	}

	podPortBindingList, err := parsePoolBindingsAnnotation(pod)
	if err != nil {
		return errors.Wrapf(err, "parse pod annotations for bindingItems failed")
	}

	// if portBinding.External changed, update pod's annotation
	changed := false
	for idx, podPortBindingItem := range podPortBindingList {
		portBindingItem, ok := portBindingItemMap[genUniqueIDOfPortBindingItem(podPortBindingItem)]
		if !ok {
			blog.Warnf("pod's portBindingItem(in annotation) not found in PortBinding, pod: %s/%s, item: %s",
				pod.GetNamespace(), pod.GetName(), genUniqueIDOfPortBindingItem(podPortBindingItem))
			continue
		}
		if portBindingItem == nil || podPortBindingItem == nil {
			blog.Warnf("nil portBindingItem, pod:%s/%s", pod.GetNamespace(), pod.GetName())
			continue
		}

		if podPortBindingItem.External != portBindingItem.External {
			podPortBindingList[idx].External = portBindingItem.External
			changed = true
		}
		if !reflect.DeepEqual(podPortBindingItem.LoadBalancerIDs, portBindingItem.LoadBalancerIDs) {
			podPortBindingList[idx].LoadBalancerIDs = portBindingItem.LoadBalancerIDs
			changed = true
		}
		if !reflect.DeepEqual(podPortBindingItem.PoolItemLoadBalancers, portBindingItem.PoolItemLoadBalancers) {
			podPortBindingList[idx].PoolItemLoadBalancers = portBindingItem.PoolItemLoadBalancers
			changed = true
		}
	}
	if changed {
		blog.Info("pod[%s/%s] PortBindingItem.External changed", pod.GetNamespace(), pod.GetName())
		if err := pbh.patchPodBindingAnnotation(pod, podPortBindingList); err != nil {
			return errors.Wrapf(err, "patch pod[%s/%s] for binding annotation failed",
				pod.GetNamespace(), pod.GetName())
		}
	}

	return nil
}

// patchPortBindingAnnotation patch annotation to portbinding
func (pbh *portBindingHandler) patchPortBindingAnnotation(
	portbinding *networkextensionv1.PortBinding, notReadyTimestamp string,
) error {
	patchStruct := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				// 记录portBinding最近一次变为NotReady的时间
				constant.AnnotationForPortBindingNotReadyTimestamp: notReadyTimestamp,
			},
		},
	}
	patchBytes, err := json.Marshal(patchStruct)
	if err != nil {
		return errors.Wrapf(err, "marshal patchStruct for portbinding '%s/%s' failed", portbinding.GetNamespace(),
			portbinding.GetName())
	}
	rawPatch := client.RawPatch(k8stypes.MergePatchType, patchBytes)
	updatePortBinding := &networkextensionv1.PortBinding{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      portbinding.GetName(),
			Namespace: portbinding.GetNamespace(),
		},
	}
	if err := pbh.k8sClient.Patch(context.Background(), updatePortBinding, rawPatch, &client.PatchOptions{}); err != nil {
		return errors.Wrapf(err, "patch portbinding %s/%s annotation failed, patcheStruct: %s",
			portbinding.GetNamespace(), portbinding.GetName(), string(patchBytes))
	}
	return nil
}

// 处理portBinding状态变化
func (pbh *portBindingHandler) postPortBindingUpdateStatus(rawStatus, updateStatus string,
	portBinding *networkextensionv1.PortBinding) error {
	if rawStatus == updateStatus {
		return nil
	}

	// 如果portBinding状态由NotReady/nil转为Ready,则统计Ready时间并清理NotReady时间戳
	if updateStatus == constant.PortBindingStatusReady {
		if notReadyTimeStr, ok := portBinding.Annotations[constant.
			AnnotationForPortBindingNotReadyTimestamp]; ok && notReadyTimeStr != "" {
			if notReadyTime, err := time.Parse(time.RFC3339Nano, notReadyTimeStr); err != nil {
				blog.Warnf("parse not ready timestamp failed, err: %s", err.Error())
			} else {
				// 上报绑定时间到Metric
				metrics.ReportPortBindMetric(notReadyTime)
			}
			if err := pbh.patchPortBindingAnnotation(portBinding, ""); err != nil {
				blog.Warnf(err.Error())
				return err
			}
		}
	} else if updateStatus == constant.PortBindingStatusNotReady {
		// 如果portBinding状态由Ready/nil转为Not Ready,则设置NotReady时间戳
		if err := pbh.patchPortBindingAnnotation(portBinding, time.Now().Format(time.RFC3339Nano)); err != nil {
			blog.Warnf(err.Error())
			return err
		}
	}

	return nil
}
