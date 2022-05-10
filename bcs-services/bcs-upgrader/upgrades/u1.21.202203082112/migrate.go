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

import (
	"encoding/json"
	"regexp"
	"strconv"
	"unicode/utf8"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-upgrader/app/options"
)

type MigrateHandle interface {
	getMigrateProjectData() error
	getMigrateClusterData() error
	getMigrateNode() error
	migrateProject() error
	migrateCluster() error
	migrateNode() error
	run() error
}

type migrateHandle struct {
	conf     options.HttpCliConfig
	ccMgr    CcManager
	cmMgr    CmManager
	projects map[string]project
	clusters map[string]cluster
	nodes    []node
}

func NewMigrateHandle(conf options.HttpCliConfig) MigrateHandle {
	return &migrateHandle{
		conf:     conf,
		projects: make(map[string]project),
		clusters: make(map[string]cluster),
	}
}

func (h *migrateHandle) run() error {
	ccMgr := NewCcManager(h.conf.CcHOST)
	err := ccMgr.setToken(h.conf.BkAppSecret, h.conf.SsmHost)
	if err != nil {
		blog.Errorf("get ssm token failed, err : %v.", err)
		return err
	}
	h.ccMgr = ccMgr

	cmMgr := NewCmManager(h.conf.CmHost, h.conf.GatewayToken)
	h.cmMgr = cmMgr

	err = h.getMigrateProjectData()
	if err != nil {
		blog.Errorf("get cc project data failed, err : %v.", err)
		return err
	}

	err = h.getMigrateClusterData()
	if err != nil {
		blog.Errorf("get cc cluster data failed, err : %v.", err)
		return err
	}

	err = h.getMigrateNode()
	if err != nil {
		blog.Errorf("get cc node data failed, err : %v.", err)
		return err
	}

	err = h.migrateProject()
	if err != nil {
		blog.Errorf("migrate project data failed, err : %v.", err)
		return err
	}
	err = h.migrateCluster()
	if err != nil {
		blog.Errorf("migrate cluster data failed, err : %v.", err)
		return err
	}

	err = h.migrateNode()
	if err != nil {
		blog.Errorf("migrate node data failed, err : %v.", err)
		return err
	}

	return nil
}

func (h *migrateHandle) getMigrateProjectData() error {
	projects, err := h.ccMgr.getAllProject()
	if err != nil {
		return err
	}

	for _, p := range projects {

		if p.ProjectType == 0 {
			p.ProjectType = defaultProjectType
		}

		projectTMP := project{
			ProjectID:   p.ProjectID,
			Name:        p.Name,
			EnglishName: p.EnglishName,
			Creator:     p.Creator,
			ProjectType: p.ProjectType,
			UseBKRes:    p.UseBk,
			Description: p.Description,
			IsOffline:   p.IsOfflined,
			BgName:      p.BgName,
			DeptName:    p.DeptName,
			CenterName:  p.CenterName,
			IsSecret:    p.IsSecrecy,
			Updater:     p.Updator,
			Kind:        defaultKind,
			DeployType:  defaultDeployType,
			BusinessID:  strconv.Itoa(p.CcAppId),
			BgID:        strconv.Itoa(p.BgID),
			DeptID:      strconv.Itoa(p.BgID),
			CenterID:    strconv.Itoa(p.CenterID),
		}

		h.projects[p.ProjectID] = projectTMP
	}

	return nil
}

func (h *migrateHandle) getMigrateClusterData() error {
	clusters, err := h.ccMgr.getAllCluster()
	if err != nil {
		blog.Errorf("get all clusters failed, err : %v.", err)
		return err
	}

	for _, clusterVal := range clusters {
		_, ok := h.projects[clusterVal.ProjectID]
		if !ok {
			continue
		}
		for _, list := range clusterVal.ClusterList {
			clusterTMP, err := h.genCmCreateClusters(clusterVal.ProjectID, list.ClusterID)
			if err != nil {
				blog.Errorf("gen clusters data failed, err : %v.", err)
				continue
			}
			h.clusters[clusterTMP.ClusterID] = *clusterTMP
		}
	}

	return nil
}

func (h *migrateHandle) getMigrateNode() error {

	nodes, err := h.ccMgr.getAllNode()
	if err != nil {
		blog.Errorf("get all node failed, err : %v.", err)
		return err
	}

	for _, n := range nodes {
		_, ok := h.projects[n.ProjectId]
		if !ok {
			continue
		}
		_, ok = h.clusters[n.ClusterId]
		if !ok {
			continue
		}

		h.nodes = append(h.nodes, node{
			ProjectID: n.ProjectId,
			ClusterID: n.ClusterId,
			InnerIP:   n.InnerIp,
			Creator:   n.Creator,
		})
	}

	return nil
}

func (h *migrateHandle) migrateProject() error {

	for _, p := range h.projects {
		_, err := h.cmMgr.findProject(p.ProjectID)
		if err != nil {
			// cm 不存在，需要创建
			cmP := cc2CmProject(p)
			err = h.cmMgr.createProject(cmP)
			if err != nil {
				blog.Warnf("project(%s) migrate failed, err : %v.", p.ProjectID, err)
				continue
			}
		}
	}

	return nil
}

func (h *migrateHandle) migrateCluster() error {

	for _, c := range h.clusters {
		_, err := h.cmMgr.findCluster(c.ClusterID)
		if err != nil {
			// 需要创建
			cmCluster := cc2CmCluster(c)
			err = h.cmMgr.createClusters(cmCluster)
			if err != nil {
				blog.Errorf("migrate cluster(%s) failed, err : %v.", c.ClusterID, err)
				continue
			}
		}
	}

	return nil
}

func (h *migrateHandle) genCmCreateClusters(projectID, clusterID string) (*cluster, error) {

	ccCluster, err := h.ccMgr.clusterInfo(projectID, clusterID)
	if err != nil {
		return nil, err
	}

	configVersion, err := h.ccMgr.versionConfig(clusterID)
	if err != nil {
		return nil, err
	}

	allMaster, err := h.ccMgr.getAllMaster()
	var masters []string
	if err == nil {
		for _, masterTMP := range allMaster {
			if masterTMP.ClusterId == clusterID {
				masters = append(masters, masterTMP.InnerIp)
			}
		}
	}

	var versionConfigure ccversionConfigure
	err = json.Unmarshal([]byte(configVersion.Configure), &versionConfigure)
	if err != nil {
		blog.Errorf("cluster(%s) config version deJson failed, err: %v", clusterID, err)
		return nil, err
	}

	projectCurr := h.projects[projectID]

	return &cluster{
		ClusterID:            ccCluster.ClusterID,
		ClusterName:          ccCluster.Name,
		ManageType:           "INDEPENDENT_CLUSTER",
		Provider:             defaultProvider,
		VpcID:                versionConfigure.VpcID,
		ProjectID:            ccCluster.ProjectId,
		BusinessID:           projectCurr.BusinessID,
		Environment:          ccCluster.Environment,
		EngineType:           ccCluster.Type,
		ClusterType:          "single",
		Region:               strconv.Itoa(ccCluster.AreaId),
		IsExclusive:          false,
		Creator:              ccCluster.Creator,
		NetworkSettings:      createClustersNetworkSettings{},
		ClusterBasicSettings: createClustersClusterBasicSettings{},
		AreaId:               ccCluster.AreaId,
		Master:               masters,
	}, nil
}

func (h *migrateHandle) migrateNode() error {

	for _, data := range h.nodes {
		nodeTMP := cc2CmNode(data)

		err := h.cmMgr.createNode(nodeTMP)
		if err != nil {
			blog.Warnf("migrate node(%s) failed, err : %v.", data.InnerIP, err)
			continue
		}

	}

	return nil
}

// 数据转换
func cc2CmProject(data project) cmCreateProject {
	return cmCreateProject{
		ProjectID:   data.ProjectID,
		Name:        data.Name,
		EnglishName: data.EnglishName,
		Creator:     data.Creator,
		ProjectType: data.ProjectType,
		UseBKRes:    data.UseBKRes,
		Description: data.Description,
		IsOffline:   data.IsOffline,
		Kind:        data.Kind,
		BusinessID:  data.BusinessID,
		DeployType:  data.DeployType,
		BgID:        data.BgID,
		BgName:      data.BgName,
		DeptID:      data.DeptID,
		DeptName:    data.DeptName,
		CenterID:    data.CenterID,
		CenterName:  data.CenterName,
		IsSecret:    data.IsSecret,
	}
}

func cc2CmCluster(data cluster) cmCreateCluster {

	if l := utf8.RuneCountInString(data.Region); l < 2 || l > 100 {
		data.Region = defaultClusterRegion
	}

	if regexp.MustCompile("^[0-9a-zA-Z-]+$").MatchString(data.Region) {
		data.Region = defaultClusterRegion
	}

	return cmCreateCluster{
		ClusterID:           data.ClusterID,
		ClusterName:         data.ClusterName,
		Provider:            data.Provider,
		Region:              data.Region,
		VpcID:               data.VpcID,
		ProjectID:           data.ProjectID,
		BusinessID:          data.BusinessID,
		Environment:         data.Environment,
		EngineType:          data.EngineType,
		IsExclusive:         data.IsExclusive,
		ClusterType:         data.ClusterType,
		FederationClusterID: data.FederationClusterID,
		Creator:             data.Creator,
		OnlyCreateInfo:      defaultClusterOnlyCreateInfo,
		CloudID:             data.CloudID,
		ManageType:          data.ManageType,
		Master:              data.Master,
		Nodes:               data.Node,
		SystemReinstall:     data.SystemReinstall,
		InitLoginPassword:   data.InitLoginPassword,
		NetworkType:         data.NetworkType,
		ClusterBasicSettings: clusterBasicSettings{
			Version: defaultClusterBasicSettingsVersion,
		},
	}
}

func cc2CmNode(data node) cmCreateNode {
	return cmCreateNode{
		ClusterID:         data.ClusterID,
		Nodes:             []string{data.InnerIP},
		InitLoginPassword: data.InitLoginPassword,
		NodeGroupID:       data.NodeGroupID,
		OnlyCreateInfo:    defaultNodeOnlyCreateInfo,
		Operator:          data.Creator,
	}
}
