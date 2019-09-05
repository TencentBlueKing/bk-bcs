/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package discovery

import (
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/pkg/queue"
	v1 "bk-bcs/bmsf-mesh/pkg/apis/mesh/v1"

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
