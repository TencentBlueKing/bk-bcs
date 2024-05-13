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

package eip

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/qcloud-eip/conf"
)

type instanceClient struct {
	conf   *conf.NetConf
	client *cvm.Client
}

func newInstanceClient(conf *conf.NetConf) *instanceClient {
	credential := common.NewCredential(
		conf.Secret,
		conf.UUID,
	)
	cpf := profile.NewClientProfile()

	// set tencentcloud domain
	if len(conf.TencentCloudCVMDomain) != 0 {
		cpf.HttpProfile.Endpoint = conf.TencentCloudCVMDomain
	}

	client, err := cvm.NewClient(credential, conf.Region, cpf)
	if err != nil {
		blog.Errorf("new instance client failed, err %s", err.Error())
		return nil
	}
	return &instanceClient{
		conf:   conf,
		client: client,
	}
}

func (c *instanceClient) describeInstanceByIP(ip string) (*cvm.Instance, error) {
	request := cvm.NewDescribeInstancesRequest()
	request.Limit = common.Int64Ptr(1)
	request.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("private-ip-address"),
			Values: common.StringPtrs([]string{ip}),
		},
	}
	blog.V(3).Infof("send request %s", request.ToJsonString())
	response, err := c.client.DescribeInstances(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		blog.Errorf("describe instances by ip %s failed, err %s", ip, err)
		return nil, fmt.Errorf("describe instances by ip %s failed, err %s", ip, err)
	}
	if err != nil {
		blog.Errorf("describe instances by ip %s failed, err %s", ip, err.Error())
		return nil, fmt.Errorf("describe instances by ip %s failed, err %s", ip, err.Error())
	}
	blog.V(3).Infof("receive response %s", response.ToJsonString())
	if *response.Response.TotalCount == 0 {
		return nil, fmt.Errorf("cannot found cvm by ip %s, please check if the cvm is in the region of your account", ip)
	}
	return response.Response.InstanceSet[0], nil
}
