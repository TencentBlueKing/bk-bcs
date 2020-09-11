/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tencentcloud

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

func (c *Client) describeSubnets(subnetIDs []string) ([]*vpc.Subnet, error) {
	req := vpc.NewDescribeSubnetsRequest()
	req.SubnetIds = common.StringPtrs(subnetIDs)

	blog.V(3).Infof("DescribeSubnets req: %s", req.ToJsonString())

	resp, err := c.vpcClient.DescribeSubnets(req)
	if err != nil {
		return nil, fmt.Errorf("DescribeSubnets failed, err %s", err.Error())
	}

	blog.V(3).Infof("DescribeSubnets resp: %s", resp.ToJsonString())
	return resp.Response.SubnetSet, nil
}
