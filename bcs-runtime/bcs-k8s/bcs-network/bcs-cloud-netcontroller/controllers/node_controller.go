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
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/pkg/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"

	"github.com/go-logr/logr"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NodeReconciler reconciler for k8s node
type NodeReconciler struct {
	// Client client for reconciler
	Client client.Client
	Log    logr.Logger

	// Option option for bcs-cloud-netcontroller
	Option *option.ControllerOption

	// CloudClient client for cloud
	CloudClient cloud.Interface

	NodeEventer record.EventRecorder
}

// getNodePredicate filter listener events
func getNodePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newNode, okNew := e.ObjectNew.(*k8scorev1.Node)
			oldNode, okOld := e.ObjectOld.(*k8scorev1.Node)
			if !okNew || !okOld {
				return false
			}
			if newNode.DeletionTimestamp != nil {
				return true
			}
			if reflect.DeepEqual(newNode.GetLabels(), oldNode.GetLabels()) {
				blog.V(5).Infof("node %s/%s updated, but labels not change",
					oldNode.GetName(), oldNode.GetNamespace())
				return false
			}
			return true
		},
	}
}

// Reconcile reconcile k8s node info
func (nr *NodeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	tmpNode := &k8scorev1.Node{}
	if err := nr.Client.Get(context.Background(), req.NamespacedName, tmpNode); err != nil {
		if k8serrors.IsNotFound(err) {
			// node is deleted, ensure crd delete
			if inErr := nr.ensureNodeDelete(req.NamespacedName); inErr != nil {
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: time.Duration(5 * time.Second),
				}, nil
			}
			return ctrl.Result{}, nil
		}
	}
	if err := nr.ensureNodeUpdate(tmpNode); err != nil {
		blog.Warnf("ensure node %s labels %v update failed, err %s",
			tmpNode.GetName(), tmpNode.GetLabels(), err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 10 * time.Second,
		}, nil
	}

	return ctrl.Result{}, nil
}

// ensure node network delete
func (nr *NodeReconciler) ensureNodeDelete(namespacedName k8stypes.NamespacedName) error {
	blog.Infof("node %s deleted", namespacedName.String())
	tmpNodeNetwork := &cloudv1.NodeNetwork{}
	if err := nr.Client.Get(context.Background(), k8stypes.NamespacedName{
		Namespace: constant.CloudCrdNamespaceBcsSystem,
		Name:      namespacedName.Name,
	}, tmpNodeNetwork); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("no nodenetwork found, do nothing")
			return nil
		}
	}
	if err := nr.Client.Delete(context.Background(), tmpNodeNetwork); err != nil {
		blog.Warnf("delete node network %s/%s failed, err %s",
			tmpNodeNetwork.GetName(), tmpNodeNetwork.GetNamespace(), err.Error())
		return err
	}
	blog.Infof("trigger delete node network %s/%s", tmpNodeNetwork.GetName(), tmpNodeNetwork.GetNamespace())
	return nil
}

// ensure node network update
func (nr *NodeReconciler) ensureNodeUpdate(node *k8scorev1.Node) error {
	blog.Infof("node %s label updated", node.GetName())
	tmpNodeNetwork := &cloudv1.NodeNetwork{}
	foundNodeNetwork := true
	if err := nr.Client.Get(context.Background(), k8stypes.NamespacedName{
		Namespace: constant.CloudCrdNamespaceBcsSystem,
		Name:      node.GetName(),
	}, tmpNodeNetwork); err != nil {
		if k8serrors.IsNotFound(err) {
			foundNodeNetwork = false
		} else {
			blog.Warnf("get node network failed, err %s", err.Error())
			return err
		}
	}
	// node is created or updated
	nodeLabels := node.GetLabels()
	if nodeLabels != nil {
		labelValue, hasLabel := nodeLabels[constant.NodeLabelKeyForNodeNetwork]
		if hasLabel && labelValue == strconv.FormatBool(true) {
			// node need node network
			if foundNodeNetwork {
				// update node network crd
				if err := nr.updateNodeNetwork(node, tmpNodeNetwork); err != nil {
					return err
				}
				return nil
			}
			// create node network crd
			if err := nr.createNodeNetwork(node); err != nil {
				return err
			}
			return nil
		}
	}
	// node doesn't need network
	if foundNodeNetwork {
		if err := nr.Client.Delete(context.Background(), tmpNodeNetwork); err != nil {
			blog.Warnf("delete node network %s/%s failed, err %s",
				tmpNodeNetwork.GetName(), tmpNodeNetwork.GetNamespace(), err.Error())
			return err
		}
		blog.Infof("trigger delete node network %s/%s", tmpNodeNetwork.GetName(), tmpNodeNetwork.GetNamespace())
		return nil
	}
	blog.Infof("no nodenetwork found, do nothing")
	return nil
}

// update crd NodeNetwork
func (nr *NodeReconciler) updateNodeNetwork(node *k8scorev1.Node, nodeNet *cloudv1.NodeNetwork) error {
	// get eni number of node
	eniNum, err := nr.getEniNumberFromNodeLabels(node)
	if err != nil {
		blog.Warnf("get eni number from node labels falied, err %s", err.Error())
		return err
	}
	if eniNum == nodeNet.Spec.ENINum {
		blog.Infof("eni num not change, do nothing")
		return nil
	}
	eniLimit, _, err := nr.CloudClient.GetENILimit(nodeNet.Spec.VM.InstanceIP)
	if err != nil {
		blog.Warnf("get eni limit of instance %s failed, err %s", nodeNet.Spec.VM.InstanceIP, err.Error())
		return err
	}
	if eniNum > eniLimit-1 {
		blog.Warnf("request extra eni number %d exceed node limit %d", eniNum, eniLimit-1)
		return fmt.Errorf("request extra eni number %d exceed node limit %d", eniNum, eniLimit-1)
	}
	nodeNet.Spec.ENINum = eniNum
	if err := nr.Client.Update(context.Background(), nodeNet); err != nil {
		blog.Warnf("update node network %s/%s failed, err %s", nodeNet.GetName(), nodeNet.GetNamespace(), err.Error())
		return err
	}
	return nil
}

func (nr *NodeReconciler) getEniNumberFromNodeLabels(node *k8scorev1.Node) (int, error) {
	labelValue, ok := node.GetLabels()[constant.NodeLabelKeyFroNodeNetworkEniNum]
	if !ok {
		return 1, nil
	}
	num, err := strconv.Atoi(labelValue)
	if err != nil {
		return 0, err
	}
	return num, nil
}

// create crd NodeNetwork
func (nr *NodeReconciler) createNodeNetwork(node *k8scorev1.Node) error {
	nodeAddr := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == k8scorev1.NodeInternalIP {
			nodeAddr = addr.Address
			break
		}
	}
	if len(nodeAddr) == 0 {
		blog.Warnf("node %s has no internal ip", node.GetName())
		return fmt.Errorf("node %s has no internal ip", node.GetName())
	}
	// get vm info for node
	nodeVMInfo, err := nr.CloudClient.GetVMInfo(nodeAddr)
	if err != nil {
		blog.Warnf("get vm node info by addr %s failed, err %s", nodeAddr, err.Error())
		return err
	}
	// get eni number of node
	eniNum, err := nr.getEniNumberFromNodeLabels(node)
	if err != nil {
		blog.Warnf("get eni number from node labels falied, err %s", err.Error())
		return err
	}
	eniLimit, ipLimit, err := nr.CloudClient.GetENILimit(nodeVMInfo.InstanceIP)
	if err != nil {
		blog.Warnf("get eni limit of instance %s failed, err %s", nodeVMInfo.InstanceIP, err.Error())
		return err
	}
	if eniNum > eniLimit-1 {
		blog.Warnf("request extra eni number %d exceed node limit %d", eniNum, eniLimit-1)
		return fmt.Errorf("request extra eni number %d exceed node limit %d", eniNum, eniLimit-1)
	}
	// construct new node network
	newNodeNetwork := &cloudv1.NodeNetwork{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: cloudv1.SchemeGroupVersion.Version,
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      node.GetName(),
			Namespace: constant.CloudCrdNamespaceBcsSystem,
		},
		Spec: cloudv1.NodeNetworkSpec{
			Cluster:     nr.Option.Cluster,
			Hostname:    node.GetName(),
			NodeAddress: nodeVMInfo.InstanceIP,
			VM:          nodeVMInfo,
			ENINum:      eniNum,
			IPNumPerENI: ipLimit - 1,
		},
	}
	if err := nr.Client.Create(context.Background(), newNodeNetwork); err != nil {
		blog.Warnf("create nodenetwork crd %s/%s failed, err %s",
			newNodeNetwork.GetName(), newNodeNetwork.GetNamespace(), err.Error())
		return err
	}
	return nil
}

// SetupWithManager set node reconciler
func (nr *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8scorev1.Node{}).
		WithEventFilter(getNodePredicate()).
		Owns(&cloudv1.NodeNetwork{}).
		Complete(nr)
}
