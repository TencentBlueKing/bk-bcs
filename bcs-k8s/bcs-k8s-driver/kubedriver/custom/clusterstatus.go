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

package custom

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/client"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-driver/kubedriver/options"

	restful "github.com/emicklei/go-restful"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ClusterStatusAPIHandler cluster node http api implementation
type ClusterStatusAPIHandler struct {
	clientSet *kubernetes.Clientset
}

// Handler http implementation
func (csh *ClusterStatusAPIHandler) Handler(request *restful.Request, response *restful.Response) {
	nodes, err := csh.clientSet.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		CustomServerErrorResponse(response, "Get node list failed")
		return
	}

OutLoop:
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status != "True" {
				continue OutLoop
			}
		}
	}
}

// Config config kube clientset
func (csh *ClusterStatusAPIHandler) Config(KubeMasterURL string, TLSConfig options.TLSConfig) error {
	csh.clientSet = client.NewClientSet(KubeMasterURL, TLSConfig)
	if csh.clientSet == nil {
		return fmt.Errorf("failed to get k8s clientSet")
	}
	return nil
}
