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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	v1 "k8s.io/api/core/v1"
)

const (
	// ListK8SNodePath original kubernetes path with bcs prefix
	ListK8SNodePath = "/clusters/%s/api/v1/nodes?labelSelector=node-role.kubernetes.io/master!=true"
	// UpdateK8SNodePath original kubernetes path with bcs prefix
	UpdateK8SNodePath = "/clusters/%s/api/v1/nodes/%s"
)

//Client for cluster operation
type Client interface {
	ListClusterNodes(clusterID string) ([]*Node, error)
	UpdateNodeLabels(clusterID, nodeName string, labels map[string]string) error
}

// ClientOptions option for actully api-gateway client
type ClientOptions struct {
	Endpoint string
	Token    string
	Sender   Requester
}

// apiGateway bcs apiGateway client
type apiGateway struct {
	opt           *ClientOptions
	defaultHeader map[string]string
}

// NewClient new bcs apiGateway client
func NewClient(opts *ClientOptions) Client {
	header := make(map[string]string)
	header["Authorization"] = fmt.Sprintf("Bear %s", opts.Token)
	if opts.Sender == nil {
		opts.Sender = NewRequester()
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
		blog.Errorf("ListClusterNodes error, clusterID: %s, err: %s", clusterID, err.Error())
		return nil, fmt.Errorf("network error happened %s", err.Error())
	}
	var k8sNodeList v1.NodeList
	if err := json.Unmarshal(rawResponse, &k8sNodeList); err != nil {
		blog.Errorf("Decode NodeList response failed %s, raw response %s", err.Error(), string(rawResponse))
		return nil, fmt.Errorf("decode NodeList failed: %s", err.Error())
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
		node := &Node{
			Name:   k8sNode.Name,
			IP:     nodeIP,
			Labels: k8sNode.ObjectMeta.Labels,
		}
		nodeList = append(nodeList, node)
	}
	return nodeList, nil
}

// UpdateNodeLabels update node labels by clusterID and nodeName
func (c *apiGateway) UpdateNodeLabels(clusterID, nodeName string, labels map[string]string) error {
	url := fmt.Sprintf("%s%s", c.opt.Endpoint, fmt.Sprintf(UpdateK8SNodePath, clusterID, nodeName))
	labelMap := map[string]interface{}{"labels": labels}
	metaData := map[string]interface{}{"metadata": labelMap}
	metaDataStr, err := json.Marshal(metaData)
	if err != nil {
		blog.Errorf("json marshal metadata to byte err, clusterID:%s, nodeName:%s, err:%v",
			clusterID, nodeName, err)
		return err
	}
	cloneHeader := CloneMap(c.defaultHeader)
	cloneHeader["Content-Type"] = "application/merge-patch+json"
	cloneHeader["Accept"] = "application/json"
	_, err = c.opt.Sender.DoPatchRequest(url, cloneHeader, metaDataStr)
	if err != nil {
		blog.Errorf("UpdateNodeLabels error, clusterID: %s, nodeName:%s, err: %s", clusterID, nodeName, err.Error())
		return fmt.Errorf("UpdateNodeLabels %s/%s failed %s", clusterID, nodeName, err.Error())
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
