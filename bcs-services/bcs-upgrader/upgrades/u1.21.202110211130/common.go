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

package u1_21_202110211130

const (
	ALLPROJECTPATH          = "/projects?access_token=%s"
	ALLCLUSTERPATH          = "/v1/projects/resource?access_token=%s"
	SEARCHCLUSTERCONFIGPATH = "/v1/clusters/%s/cluster_config?access_token=%s"
	VERSIONCONFIGPATH       = "/v1/clusters/%s/cluster_config?access_token=%s"
	CLUSTERINFOPATH         = "/projects/%s/clusters/%s?access_token=%s"
	AllNodeListPath         = "/v1/nodes/all_node_list/?access_token=%s"
	ALLMASTERLISTPATH       = "/v1/masters/all_master_list/?desire_all_data=1&access_token=%s"

	CreateProjectPath = "/clustermanager/v1/project"
	ProjectPath       = "/clustermanager/v1/project/%s"
	ClusterHost       = "/clustermanager/v1/cluster/%s"
	NODEHOST          = "/clustermanager/v1/cluster/%s/node"
)
