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

package appnode

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

// Add creates a new AppNode Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, q queue.Queue) error {
	return add(mgr, newReconciler(mgr, q))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, q queue.Queue) reconcile.Reconciler {
	r := &ReconcileAppNode{
		Client:     mgr.GetClient(),
		localCache: mgr.GetCache(),
		scheme:     mgr.GetScheme(),
		eventQ:     q,
	}
	//starting gorutine for event handling
	go r.handleQueue()
	return r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("appnode-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to AppNode
	err = c.Watch(&source.Kind{Type: &meshv1.AppNode{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileAppNode{}

// ReconcileAppNode reconciles a AppNode object
type ReconcileAppNode struct {
	//client for data operation
	client.Client
	//cache for reading data from locally
	localCache cache.Cache
	scheme     *runtime.Scheme
	eventQ     queue.Queue
}

// Reconcile reads that state of the cluster for a AppNode object and makes changes based on the state read
// and what is in the AppNode.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mesh.bmsf.tencent.com,resources=appnodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mesh.bmsf.tencent.com,resources=appnodes/status,verbs=get;update;patch
func (r *ReconcileAppNode) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the AppNode instance
	instance := &meshv1.AppNode{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			blog.Infof("AppNode %s reconcile, but data NotFound. delete event confirm", request.NamespacedName.String())
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		blog.Errorf("AppNode %s reconcile got err when reading cache, %s", request.NamespacedName.String(), err.Error())
		return reconcile.Result{}, err
	}
	blog.Infof("AppNode %s reconcile, Add/Update event confirm", request.NamespacedName.String())
	return reconcile.Result{}, nil
}

func (r *ReconcileAppNode) handleQueue() {
	ech, err := r.eventQ.GetChannel()
	if err != nil {
		blog.Errorf("ReconcileAppNode get event queue failed, %s", err)
		return
	}
	for {
		select {
		case event, ok := <-ech:
			if !ok {
				blog.Errorf("ReconcileAppNode get error event type.")
				return
			}
			node, tok := event.Data.(*meshv1.AppNode)
			if !tok {
				blog.Errorf("ReconcileAppNode get unknown data info. discard.")
				continue
			}
			switch event.Type {
			case watch.EventAdded:
				r.onAdd(node)
			case watch.EventDeleted:
				r.onDelete(node)
			case watch.EventUpdated:
				r.onUpdate(node)
			default:
				blog.Warnf("ReconcileAppNode get unknown Event Type: %s.", event.Type)
			}
			//todo(DeveloperJim) default info for timeout?
		}
	}
}

// onAdd add new AppSvc to kube-apiserver
// todo(DeveloperJim): push event queue back when operation failed?
func (r *ReconcileAppNode) onAdd(node *meshv1.AppNode) {
	instance := &meshv1.AppNode{}
	key, kerr := client.ObjectKeyFromObject(node)
	if kerr != nil {
		blog.Errorf("ReconcileAppNode formate %s/%s to Object key failed, %s", node.GetNamespace(), node.GetName(), kerr.Error())
		return
	}
	if err := ns.CheckNamespace(r.localCache, r, node.GetNamespace()); err != nil {
		blog.Errorf("ReconcileAppNode checks namespace %s failed, %s", node.GetNamespace(), err.Error())
		return
	}
	err := r.Get(context.TODO(), key, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, create new one directly
			if err := r.Create(context.TODO(), node); err != nil {
				blog.Errorf("ReconcileAppNode create new AppNode %s failed, %s", key.String(), err.Error())
				return
			}
			blog.Infof("ReconcileAppNode creat new AppNode %s on EventAdded successfully", key.String())
			return
		}
		// Error reading the object
		blog.Errorf("AppNode reads local cache %s failed, %s", key.String(), err.Error())
		return
	}
	//get exist data, ready to update
	if reflect.DeepEqual(instance.Spec, node.Spec) {
		blog.Warnf("ReconcileAppNode get deepEqual in EventAdded, key: %s", key.String())
		return
	}
	//fix(DeveloperJim): change Spec data & Status
	instance.Spec = node.Spec
	instance.Status.LastUpdateTime = metav1.Now()
	//ready to Udpate
	if err := r.Update(context.TODO(), instance); err != nil {
		blog.Errorf("ReconcileAppNode update %s in EventAdded failed, %s", key.String(), err.Error())
		return
	}
	if err := r.Status().Update(context.TODO(), instance); err != nil {
		blog.Errorf("ReconcileAppNode update %s Status in EventAdded failed, %s", key.String(), err.Error())
		return
	}
	blog.Warnf("ReconcileAppNode update %s successfully in EventAdded, maybe TaskGroup cache lost in cluster", key.String())
}

func (r *ReconcileAppNode) onUpdate(node *meshv1.AppNode) {
	instance := &meshv1.AppNode{}
	key, kerr := client.ObjectKeyFromObject(node)
	if kerr != nil {
		blog.Errorf("ReconcileAppNode formate %s/%s to Object key failed, %s", node.GetNamespace(), node.GetName(), kerr.Error())
		return
	}
	err := r.Get(context.TODO(), key, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, create new one directly
			if err := r.Create(context.TODO(), node); err != nil {
				blog.Errorf("ReconcileAppNode create new AppNode %s on EventUpated failed, %s", key.String(), err.Error())
				return
			}
			blog.Warnf("ReconcileAppNode creat new AppNode %s on EventUpated successfully, maybe local cache lost data", key.String())
			return
		}
		// Error reading the object
		blog.Errorf("AppNode reads local cache %s on EventUpdated failed, %s", key.String(), err.Error())
		return
	}
	//get exist data, ready to update
	if reflect.DeepEqual(instance.Spec, node.Spec) {
		blog.Warnf("ReconcileAppNode get deepEqual in EventAdded, key: %s", key.String())
		return
	}
	//fix(DeveloperJim): change Spec data & Status
	instance.Spec = node.Spec
	instance.Status.LastUpdateTime = metav1.Now()
	//ready to Udpate Spec
	if err := r.Update(context.TODO(), instance); err != nil {
		blog.Errorf("ReconcileAppNode update %s in EventUpdated failed, %s", key.String(), err.Error())
		return
	}
	if err := r.Status().Update(context.TODO(), instance); err != nil {
		blog.Errorf("ReconcileAppNode update %s Status in EventUpdated failed, %s", key.String(), err.Error())
		return
	}
	blog.Infof("ReconcileAppNode update %s successfully", key.String())
}

func (r *ReconcileAppNode) onDelete(node *meshv1.AppNode) {
	//ready to Udpate
	if err := r.Delete(context.TODO(), node); err != nil {
		blog.Errorf("ReconcileAppNode DELETE %s/%s failed, %s", node.GetNamespace(), node.GetName(), err.Error())
		return
	}
	blog.Infof("ReconcileAppNode DELETE %s/%s successfully", node.GetNamespace(), node.GetName())
}
