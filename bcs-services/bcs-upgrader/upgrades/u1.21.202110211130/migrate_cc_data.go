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
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/upgrader"

	mapset "github.com/deckarep/golang-set"
)

// migrateCCData 迁移cc中的数据 (project,cluster,node)
func migrateCCData(ctx context.Context, helper upgrader.UpgradeHelper) error {

	ccToken, err := getCCToken()
	if err != nil {
		blog.Errorf("get cc token failed, err: %v", err)
		return err
	}
	CCTOKEN = ccToken

	allProject, err := getAllProject(helper)
	if err != nil {
		blog.Errorf("get ccToken failed,err: %v", err)
		return err
	}

	for _, project := range allProject {
		if err = migrateProjectData(project, helper); err != nil {
			blog.Errorf("migrate project(%s) data failed, err: %v", project.ProjectID, err)
			continue
		}

		ccAppID := strconv.Itoa(project.CcAppId)
		if err = migrateClusterData(project.ProjectID, ccAppID, helper); err != nil {
			blog.Errorf("migrate project(%s) data success ,but migrate cluster failed, err %v",
				project.ProjectID, err)
			continue
		}
	}
	return nil
}

func migrateProjectData(data ccProject, helper upgrader.UpgradeHelper) error {

	bcsProject, err := findProject(data.ProjectID, helper)
	if err != nil {
		// TODO log
		blog.Errorf("get bcs project data failed, err: %s", err)

		project, err := data2BCSProject(data)
		if err != nil {
			blog.Errorf("migrate project(%s) data failed, err: %v", data.ProjectID, err)
			return err
		}
		return createProject(*project, helper)
	}

	isUpdate, updateProjectData, err := diffProject(data, *bcsProject)
	if err != nil {
		blog.Errorf("diff project data failed, err: %v", err)
	}
	if !isUpdate {
		return nil
	}

	return updateProject(*updateProjectData, helper)
}

func migrateClusterData(projectID, ccAppID string, helper upgrader.UpgradeHelper) error {

	allClusterData, err := allCluster(helper)
	if err != nil {
		blog.Errorf("get cc cluster data failed, err: %v", err)
		return err
	}

	for _, clusterData := range allClusterData {
		if clusterData.ID != projectID {
			continue
		}
		for _, list := range clusterData.ClusterList {

			cluster, err := genCluster(projectID, list.ID, ccAppID, helper)
			if err != nil {
				blog.Errorf("gen cluster(%s) data failed, err: %v", list.ID, err)
				continue
			}

			bcsCluster, err := findCluster(list.ID, helper)
			if err != nil {
				// TODO log
				blog.Errorf("get bcs cluster(%s) data failed, err: %v", list.ID, err)

				if err = createClusters(*cluster, helper); err != nil {
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
				if err = updateCluster(*upData, helper); err != nil {
					blog.Errorf("migrate cluster(%s) data failed, err: %v", list.ID, err)
					continue
				}
			}

			if err = migrateNodeData(projectID, list.ID, helper); err != nil {
				blog.Errorf("migrate cluster(%s) node failed, err: %v", list.ID, err)
			}
		}
	}

	return nil
}

func migrateNodeData(projectID, clusterID string, helper upgrader.UpgradeHelper) error {

	nodeList, err := allNodeList(helper)
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

	if len(ccNodeIP) < 1 {
		return nil
	}

	bcsNode, err := findClusterNode(clusterID, helper)
	if err != nil {
		// TODO log
		blog.Errorf("get bcs cluster(%s) node failed, err: %v", clusterID, err)

		createNodeData := &reqCreateNode{
			ClusterID:         clusterID,
			Nodes:             ccNodeIP,
			InitLoginPassword: "",
			NodeGroupID:       "",
			OnlyCreateInfo:    true,
		}

		return createNode(*createNodeData, helper)
	}
	bcsNodeIP := make([]string, 0)
	for _, data := range bcsNode {
		bcsNodeIP = append(bcsNodeIP, data.InnerIP)
	}

	// 比较node
	createNodeData, deleteNodeData := diffNode(ccNodeIP, bcsNodeIP, clusterID)
	if createNodeData != nil {
		err = createNode(*createNodeData, helper)
		if err != nil {
			blog.Errorf("create node data failed, clusterID, err: %v", clusterID, err)
		}
	}
	if deleteNodeData != nil {
		err = deleteNode(*deleteNodeData, helper)
		if err != nil {
			blog.Errorf("delete node data failed, clusterID, err: %v", clusterID, err)
		}
	}

	return nil

}

func diffCluster(ccData bcsReqCreateCluster, bcsData bcsRespFindCluster) (bool, *bcsReqUpdateCluster, error) {

	// 对比基础数据
	if ccData.bcsClusterBase != bcsData.bcsClusterBase {
		clusters := &bcsReqUpdateCluster{
			bcsClusterBase:          ccData.bcsClusterBase,
			NetworkSettings:         ccData.NetworkSettings,
			ClusterBasicSettings:    ccData.ClusterBasicSettings,
			Updater:                 ccData.Creator,
			Master:                  ccData.Master,
			Node:                    ccData.Node,
			Labels:                  bcsData.Labels,
			BcsAddons:               bcsData.BcsAddons,
			ExtraAddons:             bcsData.ExtraAddons,
			ClusterAdvanceSettings:  bcsData.ClusterAdvanceSettings,
			NodeSettings:            bcsData.NodeSettings,
			AutoGenerateMasterNodes: bcsData.AutoGenerateMasterNodes,
			Instances:               bcsData.Instances,
			ExtraInfo:               bcsData.ExtraInfo,
			MasterInstanceID:        bcsData.MasterInstanceID,
			Status:                  bcsData.Status,
			SystemID:                bcsData.SystemID,
		}
		return true, clusters, nil
	}

	if len(bcsData.Master) == len(ccData.Master) {
		return false, nil, nil
	}

	// 对比master
	bcsMasterIPMap := make(map[string]string)
	for _, master := range bcsData.Master {
		bcsMasterIPMap[master.InnerIP] = master.InnerIP
	}

	for _, ip := range ccData.Master {
		if _, ok := bcsMasterIPMap[ip]; !ok {
			clusters := &bcsReqUpdateCluster{
				bcsClusterBase:          ccData.bcsClusterBase,
				NetworkSettings:         ccData.NetworkSettings,
				ClusterBasicSettings:    ccData.ClusterBasicSettings,
				Updater:                 ccData.Creator,
				Master:                  ccData.Master,
				Node:                    ccData.Node,
				Labels:                  bcsData.Labels,
				BcsAddons:               bcsData.BcsAddons,
				ExtraAddons:             bcsData.ExtraAddons,
				ClusterAdvanceSettings:  bcsData.ClusterAdvanceSettings,
				NodeSettings:            bcsData.NodeSettings,
				AutoGenerateMasterNodes: bcsData.AutoGenerateMasterNodes,
				Instances:               bcsData.Instances,
				ExtraInfo:               bcsData.ExtraInfo,
				MasterInstanceID:        bcsData.MasterInstanceID,
				Status:                  bcsData.Status,
				SystemID:                bcsData.SystemID,
			}
			return true, clusters, nil
		}
	}

	return false, nil, nil
}

func genCluster(projectID, clusterID, ccAppID string, helper upgrader.UpgradeHelper) (*bcsReqCreateCluster, error) {
	ccCluster, err := clusterInfo(projectID, clusterID, helper)
	if err != nil {
		blog.Errorf("get cc cluster(%s) data failed, err: %v", clusterID, err)
		return nil, err
	}

	masterList, err := allMasterList(helper)
	if err != nil {
		blog.Errorf("get cc cluster(%s) master List failed, err: %v", clusterID, err)
		return nil, err
	}
	masterIP := make([]string, 0)
	for _, data := range masterList {
		if data.ClusterId == clusterID {
			masterIP = append(masterIP, data.InnerIp)
		}
	}

	nodeList, err := allNodeList(helper)
	if err != nil {
		blog.Errorf("get cc cluster(%s) node failed, err: %v", clusterID, err)
		return nil, err
	}
	nodeIP := make([]string, 0)
	for _, data := range nodeList {
		if data.ClusterId == clusterID {
			nodeIP = append(nodeIP, data.InnerIp)
		}
	}

	configVersion, err := versionConfig(clusterID, helper)
	if err != nil {
		blog.Errorf("get cc cluster(%s) config version failed, err: %v", clusterID, err)
		return nil, err
	}

	versionConfigure := new(versionConfigure)
	err = json.Unmarshal([]byte(configVersion.Configure), versionConfigure)
	if err != nil {
		blog.Errorf("config version deJson failed, err: %v", clusterID, err)
		return nil, err
	}

	cluster := &bcsReqCreateCluster{
		bcsClusterBase: bcsClusterBase{
			ClusterID:           ccCluster.ClusterID,
			ManageType:          "INDEPENDENT_CLUSTER",
			ClusterName:         ccCluster.Name,
			Provider:            "bcs",
			Region:              "22", // TODO 待定
			VpcID:               versionConfigure.VpcID,
			ProjectID:           ccCluster.ProjectId,
			BusinessID:          ccAppID,
			Environment:         ccCluster.Environment,
			EngineType:          "k8s",
			IsExclusive:         false,
			ClusterType:         "single",
			FederationClusterID: "",
			OnlyCreateInfo:      true,
		},
		Creator: ccCluster.Creator,
		Master:  masterIP,
		Node:    nodeIP,
		NetworkSettings: createClustersNetworkSettings{
			ClusterIPv4CIDR: "",
			ServiceIPv4CIDR: "",
			MaxNodePodNum:   "",
			MaxServiceNum:   "",
		},
		ClusterBasicSettings: createClustersClusterBasicSettings{
			OS:          "",
			Version:     "1.12.3", // 默认版本
			ClusterTags: map[string]string{},
		},
	}

	return cluster, nil
}

func data2BCSProject(ccProject ccProject) (*bcsProject, error) {

	var kind string
	businessID := strconv.Itoa(ccProject.CcAppId)
	bgID := strconv.Itoa(ccProject.BgID)
	deptID := strconv.Itoa(ccProject.DeptID)
	centerID := strconv.Itoa(ccProject.CenterID)

	switch ccProject.Kind {
	case 1:
		kind = "k8s"
		break
	case 2:
		kind = "mesos"
		break
	default:
		return nil, fmt.Errorf("")
	}
	// TODO 当前cc返回DeployType为null，暂全部替换为“1”
	ccProject.DeployType = "1"
	deployType, err := strconv.Atoi(ccProject.DeployType)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	project := bcsProject{
		ProjectID:   ccProject.ProjectID,
		Name:        ccProject.Name,
		EnglishName: ccProject.EnglishName,
		Creator:     ccProject.Creator,
		ProjectType: 1, // TODO 此项待确认
		UseBKRes:    ccProject.UseBk,
		Description: ccProject.Description,
		IsOffline:   ccProject.IsOfflined,
		Kind:        kind,
		BusinessID:  businessID,
		DeployType:  deployType, // TODO 此项待定 deployType
		BgID:        bgID,
		BgName:      ccProject.BgName,
		DeptID:      deptID,
		DeptName:    ccProject.DeptName,
		CenterID:    centerID,
		CenterName:  ccProject.CenterName,
		IsSecret:    ccProject.IsSecrecy,
	}

	return &project, err
}

func diffProject(ccData ccProject, bcsData bcsProject) (isUpdate bool, project *bcsProject, err error) {

	var kind string
	switch ccData.Kind {
	case 1:
		kind = "k8s"
		break
	case 2:
		kind = "mesos"
		break
	default:
		return isUpdate, nil, fmt.Errorf("kind(%d) failed", ccData.Kind)
	}

	businessID := strconv.Itoa(ccData.CcAppId)
	bgID := strconv.Itoa(ccData.BgID)
	deptID := strconv.Itoa(ccData.DeptID)
	centerID := strconv.Itoa(ccData.CenterID)
	ccData.DeployType = "1" // TODO 当前cc返回DeployType为null，暂全部替换为“1”
	deployType, err := strconv.Atoi(ccData.DeployType)
	if err != nil {
		return isUpdate, nil, fmt.Errorf("deployType(%s) failed", ccData.DeployType)
	}

	project = &bcsProject{
		ProjectID:   ccData.ProjectID,
		Name:        ccData.Name,
		Updater:     ccData.Creator,
		ProjectType: ccData.ProjectType,
		UseBKRes:    ccData.UseBk,
		Description: ccData.Description,
		IsOffline:   ccData.IsOfflined,
		Kind:        kind,
		DeployType:  deployType,
		BgID:        bgID,
		BgName:      ccData.BgName,
		DeptID:      deptID,
		DeptName:    ccData.DeptName,
		CenterID:    centerID,
		CenterName:  ccData.CenterName,
		IsSecret:    bcsData.IsSecret,
		BusinessID:  businessID,
		Credentials: bcsData.Credentials,
	}
	if *project == bcsData {
		return false, nil, nil
	}

	return true, project, nil
}

func diffNode(ccNodeIPS, bcsNodeIPS []string, clusterID string) (createNode *reqCreateNode, deleteNode *reqDeleteNode) {

	alreadySet := mapset.NewSet()
	for _, ip := range bcsNodeIPS {
		alreadySet.Add(ip)
	}
	newSet := mapset.NewSet()
	for _, ip := range ccNodeIPS {
		newSet.Add(ip)
	}

	toCreateSet := newSet.Difference(alreadySet)
	toDeleteSet := alreadySet.Difference(newSet)
	toCreateIt := toCreateSet.Iterator()
	toDeleteIt := toDeleteSet.Iterator()
	var toCreateArray, toDeleteArray []string
	for elem := range toCreateIt.C {
		toCreateArray = append(toCreateArray, elem.(string))
	}
	for elem := range toDeleteIt.C {
		toDeleteArray = append(toDeleteArray, elem.(string))
	}

	if len(toCreateArray) != 0 {
		createNode = &reqCreateNode{
			ClusterID:         clusterID,
			Nodes:             toCreateArray,
			InitLoginPassword: "",
			NodeGroupID:       "",
			OnlyCreateInfo:    true,
		}
	}

	if len(toDeleteArray) != 0 {
		deleteNode = &reqDeleteNode{
			ClusterID:      clusterID,
			Nodes:          toDeleteArray,
			DeleteMode:     "",
			IsForce:        false, // TODO 参数待确认
			Operator:       "",
			OnlyDeleteInfo: false, // TODO 参数待确认
		}
	}

	return createNode, deleteNode
}
