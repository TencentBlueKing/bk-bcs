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

package data

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr"
	clustermanager "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/clustermgr/clustermanagerv4"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/resourcemgr"
)

// Service interface
type Service interface {
	ListDevicePoolOperationData(ctx context.Context) ([]*DevicePoolOperationData, error)
}

type dataService struct {
	cmClient clustermgr.Client
	rmClient resourcemgr.Client
}

// NewDataService new service
func NewDataService(cmCli clustermgr.Client, rmCli resourcemgr.Client) Service {
	return &dataService{cmClient: cmCli, rmClient: rmCli}
}

// ListDevicePoolOperationData list device pool operation data
func (s *dataService) ListDevicePoolOperationData(ctx context.Context) ([]*DevicePoolOperationData, error) {
	ngList, err := s.cmClient.ListAllNodeGroups(ctx)
	if err != nil {
		blog.Errorf("list all nodeGroup failed:%s", err.Error())
		return nil, fmt.Errorf(err.Error())
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
	poolList, err := s.rmClient.ListDevicePool(ctx, []string{"self"})
	if err != nil {
		blog.Errorf("list all device pool failed:%s", err.Error())
		return nil, fmt.Errorf(err.Error())
	}
	result := make([]*DevicePoolOperationData, 0)
	for _, pool := range poolList {
		for _, consumer := range pool.AllowedDeviceConsumer {
			ngs := ngMap[consumer]
			if ngs == nil {
				continue
			}
			for _, ng := range ngs {
				item := &DevicePoolOperationData{
					PoolID:       *pool.Id,
					PoolName:     *pool.Name,
					BusinessID:   strconv.FormatInt(*pool.BaseConfig.BusinessID, 10),
					InstanceType: *pool.BaseConfig.InstanceType,
					BusinessName: "",
				}
				consumerData := &ConsumerData{
					ConsumerID:    ng.ConsumerID,
					ClusterID:     ng.ClusterID,
					NodeGroup:     ng.NodeGroupID,
					NodeGroupName: ng.Name,
					ProjectID:     ng.ProjectID,
					ConsumeNum:    int(ng.AutoScaling.DesiredSize),
				}
				cluster, getClusterErr := s.cmClient.GetCluster(ctx, ng.ClusterID)
				if getClusterErr == nil {
					consumerData.ClusterName = cluster.ClusterName
					consumerData.BusinessID = cluster.BusinessID
				} else {
					blog.Errorf(getClusterErr.Error())
				}
				item.ConsumeDetails = consumerData
				result = append(result, item)
			}
		}
	}
	return result, nil
}
