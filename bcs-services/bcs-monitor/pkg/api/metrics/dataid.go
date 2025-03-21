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

package metrics

import (
	"context"

	bkmonitor_client "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// GetClusterEventDataIdReq xxx
type GetClusterEventDataIdReq struct {
	ProjectCode string `json:"projectCode" in:"path=projectCode" validate:"required"`
	ClusterId   string `json:"clusterId" in:"path=clusterId" validate:"required"`
}

// GetClusterEventDataId 获取集群事件数据ID
// @Summary 获取集群事件数据ID
// @Tags    Metrics
// @Success 200 {string} string
// @Router  /event_data_id [GET]
func GetClusterEventDataId(c context.Context, req *GetClusterEventDataIdReq) (*bkmonitor_client.ClusterDataID, error) {
	return bkmonitor_client.GetClusterEventDataID(c, config.G.BKMonitor.MetadataURL, req.ClusterId)
}
