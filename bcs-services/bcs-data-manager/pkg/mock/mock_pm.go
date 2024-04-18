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

package mock

import (
	"context"
	"encoding/json"

	pm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsproject"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MockPmClient struct {
	mock.Mock
}

func NewMockPmClient() bcsproject.BcsProjectManagerClient {
	return &MockPmClient{}
}

func (m *MockPmClient) GetBcsProjectManagerConn() (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	var conn *grpc.ClientConn
	conn, _ = grpc.Dial("127.0.0.1", opts...)
	return conn, nil
}

func (m *MockPmClient) NewGrpcClientWithHeader(ctx context.Context, conn *grpc.ClientConn) *bcsproject.BcsProjectClientWithHeader {
	return &bcsproject.BcsProjectClientWithHeader{
		Cli: NewMockPm(),
		Ctx: ctx,
	}
}

// MockPm mock project manager
type MockPm struct {
	mock.Mock
}

func NewMockPm() pm.BCSProjectClient {
	return &MockPm{}
}

func (m *MockPm) CreateProject(ctx context.Context, in *pm.CreateProjectRequest, opts ...grpc.CallOption) (*pm.ProjectResponse, error) {
	return nil, nil
}
func (m *MockPm) GetProject(ctx context.Context, in *pm.GetProjectRequest, opts ...grpc.CallOption) (*pm.ProjectResponse, error) {
	rawProject1 := []byte("{\"code\":0,\"message\":\"success\",\"data\":{\"createTime\":\"2018-02-05T16:33:53Z\",\"updateTime\":\"2019-04-16T10:57:31Z\",\"creator\":\"bellkeyang\",\"updater\":\"sundytian\",\"managers\":\"bellkeyang;sundytian\",\"projectID\":\"b37778ec757544868a01e1f01f07037f\",\"name\":\"K8S容器服务测试\",\"projectCode\":\"k8stest\",\"useBKRes\":true,\"description\":\"K8S容器服务测试，勿动大大的\",\"isOffline\":false,\"kind\":\"k8s\",\"businessID\":\"100148\",\"isSecret\":false,\"projectType\":5,\"deployType\":2,\"BGID\":\"956\",\"BGName\":\"IEG互动娱乐事业群\",\"deptID\":\"25923\",\"deptName\":\"技术运营部\",\"centerID\":\"26050\",\"centerName\":\"蓝鲸产品中心\"},\"requestID\":\"68197e5e660449e1ae2ff44102af4ffe\",\"webAnnotations\":{\"perms\":null}}")
	projectRsp1 := &pm.ProjectResponse{}
	json.Unmarshal(rawProject1, projectRsp1)
	rawProject2 := []byte("{\"code\":0,\"message\":\"success\",\"data\":{\"createTime\":\"2019-07-23T11:10:41Z\",\"updateTime\":\"2019-07-23T07:11:29Z\",\"creator\":\"bellkeyang\",\"updater\":\"bellkeyang\",\"managers\":\"bellkeyang\",\"projectID\":\"ab2b254938e84f6b86b466cc22e730b1\",\"name\":\"mesos服务测试\",\"projectCode\":\"mesostest\",\"useBKRes\":false,\"description\":\"暂无\",\"isOffline\":false,\"kind\":\"mesos\",\"businessID\":\"100148\",\"isSecret\":false,\"projectType\":5,\"deployType\":1,\"BGID\":\"956\",\"BGName\":\"IEG互动娱乐事业群\",\"deptID\":\"25923\",\"deptName\":\"技术运营部\",\"centerID\":\"26050\",\"centerName\":\"蓝鲸产品中心\"},\"requestID\":\"c76287037f544684915c8581ced1eada\",\"webAnnotations\":{\"perms\":null}}")
	projectRsp2 := &pm.ProjectResponse{}
	json.Unmarshal(rawProject2, projectRsp2)
	m.On("GetProject", &pm.GetProjectRequest{ProjectIDOrCode: "b37778ec757544868a01e1f01f07037f"}).
		Return(projectRsp1, nil)
	m.On("GetProject", &pm.GetProjectRequest{ProjectIDOrCode: "ab2b254938e84f6b86b466cc22e730b1"}).
		Return(projectRsp2, nil)
	args := m.Called(in)
	return args.Get(0).(*pm.ProjectResponse), args.Error(1)
}
func (m *MockPm) UpdateProject(ctx context.Context, in *pm.UpdateProjectRequest, opts ...grpc.CallOption) (*pm.ProjectResponse, error) {
	return nil, nil
}
func (m *MockPm) DeleteProject(ctx context.Context, in *pm.DeleteProjectRequest, opts ...grpc.CallOption) (*pm.ProjectResponse, error) {
	return nil, nil
}
func (m *MockPm) ListProjects(ctx context.Context, in *pm.ListProjectsRequest, opts ...grpc.CallOption) (*pm.ListProjectsResponse, error) {
	return nil, nil
}
func (m *MockPm) ListAuthorizedProjects(ctx context.Context, in *pm.ListAuthorizedProjReq, opts ...grpc.CallOption) (*pm.ListAuthorizedProjResp, error) {
	return nil, nil
}
