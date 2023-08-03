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
 *
 */

package api

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"github.com/aws/aws-sdk-go/service/iam"
)

// IAMClient aws iam client
type IAMClient struct {
	iamClient *iam.IAM
}

// NewIAMClient init aws iam client
func NewIAMClient(opt *cloudprovider.CommonOption) (*IAMClient, error) {
	sess, err := NewSession(opt)
	if err != nil {
		return nil, err
	}

	return &IAMClient{
		iamClient: iam.New(sess),
	}, nil
}

// GetRole gets the iam role
func (c *IAMClient) GetRole(input *iam.GetRoleInput) (*iam.Role, error) {
	blog.Infof("GetRole input: %", utils.ToJSONString(input))
	output, err := c.iamClient.GetRole(input)
	if err != nil {
		blog.Errorf("GetRole failed: %v", err)
		return nil, err
	}
	if output == nil || output.Role == nil {
		blog.Errorf("GetRole lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("GetRole %s successful: %", *input.RoleName)

	return output.Role, nil
}
