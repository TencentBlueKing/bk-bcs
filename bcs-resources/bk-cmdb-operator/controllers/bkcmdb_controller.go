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

package controllers

import (
	"context"

	bkcmdbv1 "github.com/Tencent/bk-bcs/bcs-resources/bk-cmdb-operator/api/v1"
	cmdbClient "github.com/Tencent/bk-bcs/bcs-resources/bk-cmdb-operator/kube/client"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	extensionsV1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	ownerKey = ".metadata.controller"
	apiGVStr = bkcmdbv1.GroupVersion.String()
)

// BkcmdbReconciler reconciles a Bkcmdb object
type BkcmdbReconciler struct {
	Client cmdbClient.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type reconcileFun func(cluster *bkcmdbv1.Bkcmdb) error

// +kubebuilder:rbac:groups=bkcmdb.bkbcs.tencent.com,resources=bkcmdbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=bkcmdb.bkbcs.tencent.com,resources=bkcmdbs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=secrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=configmaps/status,verbs=get;update;patch
func (r *BkcmdbReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	//log = r.Log.WithValues("bkcmdb", req.NamespacedName)

	var instance bkcmdbv1.Bkcmdb
	if err := r.Client.Get(ctx, req.NamespacedName, &instance); err != nil {
		r.Log.Error(err, "unable to fetch bkcmdb")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// if not specified mongodb, then create it
	if instance.Spec.MongoDb == nil {
		if err := r.reconcileMongoDb(&instance); err != nil {
			r.Log.Error(err, "unable to reconcile MongoDb")
			return ctrl.Result{}, err
		}
	}

	// if not specified redis, then create it
	if instance.Spec.Redis == nil {
		if err := r.reconcileRedis(&instance); err != nil {
			r.Log.Error(err, "unable to reconcile redis")
			return ctrl.Result{}, err
		}
	}

	// if not specified zookeeper, then create it
	if instance.Spec.Zookeeper == nil {
		if err := r.reconcileZookeeper(&instance); err != nil {
			r.Log.Error(err, "unable to reconcile zookeeper")
			return ctrl.Result{}, err
		}
	}

	// reconcile all bk-cmdb resources
	for _, fun := range []reconcileFun{
		r.reconcileConfigMap,
		r.reconcileAdminServer,
		r.reconcileApiServer,
		r.reconcileCoreService,
		r.reconcileDataCollection,
		r.reconcileEventServer,
		r.reconcileHostServer,
		r.reconcileOperationServer,
		r.reconcileProcServer,
		r.reconcileTaskServer,
		r.reconcileTmServer,
		r.reconcileTopoServer,
		r.reconcileWebServer,
		r.reconcileJob,
	} {
		if err := fun(&instance); err != nil {
			r.Log.Error(err, "unable to reconcile bkcmdb")
			return reconcile.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *BkcmdbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// set services index filed
	if err := mgr.GetFieldIndexer().IndexField(&v1.Service{}, ownerKey, func(rawObj runtime.Object) []string {
		svc := rawObj.(*v1.Service)
		owner := metav1.GetControllerOf(svc)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Bkcmdb" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// set deployments index filed
	if err := mgr.GetFieldIndexer().IndexField(&appsv1.Deployment{}, ownerKey, func(rawObj runtime.Object) []string {
		deploy := rawObj.(*appsv1.Deployment)
		owner := metav1.GetControllerOf(deploy)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Bkcmdb" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// set statefulsets index filed
	if err := mgr.GetFieldIndexer().IndexField(&appsv1.StatefulSet{}, ownerKey, func(rawObj runtime.Object) []string {
		sts := rawObj.(*appsv1.StatefulSet)
		owner := metav1.GetControllerOf(sts)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Bkcmdb" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// set secrets index filed
	if err := mgr.GetFieldIndexer().IndexField(&v1.Secret{}, ownerKey, func(rawObj runtime.Object) []string {
		secret := rawObj.(*v1.Secret)
		owner := metav1.GetControllerOf(secret)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Bkcmdb" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// set jobs index filed
	if err := mgr.GetFieldIndexer().IndexField(&batchV1.Job{}, ownerKey, func(rawObj runtime.Object) []string {
		job := rawObj.(*batchV1.Job)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Bkcmdb" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// set ingresses index filed
	if err := mgr.GetFieldIndexer().IndexField(&extensionsV1beta1.Ingress{}, ownerKey, func(rawObj runtime.Object) []string {
		ingress := rawObj.(*extensionsV1beta1.Ingress)
		owner := metav1.GetControllerOf(ingress)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Bkcmdb" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// set configmaps index filed
	if err := mgr.GetFieldIndexer().IndexField(&v1.ConfigMap{}, ownerKey, func(rawObj runtime.Object) []string {
		cm := rawObj.(*v1.ConfigMap)
		owner := metav1.GetControllerOf(cm)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Bkcmdb" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&bkcmdbv1.Bkcmdb{}).
		Owns(&v1.Secret{}).
		Owns(&v1.Service{}).
		Owns(&v1.ConfigMap{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&batchV1.Job{}).
		Owns(&extensionsV1beta1.Ingress{}).
		Complete(r)
}
