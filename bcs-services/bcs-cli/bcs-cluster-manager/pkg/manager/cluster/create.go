package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Create 创建集群
func (c *ClusterMgr) Create(req manager.CreateClusterReq) (resp manager.CreateClusterResp, err error) {
	servResp, err := c.client.CreateCluster(c.ctx, &clustermanager.CreateClusterReq{
		ProjectID:   req.ProjectID,
		BusinessID:  req.BusinessID,
		EngineType:  req.EngineType,
		IsExclusive: req.IsExclusive,
		ClusterType: req.ClusterType,
		Creator:     req.Creator,
		ManageType:  req.ManageType,
		ClusterName: req.ClusterName,
		Environment: req.Environment,
		Provider:    req.Provider,
		Description: req.Description,
		ClusterBasicSettings: &clustermanager.ClusterBasicSetting{
			Version: req.ClusterBasicSettings.Version,
		},
		NetworkType: req.NetworkType,
		Region:      req.Region,
		VpcID:       req.VpcID,
		NetworkSettings: &clustermanager.NetworkSetting{
			MaxNodePodNum: req.NetworkSettings.MaxNodePodNum,
			MaxServiceNum: req.NetworkSettings.MaxServiceNum,
			CidrStep:      req.NetworkSettings.CidrStep,
		},
		Master: req.Master,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	resp.ClusterID = servResp.Data.ClusterID
	resp.TaskID = servResp.Task.TaskID

	return
}
