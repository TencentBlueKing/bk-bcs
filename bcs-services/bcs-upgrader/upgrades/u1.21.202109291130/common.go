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

package u1_21_202109291130

import (
	"net/http"
)

var CCTOKEN = "hKMCPNsEsqNyPU9iqwtNkho6dwncdY"

const (
	// TODO 此项改为读取配置文件
	CCHOST      = "http://paas-dev.bktencent.com/api/apigw/bcs-cc/prod"
	BKAPPSECRET = "e52cb30c-9ee6-4861-81f8-99db4658d3bc"

	ALLPROJECTPATH          = CCHOST + "/projects?access_token=%s"
	ALLCLUSTERPATH          = CCHOST + "/v1/projects/resource?access_token=%s"
	SEARCHCLUSTERCONFIGPATH = CCHOST + "/v1/clusters/%s/cluster_config?access_token=%s"
	VERSIONCONFIGPATH       = CCHOST + "/v1/clusters/%s/cluster_config?access_token=%s"
	CLUSTERINFOPATH         = CCHOST + "/projects/%s/clusters/%s?access_token=%s"
	AllNodeListPath         = CCHOST + "/v1/nodes/all_node_list/?access_token=%s"
	ALLMASTERLISTPATH       = CCHOST + "/v1/masters/all_master_list/?desire_all_data=1&access_token=%s"

	// TODO 获取cc token的url,改为读取配置
	GetCCTokenPath    = "http://bkssm.service.consul:5000/api/v1/auth/access-tokens"
	BCSHOST           = "https://selftest-api-gateway.bk.tencent.com:31443/bcsapi/v4"
	BCSTOKEN          = "g8Y9wYrT97kERysMDjMy1Gvdq3nI6Tid"
	CreateProjectPath = BCSHOST + "/clustermanager/v1/project"
	ProjectPath       = BCSHOST + "/clustermanager/v1/project/%s"
	ClusterHost       = BCSHOST + "/clustermanager/v1/cluster/%s"
	NODEHOST          = BCSHOST + "/clustermanager/v1/cluster/%s/node"
)

func TokenHeader() http.Header {

	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", "Bearer "+CCTOKEN)

	return header
}
