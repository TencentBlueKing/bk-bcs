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

	"github.com/aws/aws-sdk-go/service/autoscaling"
)

// AutoScalingClient aws auto scaling client
type AutoScalingClient struct {
	asClient *autoscaling.AutoScaling
}

// NewAutoScalingClient init autoscaling client
func NewAutoScalingClient(opt *cloudprovider.CommonOption) (*AutoScalingClient, error) {
	sess, err := NewSession(opt)
	if err != nil {
		return nil, err
	}

	return &AutoScalingClient{
		asClient: autoscaling.New(sess),
	}, nil
}

// DescribeAutoScalingGroups describes AutoScalingGroups
func (as *AutoScalingClient) DescribeAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) ([]*autoscaling.Group, error) {
	blog.Infof("DescribeAutoScalingGroups input: %", utils.ToJSONString(input))
	output, err := as.asClient.DescribeAutoScalingGroups(input)
	if err != nil {
		blog.Errorf("DescribeAutoScalingGroups failed: %v", err)
		return nil, err
	}
	if output == nil || output.AutoScalingGroups == nil {
		blog.Errorf("DescribeAutoScalingGroups lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("DescribeAutoScalingGroups %s successful: %", utils.ToJSONString(input))

	return output.AutoScalingGroups, nil
}

// SetDesiredCapacity describes AutoScalingGroups
func (as *AutoScalingClient) SetDesiredCapacity(asgName string, capacity int64) error {
	blog.Infof("SetDesiredCapacity set autoScalingGroup[%s] capacity to %d", asgName, capacity)
	_, err := as.asClient.SetDesiredCapacity(
		&autoscaling.SetDesiredCapacityInput{
			AutoScalingGroupName: &asgName,
			DesiredCapacity:      &capacity,
		},
	)
	if err != nil {
		blog.Errorf("SetDesiredCapacity failed: %v", err)
		return err
	}
	blog.Infof("SetDesiredCapacity for %s successful, capacity %d", asgName, capacity)

	return nil
}

// TerminateInstanceInAutoScalingGroup terminates instance in AutoScalingGroups
func (as *AutoScalingClient) TerminateInstanceInAutoScalingGroup(
	input *autoscaling.TerminateInstanceInAutoScalingGroupInput) (*autoscaling.Activity, error) {
	blog.Infof("TerminateInstanceInAutoScalingGroup input: %", utils.ToJSONString(input))
	output, err := as.asClient.TerminateInstanceInAutoScalingGroup(input)
	if err != nil {
		blog.Errorf("TerminateInstanceInAutoScalingGroup failed: %v", err)
		return nil, err
	}
	blog.Infof("TerminateInstanceInAutoScalingGroup instance %s successful", input.InstanceId)

	return output.Activity, nil
}

// DetachInstances detach instances in AutoScalingGroups
func (as *AutoScalingClient) DetachInstances(input *autoscaling.DetachInstancesInput) ([]*autoscaling.Activity, error) {
	blog.Infof("DetachInstances input: %", utils.ToJSONString(input))
	output, err := as.asClient.DetachInstances(input)
	if err != nil {
		blog.Errorf("DetachInstances failed: %v", err)
		return nil, err
	}
	blog.Infof("DetachInstances instances %v for group %s successful", input.InstanceIds, input.AutoScalingGroupName)

	return output.Activities, nil
}
