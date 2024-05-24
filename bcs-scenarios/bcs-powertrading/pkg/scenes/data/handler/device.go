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

// Package handler xxx
package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr"
	clustermanager "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr/clustermanagerv4"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/scenes"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/storage"
)

// DeviceDataHandler device operation data handler
type DeviceDataHandler struct {
	Opts *scenes.Options
}

const (
	// MinuteTimeFormat minute bucket time format
	MinuteTimeFormat = "2006-01-02 15:04:00"
)

// Init init handler
func (h *DeviceDataHandler) Init() {
	blog.Infof("init device data handler")
}

// Run run handler
func (h *DeviceDataHandler) Run(ctx context.Context) {
	tick := time.NewTicker(time.Second * time.Duration(h.Opts.Interval))
	for {
		select {
		case now := <-tick.C:
			// main loops
			blog.Infof("############## device data handler ticker: %s################", now.Format(time.RFC3339))
			h.controllerLoops(ctx)
		case <-ctx.Done():
			blog.Infof("device data handler is asked to exit")
			return
		}
	}
}

// NOCC:golint/funlen(设计如此)
// nolint
func (h *DeviceDataHandler) controllerLoops(ctx context.Context) {
	poolList, err := h.Opts.ResourceMgrCli.ListDevicePool(ctx, []string{"self", "cr"})
	if err != nil {
		blog.Errorf("ListDevicePool failed:%s", err.Error())
		return
	}
	ngMap, err := h.getNodegroupMap(ctx)
	if err != nil {
		return
	}
	recordTime := time.Now().Format(MinuteTimeFormat)
	for _, pool := range poolList {
		blog.Infof("begin to check pool %s", *pool.Id)
		rsp, listErr := h.Opts.ResourceMgrCli.ListDeviceByPool(ctx, 5000, []string{*pool.Id})
		if listErr != nil {
			blog.Errorf("list device error, pool:%s, err:%s", *pool.Id, listErr.Error())
			return
		}
		blog.Infof("device count:%d", len(rsp))
		for _, device := range rsp {
			deviceInfo := &storage.DeviceOperationData{
				DeviceID:     *device.Id,
				PoolID:       *device.DevicePoolID,
				PoolName:     *pool.Name,
				AssetID:      *device.Info.AssetID,
				InnerIP:      *device.Info.InnerIP,
				InstanceType: *device.Info.InstanceType,
				DeviceStatus: *device.Status,
				Source:       *pool.Type,
				Message:      "",
				CheckTime:    time.Now(),
				RecordTime:   recordTime,
			}
			if *device.Status == "CONSUMED" {
				deviceInfo.ConsumerID = *device.LastConsumerID
				deviceInfo.ConsumerID = *device.LastConsumerID
				// 获取ng信息
				if ngMap[*device.LastConsumerID] == nil {
					deviceInfo.Message = fmt.Sprintf("cannot find ng by LastConsumerID %s", *device.LastConsumerID)
					createErr := h.Opts.Storage.CreateDeviceData(ctx, deviceInfo, &storage.CreateOptions{})
					if createErr != nil {
						blog.Errorf("create device record failed, id:%s, error:%s", deviceInfo.DeviceID,
							createErr.Error())
					}
					continue
				}
				ng := ngMap[*device.LastConsumerID]
				if len(ng) != 1 {
					for _, nodegroup := range ng {
						deviceInfo.ShouldConsumedNodeGroup += nodegroup.NodeGroupID + ","
						deviceInfo.ShouldConsumedClusterID += nodegroup.ClusterID + ","
					}
				} else {
					deviceInfo.ShouldConsumedNodeGroup += ng[0].NodeGroupID
					deviceInfo.ShouldConsumedClusterID += ng[0].ClusterID
				}
				// 获取节点信息
				blog.Infof("device:%v", deviceInfo)
				cmNodeInfo, infoErr := h.Opts.ClusterMgrCli.GetNodeDetail(ctx, *device.Info.InnerIP)
				if infoErr != nil || cmNodeInfo == nil {
					blog.Errorf("get node detail from cluster manager failed, ip:%s, err:%v, info:%v",
						*device.Info.InnerIP, infoErr, cmNodeInfo)
					deviceInfo.Message = fmt.Sprintf("cannot find node detail by ip %s", *device.Info.InnerIP)
					createErr := h.Opts.Storage.CreateDeviceData(ctx, deviceInfo, &storage.CreateOptions{})
					if createErr != nil {
						blog.Errorf("create device record failed, id:%s, error:%s", deviceInfo.DeviceID,
							createErr.Error())
					}
					continue
				}
				deviceInfo.RealNodeGroup = cmNodeInfo.NodeGroupID
				deviceInfo.RealClusterID = cmNodeInfo.ClusterID
				consumeCheck(deviceInfo)
				// 检查node状态
				nodeStatus, checkNodeStatusErr := h.checkNodeStatus(ctx, cmNodeInfo, deviceInfo)
				if checkNodeStatusErr != nil {
					continue
				}
				deviceInfo.NodeStatus = nodeStatus.Status
				businessCheckErr := h.businessCheck(ctx, deviceInfo, nodeStatus)
				if businessCheckErr != nil {
					continue
				}
			} else {
				cmNodeInfo, infoErr := h.Opts.ClusterMgrCli.GetNodeDetail(ctx, *device.Info.InnerIP)
				if infoErr != nil {
					deviceInfo.Message = fmt.Sprintf("get node detail from cm failed:%s", err.Error())
					createErr := h.Opts.Storage.CreateDeviceData(ctx, deviceInfo, &storage.CreateOptions{})
					if createErr != nil {
						blog.Errorf("create device record failed, id:%s, error:%s", deviceInfo.DeviceID,
							createErr.Error())
					}
					continue
				}
				if cmNodeInfo != nil {
					deviceInfo.RealNodeGroup = cmNodeInfo.NodeGroupID
					deviceInfo.RealClusterID = cmNodeInfo.ClusterID
					consumeCheck(deviceInfo)
				} else {
					consumeCheck(deviceInfo)
					createErr := h.Opts.Storage.CreateDeviceData(ctx, deviceInfo, &storage.CreateOptions{})
					if createErr != nil {
						blog.Errorf("create device record failed, id:%s, error:%s", deviceInfo.DeviceID,
							createErr.Error())
					}
					continue
				}
				// 检查node状态
				nodeStatus, checkNodeStatusErr := h.checkNodeStatus(ctx, cmNodeInfo, deviceInfo)
				if checkNodeStatusErr != nil {
					continue
				}
				deviceInfo.NodeStatus = nodeStatus.Status
				businessCheckErr := h.businessCheck(ctx, deviceInfo, nodeStatus)
				if businessCheckErr != nil {
					continue
				}
			}
		}
	}
}

func consumeCheck(deviceInfo *storage.DeviceOperationData) {
	deviceInfo.ConsumeCheck = true
	if !strings.Contains(deviceInfo.ShouldConsumedNodeGroup, deviceInfo.RealNodeGroup) {
		deviceInfo.ConsumeCheck = false
		return
	}
	if !strings.Contains(deviceInfo.ShouldConsumedClusterID, deviceInfo.RealClusterID) {
		deviceInfo.ConsumeCheck = false
		return
	}
}

func (h *DeviceDataHandler) getNodegroupMap(ctx context.Context) (map[string][]*clustermanager.NodeGroup, error) {
	ngList, err := h.Opts.ClusterMgrCli.ListAllNodeGroups(ctx)
	if err != nil {
		blog.Errorf("ListAllNodeGroups failed:%s", err.Error())
		return nil, err
	}
	ngMap := make(map[string][]*clustermanager.NodeGroup)
	for _, ng := range ngList {
		if ngMap[ng.ConsumerID] != nil {
			ngMap[ng.ConsumerID] = append(ngMap[ng.ConsumerID], ng)
		} else {
			ngs := make([]*clustermanager.NodeGroup, 0)
			ngs = append(ngs, ng)
			ngMap[ng.ConsumerID] = ngs
		}
	}
	return ngMap, nil
}

func (h *DeviceDataHandler) checkNodeStatus(ctx context.Context, cmNodeInfo *clustermanager.Node,
	deviceInfo *storage.DeviceOperationData) (*clustermgr.Node, error) {
	nodeName := cmNodeInfo.InnerIP
	if cmNodeInfo.NodeType == "IDC" {
		nodeName = "node-" + cmNodeInfo.InnerIP
	}
	nodeStatus, nodeStatusErr := h.Opts.ClusterMgrCli.GetNode(nodeName, cmNodeInfo.ClusterID)
	if nodeStatusErr != nil || nodeStatus == nil {
		deviceInfo.Message = fmt.Sprintf("get node status failed:%v, node:%v", nodeStatusErr, nodeStatus)
		blog.Errorf("get node status failed:%v, node:%v", nodeStatusErr, nodeStatus)
		createErr := h.Opts.Storage.CreateDeviceData(ctx, deviceInfo, &storage.CreateOptions{})
		if createErr != nil {
			blog.Errorf("create device record failed, id:%s, error:%s", deviceInfo.DeviceID,
				createErr.Error())
		}
		return nil, fmt.Errorf("get node status failed:%s, node:%v", nodeStatusErr, nodeStatus)
	}
	return nodeStatus, nil
}

func (h *DeviceDataHandler) businessCheck(ctx context.Context, deviceInfo *storage.DeviceOperationData,
	nodeStatus *clustermgr.Node) error {
	deviceInfo.DeviceBusinessID = nodeStatus.Labels["bkcmdb.tencent.com/bk-biz-id"]
	if deviceInfo.DeviceBusinessID != "" {
		ccInfo, ccErr := h.Opts.BkccCli.ListHostByCC(ctx, []string{deviceInfo.InnerIP}, deviceInfo.DeviceBusinessID)
		if ccErr != nil {
			blog.Errorf("get node info from cc error:%s", ccErr.Error())
			createErr := h.Opts.Storage.CreateDeviceData(ctx, deviceInfo, &storage.CreateOptions{})
			if createErr != nil {
				blog.Errorf("create device record failed, id:%s, error:%s", deviceInfo.DeviceID,
					createErr.Error())
			}
			return ccErr
		}
		if len(ccInfo) == 1 {
			deviceInfo.RealBusinessID = deviceInfo.DeviceBusinessID
			deviceInfo.BusinessCheck = true
		} else {
			deviceInfo.BusinessCheck = false
		}
	}
	createErr := h.Opts.Storage.CreateDeviceData(ctx, deviceInfo, &storage.CreateOptions{})
	if createErr != nil {
		blog.Errorf("create device record failed, id:%s, error:%s", deviceInfo.DeviceID,
			createErr.Error())
		return createErr
	}
	return nil
}
