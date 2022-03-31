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

package u1x21x202203082112

const (
	//
	defaultProjectType                 = 1
	defaultKind                        = "k8s"
	defaultDeployType                  = 1
	defaultClusterRegion               = "22"
	defaultClusterOnlyCreateInfo       = true
	defaultClusterBasicSettingsVersion = "1.12.3"
	defaultProvider                    = "bcs"
	defaultNodeOnlyCreateInfo          = true

	// ccAllProjectPath :cc get all project
	ccAllProjectPath = "/projects?access_token=%s"
	// ccAllClusterPath :cc get all cluster
	ccAllClusterPath = "/v1/projects/resource?access_token=%s"
	// ccSearchClusterConfigPath :cc get cluster config
	ccSearchClusterConfigPath = "/v1/clusters/%s/cluster_config?access_token=%s"
	//ccVersionConfigPath :cc cluster version config
	ccVersionConfigPath = "/v1/clusters/%s/cluster_config?access_token=%s"
	//ccClusterInfoPath :cc get cluster info
	ccClusterInfoPath = "/projects/%s/clusters/%s?access_token=%s"
	//ccAllNodeListPath :cc get all node
	ccAllNodeListPath = "/v1/nodes/all_node_list/?access_token=%s"
	//ccAllMasterListPath :cc get all master
	ccAllMasterListPath = "/v1/masters/all_master_list/?desire_all_data=1&access_token=%s"
	// cmCreateProjectPath : cm project path post
	cmCreateProjectPath = "/clustermanager/v1/project"
	// cmProjectPath :cm get|put project
	cmProjectPath = "/clustermanager/v1/project/%s"
	// cmCreateProjectPath : cm project path post
	cmCreateClusterPath = "/clustermanager/v1/cluster"
	// cmClusterHost :cm cluster path get|post|put
	cmClusterHost = "/clustermanager/v1/cluster/%s"
	// cmNodeHost :cm cluster path get|post|put
	cmNodeHost = "/clustermanager/v1/cluster/%s/node"
)
