/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package rssche

import ()

// Resource is the schedulability objects, such as service instance, connection or static resource.
// Impl your own Resource struct, add to a scheduler, get the resource by Query func.
type Resource interface {
	// Update updates your local storage by remote data(etcd).
	// Create a scheduler which can get KV data from etcd cluster, and save to your local storage by update func.
	Update(updates []*Update) error

	// Query can get the resource list from local storage which update by remote data(etcd).
	// Build your own Resource struct with a sort like func to order the resource list, as a load balancer.
	Query() ([]interface{}, error)
}
