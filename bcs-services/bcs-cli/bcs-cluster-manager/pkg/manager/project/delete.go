package project

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *ProjectMgr) Delete(req manager.DeleteProjectReq) error {
	resp, err := c.client.DeleteProject(c.ctx, &clustermanager.DeleteProjectRequest{
		ProjectID: req.ProjectID,
		IsForce:   req.IsForce,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
