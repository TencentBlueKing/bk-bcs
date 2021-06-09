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

package appsvc

import (
	"context"
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/queue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
	"github.com/Tencent/bk-bcs/bmsf-mesh/bmsf-mesos-adapter/controller/ns"
	meshv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/mesh/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AppSvc Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, q queue.Queue) error {
	return add(mgr, newReconciler(mgr, q))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, q queue.Queue) reconcile.Reconciler {
	r := &ReconcileAppSvc{
		Client:     mgr.GetClient(),
		localCache: mgr.GetCache(),
		scheme:     mgr.GetScheme(),
		eventQ:     q,
	}
	//starting goroutine for event handling
	go r.handleQueue()
	return r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("appsvc-adaptor-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to AppSvc
	err = c.Watch(&source.Kind{Type: &meshv1.AppSvc{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by AppSvc - change this for objects you create
	return nil
}

var _ reconcile.Reconciler = &ReconcileAppSvc{}

// ReconcileAppSvc reconciles a AppSvc object
type ReconcileAppSvc struct {
	client.Client
	//cache for reading data from locally
	localCache cache.Cache
	scheme     *runtime.Scheme
	eventQ     queue.Queue
}

// Reconcile reads that state of the cluster for a AppSvc object and makes changes based on the state read
// and what is in the AppSvc.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mesh.bmsf.tencent.com,resources=appsvcs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mesh.bmsf.tencent.com,resources=appsvcs/status,verbs=get;update;patch
func (r *ReconcileAppSvc) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the AppSvc instance
	instance := &meshv1.AppSvc{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	blog.Infof("AppSvc %s reconcile", request.NamespacedName.String())
	return reconcile.Result{}, nil
}

func (r *ReconcileAppSvc) handleQueue() {
	ech, err := r.eventQ.GetChannel()
	if err != nil {
		blog.Errorf("ReconcileAppSvc get event queue failed, %s", err)
		return
	}
	for {
		select {
		case event, ok := <-ech:
			if !ok {
				blog.Errorf("ReconcilerAppSvc get error event type.")
				return
			}
			svc, tok := event.Data.(*meshv1.AppSvc)
			if !tok {
				blog.Errorf("ReconcileAppSvc get unknown data info. discard.")
				continue
			}
			switch event.Type {
			case watch.EventAdded:
				r.onAdd(svc)
			case watch.EventDeleted:
				r.onDelete(svc)
			case watch.EventUpdated:
				r.onUpdate(svc)
			default:
				blog.Warnf("ReconcilerAppSvc get unknown Event Type: %s.", event.Type)
			}
			//todo(DeveloperJim) default info for timeout?
		}
	}
}

// onAdd add new AppSvc to kube-apiserver
// todo(DeveloperJim): push event queue back when operation failed?
func (r *ReconcileAppSvc) onAdd(svc *meshv1.AppSvc) {
	instance := &meshv1.AppSvc{}
	key, kerr := client.ObjectKeyFromObject(svc)
	if kerr != nil {
		blog.Errorf("ReconcileAppSvc formate %s/%s to Object key failed, %s", svc.GetNamespace(), svc.GetName(), kerr.Error())
		return
	}
	if err := ns.CheckNamespace(r.localCache, r, svc.GetNamespace()); err != nil {
		blog.Errorf("ReconcileAppSvc checks namespace %s failed, %s", svc.GetNamespace(), err.Error())
		return
	}
	err := r.Get(context.TODO(), key, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, create new one directly
			if err := r.Create(context.TODO(), svc); err != nil {
				blog.Errorf("ReconcileAppSvc create new AppSvc %s failed, %s", key.String(), err.Error())
				return
			}
			blog.Infof("ReconcileAppSvc creat new AppSvc %s on EventAdded successfully", key.String())
			return
		}
		// Error reading the object
		blog.Errorf("AppSvc reads local cache %s failed, %s", key.String(), err.Error())
		return
	}
	//get exist data, ready to update?
	if reflect.DeepEqual(instance.Spec, svc.Spec) {
		blog.Warnf("ReconcileAppSvc get deepEqual in EventAdded, key: %s", key.String())
		return
	}
	instance.Spec = svc.Spec
	instance.Status.LastUpdateTime = metav1.Now()
	//ready to Udpate
	if err := r.Update(context.TODO(), instance); err != nil {
		blog.Errorf("ReconcileAppSvc update %s in EventAdded failed, %s", key.String(), err.Error())
		return
	}
	if err := r.Status().Update(context.TODO(), instance); err != nil {
		blog.Errorf("ReconcileAppNode update %s Status in EventAdded failed, %s", key.String(), err.Error())
		return
	}
	blog.Warnf("ReconcileAppSvc update %s successfully in EventAdded, maybe service cache lost in cluster", key.String())
}

func (r *ReconcileAppSvc) onUpdate(svc *meshv1.AppSvc) {
	instance := &meshv1.AppSvc{}
	key, kerr := client.ObjectKeyFromObject(svc)
	if kerr != nil {
		blog.Errorf("ReconcileAppSvc formate %s/%s to Object key failed, %s", svc.GetNamespace(), svc.GetName(), kerr.Error())
		return
	}
	err := r.Get(context.TODO(), key, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, create new one directly
			if err := r.Create(context.TODO(), svc); err != nil {
				blog.Errorf("ReconcileAppSvc create new AppSvc %s on EventUpated failed, %s", key.String(), err.Error())
				return
			}
			blog.Warnf("ReconcileAppSvc creat new AppSvc %s on EventUpated successfully, maybe local cache lost data", key.String())
			return
		}
		// Error reading the object
		blog.Errorf("AppSvc reads local cache %s on EventUpdated failed, %s", key.String(), err.Error())
		return
	}
	//get exist data, ready to update
	if reflect.DeepEqual(instance.Spec, svc.Spec) && reflect.DeepEqual(instance.Labels, svc.Labels) {
		blog.Warnf("ReconcileAppSvc get deepEqual in EventAdded, key: %s", key.String())
		return
	}
	instance.Spec = svc.Spec
	instance.Status.LastUpdateTime = metav1.Now()
	//ready to Udpate
	if err := r.Update(context.TODO(), instance); err != nil {
		blog.Errorf("ReconcileAppSvc update %s in EventUpdated failed, %s", key.String(), err.Error())
		return
	}
	if err := r.Status().Update(context.TODO(), instance); err != nil {
		blog.Errorf("ReconcileAppNode update %s Status in EventUpdated failed, %s", key.String(), err.Error())
		return
	}
	blog.Infof("ReconcileAppSvc update %s successfully", key.String())
}

func (r *ReconcileAppSvc) onDelete(svc *meshv1.AppSvc) {
	//ready to Udpate
	if err := r.Delete(context.TODO(), svc); err != nil {
		blog.Errorf("ReconcileAppSvc DELETE %s/%s failed, %s", svc.GetNamespace(), svc.GetName(), err.Error())
		return
	}
	blog.Infof("ReconcileAppSvc DELETE %s/%s successfully", svc.GetNamespace(), svc.GetName())
}
