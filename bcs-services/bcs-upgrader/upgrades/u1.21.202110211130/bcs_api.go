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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func createProject(project bcsProject) error {

	req, err := json.Marshal(project)
	if err != nil {
		return err
	}

	// TODO 在新版本中，创建项目url中不用带projectID，使用 CreateProjectPath
	url := fmt.Sprintf(ProjectPath, project.ProjectID)

	replyData, err := XRequest(url, http.MethodPost, TokenHeader(), bytes.NewBuffer(req))
	if err != nil {
		return err
	}

	resp := new(bcsBaseResp)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return err
	}
	if !resp.Result {
		return errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return errors.New("")
	}

	return nil
}

func findProject(projectID string) (*bcsProject, error) {

	url := fmt.Sprintf(ProjectPath, projectID)

	replyData, err := XRequest(url, http.MethodGet, TokenHeader(), nil)
	if err != nil {
		return nil, err
	}

	resp := new(respSearchProjectByID)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		//return nil, errors.New(resp.Message)
		return nil, nil
	}
	if resp.Code != 0 {
		return nil, nil
	}

	return &resp.Data, nil
}

func updateProject(project bcsProject) error {
	projectJson, err := json.Marshal(project)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(ProjectPath, project.ProjectID)

	replyData, err := XRequest(url, http.MethodPut, TokenHeader(), bytes.NewBuffer(projectJson))
	if err != nil {
		return err
	}

	resp := new(bcsBaseResp)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return err
	}
	if !resp.Result {
		return errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}

func createClusters(clusters bcsReqCreateCluster) error {
	clustersJson, err := json.Marshal(clusters)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(ClusterHost, clusters.ClusterID)

	replyData, err := XRequest(url, http.MethodPost, TokenHeader(), bytes.NewBuffer(clustersJson))
	if err != nil {
		return err
	}
	resp := new(bcsBaseResp)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return err
	}
	if !resp.Result {
		return errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}

func findCluster(clustersID string) (*bcsRespFindCluster, error) {

	url := fmt.Sprintf(ClusterHost, clustersID)

	replyData, err := XRequest(url, http.MethodGet, TokenHeader(), nil)
	if err != nil {
		return nil, err
	}

	resp := new(respFindCluster)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		//return nil, errors.New(resp.Message)
		return nil, nil
	}
	if resp.Code != 0 {
		return nil, nil
	}

	return &resp.Data, nil
}

func updateCluster(data bcsReqUpdateCluster) error {
	url := fmt.Sprintf(ClusterHost, data.ClusterID)

	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	replyData, err := XRequest(url, http.MethodPut, TokenHeader(), bytes.NewBuffer(dataJson))
	if err != nil {
		return err
	}

	resp := new(bcsBaseResp)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return err
	}
	if !resp.Result {
		return errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}

func createNode(data reqCreateNode) error {
	nodeJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(NODEHOST, data.ClusterID)

	replyData, err := XRequest(url, http.MethodPost, TokenHeader(), bytes.NewBuffer(nodeJson))
	if err != nil {
		return err
	}
	resp := new(bcsBaseResp)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return err
	}
	if !resp.Result {
		return errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}

func deleteNode(data reqDeleteNode) error {
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	url := fmt.Sprintf(NODEHOST, data.ClusterID)

	replyData, err := XRequest(url, http.MethodDelete, TokenHeader(), bytes.NewBuffer(dataJson))
	if err != nil {
		return err
	}
	resp := new(bcsBaseResp)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return err
	}
	if !resp.Result {
		return errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}

func findClusterNode(clustersID string) ([]bcsNodeListData, error) {

	url := fmt.Sprintf(NODEHOST, clustersID)

	replyData, err := XRequest(url, http.MethodGet, TokenHeader(), nil)
	if err != nil {
		return nil, err
	}
	resp := new(bcsRespNodeList)
	err = json.Unmarshal([]byte(replyData), resp)
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, errors.New(resp.Message)
	}
	if resp.Code != 0 {
		return nil, errors.New(resp.Message)
	}

	return resp.Data, nil

}
