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

package bkmonitor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/alarm"
)

// shieldType 告警屏蔽类型
type shieldType string

const (
	// scope type
	scope shieldType = "scope"
	// dimension type
	dimension shieldType = "dimension"
)

const (
	shieldDesc = "bcs-cluster-manager集群管理屏蔽主机告警"
)

// ShieldHostAlarmRequest shield host alarm request
type ShieldHostAlarmRequest struct {
	Category        string          `json:"category"`
	BkBizID         uint64          `json:"bk_biz_id"`
	BeginTime       string          `json:"begin_time"`
	EndTime         string          `json:"end_time"`
	DimensionConfig DimensionConfig `json:"dimension_config"`
	Description     string          `json:"description"`
	ShieldNotice    bool            `json:"shield_notice"`
	CycleConfig     CycleConfig     `json:"cycle_config"`
}

func buildBizHostAlarmConfig(hosts *alarm.ShieldHost) (*ShieldHostAlarmRequest, error) {
	bizID, err := strconv.ParseUint(hosts.BizID, 10, 64)
	if err != nil {
		blog.Errorf("buildBizHostAlarmConfig ParseUint bizID failed: %v", err)
		return nil, err
	}

	var (
		ipInfos = make([]IPInfo, 0)
		ips     = make([]string, 0)
	)

	for i := range hosts.HostList {
		ipInfos = append(ipInfos, IPInfo{
			Ip:      hosts.HostList[i].IP,
			CloudID: hosts.HostList[i].CloudID,
		})
		ips = append(ips, hosts.HostList[i].IP)
	}
	if len(ipInfos) == 0 || len(ips) == 0 {
		blog.Errorf("buildBizHostAlarmConfig hosts empty")
		return nil, fmt.Errorf("buildBizHostAlarmConfig ipList empty")
	}

	// build dimensionConfig
	var dimensionConfig DimensionConfig
	switch hosts.ShieldType {
	case string(scope):
		dimensionConfig = DimensionConfig{
			ScopeType: "ip",
			Target:    ipInfos,
		}
	case string(dimension):
		dimensionConfig = DimensionConfig{
			DimensionConditions: []DimensionCondition{
				buildClusterIdCondition(hosts.ClusterId), buildNodeCondition(ips)},
		}
	default:
		return nil, fmt.Errorf("buildBizHostAlarmConfig shieldType invalid")
	}

	return &ShieldHostAlarmRequest{
		Category:        hosts.ShieldType,
		BkBizID:         bizID,
		BeginTime:       time.Now().Format("2006-01-02 15:04:00"),
		EndTime:         time.Now().Add(time.Minute * 30).Format("2006-01-02 15:04:00"),
		DimensionConfig: dimensionConfig,
		Description:     shieldDesc,
		ShieldNotice:    false,
		CycleConfig:     CycleConfig{Type: 1},
	}, nil
}

// CycleConfig 屏蔽周期配置
type CycleConfig struct {
	Type uint32 `json:"type"`
}

// DimensionConfig 屏蔽维度
type DimensionConfig struct {
	ScopeType           string               `json:"scope_type,omitempty"`
	Target              []IPInfo             `json:"target,omitempty"`
	DimensionConditions []DimensionCondition `json:"dimension_conditions,omitempty"`
}

// DimensionCondition 屏蔽维度条件
type DimensionCondition struct {
	Condition string   `json:"condition"`
	Key       string   `json:"key"`
	Method    string   `json:"method"`
	Value     []string `json:"value"`
	Name      string   `json:"name"`
}

func buildClusterIdCondition(clusterID string) DimensionCondition {
	return DimensionCondition{
		Condition: "and",
		Key:       "tags.bcs_cluster_id",
		Method:    "eq",
		Value:     []string{clusterID},
		Name:      "bcs_cluster_id",
	}
}

func buildNodeCondition(nodeIPs []string) DimensionCondition {
	return DimensionCondition{
		Condition: "and",
		Key:       "tags.node",
		Method:    "eq",
		Value:     nodeIPs,
		Name:      "node",
	}
}

// IPInfo 主机信息
type IPInfo struct {
	Ip      string `json:"ip"`
	CloudID uint64 `json:"bk_cloud_id"`
}
