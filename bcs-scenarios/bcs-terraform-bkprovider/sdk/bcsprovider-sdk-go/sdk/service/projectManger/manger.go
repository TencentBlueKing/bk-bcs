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

// Package projectManger project-service
package projectManger

import (
	"bytes"
	"context"
	"fmt"
	"regexp"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/header"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/roundtrip"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// Service 对接 bcs-project-manager
type Service interface {
	// CreateProject 创建
	CreateProject(ctx context.Context, req *CreateProjectRequest) (*pb.ProjectResponse, error)

	// DeleteProject 删除
	DeleteProject(ctx context.Context, req *DeleteProjectRequest) (*pb.ProjectResponse, error)

	// UpdateProject 修改
	UpdateProject(ctx context.Context, req *UpdateProjectRequest) (*pb.ProjectResponse, error)

	// GetProject 查询
	GetProject(ctx context.Context, req *GetProjectRequest) (*pb.ProjectResponse, error)

	// ListProjects 查询
	ListProjects(ctx context.Context, req *ListProjectsRequest) (*pb.ListProjectsResponse, error)
}

// NewService return Service
func NewService(config *options.Config, roundtrip roundtrip.Client) (Service, error) {
	if config == nil {
		return nil, errors.Errorf("config cannot be empty.")
	}
	if roundtrip == nil {
		return nil, errors.Errorf("roundtrip cannot be empty.")
	}

	h := &handler{
		config: config,
		// roundtrip
		roundtrip: roundtrip,
		// api
		backendApi: map[string]string{},
	}
	if err := h.init(); err != nil {
		return nil, errors.Wrapf(err, "init handler failed")
	}

	return h, nil
}

// handler impl Service
type handler struct {
	// config 配置
	config *options.Config

	// roundtrip http client
	roundtrip roundtrip.Client

	// backendApi 后端完整路径
	backendApi map[string]string

	// projectCodeRegexp 正则
	projectCodeRegexp *regexp.Regexp
}

func (h *handler) init() error {
	regex, err := regexp.Compile(projectCodeRegexp)
	if err != nil {
		return errors.Wrapf(err, "error compiling regex: %s", projectCodeRegexp)
	}
	h.projectCodeRegexp = regex

	apis := []string{
		createProjectApi,
		deleteProjectApi,
		updateProjectApi,
		getProjectApi,
		listProjectsApi,
	}
	addr := h.config.BcsGatewayAddr

	for _, api := range apis {
		h.backendApi[api] = utils.PathJoin(addr, api)
	}

	return nil
}

// CreateProject 创建
func (h *handler) CreateProject(ctx context.Context, req *CreateProjectRequest) (*pb.ProjectResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectCode) == 0 {
		return nil, errors.Errorf("project_code connot be empty.")
	}
	if len(req.ProjectCode) < 2 || len(req.ProjectCode) > 32 {
		return nil, errors.Errorf("project_code can only be 2-32 characters long.")
	}
	if !h.projectCodeRegexp.MatchString(req.ProjectCode) { // project_code必须以小写字母开头，内容由小写字母、数字、中划线组成
		return nil, errors.Errorf("project_code must start with a lowercase letter and consist of lowercase letters," +
			" numbers, and hyphens.")
	}
	if len(req.Name) == 0 {
		return nil, errors.Errorf("project_name connot be empty.")
	}
	if len(req.BusinessID) == 0 {
		return nil, errors.Errorf("project_businessID connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 转换(提示：1、切换方法)
	body, err := h.createProjectReqToBytes(req)
	if err != nil {
		return nil, err
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Post(ctx, req.RequestID, h.createProjectApi(), body)
	if err != nil {
		return nil, errors.Wrapf(err, "get project detail failed, traceId: %s, projectCode: %s",
			req.RequestID, req.ProjectCode)
	}

	result := new(pb.ProjectResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// createProjectReqToBytes to pb struct data
func (h *handler) createProjectReqToBytes(req *CreateProjectRequest) ([]byte, error) {
	obj := &pb.CreateProjectRequest{
		Name:        req.Name,
		ProjectCode: req.ProjectCode,
		BusinessID:  req.BusinessID,
		Description: req.Description,
		Creator:     h.config.Username,
		//
		Kind: "k8s",
		//
		ProjectType: 0,
		//
		DeployType: 0,
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, obj); err != nil {
		return nil, errors.Wrapf(err, "pb.CreateProjectRequest marhsal failed.")
	}

	return body.Bytes(), nil
}

// createProjectApi api
func (h *handler) createProjectApi() string {
	return h.backendApi[createProjectApi]
}

// DeleteProject 删除
func (h *handler) DeleteProject(ctx context.Context, req *DeleteProjectRequest) (*pb.ProjectResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("project_id connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 请求
	resp, err := h.roundtrip.Delete(ctx, req.RequestID, h.deleteProjectApi(req.ProjectID), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "delete project detail failed, traceId: %s, projectId: %s",
			req.RequestID, req.ProjectID)
	}

	result := new(pb.ProjectResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// deleteProjectApi api
func (h *handler) deleteProjectApi(projectID string) string {
	return fmt.Sprintf(h.backendApi[deleteProjectApi], projectID)
}

// UpdateProject 修改
func (h *handler) UpdateProject(ctx context.Context, req *UpdateProjectRequest) (*pb.ProjectResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("project_id connot be empty.")
	}
	if len(req.Name) == 0 {
		return nil, errors.Errorf("project_name connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 转换
	body, err := h.updateProjectReqToBytes(req)
	if err != nil {
		return nil, err
	}

	// 请求
	resp, err := h.roundtrip.Put(ctx, req.RequestID, h.updateProjectApi(req.ProjectID), body)
	if err != nil {
		return nil, errors.Wrapf(err, "update project detail failed, traceId: %s, projectId: %s",
			req.RequestID, req.ProjectID)
	}

	result := new(pb.ProjectResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// updateProjectReqToBytes to pb struct data
func (h *handler) updateProjectReqToBytes(req *UpdateProjectRequest) ([]byte, error) {
	obj := &pb.UpdateProjectRequest{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		BusinessID:  req.BusinessID,
		Managers:    req.Managers,
		Creator:     req.Creator,
		// set Updater
		Updater: h.config.Username,
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, obj); err != nil {
		return nil, errors.Wrapf(err, "pb.UpdateProjectRequest marhsal failed.")
	}

	return body.Bytes(), nil
}

// updateProjectApi api
func (h *handler) updateProjectApi(projectID string) string {
	return fmt.Sprintf(h.backendApi[updateProjectApi], projectID)
}

// GetProject 查询
func (h *handler) GetProject(ctx context.Context, req *GetProjectRequest) (*pb.ProjectResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("project_id connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	resp, err := h.roundtrip.Get(ctx, req.RequestID, h.getProjectApi(req.ProjectID), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get project detail failed, traceId: %s, projectId: %s",
			req.RequestID, req.ProjectID)
	}

	result := new(pb.ProjectResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// getProjectApi api
func (h *handler) getProjectApi(projectID string) string {
	return fmt.Sprintf(h.backendApi[getProjectApi], projectID)
}

// ListProjects 查询
func (h *handler) ListProjects(ctx context.Context, req *ListProjectsRequest) (*pb.ListProjectsResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	resp, err := h.roundtrip.Get(ctx, req.RequestID, h.listProjectsApi(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "list projects failed, traceId: %s", req.RequestID)
	}

	result := new(pb.ListProjectsResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listProjectsApi api
func (h *handler) listProjectsApi() string {
	return h.backendApi[listProjectsApi]
}
