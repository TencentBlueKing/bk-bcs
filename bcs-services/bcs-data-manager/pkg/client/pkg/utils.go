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

package pkg

const (
	GetProjectUrl    = "/datamanager/v1/projects/%s?dimension=%s"
	GetClusterUrl    = "/datamanager/v1/clusters/%s?dimension=%s"
	GetNamespaceUrl  = "/datamanager/v1/clusters/%s/namespaces/%s?dimension=%s"
	GetWorkloadUrl   = "/datamanager/v1/clusters/%s/namespaces/%s/%s/%s?dimension=%s"
	ListClusterUrl   = "/datamanager/v1/projects/%s/cluster?dimension=%s&page=%s&size=%s"
	ListNamespaceUrl = "/datamanager/v1/clusters/%s/namespaces?dimension=%s&page=%s&size=%s"
	ListWorkloadUrl  = "/datamanager/v1/clusters/%s/namespaces/%s/%s?dimension=%s&page=%s&size=%s"
	PrefixUrl        = "/bcsapi/v4"
)
