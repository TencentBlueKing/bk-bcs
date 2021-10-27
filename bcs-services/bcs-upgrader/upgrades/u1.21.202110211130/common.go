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

package u1x21x202110211130

const (
	// AllProjectPath :bcs-SaaS cc get all project
	AllProjectPath = "/projects?access_token=%s"
	// AllClusterPath :bcs-SaaS cc get all cluster
	AllClusterPath = "/v1/projects/resource?access_token=%s"
	// SearchClusterConfigPath :bcs-SaaS cc get cluster config
	SearchClusterConfigPath = "/v1/clusters/%s/cluster_config?access_token=%s"
	//VersionConfigPath :bcs-SaaS cc cluster version config
	VersionConfigPath = "/v1/clusters/%s/cluster_config?access_token=%s"
	//ClusterInfoPath :bcs-SaaS cc get cluster info
	ClusterInfoPath = "/projects/%s/clusters/%s?access_token=%s"
	//AllNodeListPath :bcs-SaaS cc get all node
	AllNodeListPath = "/v1/nodes/all_node_list/?access_token=%s"
	//AllMasterListPath :bcs-SaaS cc get all master
	AllMasterListPath = "/v1/masters/all_master_list/?desire_all_data=1&access_token=%s"
	// CreateProjectPath : clusterManager project path get|put
	CreateProjectPath = "/clustermanager/v1/project"
	// ProjectPath :clusterManager get|put project
	ProjectPath = "/clustermanager/v1/project/%s"
	// ClusterHost :clusterManager cluster path get|post|put
	ClusterHost = "/clustermanager/v1/cluster/%s"
	// NodeHost :clusterManager cluster path get|post|put
	NodeHost = "/clustermanager/v1/cluster/%s/node"
)
