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
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	defaultRegion = "ap-northeast-1"
	limit         = 100
)

var nodeMgr sync.Once

func init() {
	nodeMgr.Do(func() {
		// init Node
		cloudprovider.InitNodeManager("aws", &NodeManager{})
	})
}

// GetEc2Client get ec2 client from common option
func GetEc2Client(opt *cloudprovider.CommonOption) (*ec2.EC2, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	awsConf := &aws.Config{Region: &opt.Region}
	awsConf.Credentials = credentials.NewStaticCredentials(opt.Account.SecretID, opt.Account.SecretKey, "")

	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, err
	}

	return ec2.New(sess), nil
}

// NodeManager CVM relative API management
type NodeManager struct {
}

// ListNodeInstanceType get node instance type list
func (nm *NodeManager) ListNodeInstanceType(info cloudprovider.InstanceInfo,
	opt *cloudprovider.CommonOption) ([]*proto.InstanceType, error) {
	blog.Infof("ListNodeInstanceType region: %s, nodeFamily: %s, cpu: %d, memory: %d",
		info.Region, info.NodeFamily, info.Cpu, info.Memory)

	client, err := NewEC2Client(opt)
	if err != nil {
		blog.Errorf("ListNodeInstanceType create ec2 client failed, %s", err.Error())
		return nil, err
	}

	cloudInstanceTypes := make([]*ec2.InstanceTypeInfo, 0)

	err = client.ec2Client.DescribeInstanceTypesPages(
		&ec2.DescribeInstanceTypesInput{
			MaxResults: aws.Int64(limit),
			// 过滤支持x86的机型, 适配AL2_x86_86镜像
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("processor-info.supported-architecture"),
					Values: aws.StringSlice([]string{"x86_64"}),
				},
				{
					Name:   aws.String("supported-usage-class"),
					Values: aws.StringSlice([]string{"on-demand"}),
				},
			},
		}, func(page *ec2.DescribeInstanceTypesOutput, lastPage bool) bool {
			cloudInstanceTypes = append(cloudInstanceTypes, page.InstanceTypes...)
			return !lastPage
		})
	if err != nil {
		blog.Errorf("ListNodeInstanceType DescribeInstanceTypesPages failed, %s", err.Error())
		return nil, err
	}

	instanceTypes := convertToInstanceType(cloudInstanceTypes)

	return instanceTypes, nil
}

func convertToInstanceType(cloudInstanceTypes []*ec2.InstanceTypeInfo) []*proto.InstanceType {
	instanceTypes := make([]*proto.InstanceType, 0)
	for _, v := range cloudInstanceTypes {
		t := &proto.InstanceType{}
		if v.InstanceType != nil {
			t.TypeName = *v.InstanceType
			t.NodeType = *v.InstanceType
			family := strings.Split(*v.InstanceType, ".")
			t.NodeFamily = family[0]
			t.Status = common.InstanceSell
		}
		if v.VCpuInfo != nil && v.VCpuInfo.DefaultVCpus != nil {
			t.Cpu = uint32(*v.VCpuInfo.DefaultVCpus)
		}
		if v.MemoryInfo != nil && v.MemoryInfo.SizeInMiB != nil {
			memGb := math.Ceil(float64(*v.MemoryInfo.SizeInMiB / 1024)) // nolint
			t.Memory = uint32(memGb)
		}
		if v.GpuInfo != nil && v.GpuInfo.Gpus != nil {
			var gpuCount uint32
			for _, g := range v.GpuInfo.Gpus {
				if g.Count != nil {
					gpuCount += uint32(*g.Count)
				}
			}
			t.Gpu = gpuCount
		}
		instanceTypes = append(instanceTypes, t)
	}

	return instanceTypes
}

// GetExternalNodeByIP xxx
func (nm *NodeManager) GetExternalNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListExternalNodesByIP xxx
func (nm *NodeManager) ListExternalNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	if len(ips) == 0 {
		return nil, nil
	}

	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListKeyPairs xxx
func (nm *NodeManager) ListKeyPairs(opt *cloudprovider.ListNetworksOption) ([]*proto.KeyPair, error) {
	client, err := NewEC2Client(&opt.CommonOption)
	if err != nil {
		blog.Errorf("ListKeyPairs create ec2 client failed, %s", err.Error())
		return nil, err
	}

	cloudKeyPairs, err := client.DescribeKeyPairs(&ec2.DescribeKeyPairsInput{})
	if err != nil {
		blog.Errorf("ListKeyPairs DescribeKeyPairs failed, %s", err.Error())
		return nil, err
	}

	keyPairs := make([]*proto.KeyPair, 0)
	for _, v := range cloudKeyPairs {
		k := &proto.KeyPair{
			KeyName: *v.KeyName,
			KeyID:   *v.KeyPairId,
		}
		keyPairs = append(keyPairs, k)
	}

	return keyPairs, nil
}

// GetNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodesByIP list node by IP set
func (nm *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	if len(ips) == 0 {
		return nil, nil
	}

	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodesByInstanceID list node by instanceIDs
func (nm *NodeManager) ListNodesByInstanceID(ids []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	idChunks := utils.SplitStringsChunks(ids, limit)
	nodeList := make([]*proto.Node, 0)

	blog.Infof("ListNodesByInstanceID ipChunks %+v", idChunks)
	for _, chunk := range idChunks {
		if len(chunk) > 0 {
			nodes, err := nm.transInstanceIDsToNodes(chunk, opt)
			if err != nil {
				blog.Errorf("ListNodesByInstanceID failed: %v", err)
				return nil, err
			}
			if len(nodes) == 0 {
				continue
			}

			nodeList = append(nodeList, nodes...)
		}
	}

	return nodeList, nil
}

// transInstanceIDsToNodes trans IDList to Nodes
func (nm *NodeManager) transInstanceIDsToNodes(ids []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	client, err := NewEC2Client(opt.Common)
	if err != nil {
		blog.Errorf("create ec2 client when GetNodeByIP failed, %s", err.Error())
		return nil, err
	}

	instances, err := client.DescribeInstancesPages(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice(ids)})
	if err != nil {
		blog.Errorf("ec2 client DescribeInstances[%+v] failed, %s", len(ids), err.Error())
		return nil, err
	}
	blog.Infof("ec2 client DescribeInstances len(%d) ip response num %d", len(ids), len(instances))

	if len(instances) == 0 {
		// * no data response
		return nil, nil
	}
	if len(instances) != len(ids) {
		blog.Warnf("ec2 client DescribeInstances, expect %d, but got %d", len(ids), len(instances))
	}

	blog.Infof("transInstanceIDsToNodes instances %+v", instances)
	nodeMap := make(map[string]*proto.Node)
	var nodes []*proto.Node
	for _, inst := range instances {
		node := InstanceToNode(inst)
		// clean duplicated Node if user input multiple ip that
		// belong to one cvm instance
		if _, ok := nodeMap[node.NodeID]; ok {
			continue
		}

		nodeMap[node.NodeID] = node
		// default get first privateIP
		node.InnerIP = *inst.PrivateIpAddress
		node.Region = opt.Common.Region

		// check node vpc and cluster vpc
		if !strings.EqualFold(node.VPC, opt.ClusterVPCID) {
			return nil, fmt.Errorf(cloudprovider.ErrCloudNodeVPCDiffWithClusterResponse, node.InnerIP)
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

// InstanceToNode parse Instance information in aws to Node in clustermanager
// @param Instance: aws instance information, can not be nil;
// @return Node: cluster-manager node information;
func InstanceToNode(inst *ec2.Instance) *proto.Node {
	node := &proto.Node{
		NodeID:       *inst.InstanceId,
		NodeName:     *inst.PrivateDnsName,
		InstanceType: *inst.InstanceType,
		VPC:          *inst.VpcId,
		ZoneID:       *inst.Placement.AvailabilityZone,
	}
	return node
}

// GetCVMImageIDByImageName get imageID by imageName
func (nm *NodeManager) GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	return "", cloudprovider.ErrCloudNotImplemented
}

// GetCloudRegions get cloud regions
func (nm *NodeManager) GetCloudRegions(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
	// set default region
	opt.Region = defaultRegion

	client, err := GetEc2Client(opt)
	if err != nil {
		blog.Errorf("create ec2 client when GetRegionsInfo failed: %v", err)
		return nil, err
	}

	input := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}
	output, err := client.DescribeRegions(input)
	if err != nil {
		blog.Errorf("ec2 client DescribeRegions failed: %v", err)
		return nil, err
	}

	regions := make([]*proto.RegionInfo, 0)
	for _, v := range output.Regions {
		regions = append(regions, &proto.RegionInfo{
			Region:      aws.StringValue(v.RegionName),
			RegionName:  aws.StringValue(v.RegionName),
			RegionState: aws.StringValue(v.OptInStatus),
		})
	}

	return regions, nil
}

// GetZoneList get zoneList by region
func (nm *NodeManager) GetZoneList(opt *cloudprovider.GetZoneListOption) ([]*proto.ZoneInfo, error) {
	client, err := NewEC2Client(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create ec2 client failed, err %s", err.Error())
	}
	zones, err := client.DescribeAvailabilityZones(
		&ec2.DescribeAvailabilityZonesInput{AllAvailabilityZones: aws.Bool(true)})
	if err != nil {
		return nil, fmt.Errorf("list regions failed, err %s", err.Error())
	}
	var zonesInfo []*proto.ZoneInfo
	for _, z := range zones {
		zonesInfo = append(zonesInfo, &proto.ZoneInfo{
			ZoneID:    *z.ZoneId,
			Zone:      *z.ZoneId,
			ZoneName:  *z.ZoneName,
			ZoneState: *z.State,
		})
	}

	return zonesInfo, nil
}

// ListOsImage get osimage list
func (nm *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetResourceGroups resource groups list
func (nm *NodeManager) GetResourceGroups(opt *cloudprovider.CommonOption) ([]*proto.ResourceGroupInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

func checkRoleForPolicies(client *IAMClient, roleName, roleType string) bool {
	resp, err := client.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	})
	if err != nil {
		blog.Errorf("checkRoleForPolicies ListAttachedRolePolicies failed, %s:", err.Error())
		return false
	}

	switch roleType {
	case "nodeGroup":
		index := 0
		for _, policy := range resp {
			switch *policy.PolicyArn {
			case EKSRolePolicyWorkerNode:
				index++
			case EKSRolePolicyContainerRegistryReadOnly:
				index++
			case EKSRolePolicyCNI:
				index++
			}
		}
		return index == 3
	case "cluster":
		for _, policy := range resp {
			if *policy.PolicyArn == EksClusterRole {
				return true
			}
		}
	default:
		return false
	}

	return false
}

// GetServiceRoles service roles list
func (nm *NodeManager) GetServiceRoles(opt *cloudprovider.CommonOption, roleType string) (
	[]*proto.ServiceRoleInfo, error) {
	client, err := NewIAMClient(opt)
	if err != nil {
		return nil, fmt.Errorf("GetServiceRoles create iam client failed, %s", err.Error())
	}

	roles, err := client.ListRoles(&iam.ListRolesInput{})
	if err != nil {
		return nil, fmt.Errorf("GetServiceRoles ListRoles failed, %s", err.Error())
	}

	result := make([]*proto.ServiceRoleInfo, 0)

	for _, r := range roles {
		if checkRoleForPolicies(client, *r.RoleName, roleType) {
			result = append(result, &proto.ServiceRoleInfo{
				RoleName:    *r.RoleName,
				RoleID:      *r.RoleId,
				Arn:         *r.Arn,
				Description: *r.Description,
			})
		}
	}

	return result, nil
}

// ListRuntimeInfo get runtime info list
func (nm *NodeManager) ListRuntimeInfo(opt *cloudprovider.ListRuntimeInfoOption) (map[string][]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
