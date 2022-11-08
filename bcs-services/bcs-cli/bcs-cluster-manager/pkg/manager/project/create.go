package project

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *ProjectMgr) Create(req manager.CreateProjectReq) error {
	resp, err := c.client.CreateProject(c.ctx, &clustermanager.CreateProjectRequest{
		Name:        req.Name,
		EnglishName: req.EnglishName,
		ProjectType: req.ProjectType,
		Kind:        req.Kind,
		UseBKRes:    req.UseBKRes,
		DeployType:  req.DeployType,
		BusinessID:  req.BusinessID,
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
