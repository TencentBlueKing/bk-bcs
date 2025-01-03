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

// Package api xxx
package api

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// NewASClient init as client
func NewASClient(opt *cloudprovider.CommonOption) (*ASClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	credential := common.NewCredential(opt.Account.SecretID, opt.Account.SecretKey)
	cpf := profile.NewClientProfile()

	cli, err := as.NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}

	return &ASClient{as: cli}, nil
}

// ASClient is the client for as
type ASClient struct {
	as *as.Client
}

// DescribeAutoScalingInstances describe auto scaling instances
// https://cloud.tencent.com/document/api/377/20437
func (c *ASClient) DescribeAutoScalingInstances(asgID string) ([]*AutoScalingInstances, error) {
	blog.Infof("DescribeAutoScalingInstances input: %s", asgID)
	req := as.NewDescribeAutoScalingInstancesRequest()
	req.Limit = common.Int64Ptr(limit)
	if asgID != "" {
		req.Filters = make([]*as.Filter, 0)
		req.Filters = append(req.Filters, &as.Filter{
			Name: common.StringPtr("auto-scaling-group-id"), Values: common.StringPtrs([]string{asgID})})
	}

	got, total := 0, 0
	first := true
	ins := make([]*AutoScalingInstances, 0)
	for got < total || first {
		first = false
		req.Offset = common.Int64Ptr(int64(got))
		resp, err := c.as.DescribeAutoScalingInstances(req)
		if err != nil {
			blog.Errorf("DescribeAutoScalingInstances failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeAutoScalingInstances resp is nil")
			return nil, fmt.Errorf("DescribeAutoScalingInstances resp is nil")
		}
		blog.Infof("DescribeAutoScalingInstances success, requestID: %s", resp.Response.RequestId)
		for i := range resp.Response.AutoScalingInstanceSet {
			ins = append(ins, convertASGInstance(resp.Response.AutoScalingInstanceSet[i]))
		}
		got += len(resp.Response.AutoScalingInstanceSet)
		total = int(*resp.Response.TotalCount)
	}
	return ins, nil
}

// RemoveInstances 从 asg 中删除 CVM 实例，如果实例由弹性伸缩自动创建，则实例会被销毁；如果实例系创建后加入伸缩组的，则会从伸缩组中移除，保留实例。
// https://cloud.tencent.com/document/api/377/20431
func (c *ASClient) RemoveInstances(asgID string, nodeIDs []string) (string, error) {
	blog.Infof("RemoveInstances input: %s, %v", asgID, nodeIDs)
	req := as.NewRemoveInstancesRequest()
	req.AutoScalingGroupId = &asgID
	if len(nodeIDs) > 0 {
		req.InstanceIds = common.StringPtrs(nodeIDs)
	}
	resp, err := c.as.RemoveInstances(req)
	if err != nil {
		blog.Errorf("RemoveInstances failed, err: %s", err.Error())
		return "", err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("DescribeAutoScalingInstances resp is nil")
		return "", fmt.Errorf("DescribeAutoScalingInstances resp is nil")
	}
	blog.Infof("RemoveInstances success, requestID: %s", resp.Response.RequestId)
	return *resp.Response.ActivityId, nil
}

// DetachInstances 从伸缩组移出 CVM 实例，本接口不会销毁实例。
// https://cloud.tencent.com/document/api/377/20436
func (c *ASClient) DetachInstances(asgID string, nodeIDs []string) error {
	blog.Infof("DetachInstances input: %s, %v", asgID, nodeIDs)
	req := as.NewDetachInstancesRequest()
	req.AutoScalingGroupId = &asgID
	if len(nodeIDs) > 0 {
		req.InstanceIds = common.StringPtrs(nodeIDs)
	}
	resp, err := c.as.DetachInstances(req)
	if err != nil {
		blog.Errorf("DetachInstances failed, err: %s", err.Error())
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("DetachInstances resp is nil")
		return fmt.Errorf("DetachInstances resp is nil")
	}
	blog.Infof("DetachInstances success, requestID: %s", resp.Response.RequestId)
	return nil
}

// ModifyDesiredCapacity 修改指定伸缩组的期望实例数, 无activityID
// https://cloud.tencent.com/document/api/377/20432
func (c *ASClient) ModifyDesiredCapacity(asgID string, capacity uint64) error {
	blog.Infof("ModifyDesiredCapacity input: %s, %d", asgID, capacity)
	req := as.NewModifyDesiredCapacityRequest()
	req.AutoScalingGroupId = &asgID
	req.DesiredCapacity = common.Uint64Ptr(capacity)
	resp, err := c.as.ModifyDesiredCapacity(req)
	if err != nil {
		blog.Errorf("ModifyDesiredCapacity failed, err: %s", err.Error())
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("ModifyDesiredCapacity resp is nil")
		return fmt.Errorf("ModifyDesiredCapacity resp is nil")
	}
	blog.Infof("ModifyDesiredCapacity success, requestID: %s", resp.Response.RequestId)
	return nil
}

// ModifyAutoScalingGroup 修改伸缩组的属性
// https://cloud.tencent.com/document/api/377/20433
func (c *ASClient) ModifyAutoScalingGroup(asg *as.ModifyAutoScalingGroupRequest) error {
	blog.Infof("ModifyAutoScalingGroup input: %v", utils.ToJSONString(asg))
	resp, err := c.as.ModifyAutoScalingGroup(asg)
	if err != nil {
		blog.Errorf("ModifyAutoScalingGroup failed, err: %s", err.Error())
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("ModifyAutoScalingGroup resp is nil")
		return fmt.Errorf("ModifyAutoScalingGroup resp is nil")
	}
	blog.Infof("ModifyAutoScalingGroup success, requestID: %s", resp.Response.RequestId)
	return nil
}

// ModifyLaunchConfigurationAttributes 修改启动配置的属性
// https://cloud.tencent.com/document/api/377/31298
func (c *ASClient) ModifyLaunchConfigurationAttributes(req *as.ModifyLaunchConfigurationAttributesRequest) error {
	blog.Infof("ModifyLaunchConfigurationAttributes input: %v", utils.ToJSONString(req))
	resp, err := c.as.ModifyLaunchConfigurationAttributes(req)
	if err != nil {
		blog.Errorf("ModifyLaunchConfigurationAttributes failed, err: %s", err.Error())
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("ModifyLaunchConfigurationAttributes resp is nil")
		return fmt.Errorf("ModifyLaunchConfigurationAttributes resp is nil")
	}
	blog.Infof("ModifyLaunchConfigurationAttributes success, requestID: %s", resp.Response.RequestId)
	return nil
}

// UpgradeLaunchConfiguration 升级启动配置
// https://cloud.tencent.com/document/api/377/35199
func (c *ASClient) UpgradeLaunchConfiguration(req *as.UpgradeLaunchConfigurationRequest) error {
	blog.Infof("UpgradeLaunchConfiguration input: %v", utils.ToJSONString(req))
	if *req.InternetAccessible.InternetChargeType == InternetChargeTypeBandwidthPrepaid {
		req.InternetAccessible.InternetChargeType = common.StringPtr(InternetChargeTypeBandwidthPostpaidByHour)
	}
	resp, err := c.as.UpgradeLaunchConfiguration(req)
	if err != nil {
		blog.Errorf("UpgradeLaunchConfiguration failed, err: %s", err.Error())
		return err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("UpgradeLaunchConfiguration resp is nil")
		return fmt.Errorf("UpgradeLaunchConfiguration resp is nil")
	}
	blog.Infof("UpgradeLaunchConfiguration success, requestID: %s", resp.Response.RequestId)
	return nil
}

// DescribeLaunchConfigurations describe LaunchConfigurations, when ascIDs is empty, describe all
// https://cloud.tencent.com/document/api/377/20445
func (c *ASClient) DescribeLaunchConfigurations(ascIDs []string) ([]*as.LaunchConfiguration, error) {
	blog.Infof("DescribeLaunchConfigurations input: %s", ascIDs)
	req := as.NewDescribeLaunchConfigurationsRequest()
	req.Limit = common.Uint64Ptr(limit)
	if len(ascIDs) > 0 {
		req.LaunchConfigurationIds = common.StringPtrs(ascIDs)
	}

	got, total := 0, 0
	first := true
	ins := make([]*as.LaunchConfiguration, 0)
	for got < total || first {
		first = false
		req.Offset = common.Uint64Ptr(uint64(got))
		resp, err := c.as.DescribeLaunchConfigurations(req)
		if err != nil {
			blog.Errorf("DescribeLaunchConfigurations failed, err: %s", err.Error())
			return nil, err
		}
		if resp == nil || resp.Response == nil {
			blog.Errorf("DescribeLaunchConfigurations resp is nil")
			return nil, fmt.Errorf("DescribeLaunchConfigurations resp is nil")
		}
		blog.Infof("DescribeLaunchConfigurations success, requestID: %s", resp.Response.RequestId)
		ins = append(ins, resp.Response.LaunchConfigurationSet...)
		got += len(resp.Response.LaunchConfigurationSet)
		total = int(*resp.Response.TotalCount)
	}
	return ins, nil
}

// DescribeAutoScalingGroups 查询ASG信息
func (c *ASClient) DescribeAutoScalingGroups(asgID string) (*as.AutoScalingGroup, error) {
	blog.Infof("DescribeAutoScalingGroups input: %s", asgID)
	req := as.NewDescribeAutoScalingGroupsRequest()
	if asgID != "" {
		req.AutoScalingGroupIds = append(req.AutoScalingGroupIds, common.StringPtr(asgID))
	}
	resp, err := c.as.DescribeAutoScalingGroups(req)
	if err != nil {
		blog.Errorf("DescribeAutoScalingGroups failed, err: %s", err.Error())
		return nil, err
	}
	if resp == nil || resp.Response == nil || *resp.Response.TotalCount != 1 {
		blog.Errorf("DescribeAutoScalingGroups resp is nil")
		return nil, fmt.Errorf("DescribeAutoScalingGroups resp is nil")
	}
	blog.Infof("DescribeAutoScalingGroups success, requestID: %s", resp.Response.RequestId)

	return resp.Response.AutoScalingGroupSet[0], nil
}

// ScaleOutInstances 指定数量扩容实例
func (c *ASClient) ScaleOutInstances(asgID string, scaleOutNum uint64) (string, error) {
	blog.Infof("ScaleOutInstances input: asg %s; scaleOut %v", asgID, scaleOutNum)

	req := as.NewScaleOutInstancesRequest()
	req.AutoScalingGroupId = common.StringPtr(asgID)
	req.ScaleOutNumber = common.Uint64Ptr(scaleOutNum)

	resp, err := c.as.ScaleOutInstances(req)
	if err != nil {
		blog.Errorf("ScaleOutInstances failed, err: %s", err.Error())
		return "", err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("ScaleOutInstances resp is nil")
		return "", fmt.Errorf("ScaleOutInstances resp is nil")
	}
	blog.Infof("ScaleOutInstances success, requestID: %s", resp.Response.RequestId)

	return *resp.Response.ActivityId, nil
}

// ScaleInInstances 指定数量缩容实例, 返回伸缩活动ID
func (c *ASClient) ScaleInInstances(asgID string, scaleInNum uint64) (string, error) {
	blog.Infof("ScaleInInstances input: asg %s; scaleIn %v", asgID, scaleInNum)

	req := as.NewScaleInInstancesRequest()
	req.AutoScalingGroupId = common.StringPtr(asgID)
	req.ScaleInNumber = common.Uint64Ptr(scaleInNum)

	resp, err := c.as.ScaleInInstances(req)
	if err != nil {
		blog.Errorf("ScaleInInstances failed, err: %s", err.Error())
		return "", err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("ScaleInInstances resp is nil")
		return "", fmt.Errorf("ScaleInInstances resp is nil")
	}
	blog.Infof("ScaleInInstances success, requestID: %s", resp.Response.RequestId)

	return *resp.Response.ActivityId, nil
}

// DescribeAutoScalingActivities 查询伸缩组的伸缩活动记录
func (c *ASClient) DescribeAutoScalingActivities(activityID string) (*as.Activity, error) {
	blog.Infof("DescribeAutoScalingActivities input: activityID %s; scaleIn %v", activityID)

	req := as.NewDescribeAutoScalingActivitiesRequest()
	req.ActivityIds = append(req.ActivityIds, common.StringPtr(activityID))

	resp, err := c.as.DescribeAutoScalingActivities(req)
	if err != nil {
		blog.Errorf("DescribeAutoScalingActivities failed, err: %s", err.Error())
		return nil, err
	}
	if resp == nil || resp.Response == nil || *resp.Response.TotalCount != 1 {
		blog.Errorf("DescribeAutoScalingActivities resp is nil")
		return nil, fmt.Errorf("DescribeAutoScalingActivities resp is nil")
	}
	blog.Infof("DescribeAutoScalingActivities success, requestID: %s", resp.Response.RequestId)

	return resp.Response.ActivitySet[0], nil
}
