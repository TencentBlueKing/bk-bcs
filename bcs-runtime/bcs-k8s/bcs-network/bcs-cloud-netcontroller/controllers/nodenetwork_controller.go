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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/internal/option"
	cloudAPI "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/pkg/cloud"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"

	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeNetworkReconciler reconciles a Node object
type NodeNetworkReconciler struct {
	Client client.Client
	Option *option.ControllerOption

	NodeEventer record.EventRecorder

	CloudNetClient pbcloudnet.CloudNetserviceClient
	CloudClient    cloudAPI.Interface
	processor      *Processor
}

// Reconcile reconcile node info
func (r *NodeNetworkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("node network %s event trigger", req.NamespacedName.String())
	r.processor.OnEvent(NodeNetworkEvent{
		NodeName:      req.NamespacedName.Name,
		NodeNamespace: req.NamespacedName.Namespace,
	})

	return ctrl.Result{}, nil
}

// initProcessor
func (r *NodeNetworkReconciler) initProcessor() error {
	processor := NewProcessor(r.Client, r.Option, r.CloudNetClient, r.CloudClient, r.NodeEventer)
	go func() {
		err := processor.Run(context.Background())
		if err != nil {
			blog.Errorf("processor exits, err %s", err.Error())
		}
	}()
	r.processor = processor
	return nil
}

// SetupWithManager setup reconciler
func (r *NodeNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.initProcessor(); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudv1.NodeNetwork{}).
		Complete(r)
}
