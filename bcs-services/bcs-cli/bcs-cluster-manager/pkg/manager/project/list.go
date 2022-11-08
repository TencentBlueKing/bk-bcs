package project

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

func (c *ProjectMgr) List(req manager.ListProjectReq) (resp manager.ListProjectResp, err error) {
	servResp, err := c.client.ListProject(c.ctx, &clustermanager.ListProjectRequest{})
	if err != nil {
		return
	}

	if servResp != nil && servResp.Code != 0 {
		return resp, errors.New(servResp.Message)
	}

	for _, v := range servResp.Data {
		credentials := make(map[string]manager.Credential, 0)
		for x, y := range v.Credentials {
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

		resp.Data = append(resp.Data, &manager.Project{
			ProjectID:   v.ProjectID,
			Name:        v.Name,
			EnglishName: v.EnglishName,
			Creator:     v.Creator,
			Updater:     v.Updater,
			ProjectType: v.ProjectType,
			UseBKRes:    v.UseBKRes,
			BusinessID:  v.BusinessID,
			Description: v.Description,
			IsOffline:   v.IsOffline,
			Kind:        v.Kind,
			DeployType:  v.DeployType,
			BgID:        v.BgID,
			BgName:      v.BgName,
			DeptID:      v.DeptID,
			DeptName:    v.DeptName,
			CenterID:    v.CenterID,
			CenterName:  v.CenterName,
			IsSecret:    v.IsSecret,
			Credentials: credentials,
			CreatTime:   v.CreatTime,
			UpdateTime:  v.UpdateTime,
		})
	}

	return
}
