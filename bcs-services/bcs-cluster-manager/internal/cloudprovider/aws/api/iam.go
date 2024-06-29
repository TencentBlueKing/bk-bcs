/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// IAMClient aws iam client
type IAMClient struct {
	iamClient *iam.IAM
}

// NewIAMClient init aws iam client
func NewIAMClient(opt *cloudprovider.CommonOption) (*IAMClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	awsConf := &aws.Config{}
	awsConf.Credentials = credentials.NewStaticCredentials(opt.Account.SecretID, opt.Account.SecretKey, "")

	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, err
	}

	return &IAMClient{
		iamClient: iam.New(sess),
	}, nil
}

// GetRole gets the iam role
func (c *IAMClient) GetRole(input *iam.GetRoleInput) (*iam.Role, error) {
	blog.Infof("GetRole input: %s", utils.ToJSONString(input))
	output, err := c.iamClient.GetRole(input)
	if err != nil {
		blog.Errorf("GetRole failed: %v", err)
		return nil, err
	}
	if output == nil || output.Role == nil {
		blog.Errorf("GetRole lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("GetRole %s successful: %s", *input.RoleName)

	return output.Role, nil
}

// ListRoles gets the iam roles
func (c *IAMClient) ListRoles(input *iam.ListRolesInput) ([]*iam.Role, error) {
	blog.Infof("ListRoles input: %s", utils.ToJSONString(input))
	roles := make([]*iam.Role, 0)
	err := c.iamClient.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		roles = append(roles, page.Roles...)
		return !lastPage
	})
	if err != nil {
		blog.Errorf("ListRoles failed: %v", err)
		return nil, err
	}
	blog.Infof("ListRoles %s successful: %s")

	return roles, nil
}

// ListAttachedRolePolicies gets attached role policies
func (c *IAMClient) ListAttachedRolePolicies(input *iam.ListAttachedRolePoliciesInput) ([]*iam.AttachedPolicy, error) {
	blog.Infof("ListAttachedRolePolicies input: %s", utils.ToJSONString(input))
	policy := make([]*iam.AttachedPolicy, 0)
	err := c.iamClient.ListAttachedRolePoliciesPages(input,
		func(page *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
			policy = append(policy, page.AttachedPolicies...)
			return !lastPage
		})
	if err != nil {
		blog.Errorf("ListAttachedRolePolicies failed: %v", err)
		return nil, err
	}
	blog.Infof("ListAttachedRolePolicies %s successful: %s")

	return policy, nil
}
