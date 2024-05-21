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

package xtencentcloud

import (
	"fmt"

	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tvpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// VpcClient tencent cloud vpc client
type VpcClient struct {
	// domain for tencent cloud clb service
	domain string
	// secret id for tencent cloud account
	secretID string
	// secret key for tencent cloud account
	secretKey string
	// client profile for tencent cloud sdk
	cpf *tprofile.ClientProfile
	// credential for tencent cloud sdk
	credential *tcommon.Credential
	// tencent cloud region
	region string

	cli *tvpc.Client
}

// NewClient return new client for t vpc
func NewClient(domain, secretID, secretKey, region string) (*VpcClient, error) {
	c := &VpcClient{
		domain:    domain,
		secretID:  secretID,
		secretKey: secretKey,
		region:    region,
	}

	credential := tcommon.NewCredential(
		c.secretID,
		c.secretKey,
	)
	cpf := tprofile.NewClientProfile()
	if len(c.domain) != 0 {
		cpf.HttpProfile.Endpoint = c.domain
	}
	c.credential = credential
	c.cpf = cpf

	newCli, err := tvpc.NewClient(c.credential, c.region, c.cpf)
	if err != nil {
		return nil, fmt.Errorf("create clb client for region %s failed, err %s", c.region, err.Error())
	}
	c.cli = newCli
	return c, nil
}

// DescribeAddressTemplate describe addr template
func (c *VpcClient) DescribeAddressTemplate(addrTemplateID string) (*tvpc.
	DescribeAddressTemplatesResponse, error) {
	req := tvpc.NewDescribeAddressTemplatesRequest()
	req.Filters = []*tvpc.Filter{
		{
			Name: tcommon.StringPtr("address-template-id"),
			Values: []*string{
				tcommon.StringPtr(addrTemplateID),
			},
		},
	}
	resp, err := c.cli.DescribeAddressTemplates(req)
	if err != nil {
		reqID := ""
		if resp != nil && resp.Response != nil && resp.Response.RequestId != nil {
			reqID = *resp.Response.RequestId
		}
		return nil, fmt.Errorf("describe address template failed, req: %s,req_id: %s, err: %s", addrTemplateID,
			reqID, err.Error())
	}
	return resp, nil
}

// AddTemplateMember add member to addr template
func (c *VpcClient) AddTemplateMember(addrTemplateID string, members []*tvpc.MemberInfo) error {
	req := tvpc.NewAddTemplateMemberRequest()
	req.TemplateId = tcommon.StringPtr(addrTemplateID)
	req.TemplateMember = members
	resp, err := c.cli.AddTemplateMember(req)
	if err != nil {
		reqID := ""
		if resp != nil && resp.Response != nil && resp.Response.RequestId != nil {
			reqID = *resp.Response.RequestId
		}
		return fmt.Errorf("addTemplateMember failed, req_id: %s,err: %s", reqID, err.Error())
	}

	return nil
}
