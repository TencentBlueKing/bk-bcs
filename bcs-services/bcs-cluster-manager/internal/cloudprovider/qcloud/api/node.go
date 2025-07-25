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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetCVMClient get cvm client from common option
func GetCVMClient(opt *cloudprovider.CommonOption) (*NodeClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}
	credential := common.NewCredential(opt.Account.SecretID, opt.Account.SecretKey)

	cpf := profile.NewClientProfile()
	if opt.CommonConf.CloudInternalEnable {
		cpf.HttpProfile.Endpoint = opt.CommonConf.MachineDomain
	}

	cli, err := cvm.NewClient(credential, opt.Region, cpf)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}

	return &NodeClient{client: cli}, nil
}

// NodeClient CVM relative API management
type NodeClient struct {
	client *cvm.Client
}

// DescribeZones get cloud zoneList
func (nc *NodeClient) DescribeZones() ([]*cvm.ZoneInfo, error) {
	// DescribeZones
	req := cvm.NewDescribeZonesRequest()
	resp, err := nc.client.DescribeZones(req)
	if err != nil {
		blog.Errorf("cvm client GetZoneList failed, %s", err.Error())
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client GetZoneList but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] cvm client GetZoneList response num %d",
		response.RequestId, *response.TotalCount)

	if *response.TotalCount == 0 || len(response.ZoneSet) == 0 {
		// * no data response
		return nil, nil
	}

	return response.ZoneSet, nil
}

// GetCloudRegions get regionInfo
func (nc *NodeClient) GetCloudRegions() ([]*cvm.RegionInfo, error) {
	// DescribeRegions
	req := cvm.NewDescribeRegionsRequest()
	resp, err := nc.client.DescribeRegions(req)
	if err != nil {
		blog.Errorf("cvm client DescribeRegions failed, %s", err.Error())
		return nil, err
	}

	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeRegions but lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] cvm client DescribeRegions response num %d",
		response.RequestId, *response.TotalCount)

	if *response.TotalCount == 0 || len(response.RegionSet) == 0 {
		// * no data response
		return nil, nil
	}

	return response.RegionSet, nil
}

// GetNodeInstanceByIP get specified Node by innerIP address
func (nc *NodeClient) GetNodeInstanceByIP(ip string) (*cvm.Instance, error) {
	req := cvm.NewDescribeInstancesRequest()

	var ips []*string
	ips = append(ips, common.StringPtr(ip))
	req.Filters = append(req.Filters, &cvm.Filter{
		Name:   common.StringPtr("private-ip-address"),
		Values: ips,
	})

	// DescribeInstances
	resp, err := nc.client.DescribeInstances(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstance %s failed, %s", ip, err.Error())
		return nil, err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeInstance %s but lost response information", ip)
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] cvm client DescribeInstance %s response num %d",
		response.RequestId, ip, *response.TotalCount,
	)
	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		// * no data response
		return nil, cloudprovider.ErrCloudNoHost
	}

	return response.InstanceSet[0], nil
}

// GetImageByImageID xxx
func (nc *NodeClient) GetImageByImageID(imageID string) (*cvm.Image, error) {
	req := cvm.NewDescribeImagesRequest()
	req.ImageIds = append(req.ImageIds, common.StringPtr(imageID))

	// DescribeImages
	resp, err := nc.client.DescribeImages(req)
	if err != nil {
		blog.Errorf("cvm client DescribeImages %s failed, %s", imageID, err.Error())
		return nil, err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeImages %s but lost response information", imageID)
		return nil, cloudprovider.ErrCloudLostResponse
	}

	if len(response.ImageSet) == 0 {
		blog.Errorf("cvm client DescribeImages %s failed", imageID)
		return nil, fmt.Errorf("not found image[%s]", imageID)
	}

	return response.ImageSet[0], nil
}

// ListImages get region all images
func (nc *NodeClient) ListImages() ([]*cvm.Image, error) {
	var (
		initOffset   uint64
		imageListLen = 100

		imageList = make([]*cvm.Image, 0)
	)

	for {
		if imageListLen != 100 {
			break
		}
		req := cvm.NewDescribeImagesRequest()
		req.Offset = common.Uint64Ptr(initOffset)
		req.Limit = common.Uint64Ptr(uint64(100))

		resp, err := nc.client.DescribeImages(req)
		if err != nil {
			blog.Errorf("cvm client DescribeImages failed, %s", err.Error())
			return nil, err
		}
		// check response
		response := resp.Response
		if response == nil {
			blog.Errorf("cvm client ListImages DescribeImages but lost response information")
			return nil, cloudprovider.ErrCloudLostResponse
		}

		imageList = append(imageList, response.ImageSet...)

		imageListLen = len(response.ImageSet)
		initOffset += 100
	}

	blog.Infof("ListImages get region all images successful")

	return imageList, nil
}

// GetInstancesByID get instances list by ids
func (nc *NodeClient) GetInstancesByID(ids []string) ([]*cvm.Instance, error) {
	req := cvm.NewDescribeInstancesRequest()
	req.Limit = common.Int64Ptr(limit)

	var idList []*string
	for _, id := range ids {
		idList = append(idList, common.StringPtr(id))
	}
	// instanceIDs max 100
	if len(idList) > 0 {
		req.InstanceIds = append(req.InstanceIds, idList...)
	}

	// DescribeInstances
	resp, err := nc.client.DescribeInstances(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip address failed, %s", len(ids), err.Error())
		return nil, err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip but lost response information", len(ids))
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] cvm client DescribeInstance len(%d) ip response num %d",
		*response.RequestId, len(ids), *response.TotalCount)

	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		// * no data response
		return nil, nil
	}
	if len(response.InstanceSet) != len(ids) {
		blog.Warnf("RequestId[%s] DescribeInstance, expect %d, but got %d", *response.RequestId,
			len(ids), len(response.InstanceSet))
	}

	return response.InstanceSet, nil
}

// GetInstancesByIP trans IPList to cloud Instance, filter max 5 values
func (nc *NodeClient) GetInstancesByIP(ips []string) ([]*cvm.Instance, error) {
	req := cvm.NewDescribeInstancesRequest()
	req.Limit = common.Int64Ptr(limit)

	var ipList []*string
	for _, ip := range ips {
		ipList = append(ipList, common.StringPtr(ip))
	}

	// filters values max 5
	if len(ipList) > 0 {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name:   common.StringPtr("private-ip-address"),
			Values: ipList,
		})
	}

	// DescribeInstances
	resp, err := nc.client.DescribeInstances(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip address failed, %s", len(ips), err.Error())
		return nil, err
	}
	// check response
	response := resp.Response
	if response == nil {
		blog.Errorf("cvm client DescribeInstance len(%d) ip but lost response information", len(ips))
		return nil, cloudprovider.ErrCloudLostResponse
	}
	// check response data
	blog.Infof("RequestId[%s] cvm client DescribeInstance len(%d) ip response num %d",
		*response.RequestId, len(ips), *response.TotalCount)

	if *response.TotalCount == 0 || len(response.InstanceSet) == 0 {
		// * no data response
		return nil, nil
	}

	if len(response.InstanceSet) != len(ips) {
		blog.Warnf("RequestId[%s] DescribeInstance, expect %d, but got %d", *response.RequestId,
			len(ips), len(response.InstanceSet))
	}

	return response.InstanceSet, nil
}

// DescribeInstanceTypeConfigs describe instance type configs (https://cloud.tencent.com/document/product/213/15749)
func (nc *NodeClient) DescribeInstanceTypeConfigs(filters []*Filter) ([]*cvm.InstanceTypeConfig, error) {
	blog.Infof("DescribeInstanceTypeConfigs input: %s", utils.ToJSONString(filters))

	req := cvm.NewDescribeInstanceTypeConfigsRequest()
	for _, v := range filters {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr(v.Name), Values: common.StringPtrs(v.Values)})
	}
	resp, err := nc.client.DescribeInstanceTypeConfigs(req)
	if err != nil {
		blog.Errorf("cvm client DescribeInstanceTypeConfigs failed: %v", err)
		return nil, err
	}

	if resp == nil || resp.Response == nil {
		blog.Errorf("cvm client DescribeInstanceTypeConfigs lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}

	blog.Infof("DescribeInstanceTypeConfigs success, result: %s",
		utils.ToJSONString(resp.Response.InstanceTypeConfigSet))

	return resp.Response.InstanceTypeConfigSet, nil
}

// DescribeZoneInstanceConfigInfos zone instance config infos(https://cloud.tencent.com/document/api/213/17378)
func (nc *NodeClient) DescribeZoneInstanceConfigInfos(zone, instanceFamily, instanceType string) (
	[]*cvm.InstanceTypeQuotaItem, error) {

	blog.Infof("DescribeZoneInstanceConfigInfos input: zone/%s, instanceFamily/%s, instanceType/%s", zone,
		instanceFamily, instanceType)

	req := cvm.NewDescribeZoneInstanceConfigInfosRequest()
	req.Filters = make([]*cvm.Filter, 0)
	// 按量计费
	req.Filters = append(req.Filters, &cvm.Filter{
		Name: common.StringPtr("instance-charge-type"), Values: common.StringPtrs([]string{"POSTPAID_BY_HOUR"})})
	if len(zone) > 0 {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr("zone"), Values: common.StringPtrs([]string{zone})})
	}
	if len(instanceFamily) > 0 {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr("instance-family"), Values: common.StringPtrs([]string{instanceFamily})})
	}
	if len(instanceType) > 0 {
		req.Filters = append(req.Filters, &cvm.Filter{
			Name: common.StringPtr("instance-type"), Values: common.StringPtrs([]string{instanceType})})
	}
	resp, err := nc.client.DescribeZoneInstanceConfigInfos(req)
	if err != nil {
		blog.Errorf("cvm client DescribeZoneInstanceConfigInfos failed: %v", err)
		return nil, err
	}
	if resp == nil || resp.Response == nil {
		blog.Errorf("cvm client DescribeZoneInstanceConfigInfos lost response information")
		return nil, cloudprovider.ErrCloudLostResponse
	}

	blog.Infof("DescribeZoneInstanceConfigInfos success, result: %s",
		utils.ToJSONString(resp.Response.InstanceTypeQuotaSet))
	return resp.Response.InstanceTypeQuotaSet, nil
}

// DescribeImages describe images: PRIVATE_IMAGE: 私有镜像; PUBLIC_IMAGE: 公共镜像 (腾讯云官方镜像)
// https://cloud.tencent.com/document/api/213/15715
func (nc *NodeClient) DescribeImages(imageType string) ([]*cvm.Image, error) {
	blog.Infof("DescribeImages input: %s", imageType)

	req := cvm.NewDescribeImagesRequest()
	if imageType != "" {
		req.Filters = []*cvm.Filter{
			{
				Name:   common.StringPtr("image-type"),
				Values: common.StringPtrs([]string{imageType}),
			},
		}
	}

	var (
		initOffset   uint64
		imageListLen = 100

		images = make([]*cvm.Image, 0)
	)

	for {
		if imageListLen != 100 {
			break
		}
		req.Offset = common.Uint64Ptr(initOffset)
		req.Limit = common.Uint64Ptr(uint64(100))

		resp, err := nc.client.DescribeImages(req)
		if err != nil {
			blog.Errorf("cvm client DescribeImages failed, %s", err.Error())
			return nil, err
		}
		// check response
		response := resp.Response
		if response == nil {
			blog.Errorf("cvm client ListImages DescribeImages but lost response information")
			return nil, cloudprovider.ErrCloudLostResponse
		}

		images = append(images, response.ImageSet...)

		imageListLen = len(response.ImageSet)
		initOffset += 100
	}

	blog.Infof("nodeClient DescribeImages successful")
	return images, nil
}

// DescribeKeyPairsByID describe ssh keyPairs https://cloud.tencent.com/document/product/213/15699 (max 100)
func (nc *NodeClient) DescribeKeyPairsByID(keyIDs []string) ([]*cvm.KeyPair, error) {
	req := cvm.NewDescribeKeyPairsRequest()
	if len(keyIDs) > 0 {
		req.KeyIds = common.StringPtrs(keyIDs)
	}
	req.Limit = common.Int64Ptr(limit)

	resp, err := nc.client.DescribeKeyPairs(req)
	if err != nil {
		blog.Errorf("DescribeKeyPairs[%v] failed: %v", keyIDs, err)
		return nil, err
	}
	if len(resp.Response.KeyPairSet) == 0 {
		return nil, nil
	}

	return resp.Response.KeyPairSet, nil
}

func getCvmImagesByImageType(provider string, opt *cloudprovider.CommonOption) ([]*cvm.Image, error) {
	cli, err := GetCVMClient(opt)
	if err != nil {
		return nil, fmt.Errorf("DescribeOsImages[%s] GetCVMClient failed: %v", provider, err)
	}

	cvmImages, err := cli.DescribeImages(provider)
	if err != nil {
		return nil, fmt.Errorf("DescribeOsImages[%s] DescribeImages failed: %v", provider, err)
	}

	return cvmImages, nil
}

func getCvmImageByImageName(imageName string, opt *cloudprovider.CommonOption) (*cvm.Image, error) {
	cli, err := GetCVMClient(opt)
	if err != nil {
		return nil, fmt.Errorf("getCvmImageByImageName[%s] GetCVMClient failed: %v", imageName, err)
	}

	image, exist := getImageNameCacheData(opt.Region, imageName)
	if exist {
		return image, nil
	}

	imageID, err := cli.GetImageIDByImageName(imageName, opt)
	if err != nil {
		return nil, fmt.Errorf("getCvmImageByImageName[%s] GetImageIDByImageName failed: %v", imageName, err)
	}

	cvmImage, err := cli.GetImageByImageID(imageID)
	if err != nil {
		return nil, fmt.Errorf("getCvmImageByImageName[%s] GetImageByImageID failed: %v", imageID, err)
	}

	err = setImageNameCacheData(opt.Region, imageName, cvmImage)
	if err != nil {
		blog.Errorf("getCvmImageByImageName[%s] setImageNameCacheData failed: %v", imageName, err)
	}

	return cvmImage, nil
}

// GetImageIDByImageName get imageID by imageName
func (nc *NodeClient) GetImageIDByImageName(imageName string, opt *cloudprovider.CommonOption) (string, error) {
	cloudImages, err := nc.ListImages()
	if err != nil {
		blog.Errorf("getCVMImageIDByImageName cvm ListImages %s failed, %s", imageName, err.Error())
		return "", err
	}
	var (
		imageIDList = make([]string, 0)
	)
	for _, image := range cloudImages {
		if *image.ImageName == imageName {
			imageIDList = append(imageIDList, *image.ImageId)
		}
	}
	blog.Infof("GetImageIDByImageName successful %v", imageIDList)

	if len(imageIDList) == 0 {
		return "", fmt.Errorf("GetImageIDByImageName[%s] failed: imageIDList empty", imageName)
	}

	return imageIDList[0], nil
}

// ListKeyPairs describe all ssh keyPairs https://cloud.tencent.com/document/product/213/15699
func (nc *NodeClient) ListKeyPairs() ([]*cvm.KeyPair, error) {
	var (
		keyPairs = make([]*cvm.KeyPair, 0)

		initOffset int64
		keyListLen = limit
	)

	for {
		if keyListLen != limit {
			break
		}
		req := cvm.NewDescribeKeyPairsRequest()
		req.Offset = common.Int64Ptr(initOffset)
		req.Limit = common.Int64Ptr(limit)

		resp, err := nc.client.DescribeKeyPairs(req)
		if err != nil {
			blog.Errorf("cvm client DescribeKeyPairs failed, %s", err.Error())
			continue
		}

		// check response
		response := resp.Response
		if response == nil {
			blog.Errorf("cvm client DescribeKeyPairs but lost response information")
			continue
		}
		keyPairs = append(keyPairs, response.KeyPairSet...)

		keyListLen = len(response.KeyPairSet)
		initOffset += limit
	}

	blog.Infof("ListKeyPairs successful")

	return keyPairs, nil
}

// ModifyInstancesVpcAttribute 修改实例vpc属性(vpc必须存在对应可用区的子网)
func (nc *NodeClient) ModifyInstancesVpcAttribute(vpcID string, subnet string, instanceIds []string) error {
	req := cvm.NewModifyInstancesVpcAttributeRequest()
	if len(instanceIds) > 0 {
		req.InstanceIds = common.StringPtrs(instanceIds)
	}
	req.ReserveHostName = common.BoolPtr(true)
	req.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
		VpcId:    common.StringPtr(vpcID),
		SubnetId: common.StringPtr(subnet),
	}

	resp, err := nc.client.ModifyInstancesVpcAttribute(req)
	if err != nil {
		blog.Errorf("ModifyInstancesVpcAttribute[%+v] failed: %v", instanceIds, err)
		return err
	}

	if resp.Response == nil {
		return fmt.Errorf("ModifyInstancesVpcAttribute[%+v] lost validate response", instanceIds)
	}

	blog.Infof("ModifyInstancesVpcAttribute[%+v] vpc[%s] subnet[%s] successful", instanceIds, vpcID, subnet)
	return nil
}
