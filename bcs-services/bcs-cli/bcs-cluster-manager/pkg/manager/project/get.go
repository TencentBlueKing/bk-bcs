package project

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *ProjectMgr) Get(req manager.GetProjectReq) (resp manager.GetProjectResp, err error) {
	servResp, err := c.client.GetProject(c.ctx, &clustermanager.GetProjectRequest{
		ProjectID: req.ProjectID,
	})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	credentials := make(map[string]manager.Credential, 0)
	for x, y := range servResp.Data.Credentials {
		credentials[x] = manager.Credential{
			Key:               y.Key,
			Secret:            y.Secret,
			SubscriptionID:    y.SubscriptionID,
			TenantID:          y.TenantID,
			ResourceGroupName: y.ResourceGroupName,
			ClientID:          y.ClientID,
			ClientSecret:      y.ClientSecret,
		}
	}

	resp.Data = manager.Project{
		ProjectID:   servResp.Data.ProjectID,
		Name:        servResp.Data.Name,
		EnglishName: servResp.Data.EnglishName,
		Creator:     servResp.Data.Creator,
		Updater:     servResp.Data.Updater,
		ProjectType: servResp.Data.ProjectType,
		UseBKRes:    servResp.Data.UseBKRes,
		BusinessID:  servResp.Data.BusinessID,
		Description: servResp.Data.Description,
		IsOffline:   servResp.Data.IsOffline,
		Kind:        servResp.Data.Kind,
		DeployType:  servResp.Data.DeployType,
		BgID:        servResp.Data.BgID,
		BgName:      servResp.Data.BgName,
		DeptID:      servResp.Data.DeptID,
		DeptName:    servResp.Data.DeptName,
		CenterID:    servResp.Data.CenterID,
		CenterName:  servResp.Data.CenterName,
		IsSecret:    servResp.Data.IsSecret,
		Credentials: credentials,
		CreatTime:   servResp.Data.CreatTime,
		UpdateTime:  servResp.Data.UpdateTime,
	}

	return
}
