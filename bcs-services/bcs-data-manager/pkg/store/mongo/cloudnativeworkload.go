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

package mongo

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	bcsdatamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// ModelCloudNative model for cloud native score
type ModelCloudNative struct {
	Public
	Config types.CloudNativeConfig
}

// NewModelCloudNative return a new struct of ModelCloudNative
func NewModelCloudNative(db drivers.DB, bkbaseConf *types.BkbaseConfig) *ModelCloudNative {
	return &ModelCloudNative{
		Public: Public{
			TableName: bkbaseConf.CloudNative.Bkbase.MongoTable,
			Indexes:   make([]drivers.Index, 0),
			DB:        db,
		},
		Config: bkbaseConf.CloudNative,
	}

}

// GetCloudNativeWorkloadList
func (m *ModelCloudNative) GetCloudNativeWorkloadList(ctx context.Context,
	req *bcsdatamanager.GetCloudNativeWorkloadListRequest) (*bcsdatamanager.TEGMessage, error) {
	// page info
	currentPage := int(req.GetCurrentPage())
	pageSize := int(req.GetPageSize())
	if pageSize > 10000 {
		return nil, fmt.Errorf("The max pageSize currently supported is 10000.")
	}
	if currentPage <= 0 {
		currentPage = 1
	}
	startIndex := (currentPage - 1) * pageSize

	// sort by dtEventTime descending
	timeSortParams := map[string]interface{}{
		DtEventTimeKey: DescendingKey,
	}

	timeResult := make([]map[string]string, 0)
	conds := operator.NewLeafCondition(operator.Ne, operator.M{DtEventTimeKey: ""})
	if err := m.DB.Table(m.TableName).Find(conds).WithProjection(map[string]int{
		DtEventTimeKey: 1,
	}).WithSort(timeSortParams).
		WithLimit(1).
		All(ctx, &timeResult); err != nil {
		return nil, fmt.Errorf("Get newest time failed, err: %s", err.Error())
	}
	if len(timeResult) == 0 {
		return nil, fmt.Errorf("Get newest time failed, err: time result is nil")
	}
	dtEventTime := timeResult[0][DtEventTimeKey]
	if dtEventTime == "" {
		return nil, fmt.Errorf("Get newest time failed, err: time result is empty str")
	}

	// time conditions
	timeCond := operator.NewLeafCondition(operator.Eq, operator.M{
		DtEventTimeKey: dtEventTime,
	})

	// finder
	workloads := make([]*types.TEGWorkload, 0)
	finder := m.Public.DB.Table(m.Public.TableName).Find(timeCond)

	// count workloads
	total, err := finder.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("Count workloads error, err: %s", err.Error())
	}

	// sort by _id ascending
	idSortParams := map[string]interface{}{
		"_id": AscendingKey,
	}

	// find workloads, id升序，但时间为dtEventTime的数据
	if err = finder.WithProjection(map[string]int{
		"_id":              0,
		"localTime":        0,
		"thedate":          0,
		"dtEventTime":      0,
		"dtEventTimeStamp": 0,
		"create_at":        0,
	}).WithSort(idSortParams).
		WithStart(int64(startIndex)).
		WithLimit(int64(pageSize)).
		All(ctx, &workloads); err != nil {
		return nil, fmt.Errorf("Get workloads error, err: %s", err.Error())
	}

	result := make([]*bcsdatamanager.TEGWorkload, 0)
	for _, wl := range workloads {
		result = append(result, &bcsdatamanager.TEGWorkload{
			ClusterId:        wl.ClusterId,
			Namespace:        wl.Namespace,
			WorkloadKind:     wl.WorkloadKind,
			WorkloadName:     wl.WorkloadName,
			Maintainer:       wl.Maintainer,
			BakMaintainer:    wl.BakMaintainer,
			BusinessSetId:    wl.BusinessSetId,
			BusinessId:       wl.BusinessId,
			BusinessModuleId: wl.BusinessModuleId,
			SchedulerStatus:  wl.SchedulerStatus,
			ServiceStatus:    wl.ServiceStatus,
			HpaStatus:        wl.HpaStatus,
		})
	}

	// response message
	tegMessage := &bcsdatamanager.TEGMessage{
		Data:     result,
		Platform: m.Config.Platform,
		Appid:    m.Config.AppId,
		Total:    uint32(total),
	}

	return tegMessage, nil
}
