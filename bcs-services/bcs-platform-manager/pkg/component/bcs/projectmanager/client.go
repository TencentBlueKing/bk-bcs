/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package projectmanager 项目管理服务client
package projectmanager

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// NewClient create project manager service client
func NewClient(tlsConfig *tls.Config, microRgt registry.Registry) error {
	if !discovery.UseServiceDiscovery() {
		dis := discovery.NewModuleDiscovery(constants.ProjectManagerServiceName, microRgt)
		err := dis.Start()
		if err != nil {
			return err
		}
		bcsproject.SetClientConfig(tlsConfig, dis)
	} else {
		bcsproject.SetClientConfig(tlsConfig, nil)
	}
	return nil
}

// GetProject get project from project manager
func GetProject(ctx context.Context, projectIDOrCode string) (*types.Project, error) {
	cli, close, err := bcsproject.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.Project.GetProject(ctx, &bcsproject.GetProjectRequest{
		ProjectIDOrCode: projectIDOrCode,
	})
	if err != nil {
		return nil, fmt.Errorf("GetProject error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("GetProject error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return &types.Project{
		Name:          p.Data.Name,
		ProjectId:     p.Data.ProjectID,
		Code:          p.Data.ProjectCode,
		CcBizID:       p.Data.BusinessID,
		Creator:       p.Data.Creator,
		Kind:          p.Data.Kind,
		RawCreateTime: p.Data.CreateTime,
	}, nil
}
