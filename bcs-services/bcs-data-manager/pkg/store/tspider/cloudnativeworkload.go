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

package tspider

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// ModelCloudNative model for cloud native score
type ModelCloudNative struct {
	Public
	Config types.CloudNativeConfig
}

// NewModelCloudNative return a new struct of ModelCloudNative
func NewModelCloudNative(dbs map[string]*sqlx.DB, bkbaseConf *types.BkbaseConfig) *ModelCloudNative {
	store := bkbaseConf.CloudNative.Bkbase.TspiderStore
	db, ok := dbs[store]
	if !ok {
		blog.Errorf("[cloudnative] store is not support: %s ", store)
		return &ModelCloudNative{}
	}

	return &ModelCloudNative{
		Public: Public{
			TableName: bkbaseConf.CloudNative.Bkbase.TspiderTable,
			DB:        db,
		},
		Config: bkbaseConf.CloudNative,
	}

}

// GetCloudNativeWorkloadList get cloud native workload list
func (m *ModelCloudNative) GetCloudNativeWorkloadList(ctx context.Context,
	request *datamanager.GetCloudNativeWorkloadListRequest) (*datamanager.TEGMessage, error) {
	// validate params
	if err := m.validate(request); err != nil {
		return nil, err
	}

	// check db init
	if m.DB == nil {
		return nil, fmt.Errorf("Cloud Native workload init failed, DB is empty")
	}

	// get builders
	query, count, err := m.getBuilders(request, m.Public.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get query builder, err: %s", err)
	}

	// query for data
	workloads := make([]*types.TEGWorkload, 0)
	if err = m.Public.QueryxToStruct(query, &workloads); err != nil {
		return nil, err
	}
	result := make([]*datamanager.TEGWorkload, 0)
	for _, wl := range workloads {
		result = append(result, &datamanager.TEGWorkload{
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
			//HpaStatus:        wl.HpaStatus,
		})
	}

	// count for data
	total, err := m.Public.Countx(count)
	if err != nil {
		return nil, err
	}

	tegMessage := &datamanager.TEGMessage{
		Data:     result,
		Platform: m.Config.Platform,
		Appid:    m.Config.AppId,
		Total:    uint32(total),
	}

	return tegMessage, nil
}

func (m *ModelCloudNative) validate(request *datamanager.GetCloudNativeWorkloadListRequest) error {
	pageSize := request.GetPageSize()
	if pageSize > 10000 {
		return fmt.Errorf("the max pageSize currently supported is 10000")
	}

	return nil
}

func (m *ModelCloudNative) getBuilders(
	request *datamanager.GetCloudNativeWorkloadListRequest, tableName string) (sq.SelectBuilder, sq.SelectBuilder, error) {

	// page info
	currentPage := request.GetCurrentPage()
	limit := request.GetPageSize()
	if currentPage <= 0 {
		currentPage = 1
	}
	offset := (currentPage - 1) * limit

	queryBuilder := sq.Select(types.TEGWorkloadColumns...).
		From(tableName).
		OrderBy(types.TEGWorkloadSortColumns...).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	countBuilder := sq.Select(SqlSelectCount).
		From(tableName).
		Limit(1)

	var maxTime string
	if err := m.Public.GetMax(m.TableName, DtEventTimeStampKey, &maxTime); err != nil {
		return queryBuilder, countBuilder, err
	}
	queryBuilder = queryBuilder.Where(sq.Eq{DtEventTimeStampKey: maxTime})
	countBuilder = countBuilder.Where(sq.Eq{DtEventTimeStampKey: maxTime})

	return queryBuilder, countBuilder, nil
}
