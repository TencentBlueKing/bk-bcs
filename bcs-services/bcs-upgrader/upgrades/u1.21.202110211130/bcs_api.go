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

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/upgrader"
)

func createProject(project bcsProject, helper upgrader.UpgradeHelper) error {

	projectByte, err := json.Marshal(project)
	if err != nil {
		return err
	}
	// TODO 在新版本中，创建项目url中不用带projectID，使用 CreateProjectPath
	url := fmt.Sprintf(ProjectPath, project.ProjectID)

	_, err = helper.ClusterManagerRequest(http.MethodPost, url, projectByte)
	return err
}

func findProject(projectID string, helper upgrader.UpgradeHelper) (*bcsProject, error) {

	url := fmt.Sprintf(ProjectPath, projectID)

	replyData, err := helper.ClusterManagerRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := new(bcsProject)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func updateProject(project bcsProject, helper upgrader.UpgradeHelper) error {
	projectJson, err := json.Marshal(project)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(ProjectPath, project.ProjectID)

	_, err = helper.ClusterManagerRequest(http.MethodPut, url, projectJson)
	return err

}

func createClusters(clusters bcsReqCreateCluster, helper upgrader.UpgradeHelper) error {
	clustersJson, err := json.Marshal(clusters)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(ClusterHost, clusters.ClusterID)

	_, err = helper.ClusterManagerRequest(http.MethodPost, url, clustersJson)
	return err
}

func findCluster(clustersID string, helper upgrader.UpgradeHelper) (*bcsRespFindCluster, error) {

	url := fmt.Sprintf(ClusterHost, clustersID)

	replyData, err := helper.ClusterManagerRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp := new(bcsRespFindCluster)
	err = json.Unmarshal(replyData, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func updateCluster(data bcsReqUpdateCluster, helper upgrader.UpgradeHelper) error {
	url := fmt.Sprintf(ClusterHost, data.ClusterID)

	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = helper.ClusterManagerRequest(http.MethodPut, url, dataJson)
	return err
}

func createNode(data reqCreateNode, helper upgrader.UpgradeHelper) error {
	nodeJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(NODEHOST, data.ClusterID)

	_, err = helper.ClusterManagerRequest(http.MethodPost, url, nodeJson)
	return err
}

func deleteNode(data reqDeleteNode, helper upgrader.UpgradeHelper) error {
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(NODEHOST, data.ClusterID)

	_, err = helper.ClusterManagerRequest(http.MethodDelete, url, dataJson)
	return err
}

func findClusterNode(clustersID string, helper upgrader.UpgradeHelper) ([]bcsNodeListData, error) {

	url := fmt.Sprintf(NODEHOST, clustersID)

	replyData, err := helper.ClusterManagerRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp := make([]bcsNodeListData, 0)
	err = json.Unmarshal(replyData, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil

}
