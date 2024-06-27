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

// Package clusterMangerSample 测试
package clusterMangerSample

import (
	"context"
	"log"
	"testing"

	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/clusterManger"
)

func Test_ListCloudOsImage(t *testing.T) {
	req := &pb.ListCloudOsImageRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
		Provider:  clusterManger.PublicImage,
	}

	resp, err := service.ListCloudOsImage(context.TODO(), req)
	if err != nil {
		t.Fatalf("list cloud os image failed, err: %s", err.Error())
	}

	log.Printf("list cloud os image success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_ListCloudInstanceTypes(t *testing.T) {
	req := &pb.ListCloudInstanceTypeRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.ListCloudInstanceTypes(context.TODO(), req)
	if err != nil {
		t.Fatalf("list cloud instance types failed, err: %s", err.Error())
	}

	log.Printf("list cloud instance types success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_ListCloudSecurityGroups(t *testing.T) {
	req := &pb.ListCloudSecurityGroupsRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.ListCloudSecurityGroups(context.TODO(), req)
	if err != nil {
		t.Fatalf("list cloud security groups failed, err: %s", err.Error())
	}

	log.Printf("list cloud security groups success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_GetCloudRegions(t *testing.T) {
	req := &pb.GetCloudRegionsRequest{
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.GetCloudRegions(context.TODO(), req)
	if err != nil {
		t.Fatalf("list cloud regions failed, err: %s", err.Error())
	}

	log.Printf("list cloud regions success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_GetCloudRegionZones(t *testing.T) {
	req := &pb.GetCloudRegionZonesRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.GetCloudRegionZones(context.TODO(), req)
	if err != nil {
		t.Fatalf("list cloud region zones failed, err: %s", err.Error())
	}

	log.Printf("list cloud region zones success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_ListCloudVpcs(t *testing.T) {
	req := &pb.ListCloudVpcsRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.ListCloudVpcs(context.TODO(), req)
	if err != nil {
		t.Fatalf("list cloud vpc failed, err: %s", err.Error())
	}

	log.Printf("list cloud vpc success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_ListCloudSubnets(t *testing.T) {
	req := &pb.ListCloudSubnetsRequest{
		VpcID:     vpcID,
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.ListCloudSubnets(context.TODO(), req)
	if err != nil {
		t.Fatalf("list cloud subnets failed, err: %s", err.Error())
	}

	log.Printf("list cloud subnets success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_ListKeypairs(t *testing.T) {
	req := &pb.ListKeyPairsRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.ListKeypairs(context.TODO(), req)
	if err != nil {
		t.Fatalf("list key paris failed, err: %s", err.Error())
	}

	log.Printf("list key paris success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_GetCloudAccountType(t *testing.T) {
	req := &pb.GetCloudAccountTypeRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.GetCloudAccountType(context.TODO(), req)
	if err != nil {
		t.Fatalf("get cloud account type failed, err: %s", err.Error())
	}

	log.Printf("get cloud account type success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_GetCloudBandwidthPackages(t *testing.T) {
	req := &pb.GetCloudBandwidthPackagesRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.GetCloudBandwidthPackages(context.TODO(), req)
	if err != nil {
		t.Fatalf("get cloud bgp failed, err: %s", err.Error())
	}

	log.Printf("get cloud bgp success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_ListCloudProjects(t *testing.T) {
	req := &pb.ListCloudProjectsRequest{
		Region:    region,
		AccountID: accountID,
		CloudID:   clusterManger.TencentCloud,
	}

	resp, err := service.ListCloudProjects(context.TODO(), req)
	if err != nil {
		t.Fatalf("get failed, err: %s", err.Error())
	}

	log.Printf("get cloud projects success. resp: %s", utils.ObjToPrettyJson(resp))
}
