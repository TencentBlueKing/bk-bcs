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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// EC2Client aws ec2 client
type EC2Client struct {
	ec2Client *ec2.EC2
}

// NewEC2Client init Eks client
func NewEC2Client(opt *cloudprovider.CommonOption) (*EC2Client, error) {
	sess, err := NewSession(opt)

	if err != nil {
		return nil, err
	}

	return &EC2Client{
		ec2Client: ec2.New(sess),
	}, nil
}

// DescribeAvailabilityZones describes availability zones
func (c *EC2Client) DescribeAvailabilityZones(input *ec2.DescribeAvailabilityZonesInput) ([]*ec2.AvailabilityZone, error) {
	blog.Infof("DescribeAvailabilityZones input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.DescribeAvailabilityZones(input)
	if err != nil {
		blog.Errorf("DescribeAvailabilityZones failed: %v", err)
		return nil, err
	}
	if output == nil || output.AvailabilityZones == nil {
		blog.Errorf("DescribeAvailabilityZones lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("DescribeAvailabilityZones %s successful: %s", utils.ToJSONString(input))

	return output.AvailabilityZones, nil
}

// CreateLaunchTemplate creates a LaunchTemplate
func (c *EC2Client) CreateLaunchTemplate(input *ec2.CreateLaunchTemplateInput) (*ec2.LaunchTemplate, error) {
	blog.Infof("CreateLaunchTemplate input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.CreateLaunchTemplate(input)
	if err != nil {
		blog.Errorf("CreateLaunchTemplate failed: %v", err)
		return nil, err
	}
	if output == nil || output.LaunchTemplate == nil {
		blog.Errorf("CreateLaunchTemplate created launch template but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("ec2 client CreateLaunchTemplate[%s] successful", *output.LaunchTemplate.LaunchTemplateName)

	return output.LaunchTemplate, nil
}

// DescribeLaunchTemplates describes a LaunchTemplate
func (c *EC2Client) DescribeLaunchTemplates(input *ec2.DescribeLaunchTemplatesInput) ([]*ec2.LaunchTemplate, error) {
	blog.Infof("DescribeLaunchTemplates input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.DescribeLaunchTemplates(input)
	if err != nil {
		blog.Errorf("DescribeLaunchTemplates failed: %v", err)
		return nil, err
	}
	if output == nil || output.LaunchTemplates == nil {
		blog.Errorf("DescribeLaunchTemplates lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}

	return output.LaunchTemplates, nil
}

// ModifyLaunchTemplate modify a LaunchTemplate
func (c *EC2Client) ModifyLaunchTemplate(input *ec2.ModifyLaunchTemplateInput) (*ec2.LaunchTemplate, error) {
	blog.Infof("ModifyLaunchTemplate input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.ModifyLaunchTemplate(input)
	if err != nil {
		blog.Errorf("ModifyLaunchTemplate failed: %v", err)
		return nil, err
	}
	if output == nil || output.LaunchTemplate == nil {
		blog.Errorf("ModifyLaunchTemplate lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}

	return output.LaunchTemplate, nil
}

// CreateLaunchTemplateVersion creates a versioned LaunchTemplate
func (c *EC2Client) CreateLaunchTemplateVersion(input *ec2.CreateLaunchTemplateVersionInput) (
	*ec2.LaunchTemplateVersion, error) {
	blog.Infof("CreateLaunchTemplateVersion input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.CreateLaunchTemplateVersion(input)
	if err != nil {
		blog.Errorf("CreateLaunchTemplateVersion failed: %v", err)
		return nil, err
	}
	if output == nil || output.LaunchTemplateVersion == nil {
		blog.Errorf("CreateLaunchTemplateVersion created launch template version but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("ec2 client CreateLaunchTemplateVersion[%s] version %d successful",
		*output.LaunchTemplateVersion.LaunchTemplateName, *output.LaunchTemplateVersion.VersionNumber)

	return output.LaunchTemplateVersion, nil
}

// DescribeLaunchTemplateVersions describes versioned LaunchTemplate
func (c *EC2Client) DescribeLaunchTemplateVersions(input *ec2.DescribeLaunchTemplateVersionsInput) (
	[]*ec2.LaunchTemplateVersion, error) {
	blog.Infof("DescribeLaunchTemplateVersions input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.DescribeLaunchTemplateVersions(input)
	if err != nil {
		blog.Errorf("DescribeLaunchTemplateVersions failed: %v", err)
		return nil, err
	}
	if output == nil || output.LaunchTemplateVersions == nil {
		blog.Errorf("DescribeLaunchTemplateVersions lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("DescribeLaunchTemplateVersions %s successful", *input.LaunchTemplateName)

	return output.LaunchTemplateVersions, nil
}

// DescribeImages gets image info
func (c *EC2Client) DescribeImages(input *ec2.DescribeImagesInput) ([]*ec2.Image, error) {
	blog.Infof("DescribeImages input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.DescribeImages(input)
	if err != nil {
		blog.Errorf("DescribeImages failed: %v", err)
		return nil, err
	}
	if output == nil || output.Images == nil {
		blog.Errorf("DescribeImages lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("ec2 client DescribeImages successful")

	return output.Images, nil
}

// DescribeInstances gets vm instances
func (c *EC2Client) DescribeInstances(input *ec2.DescribeInstancesInput) ([]*ec2.Instance, error) {
	blog.Infof("DescribeInstances input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.DescribeInstances(input)
	if err != nil {
		blog.Errorf("DescribeInstances failed: %v", err)
		return nil, err
	}
	if output == nil || output.Reservations == nil {
		blog.Errorf("DescribeInstances lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("ec2 client DescribeInstances successful")
	return output.Reservations[0].Instances, nil
}

// TerminateInstances terminates instances
func (c *EC2Client) TerminateInstances(input *ec2.TerminateInstancesInput) ([]*ec2.InstanceStateChange, error) {
	blog.Infof("TerminateInstances input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.TerminateInstances(input)
	if err != nil {
		blog.Errorf("TerminateInstances failed: %v", err)
		return nil, err
	}
	if output == nil || output.TerminatingInstances == nil {
		blog.Errorf("TerminateInstances lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("ec2 client TerminateInstances successful")
	return output.TerminatingInstances, nil
}

// DescribeInstanceTypes gets vm instance types
func (c *EC2Client) DescribeInstanceTypes(input *ec2.DescribeInstanceTypesInput) (
	*ec2.DescribeInstanceTypesOutput, error) {
	blog.Infof("DescribeInstanceTypes input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.DescribeInstanceTypes(input)
	if err != nil {
		blog.Errorf("DescribeInstanceTypes failed: %v", err)
		return nil, err
	}
	if output == nil || output.InstanceTypes == nil {
		blog.Errorf("DescribeInstanceTypes lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("ec2 client DescribeInstanceTypes successful")
	return output, nil
}

// DescribeKeyPairs gets key pairs
func (c *EC2Client) DescribeKeyPairs(input *ec2.DescribeKeyPairsInput) (
	[]*ec2.KeyPairInfo, error) {
	blog.Infof("DescribeKeyPairs input: %s", utils.ToJSONString(input))
	output, err := c.ec2Client.DescribeKeyPairs(input)
	if err != nil {
		blog.Errorf("DescribeKeyPairs failed: %v", err)
		return nil, err
	}
	if output == nil || output.KeyPairs == nil {
		blog.Errorf("DescribeKeyPairs lose response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	blog.Infof("ec2 client DescribeKeyPairs successful")

	return output.KeyPairs, nil
}
