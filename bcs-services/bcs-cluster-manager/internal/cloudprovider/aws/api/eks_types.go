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

// LaunchTemplate describes a launch template.
type LaunchTemplate struct {
	// The version number of the default version of the launch template.
	DefaultVersionNumber *int64 `locationName:"defaultVersionNumber" type:"long"`
	// The version number of the latest version of the launch template.
	LatestVersionNumber *int64 `locationName:"latestVersionNumber" type:"long"`
	// The ID of the launch template.
	LaunchTemplateId *string `locationName:"launchTemplateId" type:"string"`
	// The name of the launch template.
	LaunchTemplateName *string `locationName:"launchTemplateName" min:"3" type:"string"`
	// The tags for the launch template.
	Tags []*Tag `locationName:"tagSet" locationNameList:"item" type:"list"`
}

// CreateLaunchTemplateInput represents input when creating launch template
type CreateLaunchTemplateInput struct {
	// Unique, case-sensitive identifier you provide to ensure the idempotency of the request.
	ClientToken *string `type:"string"`
	// The information for the launch template.
	LaunchTemplateData *RequestLaunchTemplateData `type:"structure" required:"true" sensitive:"true"`
	// A name for the launch template.
	LaunchTemplateName *string `min:"3" type:"string" required:"true"`
	// The tags to apply to the launch template on creation
	TagSpecifications []*TagSpecification `locationName:"TagSpecification" locationNameList:"item" type:"list"`
	// A description for the first version of the launch template.
	VersionDescription *string `type:"string"`
}

// RequestLaunchTemplateData represents the information to include in the launch template.
type RequestLaunchTemplateData struct {
	// The block device mapping.
	BlockDeviceMappings []*LaunchTemplateBlockDeviceMappingRequest `locationName:"BlockDeviceMapping" locationNameList:"BlockDeviceMapping" type:"list"`
	// The Capacity Reservation targeting option
	CapacityReservationSpecification *LaunchTemplateCapacityReservationSpecificationRequest `type:"structure"`
	// Indicates whether the instance is optimized for Amazon EBS I/O
	EbsOptimized *bool `type:"boolean"`
	// An elastic GPU to associate with the instance.
	ElasticGpuSpecifications []*ElasticGpuSpecification `locationName:"ElasticGpuSpecification" locationNameList:"ElasticGpuSpecification" type:"list"`
	// The elastic inference accelerator for the instance.
	ElasticInferenceAccelerators []*LaunchTemplateElasticInferenceAccelerator `locationName:"ElasticInferenceAccelerator" locationNameList:"item" type:"list"`
	// Indicates whether an instance is enabled for hibernation
	HibernationOptions *LaunchTemplateHibernationOptionsRequest `type:"structure"`
	// The ID of the AMI.
	ImageId *string `type:"string"`
	// Indicates whether an instance stops or terminates when you initiate shutdown
	// from the instance (using the operating system command for system shutdown)
	InstanceInitiatedShutdownBehavior *string `type:"string" enum:"ShutdownBehavior"`
	// The instance type
	InstanceType *string `type:"string" enum:"InstanceType"`
	// The ID of the kernel.
	KernelId *string `type:"string"`
	// The name of the key pair.
	KeyName *string `type:"string"`
	// The license configurations.
	LicenseSpecifications []*LaunchTemplateLicenseConfigurationRequest `locationName:"LicenseSpecification" locationNameList:"item" type:"list"`
	// The metadata options for the instance
	MetadataOptions *LaunchTemplateInstanceMetadataOptionsRequest `type:"structure"`
	// The monitoring for the instance.
	Monitoring *LaunchTemplatesMonitoringRequest `type:"structure"`
	// One or more network interfaces
	NetworkInterfaces []*LaunchTemplateInstanceNetworkInterfaceSpecificationRequest `locationName:"NetworkInterface" locationNameList:"InstanceNetworkInterfaceSpecification" type:"list"`
	// The placement for the instance.
	Placement *LaunchTemplatePlacementRequest `type:"structure"`
	// The ID of the RAM disk
	RamDiskId *string `type:"string"`
	// One or more security group IDs
	SecurityGroupIds []*string `locationName:"SecurityGroupId" locationNameList:"SecurityGroupId" type:"list"`
	// [EC2-Classic, default VPC] One or more security group names
	SecurityGroups []*string `locationName:"SecurityGroup" locationNameList:"SecurityGroup" type:"list"`
	// The tags to apply to the resources during launch
	TagSpecifications []*LaunchTemplateTagSpecificationRequest `locationName:"TagSpecification" locationNameList:"LaunchTemplateTagSpecificationRequest" type:"list"`
	// The Base64-encoded user data to make available to the instance
	UserData *string `type:"string"`
}

// LaunchTemplatePlacementRequest describes the placement of an instance.
type LaunchTemplatePlacementRequest struct {
	// The affinity setting for an instance on a Dedicated Host.
	Affinity *string `type:"string"`
	// The Availability Zone for the instance.
	AvailabilityZone *string `type:"string"`
	// The name of the placement group for the instance.
	GroupName *string `type:"string"`
	// The ID of the Dedicated Host for the instance.
	HostId *string `type:"string"`
	// The ARN of the host resource group in which to launch the instances
	HostResourceGroupArn *string `type:"string"`
	// The number of the partition the instance should launch in. Valid only if
	// the placement group strategy is set to partition.
	PartitionNumber *int64 `type:"integer"`
	// Reserved for future use.
	SpreadDomain *string `type:"string"`
	// The tenancy of the instance (if the instance is running in a VPC)
	Tenancy *string `type:"string" enum:"Tenancy"`
}

// LaunchTemplateInstanceNetworkInterfaceSpecificationRequest represents the parameters for a network interface
type LaunchTemplateInstanceNetworkInterfaceSpecificationRequest struct {
	// Associates a Carrier IP address with eth0 for a new network interface
	AssociateCarrierIpAddress *bool `type:"boolean"`
	// Associates a public IPv4 address with eth0 for a new network interface
	AssociatePublicIpAddress *bool `type:"boolean"`
	// Indicates whether the network interface is deleted when the instance is terminated
	DeleteOnTermination *bool `type:"boolean"`
	// A description for the network interface
	Description *string `type:"string"`
	// The device index for the network interface attachment
	DeviceIndex *int64 `type:"integer"`
	// The IDs of one or more security groups
	Groups []*string `locationName:"SecurityGroupId" locationNameList:"SecurityGroupId" type:"list"`
	// The type of network interface
	InterfaceType *string `type:"string"`
	// The number of IPv6 addresses to assign to a network interface
	Ipv6AddressCount *int64 `type:"integer"`
	// One or more specific IPv6 addresses from the IPv6 CIDR block range of your
	// subnet. You can't use this option if you're specifying a number of IPv6 addresses.
	Ipv6Addresses []*InstanceIpv6AddressRequest `locationNameList:"InstanceIpv6Address" type:"list"`
	// The index of the network card
	NetworkCardIndex *int64 `type:"integer"`
	// The ID of the network interface
	NetworkInterfaceId *string `type:"string"`
	// The primary private IPv4 address of the network interface.
	PrivateIpAddress *string `type:"string"`
	// One or more private IPv4 addresses.
	PrivateIpAddresses []*PrivateIpAddressSpecification `locationNameList:"item" type:"list"`
	// The number of secondary private IPv4 addresses to assign to a network interface.
	SecondaryPrivateIpAddressCount *int64 `type:"integer"`
	// The ID of the subnet for the network interface.
	SubnetId *string `type:"string"`
}

// PrivateIpAddressSpecification describes a secondary private IPv4 address for a network interface.
type PrivateIpAddressSpecification struct {
	// Indicates whether the private IPv4 address is the primary private IPv4 address.
	// Only one IPv4 address can be designated as primary.
	Primary *bool `locationName:"primary" type:"boolean"`
	// The private IPv4 addresses.
	PrivateIpAddress *string `locationName:"privateIpAddress" type:"string"`
}

// InstanceIpv6AddressRequest describes an IPv6 address.
type InstanceIpv6AddressRequest struct {
	// The IPv6 address.
	Ipv6Address *string `type:"string"`
}

// LaunchTemplatesMonitoringRequest describes the monitoring for the instance.
type LaunchTemplatesMonitoringRequest struct {
	// Specify true to enable detailed monitoring. Otherwise, basic monitoring is enabled.
	Enabled *bool `type:"boolean"`
}

// LaunchTemplateInstanceMetadataOptionsRequest represents the metadata options for the instance
type LaunchTemplateInstanceMetadataOptionsRequest struct {
	// This parameter enables or disables the HTTP metadata endpoint on your instances.
	HttpEndpoint *string `type:"string" enum:"LaunchTemplateInstanceMetadataEndpointState"`
	// The desired HTTP PUT response hop limit for instance metadata requests.
	HttpPutResponseHopLimit *int64 `type:"integer"`
	// The state of token usage for your instance metadata requests
	HttpTokens *string `type:"string" enum:"LaunchTemplateHttpTokensState"`
}

// LaunchTemplateLicenseConfigurationRequest describes a license configuration.
type LaunchTemplateLicenseConfigurationRequest struct {
	// The Amazon Resource Name (ARN) of the license configuration
	LicenseConfigurationArn *string `type:"string"`
}

// LaunchTemplateHibernationOptionsRequest indicates whether the instance is configured for hibernation
type LaunchTemplateHibernationOptionsRequest struct {
	// If you set this parameter to true, the instance is enabled for hibernation
	Configured *bool `type:"boolean"`
}

// LaunchTemplateCapacityReservationSpecificationRequest describes an instance's Capacity Reservation targeting option
type LaunchTemplateCapacityReservationSpecificationRequest struct {
	// Indicates the instance's Capacity Reservation preferences
	CapacityReservationPreference *string `type:"string" enum:"CapacityReservationPreference"`
	// Information about the target Capacity Reservation or Capacity Reservation group
	CapacityReservationTarget *CapacityReservationTarget `type:"structure"`
}

// CapacityReservationTarget describes a target Capacity Reservation or Capacity Reservation group
type CapacityReservationTarget struct {
	// The ID of the Capacity Reservation in which to run the instance
	CapacityReservationId *string `type:"string"`
	// The ARN of the Capacity Reservation resource group in which to run the instance
	CapacityReservationResourceGroupArn *string `type:"string"`
}

// ElasticGpuSpecification indicates a specification for an Elastic Graphics accelerator.
type ElasticGpuSpecification struct {
	// The type of Elastic Graphics accelerator
	Type *string `type:"string" required:"true"`
}

// LaunchTemplateElasticInferenceAccelerator describes an elastic inference accelerator.
type LaunchTemplateElasticInferenceAccelerator struct {
	// The number of elastic inference accelerators to attach to the instance
	Count *int64 `min:"1" type:"integer"`
	// The type of elastic inference accelerator.
	Type *string `type:"string" required:"true"`
}

// LaunchTemplateBlockDeviceMappingRequest describes a block device mapping.
type LaunchTemplateBlockDeviceMappingRequest struct {
	// The device name (for example, /dev/sdh or xvdh)
	DeviceName *string `type:"string"`
	// Parameters used to automatically set up EBS volumes when the instance is launched
	Ebs *LaunchTemplateEbsBlockDeviceRequest `type:"structure"`
	// To omit the device from the block device mapping, specify an empty string
	NoDevice *string `type:"string"`
	// The virtual device name (ephemeralN)
	VirtualName *string `type:"string"`
}

// LaunchTemplateEbsBlockDeviceRequest describes the parameters for a block device for an EBS volume
type LaunchTemplateEbsBlockDeviceRequest struct {
	// Indicates whether the EBS volume is deleted on instance termination
	DeleteOnTermination *bool `type:"boolean"`
	// Indicates whether the EBS volume is encrypted
	Encrypted *bool `type:"boolean"`
	// The number of I/O operations per second (IOPS)
	Iops *int64 `type:"integer"`
	// The ARN of the symmetric AWS Key Management Service (AWS KMS) CMK used for encryption
	KmsKeyId *string `type:"string"`
	// The ID of the snapshot.
	SnapshotId *string `type:"string"`
	// The throughput to provision for a gp3 volume, with a maximum of 1,000 MiB/s
	Throughput *int64 `type:"integer"`
	// The size of the volume, in GiBs
	VolumeSize *int64 `type:"integer"`
	// The volume type
	VolumeType *string `type:"string" enum:"VolumeType"`
}

// TagSpecification specifies tags to apply to a resource when the resource is being created. When
// you specify a tag, you must specify the resource type to tag, otherwise the
// request will fail
type TagSpecification struct {
	// The type of resource to tag on creation.
	ResourceType *string `locationName:"resourceType" type:"string" enum:"ResourceType"`
	// The tags to apply to the resource.
	Tags []*Tag `locationName:"Tag" locationNameList:"item" type:"list"`
}

// LaunchTemplateTagSpecificationRequest the tags specification for the resources that are created during instance
// launch.
type LaunchTemplateTagSpecificationRequest struct {
	// The type of resource to tag on creation.
	ResourceType *string `locationName:"resourceType" type:"string" enum:"ResourceType"`
	// The tags to apply to the resource.
	Tags []*Tag `locationName:"Tag" locationNameList:"item" type:"list"`
}

// Tag describes a tag.
type Tag struct {
	// The key of the tag.
	Key *string `locationName:"key" type:"string"`
	// The value of the tag
	Value *string `locationName:"value" type:"string"`
}

// CreateNodegroupInput for create node group
type CreateNodegroupInput struct {
	// The AMI type for your node group
	AmiType *string `locationName:"amiType" type:"string" enum:"AMITypes"`
	// The capacity type for your node group.
	CapacityType *string `locationName:"capacityType" type:"string" enum:"CapacityTypes"`
	// Identifier that you provide to ensure the idempotency of the request
	ClientRequestToken *string `locationName:"clientRequestToken" type:"string" idempotencyToken:"true"`
	// The name of the cluster to create the node group in.
	ClusterName *string `location:"uri" locationName:"name" type:"string" required:"true"`
	// The root device disk size (in GiB) for your node group instances
	DiskSize *int64 `locationName:"diskSize" type:"integer"`
	// Image ID to create instances
	ImageID *string `locationName:"imageID" type:"string"`
	// Specify the instance types for a node group
	InstanceTypes []*string `locationName:"instanceTypes" type:"list"`
	// The Kubernetes labels to be applied to the nodes in the node group
	Labels map[string]*string `locationName:"labels" type:"map"`
	// An object representing a node group's launch template specification. If specified,
	// then do not specify instanceTypes, diskSize, or remoteAccess and make sure
	// that the launch template meets the requirements in launchTemplateSpecification.
	LaunchTemplate *LaunchTemplateSpecification `locationName:"launchTemplate" type:"structure"`
	// The Amazon Resource Name (ARN) of the IAM role to associate with your node group
	NodeRole *string `locationName:"nodeRole" type:"string" required:"true"`
	// The unique name to give your node group
	NodegroupName *string `locationName:"nodegroupName" type:"string" required:"true"`
	// The AMI version of the Amazon EKS optimized AMI to use with your node group
	ReleaseVersion *string `locationName:"releaseVersion" type:"string"`
	// The remote access configuration to use with your node group
	RemoteAccess *RemoteAccessConfig `locationName:"remoteAccess" type:"structure"`
	// The scaling configuration details for the Auto Scaling group
	ScalingConfig *NodegroupScalingConfig `locationName:"scalingConfig" type:"structure"`
	// The subnets to use for the Auto Scaling group
	Subnets []*string `locationName:"subnets" type:"list" required:"true"`
	// The metadata to apply to the node group to assist with categorization and organization
	Tags map[string]*string `locationName:"tags" min:"1" type:"map"`
	// The Kubernetes taints to be applied to the nodes in the node group
	Taints []*Taint `locationName:"taints" type:"list"`
	// The node group update configuration
	UpdateConfig *NodegroupUpdateConfig `locationName:"updateConfig" type:"structure"`
}

// LaunchTemplateSpecification represents a node group launch template specification.
// You must specify either the launch template ID or the launch template name
// in the request, but not both.
type LaunchTemplateSpecification struct {
	// The ID of the launch template
	Id *string `locationName:"id" type:"string"`
	// The name of the launch template
	Name *string `locationName:"name" type:"string"`
	// The version number of the launch template to use
	Version *string `locationName:"version" type:"string"`
}

// RemoteAccessConfig represents the remote access configuration for the managed node group
type RemoteAccessConfig struct {
	// The Amazon EC2 SSH key name that provides access for SSH communication with
	// the nodes in the managed node group
	Ec2SshKey *string `locationName:"ec2SshKey" type:"string"`
	// The security group IDs that are allowed SSH access (port 22) to the nodes
	SourceSecurityGroups []*string `locationName:"sourceSecurityGroups" type:"list"`
}

// NodegroupScalingConfig is the scaling configuration details for the Auto Scaling
// group that is associated with your node group. When creating a node group,
// you must specify all or none of the properties. When updating a node group,
// you can specify any or none of the properties.
type NodegroupScalingConfig struct {
	// The current number of nodes that the managed node group should maintain
	DesiredSize *int64 `locationName:"desiredSize" type:"integer"`
	// The maximum number of nodes that the managed node group can scale out to
	MaxSize *int64 `locationName:"maxSize" min:"1" type:"integer"`
	// The minimum number of nodes that the managed node group can scale in to
	MinSize *int64 `locationName:"minSize" type:"integer"`
}

// Taint allows a node to repel a set of pods
type Taint struct {
	// The effect of the taint
	Effect *string `locationName:"effect" type:"string" enum:"TaintEffect"`
	// The key of the taint
	Key *string `locationName:"key" min:"1" type:"string"`
	// The value of the taint
	Value *string `locationName:"value" type:"string"`
}

// NodegroupUpdateConfig is the node group update configuration.
type NodegroupUpdateConfig struct {
	// The maximum number of nodes unavailable at once during a version update
	MaxUnavailable *int64 `locationName:"maxUnavailable" min:"1" type:"integer"`
	// The maximum percentage of nodes unavailable during a version update
	MaxUnavailablePercentage *int64 `locationName:"maxUnavailablePercentage" min:"1" type:"integer"`
}
