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

package sqlstore

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

const (
	// CidrStatusAvailable available status
	CidrStatusAvailable = "available"
	// CidrStatusUsed used status
	CidrStatusUsed = "used"
	// CidrStatusReserved reserved status
	CidrStatusReserved = "reserved"
)

// CidrCount cidrInfo
type CidrCount struct {
	Count    int    `json:"count"`
	Vpc      string `json:"vpc"`
	IpNumber uint   `json:"ip_number"`
	Status   string `json:"status"`
}

// QueryTkeCidr query tke cidr info
func QueryTkeCidr(tkeCidr *models.TkeCidr) *models.TkeCidr {
	start := time.Now()
	result := models.TkeCidr{}
	GCoreDB.Where(tkeCidr).First(&result)
	if result.ID != 0 {
		return &result
	}
	metrics.ReportMysqlSlowQueryMetrics("QueryTkeCidr", metrics.Query, metrics.SucStatus, start)
	return nil
}

// SaveTkeCidr save tke cidr
func SaveTkeCidr(vpc, cidr string, ipNumber uint, status, cluster string) error {
	start := time.Now()
	tkeCidr := &models.TkeCidr{
		Vpc:      vpc,
		Cidr:     cidr,
		IpNumber: ipNumber,
		Status:   status,
		Cluster:  &cluster,
	}

	err := GCoreDB.Create(tkeCidr).Error
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("SaveTkeCidr", metrics.Create, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("SaveTkeCidr", metrics.Create, metrics.SucStatus, start)
	return nil
}

// UpdateTkeCidr update tke cidr
func UpdateTkeCidr(tkeCidr, updatedTkeCidr *models.TkeCidr) error {
	start := time.Now()
	err := GCoreDB.Model(tkeCidr).Updates(*updatedTkeCidr).Error
	if err != nil {
		metrics.ReportMysqlSlowQueryMetrics("UpdateTkeCidr", metrics.Update, metrics.ErrStatus, start)
		return err
	}
	metrics.ReportMysqlSlowQueryMetrics("UpdateTkeCidr", metrics.Update, metrics.SucStatus, start)
	return nil
}

// CountTkeCidr count tke cidr
func CountTkeCidr() []CidrCount {
	start := time.Now()
	var cidrCounts []CidrCount
	GCoreDB.Table("tke_cidrs").Select("count(*) as count, vpc, ip_number, status").Group("vpc, ip_number, status").
		Scan(&cidrCounts)
	metrics.ReportMysqlSlowQueryMetrics("CountTkeCidr", metrics.Query, metrics.SucStatus, start)
	return cidrCounts
}
