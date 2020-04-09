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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"bk-bcs/bcs-common/common/blog"
)

func (c *Client) queryInstanceInfo() error {
	req := &ec2.DescribeInstancesInput{}
	req.SetFilters([]*ec2.Filter{
		{
			Name: aws.String("private-ip-address"),
			Values: []*string{
				aws.String(c.InstanceIP),
			},
		},
	})

	blog.V(2).Infof("aws DescribeInstances request %s", req.String())

	resp, err := c.ec2client.DescribeInstances(req)
	if err != nil {
		return fmt.Errorf("describe instance failed, err %s", err.Error())
	}

	blog.V(2).Infof("aws DescribeInstances response %s", resp.String())

	if len(resp.Reservations) == 0 {
		return fmt.Errorf("no reservations info in DescribeInstances response")
	}

	reservation := resp.Reservations[0]
	if len(reservation.Instances) == 0 {
		return fmt.Errorf("no instance in reservation %s", aws.StringValue(reservation.ReservationId))
	}

	c.instance = reservation.Instances[0]
	return nil
}

func (c *Client) getAvailableSubnet(ipNum int) (string, error) {

	if len(c.SubnetIDs) == 0 {
		return aws.StringValue(c.instance.SubnetId), nil
	}

	req := &ec2.DescribeSubnetsInput{}
	req.SetSubnetIds(aws.StringSlice(c.SubnetIDs))

	blog.V(2).Infof("aws DescribeSubnets request %s", req.String())

	resp, err := c.ec2client.DescribeSubnets(req)
	if err != nil {
		return "", fmt.Errorf("describe subnets failed, err %s", err.Error())
	}

	blog.V(2).Infof("aws DescribeSubnets response %s", resp.String())

	if len(resp.Subnets) == 0 {
		return "", fmt.Errorf("no subnets in DescribeSubnets response")
	}
	for _, subnet := range resp.Subnets {
		if aws.Int64Value(subnet.AvailableIpAddressCount) > int64(ipNum+1) {
			return aws.StringValue(subnet.SubnetId), nil
		}
	}
	return "", fmt.Errorf("no found available subnet")
}
