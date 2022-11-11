package cluster

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// CheckCloudKubeConfig kubeConfig连接集群可用性检测
func (c *ClusterMgr) CheckCloudKubeconfig(req manager.CheckCloudKubeconfigReq) error {
	resp, err := c.client.CheckCloudKubeConfig(c.ctx, &clustermanager.KubeConfigReq{
		KubeConfig: req.Kubeconfig,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
