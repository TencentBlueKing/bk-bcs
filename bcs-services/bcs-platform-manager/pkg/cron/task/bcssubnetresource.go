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

// Package task xxx
package task

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cluproto "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/hibiken/asynq"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// HandleBcsSubnetResourceTask : handle bcs subnet resource task
func HandleBcsSubnetResourceTask(ctx context.Context, t *asynq.Task) error {
	blog.Infof("handle bcs subnet resource task: %s", t.Payload())
	clusters, err := clustermanager.ListCluster(ctx, &cluproto.ListClusterReq{})
	if err != nil {
		blog.Errorf("handle bcs subnet resource task: list cluster failed: %v", err)
		return err
	}
	for _, cluster := range clusters {
		// 查询子网使用率下限
		cloud, err := clustermanager.GetCloud(ctx, &cluproto.GetCloudRequest{
			CloudID: cluster.Provider,
		})
		if err != nil {
			blog.Errorf("handle bcs subnet resource task: get cloud failed: %v", err)
			return err
		}
		if cloud.NetworkInfo == nil || cloud.NetworkInfo.UnderlayRatio == 0 {
			continue
		}

		if cluster.NetworkSettings == nil {
			continue
		}
		if !cluster.NetworkSettings.EnableVPCCni {
			continue
		}
		var availableIPAddressCount uint64
		var totalIPAddressCount uint64
		var zone string
		for _, v := range cluster.NetworkSettings.EniSubnetIDs {
			// 查询需要扩容的集群及子网
			subnets, errr := clustermanager.ListCloudSubnets(ctx, &cluproto.ListCloudSubnetsRequest{
				CloudID:  cluster.Provider,
				Region:   cluster.Region,
				SubnetID: v,
				VpcID:    cluster.VpcID,
			})
			if errr != nil {
				blog.Errorf("handle bcs subnet resource task: list cloud subnets failed: %v", errr)
				return errr
			}
			if len(subnets.Data) == 0 {
				continue
			}
			availableIPAddressCount += subnets.Data[0].AvailableIPAddressCount
			totalIPAddressCount += subnets.Data[0].TotalIpAddressCount
			zone = subnets.Data[0].Zone
		}

		var usageRatio float64
		if totalIPAddressCount != 0 {
			usageRatio = float64(totalIPAddressCount-availableIPAddressCount) / float64(totalIPAddressCount) * 100
		}
		if usageRatio > float64(cloud.NetworkInfo.UnderlayRatio) {
			// 分配子网资源
			_, err = clustermanager.AddSubnetToCluster(ctx, &cluproto.AddSubnetToClusterReq{
				ClusterID: cluster.ClusterID,
				Subnet: &cluproto.SubnetSource{
					New: []*cluproto.NewSubnet{{
						Zone:  zone,
						IpCnt: uint32(config.G.TaskConf.AllocateSubnetIPCnt),
					}},
				},
				Operator: "",
			})
			if err != nil {
				blog.Errorf("handle bcs subnet resource task: add subnet to cluster failed: %v", err)
				return err
			}
		}

	}
	// Email delivery code ...
	return nil
}

// HandleVpcIPMonitorTask : handle vpc ip monitor task
func HandleVpcIPMonitorTask(ctx context.Context, t *asynq.Task) error {
	blog.Infof("handle vpc ip monitor task: %s", t.Payload())
	underlaySubnets, err := clustermanager.ListCloudVpc(ctx, &cluproto.ListCloudVPCRequest{
		NetworkType: "underlay",
	})
	if err != nil {
		blog.Errorf("handle vpc ip monitor task: list cloud vpc error: %s", err)
		return err
	}
	for _, subnet := range underlaySubnets.Data {
		if subnet.Underlay != nil {
			if subnet.Underlay.AvailableIPNum < uint32(config.G.TaskConf.RemainUnderlayIPNum) {
				// 发送告警邮件
				blog.Infof("handle vpc ip monitor task: vpcid: %s, "+
					"underlay ip available ip num: %d lt remain underlay ip num: %d", subnet.VpcID,
					subnet.Underlay.AvailableIPNum, config.G.TaskConf.RemainUnderlayIPNum)
			}
		}
	}
	overlaySubnets, err := clustermanager.ListCloudVpc(ctx, &cluproto.ListCloudVPCRequest{
		NetworkType: "overlay",
	})
	if err != nil {
		blog.Errorf("handle vpc ip monitor task: list cloud vpc error: %s", err)
		return err
	}
	for _, subnet := range overlaySubnets.Data {
		if subnet.Overlay != nil {
			if subnet.Overlay.AvailableIPNum < uint32(config.G.TaskConf.RemainOverlayIPNum) {
				// 发送告警邮件
				// 发送通知
				blog.Infof("handle vpc ip monitor task: vpcid: %s, "+
					"overlay ip available ip num: %d lt remain overlay ip num: %d", subnet.VpcID,
					subnet.Overlay.AvailableIPNum, config.G.TaskConf.RemainOverlayIPNum)
			}
		}
	}
	return nil
}
