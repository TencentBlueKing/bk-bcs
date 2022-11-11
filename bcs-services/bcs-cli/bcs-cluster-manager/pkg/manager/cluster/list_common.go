package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// ListCommon 查询公共集群及公共集群所属权限
func (c *ClusterMgr) ListCommon(req manager.ListCommonClusterReq) (resp manager.ListCommonClusterResp, err error) {
	servResp, err := c.client.ListCommonCluster(c.ctx, &clustermanager.ListCommonClusterReq{})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		master := make([]string, 0)
		for _, y := range v.Master {
			master = append(master, y.InnerIP)
		}

		resp.Data = append(resp.Data, &manager.Cluster{
			ClusterID:   v.ClusterID,
			ProjectID:   v.ProjectID,
			BusinessID:  v.BusinessID,
			EngineType:  v.BusinessID,
			IsExclusive: v.IsExclusive,
			ClusterType: v.ClusterType,
			Creator:     v.Creator,
			Updater:     v.Updater,
			ManageType:  v.ManageType,
			ClusterName: v.ClusterName,
			Environment: v.Environment,
			Provider:    v.Provider,
			Description: v.Description,
			ClusterBasicSettings: manager.ClusterBasicSettings{
				Version: v.ClusterBasicSettings.Version,
			},
			NetworkType: v.NetworkType,
			Region:      v.Region,
			VpcID:       v.VpcID,
			NetworkSettings: manager.NetworkSettings{
				CidrStep:      v.NetworkSettings.CidrStep,
				MaxNodePodNum: v.NetworkSettings.MaxNodePodNum,
				MaxServiceNum: v.NetworkSettings.MaxServiceNum,
			},
			Master: master,
		})
	}

	return
}
