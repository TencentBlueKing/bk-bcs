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
	ingresscommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	bcsnetcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// PortBindingReconciler reconciler for bcs port pool
type PortBindingReconciler struct {
	cleanInterval time.Duration
	ctx           context.Context
	k8sClient     client.Client
	poolCache     *portpoolcache.Cache
}

// NewPortBindingReconciler create PortBindingReconciler
func NewPortBindingReconciler(
	ctx context.Context, cleanInterval time.Duration,
	k8sClient client.Client, poolCache *portpoolcache.Cache) *PortBindingReconciler {
	return &PortBindingReconciler{
		ctx:           ctx,
		cleanInterval: cleanInterval,
		k8sClient:     k8sClient,
		poolCache:     poolCache,
	}
}

// Reconcile reconcile port pool
// portbinding name is same with pod name
func (pbr *PortBindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("PortBinding %+v triggered", req.NamespacedName)
	portBinding := &networkextensionv1.PortBinding{}
	isPortBindingFound := true
	if err := pbr.k8sClient.Get(pbr.ctx, req.NamespacedName, portBinding); err != nil {
		if k8serrors.IsNotFound(err) {
			isPortBindingFound = false
		} else {
			blog.Warnf("get portbinding %v failed, err %s, requeue it", req.NamespacedName, err.Error())
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 3 * time.Second,
			}, nil
		}
	}
	pod := &k8scorev1.Pod{}
	if err := pbr.k8sClient.Get(pbr.ctx, req.NamespacedName, pod); err != nil {
		if k8serrors.IsNotFound(err) {
			// if pod is not found, do clean portbinding
			if isPortBindingFound {
				blog.V(3).Infof("clean portbinding %v", req.NamespacedName)
				return pbr.cleanPortBinding(portBinding)
			}
			// if both pod and portbinding are not found, just return
			return ctrl.Result{}, nil
		}
		blog.Warnf("get pod %v failed, err %s", req.NamespacedName, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	// if pod is found but portbinding is not found, create portbinding for pod
	if !isPortBindingFound {
		blog.V(3).Infof("create portbinding %s/%s", pod.GetName(), pod.GetNamespace())
		return pbr.createPortBinding(pod)
	}
	// when statefulset pod is recreated, the old portbinding may be deleting
	if portBinding.DeletionTimestamp != nil {
		blog.V(3).Infof("found deleting portbinding, continue clean portbinding %v", req.NamespacedName)
		return pbr.cleanPortBinding(portBinding)
	}
	if len(pod.Status.PodIP) == 0 {
		blog.Warnf("pod %s/%s has not pod ip, requeue it", pod.GetName(), pod.GetNamespace())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 300 * time.Millisecond,
		}, nil
	}
	pbhandler := newPortBindingHandler(pbr.ctx, pbr.k8sClient)
	retry, err := pbhandler.ensurePortBinding(pod, portBinding)
	if err != nil {
		blog.Warnf("ensure port binding %s/%s failed, err %s",
			portBinding.GetName(), portBinding.GetNamespace(), err.Error())
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

func (pbr *PortBindingReconciler) createPortBinding(pod *k8scorev1.Pod) (ctrl.Result, error) {
	if !checkPortPoolAnnotationForPod(pod) {
		blog.Warnf("check pod %s/%s annotation for port binding failed", pod.GetName(), pod.GetNamespace())
		return ctrl.Result{}, nil
	}
	annotationValue, ok := pod.Annotations[constant.AnnotationForPortPoolBindings]
	if !ok {
		blog.Warnf("pod %s/%s has no annotation %s",
			pod.GetName(), pod.GetNamespace(), constant.AnnotationForPortPoolBindings)
		return ctrl.Result{}, nil
	}
	var portBindingList []*networkextensionv1.PortBindingItem
	if err := json.Unmarshal([]byte(annotationValue), &portBindingList); err != nil {
		blog.Warnf("internal logic err, decode value of pod %s/%s annotation %s is invalid, err %s, value %s",
			pod.GetName(), pod.GetNamespace(), constant.AnnotationForPortPoolPorts, err.Error(), annotationValue)
		return ctrl.Result{}, nil
	}
	podPortBinding := &networkextensionv1.PortBinding{}
	podPortBinding.SetName(pod.GetName())
	podPortBinding.SetNamespace(pod.GetNamespace())
	labels := make(map[string]string)
	for _, binding := range portBindingList {
		tmpKey := fmt.Sprintf(networkextensionv1.PortPoolBindingLabelKeyFromat, binding.PoolName, binding.PoolNamespace)
		labels[tmpKey] = binding.PoolItemName
	}
	if duration, ok := pod.Annotations[networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration]; ok {
		podPortBinding.SetAnnotations(map[string]string{
			networkextensionv1.PortPoolBindingAnnotationKeyKeepDuration: duration,
		})
	}
	podPortBinding.SetLabels(labels)
	podPortBinding.Finalizers = append(podPortBinding.Finalizers, constant.FinalizerNameBcsIngressController)
	podPortBinding.Spec = networkextensionv1.PortBindingSpec{
		PortBindingList: portBindingList,
	}

	if err := pbr.k8sClient.Create(context.Background(), podPortBinding, &client.CreateOptions{}); err != nil {
		blog.Warnf("failed to create port binding object, err %s", err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	return ctrl.Result{}, nil
}

func (pbr *PortBindingReconciler) cleanPortBinding(portBinding *networkextensionv1.PortBinding) (ctrl.Result, error) {
	if portBinding.Status.Status == constant.PortBindingStatusCleaned {
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
			blog.V(3).Infof("removing finalizer from port binding %s/%s",
				portBinding.GetName(), portBinding.GetNamespace())
			portBinding.Finalizers = bcsnetcommon.RemoveString(
				portBinding.Finalizers, constant.FinalizerNameBcsIngressController)
			if err := pbr.k8sClient.Update(pbr.ctx, portBinding, &client.UpdateOptions{}); err != nil {
				blog.Warnf("remote finalizer from port binding %s/%s failed, err %s",
					portBinding.GetName(), portBinding.GetNamespace(), err.Error())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 3 * time.Second,
				}, nil
			}
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
	// change port binding status to PortBindingStatusCleaned
	pbhandler := newPortBindingHandler(pbr.ctx, pbr.k8sClient)
	retry, err := pbhandler.cleanPortBinding(portBinding)
	if err != nil {
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
		WithEventFilter(getPortBindingPredicate()).
		Complete(pbr)
}

func getPortBindingPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newPoolBinding, okNew := e.ObjectNew.(*networkextensionv1.PortBinding)
			oldPoolBinding, okOld := e.ObjectOld.(*networkextensionv1.PortBinding)
			if !okNew || !okOld {
				return true
			}
			if reflect.DeepEqual(newPoolBinding.Spec, oldPoolBinding.Spec) &&
				reflect.DeepEqual(newPoolBinding.Status.PortBindingStatusList,
					oldPoolBinding.Status.PortBindingStatusList) &&
				newPoolBinding.Status.Status == oldPoolBinding.Status.Status &&
				reflect.DeepEqual(newPoolBinding.ObjectMeta, oldPoolBinding.ObjectMeta) {
				blog.V(5).Infof("portbinding %+v updated, but spec not change", newPoolBinding)
				return false
			}
			return true
		},
	}
}
