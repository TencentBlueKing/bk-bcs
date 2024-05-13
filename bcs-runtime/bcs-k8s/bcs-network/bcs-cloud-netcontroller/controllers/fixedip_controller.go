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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/grpclb"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
)

// CloudIPPredicate filter cloud ip event
type CloudIPPredicate struct {
	predicate.Funcs
}

// FixedIPReconciler clean fixed ip object
type FixedIPReconciler struct {
	Ctx context.Context

	client.Client
	Log    logr.Logger
	Option *option.ControllerOption

	cloudNetClient pbcloudnet.CloudNetserviceClient
	// ipCleaner      *IPCleaner
}

func (f *FixedIPReconciler) initCloudNetClient() error {
	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithBalancer(grpc.RoundRobin(grpclb.NewPseudoResolver(f.Option.CloudNetServiceEndpoints))),
	)
	if err != nil {
		f.Log.Error(err, "init cloud netservice client failed")
		return err
	}

	cloudNetClient := pbcloudnet.NewCloudNetserviceClient(conn)
	f.cloudNetClient = cloudNetClient
	return nil
}

// Reconcile reconcile fixed ip object
func (f *FixedIPReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (f *FixedIPReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := f.initCloudNetClient(); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudv1.CloudIP{}).
		Watches(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForObject{}).
		Complete(f)
}
