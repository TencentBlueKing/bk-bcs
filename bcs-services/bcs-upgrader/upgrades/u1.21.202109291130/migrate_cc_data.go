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
	"context"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/upgrader"
)

// migrateCCData 迁移cc中的数据 (project,cluster,node)
func migrateCCData(ctx context.Context, helper upgrader.UpgradeHelper) error {

	ccToken, err := getCCToken()
	if err != nil {
		blog.Errorf("get cc token failed, err: %v", err)
		return err
	}
	CCTOKEN = ccToken

	allProject, err := getAllProject()
	if err != nil {
		blog.Errorf("get ccToken failed,err: %v", err)
		return err
	}

	for _, project := range allProject {
		if err = migrateProjectData(project); err != nil {
			blog.Errorf("migrate project(%s) data failed, err %v", project.ProjectID, err)
			continue
		}

		ccAppID := strconv.Itoa(project.CcAppId)
		if err = migrateClusterData(project.ProjectID, ccAppID); err != nil {
			blog.Errorf("migrate project(%s) data success ,but migrate cluster failed, err %v",
				project.ProjectID, err)
			continue
		}
	}
	return nil
}

func migrateProjectData(data ccProject) error {

	bcsProject, err := findProject(data.ProjectID)
	if err != nil {
		blog.Errorf("get bcs project data failed, err: %s", err)
		return err
	}
	if bcsProject == nil {
		project, err := data2BCSProject(data)
		if err != nil {
			blog.Errorf("migrate project(%s) data failed, err: %v", data.ProjectID, err)
			return err
		}
		return createProject(*project)
	}

	isUpdate, updateProjectData, err := diffProject(data, *bcsProject)
	if err != nil {
		blog.Errorf("diff project data failed, err: %v", err)
	}
	if !isUpdate {
		return nil
	}

	return updateProject(*updateProjectData)
}

func migrateClusterData(projectID, ccAppID string) error {

	allClusterData, err := allCluster()
	if err != nil {
		blog.Errorf("get cc cluster data failed, err: %v", err)
		return err
	}

	for _, clusterData := range allClusterData {
		if clusterData.ID != projectID {
			continue
		}
		for _, list := range clusterData.ClusterList {

			cluster, err := genCluster(projectID, list.ID, ccAppID)
			if err != nil {
				blog.Errorf("gen cluster(%s) data failed, err: %v", list.ID, err)
				continue
			}

			bcsCluster, err := findCluster(list.ID)
			if err != nil {
				blog.Errorf("get bcs cluster(%s) data failed, err: %v", list.ID, err)
				continue
			}
			if bcsCluster == nil {
				if err = createClusters(*cluster); err != nil {
					blog.Errorf("migrate cluster(%s) failed, err: %v", list.ID, err)
					continue
				}
			} else {
				// 对比cluster数据
				isUpdate, upData, err := diffCluster(*cluster, *bcsCluster)
				if err != nil {
					blog.Errorf("diff cluster(%s) data failed, err: %v", list.ID, err)
					continue
				}
				if !isUpdate {
					continue
				}
				if err = updateCluster(*upData); err != nil {
					blog.Errorf("migrate cluster(%s) data failed, err: %v", list.ID, err)
					continue
				}
			}

			if err = migrateNodeData(projectID, list.ID); err != nil {
				blog.Errorf("migrate cluster(%s) node failed, err: %v", list.ID, err)
			}
		}
	}

	return nil
}

func migrateNodeData(projectID, clusterID string) error {

	nodeList, err := allNodeList()
	if err != nil {
		blog.Errorf("get cc node data failed, err: %v", err)
		return err
	}

	var ccNodeIP = make([]string, 0)
	for _, node := range nodeList {
		if node.ProjectId == projectID && node.ClusterId == clusterID {
			ccNodeIP = append(ccNodeIP, node.InnerIp)
		}
	}

	bcsNode, err := findClusterNode(clusterID)
	if err != nil {
		blog.Errorf("get bcs cluster(%s) node failed, err: %v", clusterID, err)
		return err
	}
	bcsNodeIP := make([]string, 0)
	for _, data := range bcsNode {
		bcsNodeIP = append(bcsNodeIP, data.InnerIP)
	}

	// 比较node
	createNodeData, deleteNodeData := diffNode(ccNodeIP, bcsNodeIP, clusterID)
	if createNodeData != nil {
		err = createNode(*createNodeData)
		if err != nil {
			blog.Errorf("create node data failed, clusterID, err: %v", clusterID, err)
		}
	}
	if deleteNodeData != nil {
		err = deleteNode(*deleteNodeData)
		if err != nil {
			blog.Errorf("delete node data failed, clusterID, err: %v", clusterID, err)
		}
	}

	return nil

}
