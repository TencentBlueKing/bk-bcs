/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/helmclient"

	"k8s.io/klog"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IstioClusterReconciler reconciles a IstioCluster object
type IstioClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	//helm client
	helm helmclient.HelmClient
	//
}

func NewIstioClusterReconciler()*IstioClusterReconciler{

}

func (r *IstioClusterReconciler) init(){

}

// +kubebuilder:rbac:groups=mesh.tencent.com,resources=istioclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mesh.tencent.com,resources=istioclusters/status,verbs=get;update;patch

func (r *IstioClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("istiocluster", req.NamespacedName)

	istioCluster := &meshv1.IstioCluster{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, istioCluster)
	if err!=nil {
		if errors.IsNotFound(err) {
			klog.Infof("IstioCluster %s is actually deleted", req.String())
			return ctrl.Result{}, nil
		}

		klog.Errorf("Get IstioCluster %s failed: %s", req.String(), err.Error())
		return ctrl.Result{RequeueAfter: time.Second*3}, err
	}
	finalizer := "istiocluster.finalizers.bkbcs.tencent.com"
	//in deleting
	if !istioCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		klog.Infof("IstioCluster %s in deleting, and DeletionTimestamp %s", req.String(), istioCluster.DeletionTimestamp.String())
		if containsString(istioCluster.ObjectMeta.Finalizers, finalizer) {
			//recovery fundpool
			if err := r.recoveryFundPools(supply); err != nil {
				return ctrl.Result{RequeueAfter: time.Second*3}, err
			}
			// delete finalizers
			istioCluster.ObjectMeta.Finalizers = removeString(istioCluster.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(context.Background(), istioCluster); err != nil {
				return ctrl.Result{RequeueAfter: time.Second*3}, err
			}
		}
	}

	//append finalizer
	if !containsString(istioCluster.ObjectMeta.Finalizers, finalizer) {
		istioCluster.ObjectMeta.Finalizers = append(istioCluster.ObjectMeta.Finalizers, finalizer)
	}

	return ctrl.Result{}, nil
}

func (r *IstioClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&meshv1.IstioCluster{}).
		Complete(r)
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}