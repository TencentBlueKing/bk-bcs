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
	"errors"

	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

const stsEndpoint = "sts.aliyuncs.com"

// GetCallerIdentity queries the credential identity without exposing SDK error details.
func (c *Client) GetCallerIdentity() (*sts.GetCallerIdentityResponseBody, error) {
	c.stsOnce.Do(func() {
		c.stsClient, c.stsErr = sts.NewClient(c.clientConfig(stsEndpoint))
	})
	if c.stsErr != nil {
		return nil, errors.New("alicloud account validation client initialization failed")
	}
	resp, err := c.stsClient.GetCallerIdentity()
	if err != nil {
		// SDK errors may contain request details; do not expose or log credentials.
		return nil, errors.New("alicloud account validation failed; check AccessKey credentials and STS connectivity")
	}
	if resp == nil || resp.Body == nil || tea.StringValue(resp.Body.AccountId) == "" ||
		tea.StringValue(resp.Body.IdentityType) == "" {
		return nil, errors.New("alicloud STS returned incomplete account identity")
	}
	return resp.Body, nil
}
