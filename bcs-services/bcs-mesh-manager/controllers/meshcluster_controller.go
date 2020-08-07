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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	"time"

	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"

	"k8s.io/klog"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MeshClusterReconciler reconciles a MeshCluster object
type MeshClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	//meshClusters
	meshClusters map[string]*MeshClusterManager
	//config
	conf config.Config
}

func NewMeshClusterReconciler()*MeshClusterReconciler{

}

func (r *MeshClusterReconciler) getMeshClusterManager(mCluster *meshv1.MeshCluster)(*MeshClusterManager,error){
	meshCluster,_ := r.meshClusters[mCluster.GetUuid()]
	if meshCluster!=nil {
		return meshCluster, nil
	}
	meshCluster,err := NewMeshClusterManager(r.conf, mCluster.DeepCopy())
	if err!=nil {
		klog.Errorf("NewMeshClusterManager(%s) failed: %s", mCluster.GetUuid(), err.Error())
		return nil, err
	}
	r.meshClusters[mCluster.GetUuid()] = meshCluster
	return meshCluster, nil
}

// +kubebuilder:rbac:groups=mesh.tencent.com,resources=MeshClusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mesh.tencent.com,resources=MeshClusters/status,verbs=get;update;patch

func (r *MeshClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("MeshCluster", req.NamespacedName)

	MeshCluster := &meshv1.MeshCluster{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, MeshCluster)
	if err!=nil {
		if errors.IsNotFound(err) {
			klog.Infof("MeshCluster %s is actually deleted", req.String())
			return ctrl.Result{}, nil
		}

		klog.Errorf("Get MeshCluster %s failed: %s", req.String(), err.Error())
		return ctrl.Result{RequeueAfter: time.Second*3}, err
	}
	finalizer := "MeshCluster.finalizers.bkbcs.tencent.com"
	//in deleting
	if !MeshCluster.ObjectMeta.DeletionTimestamp.IsZero() {
		klog.Infof("MeshCluster %s in deleting, and DeletionTimestamp %s", req.String(), MeshCluster.DeletionTimestamp.String())
		if containsString(MeshCluster.ObjectMeta.Finalizers, finalizer) {
			//recovery fundpool
			if err := r.recoveryFundPools(supply); err != nil {
				return ctrl.Result{RequeueAfter: time.Second*3}, err
			}
			// delete finalizers
			MeshCluster.ObjectMeta.Finalizers = removeString(MeshCluster.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(context.Background(), MeshCluster); err != nil {
				return ctrl.Result{RequeueAfter: time.Second*3}, err
			}
		}
	}

	//append finalizer
	if !containsString(MeshCluster.ObjectMeta.Finalizers, finalizer) {
		MeshCluster.ObjectMeta.Finalizers = append(MeshCluster.ObjectMeta.Finalizers, finalizer)
	}

	return ctrl.Result{}, nil
}

func (r *MeshClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&meshv1.MeshCluster{}).
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