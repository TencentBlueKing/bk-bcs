package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Update 更新集群
func (c *ClusterMgr) Update(req manager.UpdateClusterReq) error {
	resp, err := c.client.UpdateCluster(c.ctx, &clustermanager.UpdateClusterReq{
		ClusterID:   req.ClusterID,
		ProjectID:   req.ProjectID,
		BusinessID:  req.BusinessID,
		EngineType:  req.EngineType,
		IsExclusive: &wrapperspb.BoolValue{Value: req.IsExclusive},
		ClusterType: req.ClusterType,
		Updater:     req.Updater,
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
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
