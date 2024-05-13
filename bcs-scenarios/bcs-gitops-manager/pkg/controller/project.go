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

package controller

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4/bcsproject"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
)

// ProjectControl for bcs project data sync
type ProjectControl interface {
	Controller

	GetProject(ctx context.Context, projectCode string) (*bcsproject.Project, error)
}

// NewProjectController create project controller instance
func NewProjectController() ProjectControl {
	return &project{
		option: options.GlobalOptions(),
	}
}

// project for bk-bcs project information
// syncing to gitops system
type project struct {
	option *options.Options
	client bcsproject.BCSProjectClient
	conn   *grpc.ClientConn
}

// Init controller
func (control *project) Init() error {
	// if control.option.Mode == common.ModeService {
	//	return fmt.Errorf("service mode is not implenmented")
	// }
	// init with raw grpc connection
	if err := control.initClient(); err != nil {
		return err
	}
	return nil
}

// Start controller
func (control *project) Start() error {
	return nil
}

// Stop controller
func (control *project) Stop() {
	control.conn.Close() // nolint
}

// GetProject with specified project code
func (control *project) GetProject(ctx context.Context, projectCode string) (*bcsproject.Project, error) {
	// get information from project-manager
	req := &bcsproject.GetProjectRequest{ProjectIDOrCode: projectCode}
	// setting auth info
	header := metadata.New(map[string]string{"Authorization": fmt.Sprintf("Bearer %s",
		control.option.APIGatewayToken)})
	outCxt := metadata.NewOutgoingContext(ctx, header)
	resp, err := control.client.GetProject(outCxt, req)
	if err != nil {
		blog.Errorf("get project %s details from project-manager failed, %s", projectCode, err.Error())
		return nil, fmt.Errorf("request to project-manager failure %s", err.Error())
	}
	if resp.Code != 0 {
		blog.Errorf("request project-manager for %s failed, %s", projectCode, resp.Message)
		return nil, fmt.Errorf("project-manager response failure: %s", resp.Message)
	}
	if resp.Data == nil {
		blog.Warnf("no project %s in bcs-project-manager", projectCode)
		return nil, nil
	}
	blog.V(5).Infof("project %s request details: %+v", projectCode, resp.Data)
	return resp.Data, nil
}

func (control *project) initClient() error {
	// create grpc connection
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(control.option.APIGatewayToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", control.option.APIGatewayToken)
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(control.option.ClientTLS)))
	conn, err := grpc.Dial(control.option.APIGateway, opts...)
	if err != nil {
		blog.Errorf("project controller dial bcs-api-gateway %s failure, %s",
			control.option.APIGateway, err.Error())
		return err
	}
	control.client = bcsproject.NewBCSProjectClient(conn)
	control.conn = conn
	blog.Infof("project conctroller init project-manager with %s successfully", control.option.APIGateway)
	return nil
}
