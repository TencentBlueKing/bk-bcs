package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Import 导入用户集群(支持多云集群导入功能: 集群ID/kubeConfig)
func (c *ClusterMgr) Import(req manager.ImportClusterReq) error {
	resp, err := c.client.ImportCluster(c.ctx, &clustermanager.ImportClusterReq{
		ClusterID:   req.ClusterID,
		ClusterName: req.ClusterName,
		Provider:    req.Provider,
		ProjectID:   req.ProjectID,
		BusinessID:  req.BusinessID,
		Environment: req.Environment,
		EngineType:  req.EngineType,
		IsExclusive: &wrapperspb.BoolValue{
			Value: req.IsExclusive,
		},
		ClusterType: req.ClusterType,
		Creator:     "bcs",
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
