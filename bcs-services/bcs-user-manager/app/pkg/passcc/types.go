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

package passcc

// CommonResp common resp
type CommonResp struct {
	Code      uint   `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// GetProjectsNamespaces namespacs request
type GetProjectsNamespaces struct {
	DesireAllData int `json:"desire_all_data"`
}

// GetProjectsNamespacesResp namespaces resp
type GetProjectsNamespacesResp struct {
	CommonResp `json:",inline"`
	Data       ProjectsNamespacesData `json:"data"`
}

// ProjectsNamespacesData xxx
type ProjectsNamespacesData struct {
	Count   uint64             `json:"count"`
	Results []ProjectNamespace `json:"results"`
}

// ProjectNamespace xxx
type ProjectNamespace struct {
	ProjectID string `json:"project_id"`
	ClusterID string `json:"cluster_id"`
	Name      string `json:"name"`
}
