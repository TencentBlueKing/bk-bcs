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
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = "2018-05-25"

// Client tke common client
type Client struct {
	common.Client
}

// NewClientWithSecretId xxx
func NewClientWithSecretId(secretId, secretKey, region string) (client *Client, err error) {
	cpf := profile.NewClientProfile()
	client = &Client{}
	client.Init(region).WithSecretId(secretId, secretKey).WithProfile(cpf)
	return
}

// NewClient xxx
func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

// DescribeOSImages 获取OS聚合信息
func (c *Client) DescribeOSImages(request *DescribeOSImagesRequest) (response *DescribeOSImagesResponse, err error) {
	return c.DescribeOSImagesWithContext(context.Background(), request)
}

// DescribeOSImagesWithContext 获取OS聚合信息
func (c *Client) DescribeOSImagesWithContext(ctx context.Context, request *DescribeOSImagesRequest) (response *DescribeOSImagesResponse, err error) {
	if request == nil {
		request = NewDescribeOSImagesRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeOSImages require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeOSImagesResponse()
	err = c.Send(request, response)
	return
}
