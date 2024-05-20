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

// Package clustermgr xxx
package clustermgr

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"github.com/panjf2000/ants/v2"
	"github.com/patrickmn/go-cache"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/registry"
	v1 "k8s.io/api/core/v1"

	clustermanager "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr/clustermanagerv4"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
)

// Client interface
type Client interface {
	CordonNodes(ctx context.Context, InnerIPs []string, clusterID string, unCordon bool) error
	DrainNodes(ctx context.Context, InnerIPs []string, clusterID string) error
	ListAllNodeGroups(ctx context.Context) ([]*clustermanager.NodeGroup, error)
	BatchCordonNodeWithoutCluster(ctx context.Context, innerIP []string, unCordon bool) error
	BatchDrainNodeWithoutCluster(ctx context.Context, innerIP []string) error
	GetNodeDetail(ctx context.Context, ip string) (*clustermanager.Node, error)
	CordonNodeWithoutCluster(ctx context.Context, innerIP string, unCordon bool) error
	DrainNodeWithoutCluster(ctx context.Context, innerIP string) error
	GetCluster(ctx context.Context, clusterID string) (*clustermanager.Cluster, error)
	BatchUpdateNodeWithoutCluster(ctx context.Context, innerIP []string, labels map[string]string,
		annotations map[string]string) error
	UpdateNodeWithoutCluster(ctx context.Context, innerIP string, labels map[string]string,
		annotation map[string]string) error
	GetNode(ip, clusterID string) (*Node, error)
}

const (
	// ListK8SNodePath original kubernetes path with bcs prefix
	ListK8SNodePath = "/clusters/%s/api/v1/nodes?labelSelector=node-role.kubernetes.io/master!=true"
	// UpdateK8SNodePath original kubernetes path with bcs prefix
	UpdateK8SNodePath = "/clusters/%s/api/v1/nodes/%s"
	// GetK8sNodePath get node info
	GetK8sNodePath = "/clusters/%s/api/v1/nodes/%s"
)

type cmClient struct {
	client        clustermanager.ClusterManagerService
	clusterClient *ClusterClient
	concurrency   int
	token         string
	k8sNodeCache  *cache.Cache
}

// ClientOptions for local cluster-manager client
type ClientOptions struct {
	// Name for resource-manager registry
	Name string
	// Etcd endpoints information
	Etcd []string
	// EtcdConfig tls config for etcd
	EtcdConfig *tls.Config
	// ClientConfig tls config
	ClientConfig *tls.Config
	// Cache for store ResourcePool information
	// Cache storage.Storage
	Token string
}

// ClusterClient option for api-gateway client
type ClusterClient struct {
	defaultHeader map[string]string
	Endpoint      string
	Token         string
	Sender        requester.Requester
}

// Node for kubernetes node
type Node struct {
	Name        string
	IP          string
	Status      string
	Labels      map[string]string
	Annotations map[string]string
}

// NewClient new client
func NewClient(opt *ClientOptions, concurrency int, clusterClient *ClusterClient) Client {
	// init go-micro v2 client instance
	c := grpc.NewClient(
		client.Registry(etcd.NewRegistry(
			registry.Addrs(opt.Etcd...),
			registry.TLSConfig(opt.EtcdConfig)),
		),
		grpc.AuthTLS(opt.ClientConfig),
	)
	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", clusterClient.Token)
	header["X-Bcs-Client"] = "bcs-nodegroup-manager"
	clusterClient.defaultHeader = header
	// create resource-manager go-micro client api
	return &cmClient{
		client:        clustermanager.NewClusterManagerService(opt.Name, c),
		concurrency:   concurrency,
		token:         opt.Token,
		clusterClient: clusterClient,
		k8sNodeCache:  cache.New(time.Minute*10, time.Minute*120),
	}
}

// BatchCordonNodeWithoutCluster batch cordon node without clusterID
func (c *cmClient) BatchCordonNodeWithoutCluster(ctx context.Context, innerIP []string, unCordon bool) error {
	reqCtx := c.getMetadataCtx(ctx)
	wg := sync.WaitGroup{}
	errChan := make(chan string)
	finalErrMsg := ""
	go func() {
		for errMsg := range errChan {
			finalErrMsg += errMsg + "\n"
		}
	}()
	pool, err := ants.NewPool(c.concurrency)
	if err != nil {
		blog.Errorf("init new pool err:%v", err)
		return fmt.Errorf("init new pool err:%v", err)
	}
	defer pool.Release()
	for key := range innerIP {
		ip := innerIP[key]
		wg.Add(1)
		submitErr := pool.Submit(func() {
			cordonErr := c.CordonNodeWithoutCluster(reqCtx, ip, unCordon)
			if cordonErr != nil {
				errChan <- cordonErr.Error()
			}
			wg.Done()
		})
		if submitErr != nil {
			blog.Errorf("submit task to ch pool err:%v", submitErr.Error())
			wg.Done()
		}
	}
	wg.Wait()
	if finalErrMsg != "" {
		return fmt.Errorf(finalErrMsg)
	}
	return nil
}

// BatchDrainNodeWithoutCluster batch drain nodes without clusterID
func (c *cmClient) BatchDrainNodeWithoutCluster(ctx context.Context, innerIP []string) error {
	reqCtx := c.getMetadataCtx(ctx)
	wg := sync.WaitGroup{}
	errChan := make(chan string)
	pool, err := ants.NewPool(c.concurrency)
	if err != nil {
		blog.Errorf("init new pool err:%v", err)
		return fmt.Errorf("init new pool err:%v", err)
	}
	defer pool.Release()
	for key := range innerIP {
		ip := innerIP[key]
		wg.Add(1)
		submitErr := pool.Submit(func() {
			cordonErr := c.DrainNodeWithoutCluster(reqCtx, ip)
			if cordonErr != nil {
				errChan <- cordonErr.Error()
			}
			wg.Done()
		})
		if submitErr != nil {
			blog.Errorf("submit task to ch pool err:%v", submitErr.Error())
		}
	}
	wg.Wait()
	finalErrMsg := ""
	for errMsg := range errChan {
		finalErrMsg += errMsg + "\n"
	}
	if finalErrMsg != "" {
		return fmt.Errorf(finalErrMsg)
	}
	return nil
}

// CordonNodeWithoutCluster cordon node without clusterID
func (c *cmClient) CordonNodeWithoutCluster(ctx context.Context, innerIP string, unCordon bool) error {
	reqCtx := c.getMetadataCtx(ctx)
	nodeInfo, err := c.GetNodeDetail(reqCtx, innerIP)
	if err != nil {
		return err
	}
	err = c.CordonNodes(reqCtx, []string{innerIP}, nodeInfo.ClusterID, unCordon)
	if err != nil {
		return err
	}
	blog.Infof("cordon node %s success", innerIP)
	return nil
}

// DrainNodeWithoutCluster drain nodes without clusterID
func (c *cmClient) DrainNodeWithoutCluster(ctx context.Context, innerIP string) error {
	reqCtx := c.getMetadataCtx(ctx)
	nodeInfo, err := c.GetNodeDetail(reqCtx, innerIP)
	if err != nil {
		return err
	}
	err = c.DrainNodes(reqCtx, []string{innerIP}, nodeInfo.ClusterID)
	if err != nil {
		return err
	}
	blog.Infof("drain node %s success", innerIP)
	return nil
}

// CordonNodes cordon nodes
func (c *cmClient) CordonNodes(ctx context.Context, innerIPs []string, clusterID string, unCordon bool) error {
	var err error
	for _, ip := range innerIPs {
		var node Node
		if nodeCache, ok := c.k8sNodeCache.Get(fmt.Sprintf("%s-%s", clusterID, ip)); ok {
			node = nodeCache.(Node)
		} else {
			_, err = c.ListClusterNodes(clusterID)
			if err != nil {
				blog.Errorf("list cluster node error, clusterId:%s, err:%s", clusterID, err.Error())
				return fmt.Errorf("list cluster node error, clusterId:%s, err:%s", clusterID, err.Error())
			}
			if nodeCache, ok = c.k8sNodeCache.Get(fmt.Sprintf("%s-%s", clusterID, ip)); !ok {
				return fmt.Errorf("cannot find node %s in cluster %s", ip, clusterID)
			}
			node = nodeCache.(Node)
		}
		url := fmt.Sprintf("%s%s", c.clusterClient.Endpoint, fmt.Sprintf(UpdateK8SNodePath, clusterID, node.Name))
		spec := make(map[string]interface{})
		// metaDataMap := map[string]interface{}{"labels": labels, "annotations": annotations}
		// {"spec":{"unschedulable":true}}
		if !unCordon {
			spec["unschedulable"] = true
		} else {
			spec["unschedulable"] = nil
		}
		specData := map[string]interface{}{"spec": spec}
		specDataStr, marshalErr := json.Marshal(specData)
		if marshalErr != nil {
			return fmt.Errorf("json marshal metadata to byte err:%s", marshalErr.Error())
		}
		cloneHeader := CloneMap(c.clusterClient.defaultHeader)
		cloneHeader["Content-Type"] = "application/merge-patch+json"
		cloneHeader["Accept"] = "application/json"
		_, err = c.clusterClient.Sender.DoPatchRequest(url, cloneHeader, specDataStr)
		if err != nil {
			return fmt.Errorf("DoPatchRequest error: %s", err.Error())
		}
	}
	return err
}

// DrainNodes drain nodes
func (c *cmClient) DrainNodes(ctx context.Context, innerIPs []string, clusterID string) error {
	var err error
	for _, ip := range innerIPs {
		var node Node
		if nodeCache, ok := c.k8sNodeCache.Get(fmt.Sprintf("%s-%s", clusterID, ip)); ok {
			node = nodeCache.(Node)
		} else {
			_, err = c.ListClusterNodes(clusterID)
			if err != nil {
				blog.Errorf("list cluster node error, clusterId:%s, err:%s", clusterID, err.Error())
				return fmt.Errorf("list cluster node error, clusterId:%s, err:%s", clusterID, err.Error())
			}
			if nodeCache, ok = c.k8sNodeCache.Get(fmt.Sprintf("%s-%s", clusterID, ip)); !ok {
				return fmt.Errorf("cannot find node %s in cluster %s", ip, clusterID)
			}
			node = nodeCache.(Node)
		}
		reqCtx := c.getMetadataCtx(ctx)
		req := &clustermanager.DrainNodeRequest{
			InnerIPs:  []string{node.Name},
			ClusterID: clusterID,
		}
		rsp, drainErr := c.client.DrainNode(reqCtx, req)
		if drainErr != nil {
			blog.Errorf("drain node error:%s", drainErr.Error())
			return drainErr
		}
		if !rsp.Result {
			blog.Errorf("drain node error:%s", rsp.Message)
			return fmt.Errorf("drain node error:%s", rsp.Message)
		}
	}
	return err
}

// ListAllNodeGroups list all nodegroups
func (c *cmClient) ListAllNodeGroups(ctx context.Context) ([]*clustermanager.NodeGroup, error) {
	reqCtx := c.getMetadataCtx(ctx)
	req := &clustermanager.ListNodeGroupRequest{}
	rsp, err := c.client.ListNodeGroup(reqCtx, req)
	if err != nil {
		blog.Errorf("list all nodegroups error:%s", err.Error())
		return nil, err
	}
	if !rsp.Result {
		blog.Errorf("list all nodegroups error:%s", rsp.Message)
		return nil, fmt.Errorf("list all nodegroups error:%s", rsp.Message)
	}
	return rsp.Data, nil
}

// GetNodeDetail get node detail
func (c *cmClient) GetNodeDetail(ctx context.Context, ip string) (*clustermanager.Node, error) {
	reqCtx := c.getMetadataCtx(ctx)
	req := &clustermanager.GetNodeRequest{InnerIP: ip}
	rsp, err := c.client.GetNode(reqCtx, req)
	if err != nil {
		blog.Errorf("get node %s error:%s", ip, err.Error())
		return nil, err
	}
	if rsp.Code != 0 && strings.Contains(rsp.Message, "record not found") {
		return nil, nil
	}
	if !rsp.Result {
		blog.Errorf("get node %s detail error:%s", ip, rsp.Message)
		return nil, fmt.Errorf("get node %s detail error:%s", ip, rsp.Message)
	}
	if len(rsp.Data) != 0 {
		return rsp.Data[0], nil
	}
	return nil, fmt.Errorf("the results of node detail are more than one, result:%v", rsp.Data)
}

// GetCluster get cluster
func (c *cmClient) GetCluster(ctx context.Context, clusterID string) (*clustermanager.Cluster, error) {
	reqCtx := c.getMetadataCtx(ctx)
	req := &clustermanager.GetClusterReq{
		ClusterID: clusterID,
	}
	rsp, err := c.client.GetCluster(reqCtx, req)
	if err != nil {
		blog.Errorf("get cluster %s error:%s", clusterID, err.Error())
		return nil, err
	}
	if !rsp.Result {
		blog.Errorf("get cluster %s error:%s", clusterID, rsp.Message)
		return nil, fmt.Errorf("get cluster %s error:%s", clusterID, rsp.Message)
	}
	return rsp.Data, nil
}

func (c *cmClient) getMetadataCtx(ctx context.Context) context.Context {
	return metadata.NewContext(ctx, metadata.Metadata{
		middleware.InnerClientHeaderKey:    "bcs-powertrading",
		middleware.CustomUsernameHeaderKey: "bcs-powertrading",
	})
}

// ListClusterNodes list nodes by clusterID
func (c *cmClient) ListClusterNodes(clusterID string) ([]*Node, error) {
	var err error
	url := fmt.Sprintf("%s%s", c.clusterClient.Endpoint, fmt.Sprintf(ListK8SNodePath, clusterID))
	rawResponse, err := c.clusterClient.Sender.DoGetRequest(url, c.clusterClient.defaultHeader)
	if err != nil {
		return nil, fmt.Errorf("DoGetRequest error: %s", err.Error())
	}
	var k8sNodeList v1.NodeList
	if err = json.Unmarshal(rawResponse, &k8sNodeList); err != nil {
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
		c.setNodeCache(clusterID, node)
	}

	return nodeList, nil
}

// GetNode get node status by ip
func (c *cmClient) GetNode(nodeName, clusterID string) (*Node, error) {
	var err error
	url := fmt.Sprintf("%s%s", c.clusterClient.Endpoint, fmt.Sprintf(GetK8sNodePath, clusterID, nodeName))
	rawResponse, err := c.clusterClient.Sender.DoGetRequest(url, c.clusterClient.defaultHeader)
	if err != nil {
		return nil, fmt.Errorf("DoGetRequest error: %s", err.Error())
	}

	var k8sNode v1.Node
	if err = json.Unmarshal(rawResponse, &k8sNode); err != nil {
		return nil, fmt.Errorf("decode Node response failed %s, raw response %s",
			err.Error(), string(rawResponse))
	}
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
	return node, nil
}

// BatchUpdateNodeWithoutCluster batch update nodes without clusterID
func (c *cmClient) BatchUpdateNodeWithoutCluster(ctx context.Context, innerIP []string, labels map[string]string,
	annotations map[string]string) error {
	reqCtx := c.getMetadataCtx(ctx)
	wg := sync.WaitGroup{}
	errChan := make(chan string)
	finalErrMsg := ""
	go func() {
		for errMsg := range errChan {
			finalErrMsg += errMsg + "\n"
		}
	}()
	pool, err := ants.NewPool(c.concurrency)
	if err != nil {
		blog.Errorf("init new pool err:%v", err)
		return fmt.Errorf("init new pool err:%v", err)
	}
	defer pool.Release()
	for key := range innerIP {
		ip := innerIP[key]
		wg.Add(1)
		submitErr := pool.Submit(func() {
			cordonErr := c.UpdateNodeWithoutCluster(reqCtx, ip, labels, annotations)
			if cordonErr != nil {
				errChan <- cordonErr.Error()
			}
			wg.Done()
		})
		if submitErr != nil {
			blog.Errorf("submit task to ch pool err:%v", submitErr.Error())
		}
	}
	wg.Wait()
	close(errChan)
	if finalErrMsg != "" {
		return fmt.Errorf(finalErrMsg)
	}
	return nil
}

// UpdateNodeWithoutCluster update nodes without clusterID
func (c *cmClient) UpdateNodeWithoutCluster(ctx context.Context, innerIP string, labels map[string]string,
	annotation map[string]string) error {
	reqCtx := c.getMetadataCtx(ctx)
	nodeInfo, err := c.GetNodeDetail(reqCtx, innerIP)
	if err != nil {
		return err
	}
	if nodeInfo == nil {
		blog.Errorf("%s not found", innerIP)
		return fmt.Errorf("%s not found", innerIP)
	}
	err = c.UpdateNode(reqCtx, []string{innerIP}, nodeInfo.ClusterID, labels, annotation)
	if err != nil {
		return err
	}
	blog.Infof("update node %s success", innerIP)
	return nil
}

// UpdateNode update node
func (c *cmClient) UpdateNode(ctx context.Context, innerIPs []string, clusterID string, labels map[string]string,
	annotation map[string]string) error {
	var err error
	for _, ip := range innerIPs {
		var node Node
		if nodeCache, ok := c.k8sNodeCache.Get(fmt.Sprintf("%s-%s", clusterID, ip)); ok {
			node = nodeCache.(Node)
		} else {
			_, err = c.ListClusterNodes(clusterID)
			if err != nil {
				blog.Errorf("list cluster node error, clusterId:%s, err:%s", clusterID, err.Error())
				return fmt.Errorf("list cluster node error, clusterId:%s, err:%s", clusterID, err.Error())
			}
			if nodeCache, ok = c.k8sNodeCache.Get(fmt.Sprintf("%s-%s", clusterID, ip)); !ok {
				return fmt.Errorf("cannot find node %s in cluster %s", ip, clusterID)
			}
			node = nodeCache.(Node)
		}
		node = mergeLabelAndAnnotations(node, labels, annotation)
		reqCtx := c.getMetadataCtx(ctx)
		if labels != nil {
			req := &clustermanager.UpdateNodeLabelsRequest{
				Nodes: []*clustermanager.NodeLabel{{
					NodeName: node.Name,
					Labels:   node.Labels,
				}},
				ClusterID: clusterID,
			}
			rsp, updateErr := c.client.UpdateNodeLabels(reqCtx, req)
			if updateErr != nil {
				blog.Errorf("update node label error:%s", updateErr.Error())
				return updateErr
			}
			if !rsp.Result {
				blog.Errorf("update node label error:%s", rsp.Message)
				return fmt.Errorf("update node label error:%s", rsp.Message)
			}
		}
		if annotation != nil {
			req := &clustermanager.UpdateNodeAnnotationsRequest{
				Nodes: []*clustermanager.NodeAnnotation{{
					NodeName:    node.Name,
					Annotations: node.Annotations,
				}},
				ClusterID: clusterID,
			}
			rsp, updateErr := c.client.UpdateNodeAnnotations(reqCtx, req)
			if updateErr != nil {
				blog.Errorf("update node annotations error:%s", updateErr.Error())
				return updateErr
			}
			if !rsp.Result {
				blog.Errorf("update node annotations error:%s", rsp.Message)
				return fmt.Errorf("update node annotations error:%s", rsp.Message)
			}
		}
	}
	return err
}

func (c *cmClient) setNodeCache(clusterId string, node *Node) {
	c.k8sNodeCache.Set(fmt.Sprintf("%s-%s", clusterId, node.IP), *node, 1*time.Hour)
}

// CloneMap clone map
func CloneMap(initial map[string]string) map[string]string {
	cloneMap := make(map[string]string)
	for key := range initial {
		cloneMap[key] = initial[key]
	}
	return cloneMap
}

func mergeLabelAndAnnotations(node Node, newLabels map[string]string, newAnnotations map[string]string) Node {
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}
	for labelKey := range newLabels {
		if node.Labels[labelKey] != newLabels[labelKey] {
			node.Labels[labelKey] = newLabels[labelKey]
		}
	}
	for annoKey := range newAnnotations {
		if node.Annotations[annoKey] != newAnnotations[annoKey] {
			node.Annotations[annoKey] = newAnnotations[annoKey]
		}
	}
	return node
}
