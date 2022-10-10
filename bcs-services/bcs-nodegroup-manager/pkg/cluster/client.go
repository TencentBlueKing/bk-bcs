/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/cluster/requester"
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
}

// ClientOptions option for api-gateway client
type ClientOptions struct {
	Endpoint string
	Token    string
	Sender   requester.Requester
}

// apiGateway bcs apiGateway client
type apiGateway struct {
	opt           *ClientOptions
	defaultHeader map[string]string
}

// NewClient new bcs apiGateway client
func NewClient(opts *ClientOptions) Client {
	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bearer %s", opts.Token)
	if opts.Sender == nil {
		opts.Sender = requester.NewRequester()
	}
	return &apiGateway{
		opt:           opts,
		defaultHeader: header,
	}
}

// ListClusterNodes list nodes by clusterID
func (c *apiGateway) ListClusterNodes(clusterID string) ([]*Node, error) {
	url := fmt.Sprintf("%s%s", c.opt.Endpoint, fmt.Sprintf(ListK8SNodePath, clusterID))
	rawResponse, err := c.opt.Sender.DoGetRequest(url, c.defaultHeader)
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
			Name:   k8sNode.Name,
			IP:     nodeIP,
			Status: status,
			Labels: k8sNode.ObjectMeta.Labels,
		}
		nodeList = append(nodeList, node)
	}
	return nodeList, nil
}

// ListNodesByLabel list nodes by clusterID and label
func (c *apiGateway) ListNodesByLabel(clusterID string, labels map[string]interface{}) (map[string]*Node, error) {
	url := fmt.Sprintf("%s%s", c.opt.Endpoint, fmt.Sprintf(ListK8SNodePath, clusterID))
	for key, value := range labels {
		url = fmt.Sprintf("%s,%s=%s", url, key, value)
	}
	rawResponse, err := c.opt.Sender.DoGetRequest(url, c.defaultHeader)
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
			Name:   k8sNode.Name,
			IP:     nodeIP,
			Status: status,
			Labels: k8sNode.ObjectMeta.Labels,
		}
		nodeList[k8sNode.Name] = node
	}
	return nodeList, nil
}

// UpdateNodeMetadata update node labels and annotations by clusterID and nodeName
func (c *apiGateway) UpdateNodeMetadata(clusterID, nodeName string, labels map[string]interface{},
	annotations map[string]interface{}) error {
	url := fmt.Sprintf("%s%s", c.opt.Endpoint, fmt.Sprintf(UpdateK8SNodePath, clusterID, nodeName))
	metaDataMap := map[string]interface{}{"labels": labels, "annotations": annotations}
	metaData := map[string]interface{}{"metadata": metaDataMap}
	metaDataStr, err := json.Marshal(metaData)
	if err != nil {
		return fmt.Errorf("json marshal metadata to byte err:%s", err.Error())
	}
	cloneHeader := CloneMap(c.defaultHeader)
	cloneHeader["Content-Type"] = "application/merge-patch+json"
	cloneHeader["Accept"] = "application/json"
	_, err = c.opt.Sender.DoPatchRequest(url, cloneHeader, metaDataStr)
	if err != nil {
		return fmt.Errorf("DoPatchRequest error: %s", err.Error())
	}
	return nil
}

// CloneMap clone map
func CloneMap(initial map[string]string) map[string]string {
	cloneMap := make(map[string]string)
	for key := range initial {
		cloneMap[key] = initial[key]
	}
	return cloneMap
}
