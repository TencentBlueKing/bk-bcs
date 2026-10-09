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

// Package cluster xxx
package cluster

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/registry"
	v1 "k8s.io/api/core/v1"

	impl "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/metric"
)

const (
	// ListK8SNodePath original kubernetes path with bcs prefix
	ListK8SNodePath = "/clusters/%s/api/v1/nodes?labelSelector=node-role.kubernetes.io/master!=true"
	// UpdateK8SNodePath original kubernetes path with bcs prefix
	UpdateK8SNodePath = "/clusters/%s/api/v1/nodes/%s"
)

// Client for cluster operation
type Client interface {
	ListClusterNodes(clusterID string) ([]*Node, error)
	UpdateNodeMetadata(clusterID, nodeName string, labels, annotations map[string]interface{}) error
	ListNodesByLabel(clusterID string, labels map[string]interface{}) (map[string]*Node, error)
	UpdateNodegroupMax(clusterID, nodegroupID string, max int) error
	GetNodeDetail(ip string) (*impl.Node, error)
}

// nolint
// ClusterClientOptions option for api-gateway client
type ClusterClientOptions struct {
	Endpoint string
	Token    string
	Sender   requester.Requester
}

// nolint
// ClusterManagerClientOptions for cluster manager client
type ClusterManagerClientOptions struct {
	// Name for resource-manager registry
	Name string
	// Etcd endpoints information
	Etcd []string
	// EtcdConfig tls config for etcd
	EtcdConfig *tls.Config
	// ClientConfig tls config
	ClientConfig *tls.Config
	cmClient     impl.ClusterManagerService
}

// apiGateway bcs apiGateway client
type apiGateway struct {
	clusterClient        *ClusterClientOptions
	clusterManagerClient *ClusterManagerClientOptions
	defaultHeader        map[string]string
}

// NewClient new bcs apiGateway client
func NewClient(clusterOpt *ClusterClientOptions, cmClientOpt *ClusterManagerClientOptions) Client {
	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", clusterOpt.Token)
	header["X-Bcs-Client"] = "bcs-nodegroup-manager"
	if clusterOpt.Sender == nil {
		clusterOpt.Sender = requester.NewRequester()
	}
	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(cmClientOpt.Etcd...),
			registry.TLSConfig(cmClientOpt.EtcdConfig)),
		),
		grpc.AuthTLS(cmClientOpt.ClientConfig),
	)
	cmClientOpt.cmClient = impl.NewClusterManagerService(cmClientOpt.Name, c)
	return &apiGateway{
		clusterClient:        clusterOpt,
		clusterManagerClient: cmClientOpt,
		defaultHeader:        header,
	}
}

// ListClusterNodes list nodes by clusterID
func (c *apiGateway) ListClusterNodes(clusterID string) ([]*Node, error) {
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportClusterClientRequestMetric(clusterID, "http", "ListClusterNodes", err, startTime)
	}()
	url := fmt.Sprintf("%s%s", c.clusterClient.Endpoint, fmt.Sprintf(ListK8SNodePath, clusterID))
	rawResponse, err := c.clusterClient.Sender.DoGetRequest(url, c.defaultHeader)
	if err != nil {
		return nil, fmt.Errorf("DoGetRequest error: %s", err.Error())
	}
	var k8sNodeList v1.NodeList
	if err := json.Unmarshal(rawResponse, &k8sNodeList); err != nil {
		return nil, fmt.Errorf("decode NodeList response failed %s, raw response %s",
			err.Error(), string(rawResponse))
	}
	nodeList := make([]*Node, 0)
	for _, k8sNode := range k8sNodeList.Items {
		address := k8sNode.Status.Addresses
		var nodeIP string
		for _, ip := range address {
			if ip.Type == v1.NodeInternalIP {
				nodeIP = ip.Address
			}
		}
		var status string
		for _, condition := range k8sNode.Status.Conditions {
			if condition.Type == "Ready" {
				status = string(condition.Status)
			}
		}
		node := &Node{
			Name:        k8sNode.Name,
			IP:          nodeIP,
			Status:      status,
			Labels:      k8sNode.ObjectMeta.Labels,
			Annotations: k8sNode.ObjectMeta.Annotations,
		}
		nodeList = append(nodeList, node)
	}
	return nodeList, nil
}

// ListNodesByLabel list nodes by clusterID and label
func (c *apiGateway) ListNodesByLabel(clusterID string, labels map[string]interface{}) (map[string]*Node, error) {
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportClusterClientRequestMetric(clusterID, "http", "ListNodesByLabel", err, startTime)
	}()
	url := fmt.Sprintf("%s%s", c.clusterClient.Endpoint, fmt.Sprintf(ListK8SNodePath, clusterID))
	for key, value := range labels {
		url = fmt.Sprintf("%s,%s=%s", url, key, value)
	}
	blog.Infof("ListNodesByLabel url: %s", url)
	rawResponse, err := c.clusterClient.Sender.DoGetRequest(url, c.defaultHeader)
	if err != nil {
		return nil, fmt.Errorf("DoGetRequest error: %s", err.Error())
	}
	var k8sNodeList v1.NodeList
	if err := json.Unmarshal(rawResponse, &k8sNodeList); err != nil {
		return nil, fmt.Errorf("decode NodeList response failed %s, raw response %s",
			err.Error(), string(rawResponse))
	}
	nodeList := make(map[string]*Node, 0)
	for _, k8sNode := range k8sNodeList.Items {
		address := k8sNode.Status.Addresses
		var nodeIP string
		for _, ip := range address {
			if ip.Type == v1.NodeInternalIP {
				nodeIP = ip.Address
			}
		}
		var status string
		for _, condition := range k8sNode.Status.Conditions {
			if condition.Type == "Ready" {
				status = string(condition.Status)
			}
		}
		node := &Node{
			Name:        k8sNode.Name,
			IP:          nodeIP,
			Status:      status,
			Labels:      k8sNode.Labels,
			Annotations: k8sNode.Annotations,
		}
		nodeList[k8sNode.Name] = node
	}
	blog.Infof("filter nodes:%v", nodeList)
	return nodeList, nil
}

// UpdateNodeMetadata update node labels and annotations by clusterID and nodeName
func (c *apiGateway) UpdateNodeMetadata(clusterID, nodeName string, labels map[string]interface{},
	annotations map[string]interface{}) error {
	var err error
	startTime := time.Now()
	defer func() {
		metric.ReportClusterClientRequestMetric(clusterID, "http", "UpdateNodeMetadata", err, startTime)
	}()
	url := fmt.Sprintf("%s%s", c.clusterClient.Endpoint, fmt.Sprintf(UpdateK8SNodePath, clusterID, nodeName))
	metaDataMap := make(map[string]interface{})
	if labels != nil {
		metaDataMap["labels"] = labels
	}
	if annotations != nil {
		metaDataMap["annotations"] = annotations
	}
	metaData := map[string]interface{}{"metadata": metaDataMap}
	metaDataStr, err := json.Marshal(metaData)
	if err != nil {
		return fmt.Errorf("json marshal metadata to byte err:%s", err.Error())
	}
	cloneHeader := CloneMap(c.defaultHeader)
	cloneHeader["Content-Type"] = "application/merge-patch+json"
	cloneHeader["Accept"] = "application/json"
	_, err = c.clusterClient.Sender.DoPatchRequest(url, cloneHeader, metaDataStr)
	if err != nil {
		return fmt.Errorf("DoPatchRequest error: %s", err.Error())
	}
	return nil
}

// GetNodeDetail get node detail from cluster manager
func (c *apiGateway) GetNodeDetail(ip string) (*impl.Node, error) {
	ctx := metadata.NewContext(context.Background(), metadata.Metadata{
		"X-Bcs-Client":   "bcs-nodegroup-manager",
		"X-Bcs-Username": "bcs-nodegroup-manager",
	})
	nodeInfoRsp, err := c.clusterManagerClient.cmClient.GetNode(ctx, &impl.GetNodeRequest{InnerIP: ip})
	if err != nil {
		return nil, fmt.Errorf("get node info %s from cluster manager error: %s", ip, err.Error())
	}
	if !nodeInfoRsp.Result || len(nodeInfoRsp.Data) == 0 {
		return nil, fmt.Errorf("get node info %s failed. result:%t, message:%s, length of data: %d",
			ip, nodeInfoRsp.Result, nodeInfoRsp.Message, len(nodeInfoRsp.Data))
	}
	return nodeInfoRsp.Data[0], nil
}

// CloneMap clone map
func CloneMap(initial map[string]string) map[string]string {
	cloneMap := make(map[string]string)
	for key := range initial {
		cloneMap[key] = initial[key]
	}
	return cloneMap
}

// UpdateNodegroupMax update nodegroup max size
func (c *apiGateway) UpdateNodegroupMax(clusterID, nodegroupID string, max int) error {
	md := c.defaultHeader
	reqCtx := metadata.NewContext(context.Background(), md)
	nodegroupInfo, err := c.clusterManagerClient.cmClient.GetNodeGroup(reqCtx,
		&impl.GetNodeGroupRequest{NodeGroupID: nodegroupID})
	if err != nil {
		return fmt.Errorf("get nodegroup %s from cluster manager error%s", nodegroupInfo, err.Error())
	}
	updateReq := &impl.UpdateNodeGroupRequest{
		ClusterID:      clusterID,
		AutoScaling:    nodegroupInfo.Data.AutoScaling,
		OnlyUpdateInfo: true,
		Updater:        "bcs-nodegroup-manager",
	}
	updateReq.AutoScaling.MaxSize = uint32(max)
	rsp, err := c.clusterManagerClient.cmClient.UpdateNodeGroup(reqCtx, updateReq)
	if err != nil {
		return fmt.Errorf("update nodegroup %s autoscaling error:%s", nodegroupID, err.Error())
	}
	if !rsp.Result {
		return fmt.Errorf("nodegroup %s autoscaling failed. result:%t, message:%s",
			nodegroupID, rsp.Result, rsp.Message)
	}
	return nil
}
