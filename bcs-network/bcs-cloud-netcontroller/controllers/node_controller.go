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
	"reflect"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	"github.com/Tencent/bk-bcs/bcs-network/internal/grpclb"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netcontroller/internal/option"
	cloudAPI "github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netcontroller/pkg/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netcontroller/pkg/cloud/aws"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netcontroller/pkg/cloud/qcloud"
	"github.com/Tencent/bk-bcs/bcs-network/internal/constant"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

// NodeLabelChangePredicate filter node event
type NodeLabelChangePredicate struct {
	predicate.Funcs
}

// Update override update func
func (np *NodeLabelChangePredicate) Update(e event.UpdateEvent) bool {
	oldNode, ok1 := e.ObjectOld.(*corev1.Node)
	newNode, ok2 := e.ObjectNew.(*corev1.Node)
	if ok1 && ok2 {
		if !reflect.DeepEqual(oldNode.Labels, newNode.Labels) {
			return true
		}
		return false
	}
	return true
}

// NodeNetworkReconciler reconciles a Node object
type NodeNetworkReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Option *option.ControllerOption

	NodeEventer record.EventRecorder

	cloudNetClient pbcloudnet.CloudNetserviceClient
	cloudClient    cloudAPI.Interface
	processor      *Processor
}

// Reconcile reconcile node info
func (r *NodeNetworkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("event trigger", req.String(), req.NamespacedName.String())

	r.processor.OnEvent()

	return ctrl.Result{}, nil
}

// initCloudNetClient
func (r *NodeNetworkReconciler) initCloudNetClient() error {
	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithBalancer(grpc.RoundRobin(grpclb.NewPseudoResolver(r.Option.CloudNetServiceEndpoints))),
	)
	if err != nil {
		r.Log.Error(err, "init cloud netservice client failed")
		return err
	}

	cloudNetClient := pbcloudnet.NewCloudNetserviceClient(conn)
	r.cloudNetClient = cloudNetClient
	return nil
}

// initCloud init aws or tencent cloud client
func (r *NodeNetworkReconciler) initCloud() error {
	var cloudClient cloudAPI.Interface
	switch r.Option.Cloud {
	case constant.CLOUD_KIND_TENCENT:
		cloudClient = qcloud.New()
	case constant.CLOUD_KIND_AWS:
		cloudClient = aws.New()
	default:
		return fmt.Errorf("error cloud mode %s", r.Option.Cloud)
	}
	if err := cloudClient.Init(); err != nil {
		return fmt.Errorf("init cloud client failed, err %s", err.Error())
	}
	r.cloudClient = cloudClient
	return nil
}

// initProcessor
func (r *NodeNetworkReconciler) initProcessor() error {
	processor := NewProcessor(r, r.Option, r.cloudNetClient, r.cloudClient, r.NodeEventer)
	go func() {
		err := processor.Run(context.Background())
		if err != nil {
			r.Log.Error(err, "processor exits")
		}
	}()
	r.processor = processor
	return nil
}

// initLogs
func (r *NodeNetworkReconciler) initBlogs() {
	blog.InitLogs(r.Option.LogConfig)
}

// SetupWithManager setup reconciler
func (r *NodeNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.initBlogs()

	if err := r.initCloudNetClient(); err != nil {
		return err
	}
	if err := r.initCloud(); err != nil {
		return err
	}
	if err := r.initProcessor(); err != nil {
		return err
	}

	// TODO: leader election time is too long
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Watches(&source.Kind{Type: &cloudv1.NodeNetwork{}}, &handler.EnqueueRequestForObject{}).
		WithEventFilter(&NodeLabelChangePredicate{}).
		Complete(r)
}
