/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	netservicev1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/utils"
)

// BCSNetIPClaimReconciler reconciles a BCSNetIPClaim object
type BCSNetIPClaimReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetipclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetipclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetips,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetipclaims/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *BCSNetIPClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	blog.V(5).Infof("BCSNetIPClaim %+v triggered", req.Name)
	claim := &netservicev1.BCSNetIPClaim{}
	if err := r.Get(ctx, req.NamespacedName, claim); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Warnf("BCSNetIPClaim %s is deleted", req.Name)
			return ctrl.Result{}, nil
		}
		blog.Errorf("get BCSNetIPClaim %s failed, err %s", req.Name, err.Error())
		r.Recorder.Eventf(claim, v1.EventTypeWarning, "Unbound",
			"get BCSNetIPClaim %s failed, err %s", req.Name, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, err
	}

	// claim is deleted
	if claim.DeletionTimestamp != nil {
		if claim.Status.Phase == constant.BCSNetIPClaimBoundedStatus {
			if err := r.unboundIP(ctx, claim); err != nil {
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 5 * time.Second,
				}, err
			}
			blog.Errorf("can not delete claim %s in %s status", claim.Name, constant.BCSNetIPClaimBoundedStatus)
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, fmt.Errorf("can not delete claim %s in %s status", claim.Name, constant.BCSNetIPClaimBoundedStatus)
		}

		if err := r.removeFinalizerForPool(claim); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, err
		}
		return ctrl.Result{}, nil
	}

	// if doesn't has finalizer, add finalizer
	if !utils.StringInSlice(claim.GetFinalizers(), constant.FinalizerNameBcsNetserviceController) {
		if err := r.addFinalizerForPool(claim); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, nil
		}
		return ctrl.Result{}, nil
	}

	if claim.Status.Phase == constant.BCSNetIPClaimBoundedStatus ||
		claim.Status.Phase == constant.BCSNetIPClaimExpiredStatus {
		return ctrl.Result{}, nil
	}
	if claim.Status.Phase == "" {
		claim.Status.Phase = constant.BCSNetIPClaimPendingStatus
		if err := r.Status().Update(ctx, claim); err != nil {
			blog.Errorf("update BCSNetIPClaim %s/%s status failed, err %s", claim.Namespace, claim.Name, err.Error())
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, err
		}
		return ctrl.Result{}, nil
	}
	if err := r.boundBCSNetIP(ctx, claim); err != nil {
		r.Recorder.Eventf(claim, v1.EventTypeWarning, "Unbound",
			"Bound BCSNetIP failed, %s", err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, err
	}

	return ctrl.Result{}, nil
}

func (r *BCSNetIPClaimReconciler) unboundIP(ctx context.Context, claim *netservicev1.BCSNetIPClaim) error {
	netIP := &netservicev1.BCSNetIP{}
	if err := r.Get(ctx, types.NamespacedName{Name: claim.Status.BoundedIP}, netIP); err != nil {
		return err
	}
	if netIP.Status.Phase == constant.BCSNetIPActiveStatus {
		return fmt.Errorf("delete claim %s failed, bounded IP %s in Active status",
			fmt.Sprintf("%s/%s", claim.Namespace, claim.Name), claim.Status.BoundedIP)
	}
	netIP.Status = netservicev1.BCSNetIPStatus{
		Phase:      constant.BCSNetIPAvailableStatus,
		UpdateTime: metav1.Now(),
	}
	if err := r.Status().Update(ctx, netIP); err != nil {
		return err
	}

	return nil
}

func (r *BCSNetIPClaimReconciler) addFinalizerForPool(claim *netservicev1.BCSNetIPClaim) error {
	claim.Finalizers = append(claim.Finalizers, constant.FinalizerNameBcsNetserviceController)
	if err := r.Update(context.Background(), claim); err != nil {
		blog.Warnf("add finalizer for claim %s failed, err %s", claim.Name, err.Error())
	}
	blog.V(3).Infof("add finalizer for claim %s success", claim.Name)
	return nil
}

func (r *BCSNetIPClaimReconciler) removeFinalizerForPool(claim *netservicev1.BCSNetIPClaim) error {
	claim.Finalizers = utils.RemoveStringInSlice(claim.Finalizers, constant.FinalizerNameBcsNetserviceController)
	if err := r.Update(context.Background(), claim, &client.UpdateOptions{}); err != nil {
		blog.Warnf("remove finalizer for claim %s failed, err %s", claim.Name, err.Error())
		return fmt.Errorf("remove finalizer for claim %s failed, err %s", claim.Name, err.Error())
	}
	blog.V(3).Infof("remove finalizer for claim %s success", claim.Name)
	return nil
}

func (r *BCSNetIPClaimReconciler) boundBCSNetIP(ctx context.Context, claim *netservicev1.BCSNetIPClaim) error {
	if claim.Spec.BCSNetIPName != "" {
		netIP := &netservicev1.BCSNetIP{}
		if err := r.Get(ctx, types.NamespacedName{Name: claim.Spec.BCSNetIPName}, netIP); err != nil {
			return err
		}
		if netIP.Status.Phase != constant.BCSNetIPAvailableStatus {
			return fmt.Errorf("claimed IP [%s] is not available", claim.Spec.BCSNetIPName)
		}
		if err := r.boundIP(ctx, claim, *netIP); err != nil {
			return err
		}
		return nil
	}

	netIPList := &netservicev1.BCSNetIPList{}
	if err := r.List(context.Background(), netIPList); err != nil {
		message := fmt.Sprintf("get BCSNetPool list failed, %s", err.Error())
		blog.Errorf(message)
		return err
	}
	found := false
	for _, ip := range netIPList.Items {
		if ip.Status.Phase == constant.BCSNetIPAvailableStatus {
			found = true
			if err := r.boundIP(ctx, claim, ip); err != nil {
				return err
			}
			break
		}
	}
	if !found {
		return errors.New("no available IP to bound")
	}

	return nil
}

func (r *BCSNetIPClaimReconciler) boundIP(ctx context.Context, claim *netservicev1.BCSNetIPClaim,
	netIP netservicev1.BCSNetIP) error {
	netIP.Status.Phase = constant.BCSNetIPReservedStatus
	netIP.Status.IPClaimKey = fmt.Sprintf("%s/%s", claim.Namespace, claim.Name)
	netIP.Status.Fixed = true
	netIP.Status.UpdateTime = metav1.Now()
	if err := r.Status().Update(ctx, &netIP); err != nil {
		blog.Errorf("update BCSNetIP %s status failed, err %v", netIP.Name, err)
		return fmt.Errorf("update BCSNetIP %s status failed, err %v", netIP.Name, err)
	}
	netIP.Labels[constant.FixIPLabel] = "true"
	if err := r.Update(context.Background(), &netIP); err != nil {
		blog.Errorf("set IP [%s] label failed", netIP.Name)
		return fmt.Errorf("set IP [%s] label failed", netIP.Name)
	}
	claim.Status.BoundedIP = netIP.Name
	claim.Status.Phase = constant.BCSNetIPClaimBoundedStatus
	if err := r.Status().Update(ctx, claim); err != nil {
		blog.Errorf("update BCSNetIPClaim %s/%s status failed, err %v", claim.Namespace, claim.Name, err)
		return fmt.Errorf("update BCSNetIPClaim %s/%s status failed, err %v", claim.Namespace, claim.Name, err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BCSNetIPClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&netservicev1.BCSNetIPClaim{}).
		Complete(r)
}
