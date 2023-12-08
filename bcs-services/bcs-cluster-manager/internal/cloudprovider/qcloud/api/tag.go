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

package api

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tag "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tag/v20180813"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

const (
	projectLimit = 1000
)

// NewTagClient init tag client
func NewTagClient(opt *cloudprovider.CommonOption) (*TagClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	credential := common.NewCredential(opt.Account.SecretID, opt.Account.SecretKey)
	cpf := profile.NewClientProfile()
	if opt.CommonConf.CloudInternalEnable {
		cpf.HttpProfile.Endpoint = opt.CommonConf.CloudDomain
	}

	cli, err := tag.NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}

	return &TagClient{tag: cli}, nil
}

// TagClient xxx
type TagClient struct {
	tag *tag.Client
}

// ListProjects 获取云账号下所有项目列表
func (tc *TagClient) ListProjects() ([]*tag.Project, error) {
	var (
		projects       = make([]*tag.Project, 0)
		initOffset     uint64
		projectListLen = projectLimit
	)

	for {
		if projectListLen != projectLimit {
			break
		}
		req := tag.NewDescribeProjectsRequest()
		req.AllList = common.Uint64Ptr(1)
		req.Offset = common.Uint64Ptr(initOffset)
		req.Limit = common.Uint64Ptr(projectLimit)

		resp, err := tc.tag.DescribeProjects(req)
		if err != nil {
			blog.Errorf("tag client DescribeProjects failed, %s", err.Error())
			continue
		}

		// check response
		response := resp.Response
		if response == nil {
			blog.Errorf("tag client DescribeProjects but lost response information")
			continue
		}
		projects = append(projects, response.Projects...)

		projectListLen = len(response.Projects)
		initOffset += projectLimit
	}

	blog.Infof("ListKeyPairs successful")

	return projects, nil
}
