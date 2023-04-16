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
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	limit = 100
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

	awsConf := &aws.Config{}
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

// GetNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodesByIP list node by IP set
func (nm *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
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

	instances, err := client.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice(ids)})
	if err != nil {
		blog.Errorf("ec2 client DescribeInstances len(%d) ip address failed, %s", len(ids), err.Error())
		return nil, err
	}
	blog.Infof("ec2 client DescribeInstances len(%d) ip response num %d", len(ids), len(instances))

	if len(instances) == 0 {
		// * no data response
		return nil, nil
	}
	if len(instances) != len(ids) {
		blog.Warnf("ec2 client DescribeInstances, expect %d, but got %d")
	}
	zoneInfo, err := client.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{AllAvailabilityZones: aws.Bool(true)})
	if err != nil {
		blog.Errorf("ec2 client DescribeAvailabilityZones failed: %v", err)
	}
	zoneMap := make(map[string]string)
	for _, z := range zoneInfo {
		zoneMap[*z.ZoneName] = *z.ZoneId
	}

	nodeMap := make(map[string]*proto.Node)
	var nodes []*proto.Node
	for _, inst := range instances {
		node := InstanceToNode(inst, zoneMap)
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

// InstanceToNode parse Instance information in qcloud to Node in clustermanager
// @param Instance: qcloud instance information, can not be nil;
// @return Node: cluster-manager node information;
func InstanceToNode(inst *ec2.Instance, zoneInfo map[string]string) *proto.Node {
	var zoneID int
	if zoneInfo != nil {
		zoneID, _ = strconv.Atoi(zoneInfo[*inst.Placement.AvailabilityZone])
	}
	node := &proto.Node{
		NodeID:       *inst.InstanceId,
		InstanceType: *inst.InstanceType,
		CPU:          uint32(*inst.CpuOptions.CoreCount),
		GPU:          0,
		VPC:          *inst.VpcId,
		ZoneID:       *inst.Placement.AvailabilityZone,
		Zone:         uint32(zoneID),
	}
	return node
}

// GetCVMImageIDByImageName get imageID by imageName
func (nm *NodeManager) GetCVMImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	return "", cloudprovider.ErrCloudNotImplemented
}

// GetCloudRegions get cloud regions
func (nm *NodeManager) GetCloudRegions(opt *cloudprovider.CommonOption) ([]*proto.RegionInfo, error) {
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
			Region:      *v.Endpoint,
			RegionName:  *v.RegionName,
			RegionState: *v.OptInStatus,
		})
	}

	return regions, nil
}

// GetZoneList get zoneList by region
func (nm *NodeManager) GetZoneList(opt *cloudprovider.CommonOption) ([]*proto.ZoneInfo, error) {
	client, err := NewEC2Client(opt)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
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

// ListNodeInstanceType get node instance type list
func (nm *NodeManager) ListNodeInstanceType(zone, nodeFamily string, cpu, memory uint32, opt *cloudprovider.CommonOption) ([]*proto.InstanceType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListOsImage get osimage list
func (nm *NodeManager) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
