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

package clusterops

import (
	"context"
	"encoding/json"
	"fmt"
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
	NodeName  string
	Desired   bool
}

const (
	// DefaultTimeout default timeout to call k8s api
	DefaultTimeout = 10 * time.Second
)

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
	node, err := nodeCli.Get(ctx, nodeInfo.NodeName, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("ClusterUpdateScheduleNode GetClusterNode[%s] failed: %v", nodeInfo.NodeName, err)
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
		_, err = nodeCli.Patch(ctx, nodeInfo.NodeName, types.StrategicMergePatchType, patchBytes, patchOptions)
	} else {
		updateOptions := metav1.UpdateOptions{}
		_, err = nodeCli.Update(ctx, node, updateOptions)
	}
	if err != nil {
		blog.Errorf("ClusterUpdateScheduleNode CreateTwoWayMergePatch[%s] failed: %v", nodeInfo.NodeName, err)
	}

	return err
}

// ListClusterNodes query cluster all nodes
func (ko *K8SOperator) ListClusterNodes(ctx context.Context, clusterID string) ([]*v1.Node, error) {
	if ko == nil {
		return nil, ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("ListClusterNodes GetClusterClient failed: %v", err)
		return nil, err
	}

	var (
		defaultTimeout int64 = 300
		nodes          []*v1.Node
	)
	nodeList, err := clientInterface.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		TimeoutSeconds: &defaultTimeout,
	})
	if err != nil {
		blog.Errorf("ListClusterNodes ListNodes[%s] failed: %v", clusterID, err)
		return nil, err
	}

	blog.Infof("cluster[%s] ListClusterNodes successful: %v", clusterID, len(nodeList.Items))
	for i := range nodeList.Items {
		nodes = append(nodes, &nodeList.Items[i])
	}

	return nodes, nil
}

// QueryNodeOption query node option
type QueryNodeOption struct {
	ClusterID string
	NodeName  string
	NodeIP    string
}

// GetClusterNode query cluster node by nodeName or nodeIP
func (ko *K8SOperator) GetClusterNode(ctx context.Context, nodeOption QueryNodeOption) (*v1.Node, error) {
	if ko == nil {
		return nil, ErrServerNotInit
	}
	if nodeOption.ClusterID == "" || (nodeOption.NodeIP == "" && nodeOption.NodeName == "") {
		return nil, fmt.Errorf("GetClusterNode paras empty")
	}

	isName := len(nodeOption.NodeName) > 0
	var (
		nodes []*v1.Node
		err   error
	)
	nodes, err = ko.ListClusterNodes(ctx, nodeOption.ClusterID)
	if err != nil {
		blog.Errorf("GetClusterNode ListClusterNodes failed: %v", err)
		return nil, err
	}
	nodeIPsMap, nodeNamesMap := getClusterNodesMapInfo(nodes)

	if isName {
		node, ok := nodeNamesMap[nodeOption.NodeName]
		if ok {
			blog.Infof("GetClusterNode[%s] successful", nodeOption.NodeName)
			return node, nil
		}

		return nil, fmt.Errorf("GetClusterNode[%s] NodeName[%s] not found", nodeOption.ClusterID, nodeOption.NodeName)
	}

	node, ok := nodeIPsMap[nodeOption.NodeIP]
	if ok {
		blog.Infof("GetClusterNode[%s] successful", nodeOption.NodeIP)
		return node, nil
	}

	return nil, fmt.Errorf("GetClusterNode[%s] NodeIP[%s] not found", nodeOption.ClusterID, nodeOption.NodeIP)
}

// ListNodeOption list node option
type ListNodeOption struct {
	ClusterID string
	NodeIPs   []string
	NodeNames []string
}

func getClusterNodesMapInfo(nodes []*v1.Node) (map[string]*v1.Node, map[string]*v1.Node) {
	var (
		nodeIPsToNode   = make(map[string]*v1.Node, 0)
		nodeNamesToNode = make(map[string]*v1.Node, 0)
	)
	for i := range nodes {
		nodeName := nodes[i].Name
		nodeIP := ""
		for _, address := range nodes[i].Status.Addresses {
			if address.Type == v1.NodeInternalIP {
				nodeIP = address.Address
			}
		}
		if len(nodeName) > 0 {
			nodeNamesToNode[nodeName] = nodes[i]
		}
		if len(nodeIP) > 0 {
			nodeIPsToNode[nodeIP] = nodes[i]
		}
	}

	return nodeIPsToNode, nodeNamesToNode
}

// ListClusterNodesByIPsOrNames query cluster nodes by nodeNames or nodeIPs
func (ko *K8SOperator) ListClusterNodesByIPsOrNames(
	ctx context.Context, nodeOption ListNodeOption) ([]*v1.Node, error) {
	if ko == nil {
		return nil, ErrServerNotInit
	}
	if nodeOption.ClusterID == "" || (len(nodeOption.NodeIPs) == 0 && len(nodeOption.NodeNames) == 0) {
		return nil, fmt.Errorf("ListClusterNodesByIPsOrNames paras empty")
	}

	isName := len(nodeOption.NodeNames) > 0
	var (
		nodes []*v1.Node
		err   error
	)
	nodes, err = ko.ListClusterNodes(ctx, nodeOption.ClusterID)
	if err != nil {
		blog.Errorf("ListClusterNodesByIPsOrNames ListClusterNodes failed: %v", err)
		return nil, err
	}
	nodeIPsMap, nodeNamesMap := getClusterNodesMapInfo(nodes)

	var nodeList = make([]*v1.Node, 0)
	if isName {
		for _, name := range nodeOption.NodeNames {
			if node, ok := nodeNamesMap[name]; ok {
				nodeList = append(nodeList, node)
			}
		}

		blog.Infof("ListClusterNodesByIPsOrNames names:[%s] successful", nodeOption.ClusterID)
		return nodeList, nil
	}

	for _, ip := range nodeOption.NodeIPs {
		if node, ok := nodeIPsMap[ip]; ok {
			nodeList = append(nodeList, node)
		}
	}

	blog.Infof("ListClusterNodesByIPsOrNames ips:[%s] successful", nodeOption.ClusterID)
	return nodeList, nil
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
	DryRun          bool
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
		DeleteEmptyDirData:  drainHelper.DeleteLocalData,
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

// UpdateNodeLabels update node labels
func (ko *K8SOperator) UpdateNodeLabels(ctx context.Context, clusterID, nodeName string,
	labels map[string]string) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("UpdateNodeLabels GetClusterClient failed: %v", err)
		return err
	}

	nodeCli := clientInterface.CoreV1().Nodes()
	node, err := nodeCli.Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("UpdateNodeLabels GetClusterNode[%s] failed: %v", nodeName, err)
		return err
	}

	oldData, err := json.Marshal(node)
	if err != nil {
		return err
	}
	node.Labels = labels

	newData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, node)
	if patchErr == nil {
		patchOptions := metav1.PatchOptions{}
		_, err = nodeCli.Patch(ctx, nodeName, types.StrategicMergePatchType, patchBytes, patchOptions)
	} else {
		updateOptions := metav1.UpdateOptions{}
		_, err = nodeCli.Update(ctx, node, updateOptions)
	}
	if err != nil {
		blog.Errorf("UpdateNodeLabels CreateTwoWayMergePatch[%s] failed: %v", nodeName, err)
	}

	return err
}

// UpdateNodeAnnotations update node annotations
func (ko *K8SOperator) UpdateNodeAnnotations(ctx context.Context, clusterID, nodeName string,
	annotations map[string]string, merge bool) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("UpdateNodeAnnotations GetClusterClient failed: %v", err)
		return err
	}

	nodeCli := clientInterface.CoreV1().Nodes()
	node, err := nodeCli.Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("UpdateNodeAnnotations GetClusterNode[%s] failed: %v", nodeName, err)
		return err
	}

	oldData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	if merge {
		// merge annotations
		for k, v := range annotations {
			node.Annotations[k] = v
		}
	} else {
		node.Annotations = annotations
	}

	newData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, node)
	if patchErr == nil {
		patchOptions := metav1.PatchOptions{}
		_, err = nodeCli.Patch(ctx, nodeName, types.StrategicMergePatchType, patchBytes, patchOptions)
	} else {
		updateOptions := metav1.UpdateOptions{}
		_, err = nodeCli.Update(ctx, node, updateOptions)
	}
	if err != nil {
		blog.Errorf("UpdateNodeAnnotations CreateTwoWayMergePatch[%s] failed: %v", nodeName, err)
	}

	return err
}

// UpdateNodeTaints update node taints
func (ko *K8SOperator) UpdateNodeTaints(ctx context.Context, clusterID, nodeName string,
	taints []v1.Taint) error {
	if ko == nil {
		return ErrServerNotInit
	}
	clientInterface, err := ko.GetClusterClient(clusterID)
	if err != nil {
		blog.Errorf("UpdateNodeTaints GetClusterClient failed: %v", err)
		return err
	}

	nodeCli := clientInterface.CoreV1().Nodes()
	node, err := nodeCli.Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("UpdateNodeTaints GetClusterNode[%s] failed: %v", nodeName, err)
		return err
	}

	oldData, err := json.Marshal(node)
	if err != nil {
		return err
	}
	node.Spec.Taints = taints

	newData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, node)
	if patchErr == nil {
		patchOptions := metav1.PatchOptions{}
		_, err = nodeCli.Patch(ctx, nodeName, types.StrategicMergePatchType, patchBytes, patchOptions)
	} else {
		updateOptions := metav1.UpdateOptions{}
		_, err = nodeCli.Update(ctx, node, updateOptions)
	}
	if err != nil {
		blog.Errorf("UpdateNodeTaints CreateTwoWayMergePatch[%s] failed: %v", nodeName, err)
	}

	return err
}
