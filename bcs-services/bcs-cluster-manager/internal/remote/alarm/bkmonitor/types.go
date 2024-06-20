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
	ipInfos := make([]IPInfo, 0)
	for i := range hosts.HostList {
		ipInfos = append(ipInfos, IPInfo{
			Ip:      hosts.HostList[i].IP,
			CloudID: hosts.HostList[i].CloudID,
		})
	}
	if len(ipInfos) == 0 {
		blog.Errorf("buildBizHostAlarmConfig hosts empty")
		return nil, fmt.Errorf("buildBizHostAlarmConfig ipList empty")
	}

	return &ShieldHostAlarmRequest{
		Category:  string(scope),
		BkBizID:   bizID,
		BeginTime: time.Now().Format("2006-01-02 15:04:00"),
		EndTime:   time.Now().Add(time.Minute * 30).Format("2006-01-02 15:04:00"),
		DimensionConfig: DimensionConfig{
			ScopeType: "ip",
			Target:    ipInfos,
		},
		Description:  shieldDesc,
		ShieldNotice: false,
		CycleConfig:  CycleConfig{Type: 1},
	}, nil
}

// CycleConfig 屏蔽周期配置
type CycleConfig struct {
	Type uint32 `json:"type"`
}

// DimensionConfig 屏蔽维度
type DimensionConfig struct {
	ScopeType string   `json:"scope_type"`
	Target    []IPInfo `json:"target"`
}

// IPInfo 主机信息
type IPInfo struct {
	Ip      string `json:"ip"`
	CloudID uint64 `json:"bk_cloud_id"`
}
