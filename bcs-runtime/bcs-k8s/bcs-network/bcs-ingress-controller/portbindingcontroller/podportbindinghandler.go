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

package portbindingcontroller

import (
	"context"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
)

// PodPortBindingHandler handle portbinding related to pods
type PodPortBindingHandler struct {
	pod *k8scorev1.Pod

	*portBindingHandler
}

func newPodPortBindingHandler(ctx context.Context, k8sClient client.Client, eventer record.EventRecorder,
	pod *k8scorev1.Pod) *PodPortBindingHandler {
	ppbh := &PodPortBindingHandler{pod: pod}
	ppbh.portBindingHandler = newPortBindingHandler(ctx, k8sClient, eventer)
	ppbh.portBindingHandler.generateTargetGroup = ppbh.generateTargetGroup
	ppbh.portBindingHandler.postPortBindingUpdate = ppbh.postPortBindingUpdate
	ppbh.portBindingHandler.postPortBindingClean = ppbh.postPortBindingClean
	ppbh.portBindingHandler.portBindingType = networkextensionv1.PortBindingTypePod

	return ppbh
}

func (p *PodPortBindingHandler) generateTargetGroup(item *networkextensionv1.PortBindingItem) *networkextensionv1.
	ListenerTargetGroup {
	if p.pod == nil {
		blog.Warnf("%v", errors.New("empty node"))
		return nil
	}
	backend := networkextensionv1.ListenerBackend{
		IP:     p.pod.Status.PodIP,
		Port:   item.RsStartPort,
		Weight: networkextensionv1.DefaultWeight,
	}
	if hostPort := generator.GetPodHostPortByPort(p.pod, int32(item.RsStartPort)); item.HostPort &&
		hostPort != 0 {
		backend.IP = p.pod.Status.HostIP
		backend.Port = int(hostPort)
	}
	return &networkextensionv1.ListenerTargetGroup{
		TargetGroupProtocol: item.Protocol,
		Backends:            []networkextensionv1.ListenerBackend{backend},
	}
}

func (p *PodPortBindingHandler) postPortBindingUpdate(portBinding *networkextensionv1.PortBinding) error {
	if p.pod == nil {
		err := errors.New("empty pod")
		blog.Warnf("%v", err)
		return err
	}
	if err := p.updatePodCondition(p.pod, portBinding.Status.Status); err != nil {
		return err
	}
	if err := p.patchPodAnnotation(p.pod, portBinding.Status.Status); err != nil {
		return err
	}

	if err := p.ensurePod(p.pod, portBinding); err != nil {
		return errors.Wrapf(err, "ensurePod[%s/%s] failed", p.pod.GetNamespace(), p.pod.GetName())
	}
	p.recordEvent(portBinding, k8scorev1.EventTypeNormal, ReasonPortBindingUpdatePodSuccess,
		MsgPortBindingUpdatePodSuccess)
	return nil
}

func (p *PodPortBindingHandler) postPortBindingClean(portBinding *networkextensionv1.PortBinding) error {
	// 使用statefulset + 端口保留的情况下， 清理注解可能导致误清理新建Pod的注解
	// 由于Pod只在创建时会分配端口，因此不需要清理注解
	// pod := &k8scorev1.Pod{}
	// if err := p.k8sClient.Get(p.ctx, k8stypes.NamespacedName{Namespace: portBinding.GetNamespace(),
	// 	Name: portBinding.GetName()}, pod); err != nil {
	// 	if k8serrors.IsNotFound(err) {
	// 		blog.Infof("pod '%s/%s' has been deleted, do not clean annotation", portBinding.GetNamespace(),
	// 			portBinding.GetName())
	// 		return nil
	// 	}
	// 	blog.Warnf("get pod '%s/%s' failed, err %s", portBinding.GetNamespace(), portBinding.GetName(),
	// 		err.Error())
	// 	return errors.Wrapf(err, "get pod '%s/%s' failed", portBinding.GetNamespace(), portBinding.GetName())
	// }
	//
	// delete(pod.Annotations, constant.AnnotationForPortPoolBindings)
	// delete(pod.Annotations, constant.AnnotationForPortPoolBindingStatus)
	// if err := p.k8sClient.Update(context.TODO(), pod, &client.UpdateOptions{}); err != nil {
	// 	blog.Warnf("remove annotation from pod %s/%s failed, err %s", portBinding.GetName(),
	// 		portBinding.GetNamespace(), err.Error())
	// 	return errors.Wrapf(err, "remove annotation from pod %s/%s failed", portBinding.GetName(),
	// 		portBinding.GetNamespace())
	// }
	return nil
}

// updatePodCondition 在pod.condition上记录portBinding的绑定状态
func (p *PodPortBindingHandler) updatePodCondition(pod *k8scorev1.Pod, status string) error {
	if _, ok := pod.Annotations[constant.AnnotationForPortPoolReadinessGate]; !ok {
		return nil
	}
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newPod := &k8scorev1.Pod{}
		if err := p.k8sClient.Get(p.ctx, k8stypes.NamespacedName{Namespace: pod.GetNamespace(),
			Name: pod.GetName()}, newPod); err != nil {
			return err
		}

		found := false
		for i, condition := range newPod.Status.Conditions {
			if condition.Type == constant.ConditionTypeBcsIngressPortBinding {
				if condition.Status == k8scorev1.ConditionFalse {
					if status == constant.PortBindingStatusReady {
						newPod.Status.Conditions[i].Status = k8scorev1.ConditionTrue
						newPod.Status.Conditions[i].Reason = constant.ConditionReasonReadyBcsIngressPortBinding
						newPod.Status.Conditions[i].Message = constant.ConditionMessageReadyBcsIngressPortBinding
					} else {
						newPod.Status.Conditions[i].Status = k8scorev1.ConditionFalse
						newPod.Status.Conditions[i].Reason = constant.ConditionReasonNotReadyBcsIngressPortBinding
						newPod.Status.Conditions[i].Message = constant.ConditionMessageNotReadyBcsIngressPortBinding
					}
				}
				found = true
				break
			}
		}
		if !found && status == constant.PortBindingStatusReady {
			newPod.Status.Conditions = append(newPod.Status.Conditions, k8scorev1.PodCondition{
				Type:    constant.ConditionTypeBcsIngressPortBinding,
				Status:  k8scorev1.ConditionTrue,
				Reason:  constant.ConditionReasonReadyBcsIngressPortBinding,
				Message: constant.ConditionMessageReadyBcsIngressPortBinding,
			})
		}
		if err := p.k8sClient.Status().Update(context.Background(), newPod, &client.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = errors.Wrapf(err, "update pod %s/%s condition failed", pod.GetNamespace(), pod.GetName())
		blog.Warnf("%s", err.Error())
		return err
	}
	return nil
}

func (p *PodPortBindingHandler) patchPodAnnotation(pod *k8scorev1.Pod, status string) error {
	rawPatch := client.RawPatch(k8stypes.MergePatchType, []byte(
		"{\"metadata\":{\"annotations\":{\""+constant.AnnotationForPortPoolBindingStatus+
			"\":\""+status+"\"}}}"))
	updatePod := &k8scorev1.Pod{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		},
	}
	if err := p.k8sClient.Patch(context.Background(), updatePod, rawPatch, &client.PatchOptions{}); err != nil {
		err = errors.Wrapf(err, "patch pod %s/%s annotation status failed", pod.GetName(), pod.GetNamespace())
		blog.Errorf("%v", err)
		return err
	}
	return nil
}

// ensurePod update pod annotation if portBinding related field changed
func (p *PodPortBindingHandler) ensurePod(pod *k8scorev1.Pod, portBinding *networkextensionv1.PortBinding) error {
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
		if err := p.patchPodBindingAnnotation(pod, podPortBindingList); err != nil {
			return err
		}
	}

	return nil
}
