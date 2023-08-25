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
	"fmt"
	"strings"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	netservicev1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/utils"
)

// BCSNetPoolReconciler reconciles a BCSNetPool object
type BCSNetPoolReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	IPFilter *IPFilter
}

//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetpools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetpools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetips,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetips/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetipclaims/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;create;update;delete
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=pods,verbs=get;list
//+kubebuilder:rbac:groups=netservice.bkbcs.tencent.com,resources=bcsnetpools/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *BCSNetPoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	blog.V(5).Infof("BCSNetPool %+v triggered", req.Name)
	netPool := &netservicev1.BCSNetPool{}
	if err := r.Get(ctx, req.NamespacedName, netPool); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("BCSNetPool %s is deleted", req.Name)
			return ctrl.Result{}, nil
		}
		blog.Errorf("get BCSNetPool %s failed, err %s", req.Name, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, err
	}

	// netPool is deleted
	if netPool.DeletionTimestamp != nil {
		netIPList := &netservicev1.BCSNetIPList{}
		if err := r.listIPWithSelector(ctx, netIPList, map[string]string{"pool": netPool.Name}); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, err
		}

		for _, ip := range netIPList.Items {
			if ip.Status.Phase == constant.BCSNetIPActiveStatus {
				blog.Errorf("can not perform operation for pool %s, active IP %s exists", netPool.Name, ip)
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 5 * time.Second,
				}, fmt.Errorf("can not perform operation for pool %s, active IP %s exists", netPool.Name, ip)
			}
			if err := r.Delete(ctx, &ip); err != nil {
				blog.Errorf("delete BCSNetIP %s failed, err %s", req.Name, err.Error())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 5 * time.Second,
				}, err
			}
		}

		if err := r.removeFinalizerForPool(netPool); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, err
		}
		return ctrl.Result{}, nil
	}

	// if doesn't has finalizer, add finalizer
	if !utils.StringInSlice(netPool.GetFinalizers(), constant.FinalizerNameBcsNetserviceController) {
		if err := r.addFinalizerForPool(netPool); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, nil
		}
		return ctrl.Result{}, nil
	}

	if netPool.Status.Phase == "" {
		blog.Infof("initializing BCSNetPool %s", req.Name)
		if err := r.updatePoolStatus(ctx, netPool, constant.BCSNetPoolInitializingStatus); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, err
		}
		return ctrl.Result{}, nil
	}

	result, err := r.syncBCSNetIP(ctx, netPool)
	if err != nil {
		return result, err
	}

	if err := r.syncReservedBCSNetIP(ctx, netPool); err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, err
	}

	if netPool.Status.Phase != constant.BCSNetPoolNormalStatus {
		if err := r.updatePoolStatus(ctx, netPool, constant.BCSNetPoolNormalStatus); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *BCSNetPoolReconciler) updatePoolStatus(ctx context.Context, netPool *netservicev1.BCSNetPool, status string) error {
	netPool.Status.Phase = status
	netPool.Status.UpdateTime = metav1.Now()
	if err := r.Status().Update(ctx, netPool); err != nil {
		blog.Errorf("update BCSNetPool %s status failed, err %s", netPool.Name, err.Error())
		return err
	}
	blog.Infof("BCSNetPool %s status update success", netPool.Name)
	return nil
}

func (r *BCSNetPoolReconciler) addFinalizerForPool(netPool *netservicev1.BCSNetPool) error {
	netPool.Finalizers = append(netPool.Finalizers, constant.FinalizerNameBcsNetserviceController)
	if err := r.Update(context.Background(), netPool); err != nil {
		blog.Warnf("add finalizer for netPool %s failed, err %s", netPool.Name, err.Error())
	}
	blog.V(3).Infof("add finalizer for netPool %s success", netPool.Name)
	return nil
}

func (r *BCSNetPoolReconciler) removeFinalizerForPool(netPool *netservicev1.BCSNetPool) error {
	netPool.Finalizers = utils.RemoveStringInSlice(netPool.Finalizers, constant.FinalizerNameBcsNetserviceController)
	if err := r.Update(context.Background(), netPool, &client.UpdateOptions{}); err != nil {
		blog.Warnf("remove finalizer for netPool %s failed, err %s", netPool.Name, err.Error())
		return fmt.Errorf("remove finalizer for netPool %s failed, err %s", netPool.Name, err.Error())
	}
	blog.V(3).Infof("remove finalizer for netPool %s success", netPool.Name)
	return nil
}

func (r *BCSNetPoolReconciler) syncBCSNetIP(ctx context.Context, netPool *netservicev1.BCSNetPool) (ctrl.Result, error) {
	blog.Infof("syncing BCSNetIP...")
	// create BCSNetIP based on BCSNetPool if not exists
	for _, ip := range netPool.Spec.AvailableIPs {
		if err := r.createBCSNetIP(ctx, netPool, ip); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, err
		}
	}

	err := r.deleteBCSNetIP(ctx, netPool)
	if err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, err
	}

	return ctrl.Result{}, nil
}

func (r *BCSNetPoolReconciler) syncReservedBCSNetIP(ctx context.Context, netPool *netservicev1.BCSNetPool) error {
	netIPList := &netservicev1.BCSNetIPList{}
	if err := r.listIPWithSelector(ctx, netIPList, map[string]string{
		"pool": netPool.Name, constant.FixIPLabel: "true"}); err != nil {
		return err
	}
	var ipList []netservicev1.BCSNetIP
	for _, ip := range netIPList.Items {
		if ip.Status.Phase == constant.BCSNetIPReservedStatus && ip.Status.KeepDuration != "" {
			ipList = append(ipList, ip)
		}
	}

	for _, ip := range ipList {
		go r.releaseExpiredIP(ip)
	}

	return nil
}

func (r *BCSNetPoolReconciler) releaseExpiredIP(ip netservicev1.BCSNetIP) {
	duration, err := time.ParseDuration(ip.Status.KeepDuration)
	if err != nil {
		blog.Errorf("invalid keep duration %s for fixed IP %s", ip.Status.KeepDuration, ip.Name)
		return
	}

	time.Sleep(duration)
	currentIP := &netservicev1.BCSNetIP{}
	if err := r.Get(context.Background(), types.NamespacedName{Name: ip.Name}, currentIP); err != nil {
		blog.Errorf("get BCSNetIP failed, %s", err.Error())
		return
	}
	if currentIP.Status.UpdateTime.Add(duration).Before(time.Now()) {
		if currentIP.Status.Phase == constant.BCSNetIPReservedStatus {
			// update claim status first
			if err := r.updateIPClaimStatus(context.Background(), currentIP); err != nil {
				return
			}

			currentIP.Status = netservicev1.BCSNetIPStatus{
				Phase:      constant.BCSNetIPAvailableStatus,
				UpdateTime: metav1.Now(),
			}
			if err := r.Status().Update(context.Background(), currentIP); err != nil {
				blog.Errorf("update BCSNetPool %s status failed, err %s", ip.Name, err.Error())
				return
			}
			currentIP.Labels[constant.FixIPLabel] = "false"
			if err := r.Update(context.Background(), currentIP); err != nil {
				blog.Errorf("set IP [%s] label failed", currentIP.Name)
				return
			}

			blog.V(5).Infof("released IP %s", currentIP.Name)
		}
	}
}

func (r *BCSNetPoolReconciler) updateIPClaimStatus(ctx context.Context, ip *netservicev1.BCSNetIP) error {
	claimInfo := strings.Split(ip.Status.IPClaimKey, "/")
	if len(claimInfo) != 2 {
		blog.Errorf("get claim for IP %s failed, invalid claimKey", ip.Name)
		return fmt.Errorf("get claim for IP %s failed, invalid claimKey", ip.Name)
	}
	claim := &netservicev1.BCSNetIPClaim{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: claimInfo[0], Name: claimInfo[1]}, claim); err != nil {
		blog.Errorf("get claim for IP %s failed, %v", ip.Name, err)
		return fmt.Errorf("get claim for IP %s failed, %v", ip.Name, err)
	}
	claim.Status = netservicev1.BCSNetIPClaimStatus{
		Phase: constant.BCSNetIPClaimExpiredStatus,
	}
	if err := r.Status().Update(context.Background(), claim); err != nil {
		blog.Errorf("update claim status %s failed, %v", claim.Name, err)
		return fmt.Errorf("update claim status %s failed, %v", claim.Name, err)
	}
	return nil
}

// createBCSNetIP creates IP for a Pool
func (r *BCSNetPoolReconciler) createBCSNetIP(ctx context.Context, netPool *netservicev1.BCSNetPool, ip string) error {
	netIP := &netservicev1.BCSNetIP{}
	if err := r.Get(ctx, types.NamespacedName{Name: ip}, netIP); err != nil {
		if k8serrors.IsNotFound(err) {
			newNetIP := &netservicev1.BCSNetIP{
				ObjectMeta: metav1.ObjectMeta{
					Name:   ip,
					Labels: map[string]string{"pool": netPool.Name, constant.FixIPLabel: "false"},
				},
				Spec: netservicev1.BCSNetIPSpec{
					Net:     netPool.Spec.Net,
					Mask:    netPool.Spec.Mask,
					Gateway: netPool.Spec.Gateway,
				},
			}
			if err := r.Create(ctx, newNetIP); err != nil {
				blog.Errorf("create BCSNetIP %s failed, err %s", ip, err.Error())
				return err
			}
			blog.Infof("BCSNetIP %s created successfully", ip)

			newNetIP.Status.Phase = constant.BCSNetIPAvailableStatus
			if err := r.Status().Update(ctx, newNetIP); err != nil {
				blog.Errorf("update BCSNetIP %s status failed, err %s", ip, err.Error())
				return err
			}
			blog.Infof("BCSNetIP %s status update successfully", ip)
			return nil
		}
		return err
	}
	return nil
}

// deleteBCSNetIP deletes IP not belongs to any Pools anymore
func (r *BCSNetPoolReconciler) deleteBCSNetIP(ctx context.Context, netPool *netservicev1.BCSNetPool) error {
	netIPList := &netservicev1.BCSNetIPList{}
	if err := r.listIPWithSelector(ctx, netIPList, map[string]string{"pool": netPool.Name}); err != nil {
		return err
	}

	curIPlist := make(map[string]string)
	for _, v := range netIPList.Items {
		curIPlist[v.Name] = v.Status.Phase
	}

	delIPList := make(map[string]string)
	newIPMap := make(map[string]bool)
	for _, v := range netPool.Spec.AvailableIPs {
		newIPMap[v] = true
	}

	for k, v := range curIPlist {
		if _, exists := newIPMap[k]; !exists {
			delIPList[k] = v
		}
	}

	for k, v := range delIPList {
		if v == constant.BCSNetIPActiveStatus {
			return fmt.Errorf("can not delete IP %s in actvie status", k)
		}

		if err := r.Delete(ctx, &netservicev1.BCSNetIP{ObjectMeta: metav1.ObjectMeta{Name: k}}); err != nil {
			blog.Errorf("delete ip %s failed, err %s", k, err.Error())
			return err
		}
	}

	return nil
}

func (r *BCSNetPoolReconciler) listIPWithSelector(ctx context.Context, netIPList *netservicev1.BCSNetIPList,
	labels map[string]string) error {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: labels,
	})
	if err != nil {
		blog.Errorf("unable to convert label selector, err %s", err.Error())
		return err
	}

	if err := r.List(ctx, netIPList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		blog.Errorf("get ip list failed, err %s", err.Error())
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BCSNetPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&netservicev1.BCSNetPool{}).
		Watches(&source.Kind{Type: &netservicev1.BCSNetIP{}}, r.IPFilter).
		Complete(r)
}
