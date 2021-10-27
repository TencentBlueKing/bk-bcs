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

package serviceclient

const (
	/*
		Registry for multiple system adapter
	*/

	// ServiceRegistryCustom for mesos-adapter
	ServiceRegistryCustom = "custom"
	// ServiceRegistryKubernetes for bcs kubernetes
	ServiceRegistryKubernetes = "kubernetes"
	// ServiceRegistryMesos for mesos etcd storage
	ServiceRegistryMesos = "mesos"
)

// Client service
type Client interface {
	// GetAppService get service by specified namespace & name
	GetAppService(ns, name string) (*AppService, error)
	// ListAppService list all service in cache, filter by Label
	// selector comes from Set.AsSelector() see: k8s.io/apimachinery/pkg/labels.Set
	ListAppService(labels map[string]string) ([]*AppService, error)
	// ListAppServiceFromStatefulSet list app services, for each stateful node, generate a AppService object
	ListAppServiceFromStatefulSet(ns, name string) ([]*AppService, error)
	// Close client, clean resource
	Close()
}
