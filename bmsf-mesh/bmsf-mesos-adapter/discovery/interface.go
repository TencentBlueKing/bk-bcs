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

package discovery

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/queue"
	v1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/mesh/v1"

	"k8s.io/apimachinery/pkg/labels"
)

// Cluster maintance resource for specified cluster, offer cluster api for
// getting & listing all discovery datas
type Cluster interface {
	// GetName get cluster name
	GetName() string
	// Run starting cluster all discovery goroutine for caching data
	Run()
	// Stop stop cluster
	Stop()
	// AppSvcs get controller of AppSvc
	AppSvcs() AppSvcController
	// AppNodes get controller of AppNode
	AppNodes() AppNodeController
}

// AppSvcController controller for AppSvc encapsulation
type AppSvcController interface {
	// GetAppSvc get specified AppSvc by namespace, name
	GetAppSvc(ns, name string) (*v1.AppSvc, error)
	// ListAppSvcs List all AppSvc datas
	ListAppSvcs(selector labels.Selector) ([]*v1.AppSvc, error)
	// RegisterAppSvcHandler register event callback for AppSvc
	RegisterAppSvcQueue(handler queue.Queue)
}

// AppNodeController controller for AppNode encapsulation
type AppNodeController interface {
	// ListAppNodes list all appNode datas
	ListAppNodes(selector labels.Selector) ([]*v1.AppNode, error)
	// GetAppNode get specified AppNode by namespace, name
	GetAppNode(ns, name string) (*v1.AppNode, error)
	// RegisterAppNodeHandler register event callback for AppNode
	RegisterAppNodeQueue(handler queue.Queue)
}
