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

// Package portbindingcontroller controller for portbinding
package portbindingcontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	ingresscommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/utils"
	bcsnetcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// PortBindingReconciler reconciler for bcs port pool
type PortBindingReconciler struct {
	cleanInterval time.Duration
	ctx           context.Context
	k8sClient     client.Client
	poolCache     *portpoolcache.Cache
	eventer       record.EventRecorder

	nodePortBindingNs string
	nodeBindCache     *NodePortBindingCache
}

// NewPortBindingReconciler create PortBindingReconciler
func NewPortBindingReconciler(ctx context.Context, cleanInterval time.Duration, k8sClient client.Client,
	poolCache *portpoolcache.Cache, eventer record.EventRecorder, nodePortBindingNs string,
	nodeBindCache *NodePortBindingCache) *PortBindingReconciler {
	return &PortBindingReconciler{
		ctx:               ctx,
		cleanInterval:     cleanInterval,
		k8sClient:         k8sClient,
		poolCache:         poolCache,
		eventer:           eventer,
		nodePortBindingNs: nodePortBindingNs,
		nodeBindCache:     nodeBindCache,
	}
}

// Reconcile reconcile port pool
// portbinding name is same with pod name
// nolint  funlen
func (pbr *PortBindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("PortBinding %+v triggered", req.NamespacedName) // found same namespacedName pod & node

	portBinding, pod, node, err := pbr.getResource(req.NamespacedName, req.Name)
	if err != nil {
		blog.Warnf("get pod %v or node %v failed, err %s, requeue it", req.NamespacedName, req.NamespacedName, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}

	var portBindingType string
	if pod != nil && checkPortPoolAnnotation(pod.Annotations) {
		portBindingType = networkextensionv1.PortBindingTypePod
	} else if node != nil && checkPortPoolAnnotation(node.Annotations) {
		portBindingType = networkextensionv1.PortBindingTypeNode
	} else if portBinding != nil {
		// if entity not found and portbinding found, do clean portbinding
		blog.V(3).Infof("clean portbinding %v", req.NamespacedName)
		return pbr.cleanPortBinding(portBinding)
	} else { // nolint
		// if both entity and portbinding are not found, just return
		return ctrl.Result{}, nil
	}

	if portBinding != nil {
		// when statefulset pod is recreated, the old portbinding may be deleting
		if portBinding.DeletionTimestamp != nil {
			blog.V(3).Infof("found deleting portbinding, continue clean portbinding %v", req.NamespacedName)
			return pbr.cleanPortBinding(portBinding)
		}

		// portBinding创建后， 标记使用的端口（定时checker认为没有标记的端口已经泄露， 等待一段时间后释放）
		for _, portBindingItem := range portBinding.Spec.PortBindingList {
			pbr.poolCache.SetPortBindingUsed(portBindingItem, portBindingType, portBinding.GetNamespace(),
				portBinding.GetName())
		}
	}

	var pbhandler iPortBindingHandler
	switch portBindingType {
	case networkextensionv1.PortBindingTypePod:
		metrics.ReportPortAllocate(pod.GetName(), pod.GetNamespace(), true)
		valid, res, err1 := pbr.podPortBindingPreCheck(portBindingType, portBinding, pod)
		if !valid {
			return res, err1
		}

		pbhandler = newPodPortBindingHandler(pbr.ctx, pbr.k8sClient, pbr.eventer, pod)
	case networkextensionv1.PortBindingTypeNode:
		if !pbr.checkPortAllocate(node) {
			return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 10}, nil
		}
		metrics.ReportPortAllocate(node.GetName(), node.GetNamespace(), true)
		valid, res, err1 := pbr.nodePortBindingPreCheck(portBindingType, portBinding, node)
		if !valid {
			return res, err1
		}
		pbhandler = newNodePortBindingHandler(pbr.ctx, pbr.k8sClient, pbr.eventer, node, pbr.nodeBindCache)
	default:
		blog.Warnf("unknown portBindingType: %s", portBindingType)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}

	retry, err := pbhandler.ensurePortBinding(portBinding)
	if err != nil {
		blog.Warnf("ensure port binding %s/%s failed, err %s",
			portBinding.GetName(), portBinding.GetNamespace(), err.Error())
		pbr.recordEvent(portBinding, k8scorev1.EventTypeWarning, ReasonPortBindingEnsureFailed,
			fmt.Sprintf(MsgPortBindingEnsureFailed, err.Error()))
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	if retry {
		blog.Infof("retry to wait portbinding finished")
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	blog.Infof("ensure port binding %s/%s successfully", portBinding.GetName(), portBinding.GetNamespace())
	return ctrl.Result{}, nil
}

func (pbr *PortBindingReconciler) getResource(namespacedName k8stypes.NamespacedName,
	name string) (*networkextensionv1.PortBinding, *k8scorev1.Pod, *k8scorev1.Node, error) {
	pb := &networkextensionv1.PortBinding{}
	var portBinding *networkextensionv1.PortBinding = nil
	if err := pbr.k8sClient.Get(pbr.ctx, namespacedName, pb); err != nil {
		// nolint
		if !k8serrors.IsNotFound(err) {
			return nil, nil, nil, fmt.Errorf("get portbinding %v failed, err %s, requeue it", namespacedName, err.Error())
		}
	} else {
		portBinding = pb.DeepCopy()
	}
	// found same namespacedName pod & node
	pod := &k8scorev1.Pod{}
	if err := pbr.k8sClient.Get(pbr.ctx, namespacedName, pod); err != nil {
		// nolint
		if !k8serrors.IsNotFound(err) {
			return nil, nil, nil, fmt.Errorf("get pod %v failed, err %s", namespacedName, err.Error())
		}
	} else {
		return portBinding, pod, nil, nil
	}
	node := &k8scorev1.Node{}
	if err := pbr.k8sClient.Get(pbr.ctx, k8stypes.NamespacedName{Name: name}, node); err != nil {
		// nolint
		if !k8serrors.IsNotFound(err) {
			return nil, nil, nil, fmt.Errorf("get node %v failed, err %s", name, err.Error())
		}
	} else {
		// node's namespace is empty, set to match node portbindings' namespace
		node.SetNamespace(pbr.nodePortBindingNs)
		return portBinding, nil, node, nil
	}

	return portBinding, nil, nil, nil
}

func (pbr *PortBindingReconciler) createPortBinding(portBindingType, namespace, name string,
	annotations map[string]string) (ctrl.Result, error) {
	annotationValue, ok := annotations[constant.AnnotationForPortPoolBindings]
	if !ok {
		blog.Warnf("%s '%s/%s' has no annotation %s",
			portBindingType, namespace, name, constant.AnnotationForPortPoolBindings)
		return ctrl.Result{}, nil
	}
	var portBindingList []*networkextensionv1.PortBindingItem
	if err := json.Unmarshal([]byte(annotationValue), &portBindingList); err != nil {
		blog.Warnf("internal logic err, decode value of %s '%s/%s' annotation %s is invalid, err %s, value %s",
			portBindingType, namespace, name, constant.AnnotationForPortPoolPorts, err.Error(), annotationValue)
		return ctrl.Result{}, nil
	}
	portBinding := &networkextensionv1.PortBinding{}
	portBinding.SetName(name)
	portBinding.SetNamespace(namespace)
	labels := make(map[string]string)
	for _, binding := range portBindingList {
		tmpKey := utils.GenPortBindingLabel(binding.PoolName, binding.PoolNamespace)
		labels[tmpKey] = binding.PoolItemName
	}
	labels[networkextensionv1.PortBindingTypeLabelKey] = portBindingType
	portBindingAnnotation := make(map[string]string)
	portBindingAnnotation[constant.AnnotationForPortBindingNotReadyTimestamp] = time.Now().Format(time.RFC3339Nano)
	if duration, ok := annotations[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration]; ok {
		portBindingAnnotation[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration] = duration
	}
	portBinding.SetLabels(labels)
	portBinding.SetAnnotations(portBindingAnnotation)
	portBinding.Finalizers = append(portBinding.Finalizers, constant.FinalizerNameBcsIngressController)
	portBinding.Spec = networkextensionv1.PortBindingSpec{
		PortBindingList: portBindingList,
	}
	portBinding.Status.PortBindingType = portBindingType
	portBinding.Status.Status = constant.PortBindingStatusNotReady

	if err := pbr.k8sClient.Create(context.Background(), portBinding, &client.CreateOptions{}); err != nil {
		blog.Warnf("failed to create port binding object, err %s", err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	pbr.recordEvent(portBinding, k8scorev1.EventTypeNormal, ReasonPortBindingCreatSuccess, MsgPortBindingCreateSuccess)
	return ctrl.Result{}, nil
}

// cleanPortBinding clean portbinding resource
// 删除portBinding顺序
// 1. 删除Pod，进入clean流程
// 2. 根据portBinding的item，清理相关的listener资源
// 3. 等待所有item清理完毕后，记portBinding status为cleaned
// 4. delete portBinding（加上DeletionTimeStamp）
// 5. 移除portBinding Finalizers, 并从缓存中释放占用的端口
func (pbr *PortBindingReconciler) cleanPortBinding(portBinding *networkextensionv1.PortBinding) (ctrl.Result, error) {
	if portBinding.Status.Status == constant.PortBindingStatusCleaned {
		// 支持绑定端口保留，如果在expired内pod重新创建，还会复用相同的portBinding数据
		expired, err := isPortBindingExpired(portBinding)
		if !expired && err == nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: pbr.cleanInterval,
			}, nil
		}
		if err != nil {
			blog.Warnf("check port binding expire time failed, err %s", err.Error())
		}
		if portBinding.DeletionTimestamp != nil {
			if len(portBinding.Finalizers) != 1 || portBinding.Finalizers[0] != constant.FinalizerNameBcsIngressController {
				blog.Infof("finalizers of port binding %s/%s is not cleaned, skip clean port binding",
					portBinding.GetNamespace(), portBinding.GetName())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 3 * time.Second,
				}, nil
			}
			blog.V(3).Infof("removing finalizer from port binding %s/%s",
				portBinding.GetName(), portBinding.GetNamespace())
			portBinding.Finalizers = bcsnetcommon.RemoveString(
				portBinding.Finalizers, constant.FinalizerNameBcsIngressController)
			if err := pbr.k8sClient.Update(pbr.ctx, portBinding, &client.UpdateOptions{}); err != nil {
				blog.Warnf("remove finalizer from port binding %s/%s failed, err %s",
					portBinding.GetName(), portBinding.GetNamespace(), err.Error())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 3 * time.Second,
				}, nil
			}
			metrics.CleanPortAllocateMetric(portBinding.GetName(), portBinding.GetNamespace())
			pbr.poolCache.Lock()
			defer pbr.poolCache.Unlock()
			for _, portBindingItem := range portBinding.Spec.PortBindingList {
				poolKey := ingresscommon.GetNamespacedNameKey(portBindingItem.PoolName, portBindingItem.PoolNamespace)
				blog.Infof("release portbinding %s %s %s %d %d from cache",
					poolKey, portBindingItem.GetKey(), portBindingItem.Protocol,
					portBindingItem.StartPort, portBindingItem.EndPort)
				pbr.poolCache.ReleasePortBinding(
					poolKey, portBindingItem.GetKey(), portBindingItem.Protocol,
					portBindingItem.StartPort, portBindingItem.EndPort)
			}

		} else {
			blog.V(3).Infof("delete port binding %s/%s from apiserver",
				portBinding.GetName(), portBinding.GetNamespace())
			if err := pbr.k8sClient.Delete(pbr.ctx, portBinding, &client.DeleteOptions{}); err != nil {
				blog.Warnf("delete port binding %s/%s from apiserver failed, err %s",
					portBinding.GetName(), portBinding.GetNamespace(), err.Error())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 3 * time.Second,
				}, nil
			}
		}

		return ctrl.Result{}, nil
	}
	var pbhandler iPortBindingHandler
	switch portBinding.Status.PortBindingType {
	case networkextensionv1.PortBindingTypeNode:
		pbhandler = newNodePortBindingHandler(pbr.ctx, pbr.k8sClient, pbr.eventer, nil, pbr.nodeBindCache)
	case networkextensionv1.PortBindingTypePod:
		pbhandler = newPodPortBindingHandler(pbr.ctx, pbr.k8sClient, pbr.eventer, nil)
	default:
		// support low version, use pod portbinding handler as default
		pbhandler = newPodPortBindingHandler(pbr.ctx, pbr.k8sClient, pbr.eventer, nil)
	}
	// change port binding status to PortBindingStatusCleaned
	retry, err := pbhandler.cleanPortBinding(portBinding)
	if err != nil {
		pbr.recordEvent(portBinding, k8scorev1.EventTypeWarning, ReasonPortBindingCleanFailed,
			fmt.Sprintf(MsgPortBindingCleanFailed, err.Error()))
		blog.Warnf("delete port binding %s/%s failed, err %s",
			portBinding.GetName(), portBinding.GetNamespace(), err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	if retry {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (pbr *PortBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.PortBinding{}).
		Watches(&source.Kind{Type: &k8scorev1.Pod{}}, NewPodFilter(mgr.GetClient())).
		Watches(&source.Kind{Type: &k8scorev1.Node{}}, NewNodeFilter(mgr.GetClient(), pbr.nodePortBindingNs)).
		WithEventFilter(pbr.getPortBindingPredicate()).
		Complete(pbr)
}

func (pbr *PortBindingReconciler) getPortBindingPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newPoolBinding, okNew := e.ObjectNew.(*networkextensionv1.PortBinding)
			oldPoolBinding, okOld := e.ObjectOld.(*networkextensionv1.PortBinding)
			if !okNew || !okOld {
				return true
			}
			// todo 确认status的比较是否需要？
			if reflect.DeepEqual(newPoolBinding.Spec, oldPoolBinding.Spec) &&
				reflect.DeepEqual(newPoolBinding.Status.PortBindingStatusList,
					oldPoolBinding.Status.PortBindingStatusList) &&
				newPoolBinding.Status.Status == oldPoolBinding.Status.Status &&
				reflect.DeepEqual(newPoolBinding.DeletionTimestamp, oldPoolBinding.DeletionTimestamp) &&
				reflect.DeepEqual(newPoolBinding.Finalizers, oldPoolBinding.Finalizers) {
				blog.V(5).Infof("portbinding %+v updated, but spec not change", newPoolBinding)
				return false
			}
			return true
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			portBinding, ok := deleteEvent.Object.(*networkextensionv1.PortBinding)
			if !ok {
				return true
			}

			pod := &k8scorev1.Pod{}
			if err := pbr.k8sClient.Get(pbr.ctx, k8stypes.NamespacedName{
				Namespace: portBinding.GetNamespace(),
				Name:      portBinding.GetName(),
			}, pod); err != nil {
				if k8serrors.IsNotFound(err) {
					// pod已被删除，portBinding被删除符合预期
					return false
				}
				blog.Warnf("get pod '%s/%s' failed, err: %s", portBinding.GetNamespace(), portBinding.GetName(),
					err.Error())
				return true
			}

			blog.Infof("portBinding '%s/%s' is deleted while pod exist, push reconcile", portBinding.GetNamespace(),
				portBinding.GetName())
			return true
		},
	}
}

func (pbr *PortBindingReconciler) checkPortAllocate(node *k8scorev1.Node) bool {
	// 节点和Pod的webhook策略不同。为了避免节点加入失败，即使端口分配失败，也允许节点变更
	// 所以需要检查节点是否成功分配了端口
	if _, ok := node.Annotations[constant.AnnotationForPortPoolBindings]; !ok {
		blog.Errorf("node %s has not port allocate annotation %s", node.GetName(),
			constant.AnnotationForPortPoolPorts)
		metrics.ReportPortAllocate(node.GetName(), node.GetNamespace(), false)

		needPatch := true
		if notReadyTimeStr, timeOk := node.Annotations[constant.
			AnnotationForPortBindingNotReadyTimestamp]; timeOk && notReadyTimeStr != "" {
			if notReadyTime, inErr := time.Parse(time.RFC3339Nano, notReadyTimeStr); inErr != nil {
				blog.Warnf("parse not ready timestamp on node %s failed, err: %s", node.GetName(), inErr.Error())
			} else if time.Since(notReadyTime) < time.Second*10 {
				// 距离上次刷新时间未超过10秒
				needPatch = false
			}
		}
		if needPatch {
			// 更新注解触发mutate逻辑
			blog.Infof("patch node %s annotation %s", node.GetName(), constant.AnnotationForPortBindingNotReadyTimestamp)
			if err := utils.PatchNodeAnnotation(pbr.ctx, pbr.k8sClient, node, map[string]interface{}{
				constant.AnnotationForPortBindingNotReadyTimestamp: time.Now().Format(time.RFC3339Nano),
			}); err != nil {
				blog.Errorf("patch node %s annotation failed, err: %s", node.GetName(), err.Error())
			}
		}
		return false
	}
	return true
}

func (pbr *PortBindingReconciler) isNodeValid(node *k8scorev1.Node) bool {
	isNodeIPFound := false
	for _, address := range node.Status.Addresses {
		if address.Type == k8scorev1.NodeInternalIP && len(address.Address) != 0 {
			isNodeIPFound = true
			break
		}
	}
	if !isNodeIPFound {
		blog.Warnf("node %s/%s has not pod ip, requeue it", node.GetName(), node.GetNamespace())
		return false
	}

	return true

}

func (pbr *PortBindingReconciler) podPortBindingPreCheck(portBindingType string, portBinding *networkextensionv1.
PortBinding, pod *k8scorev1.Pod) (bool, ctrl.Result, error) {
	if portBinding == nil {
		if pod.Status.Phase == k8scorev1.PodFailed {
			blog.Infof("pod '%s/%s' is failed, reason: %s, msg: %s, no need to handle it", pod.GetNamespace(),
				pod.GetName(), pod.Status.Reason, pod.Status.Message)
			return false, ctrl.Result{}, nil
		}

		res, err := pbr.createPortBinding(portBindingType, pod.GetNamespace(), pod.GetName(), pod.GetAnnotations())
		return false, res, err
	}

	// pod状态成为Failed后，需要删除对应PortBinding， 避免端口持续被占用无法释放
	if pod.Status.Phase == k8scorev1.PodFailed {
		blog.Infof("pod '%s/%s' is failed, reason: %s, msg: %s, so clean portbinding", pod.GetNamespace(),
			pod.GetName(), pod.Status.Reason, pod.Status.Message)
		res, err := pbr.cleanPortBinding(portBinding)
		return false, res, err
	}

	if len(pod.Status.PodIP) == 0 {
		blog.V(5).Infof("pod %s/%s has not pod ip, requeue it", pod.GetName(), pod.GetNamespace())
		return false, ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}

	return true, ctrl.Result{}, nil
}

func (pbr *PortBindingReconciler) nodePortBindingPreCheck(portBindingType string, portBinding *networkextensionv1.
PortBinding, node *k8scorev1.Node) (bool, ctrl.Result, error) {
	if portBinding == nil {
		res, err := pbr.createPortBinding(portBindingType, node.GetNamespace(), node.GetName(), node.GetAnnotations())
		return false, res, err
	}

	if !pbr.isNodeValid(node) {
		return false, ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	return true, ctrl.Result{}, nil
}
