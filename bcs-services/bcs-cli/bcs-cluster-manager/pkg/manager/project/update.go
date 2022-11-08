package project

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (c *ProjectMgr) Update(req manager.UpdateProjectReq) error {
	credentials := make(map[string]*clustermanager.Credential, 0)
	for x, y := range req.Credentials {
		credentials[x] = &clustermanager.Credential{
			Key:               y.Key,
			Secret:            y.Secret,
			SubscriptionID:    y.SubscriptionID,
			TenantID:          y.TenantID,
			ResourceGroupName: y.ResourceGroupName,
			ClientID:          y.ClientID,
			ClientSecret:      y.ClientSecret,
		}
	}

	resp, err := c.client.UpdateProject(c.ctx, &clustermanager.UpdateProjectRequest{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Updater:     "bcs",
		ProjectType: req.ProjectType,
		UseBKRes: &wrapperspb.BoolValue{
			Value: req.UseBKRes,
		},
		Description: req.Description,
		IsOffline: &wrapperspb.BoolValue{
			Value: req.IsOffline,
		},
		Kind:       req.Kind,
		DeployType: req.DeployType,
		BgID:       req.BgID,
		BgName:     req.BgName,
		DeptID:     req.DeptID,
		DeptName:   req.DeptName,
		CenterID:   req.CenterID,
		CenterName: req.CenterName,
		IsSecret: &wrapperspb.BoolValue{
			Value: req.IsSecret,
		},
		Credentials: credentials,
		BusinessID:  req.BusinessID,
	})
	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
