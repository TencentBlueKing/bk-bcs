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

package clusterops

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/drain"
)

// NodeInfo node info
type NodeInfo struct {
	ClusterID string
	NodeIP    string
	Desired   bool
}

// ClusterUpdateScheduleNode uncordon node or cordon node for desired status
func (ko *K8SOperator) ClusterUpdateScheduleNode(ctx context.Context, nodeInfo NodeInfo) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(nodeInfo.ClusterID)
	if err != nil {
		blog.Errorf("ClusterUpdateScheduleNode GetClusterClient failed: %v", err)
		return err
	}

	nodeCli := clientInterface.CoreV1().Nodes()
	node, err := nodeCli.Get(ctx, nodeInfo.NodeIP, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("ClusterUpdateScheduleNode GetClusterNode[%s] failed: %v", nodeInfo.NodeIP, err)
		return err
	}

	oldData, err := json.Marshal(node)
	if err != nil {
		return err
	}
	node.Spec.Unschedulable = nodeInfo.Desired

	newData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, node)
	if patchErr == nil {
		patchOptions := metav1.PatchOptions{}
		_, err = nodeCli.Patch(ctx, nodeInfo.NodeIP, types.StrategicMergePatchType, patchBytes, patchOptions)
	} else {
		updateOptions := metav1.UpdateOptions{}
		_, err = nodeCli.Update(ctx, node, updateOptions)
	}
	if err != nil {
		blog.Errorf("ClusterUpdateScheduleNode CreateTwoWayMergePatch[%s] failed: %v", nodeInfo.NodeIP, err)
	}

	return err
}

// ListClusterNodes list nodes in cluster
func (ko *K8SOperator) ListClusterNodes(ctx context.Context, clusterID string) (*v1.NodeList, error) {
	if ko == nil {
		return nil, ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("ListClusterNodes GetClusterClient failed: %v", err)
		return nil, err
	}

	return clientInterface.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}

// DrainHelper describe drain args
type DrainHelper struct {
	Force               bool
	GracePeriodSeconds  int
	IgnoreAllDaemonSets bool
	Timeout             int
	DeleteLocalData     bool
	Selector            string
	PodSelector         string

	// DisableEviction forces drain to use delete rather than evict
	DisableEviction bool

	DryRun bool

	// SkipWaitForDeleteTimeoutSeconds ignores pods that have a
	// DeletionTimeStamp > N seconds. It's up to the user to decide when this
	// option is appropriate; examples include the Node is unready and the pods
	// won't drain otherwise
	SkipWaitForDeleteTimeoutSeconds int
}

// DrainNode drain node
func (ko *K8SOperator) DrainNode(ctx context.Context, clusterID, nodeName string, drainHelper DrainHelper) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("DrainNode GetClusterClient failed: %v", err)
		return err
	}

	drainer := &drain.Helper{
		Ctx:                 ctx,
		Client:              clientInterface,
		Force:               drainHelper.Force,
		GracePeriodSeconds:  drainHelper.GracePeriodSeconds,
		IgnoreAllDaemonSets: drainHelper.IgnoreAllDaemonSets,
		Timeout:             time.Second * time.Duration(drainHelper.Timeout),
		Selector:            drainHelper.Selector,
		PodSelector:         drainHelper.PodSelector,
		DisableEviction:     drainHelper.DisableEviction,
		Out:                 io.Discard,
		ErrOut:              io.Discard,
	}
	if drainHelper.DryRun {
		drainer.DryRunStrategy = cmdutil.DryRunServer
	}
	return drain.RunNodeDrain(drainer, nodeName)
}
