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
	"github.com/golang/protobuf/ptypes/any"
	"github.com/jmoiron/sqlx"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// ModelPowerTrading model for power trading operation data
type ModelPowerTrading struct {
	// Tables map[tableIndex]Public
	Tables map[string]*Public
}

// NewModelPowerTrading new powerTrading model
func NewModelPowerTrading(dbs map[string]*sqlx.DB, bkbaseConf *types.BkbaseConfig) *ModelPowerTrading {
	pt := &ModelPowerTrading{
		Tables: map[string]*Public{},
	}

	// create tables for powertrading
	for _, item := range bkbaseConf.PowerTrading {
		index, table, store := item.Index, item.TspiderTable, item.TspiderStore
		if index == "" || table == "" || store == "" {
			blog.Warnf("tspider set error when init powertrading model: %s/%s/%s", index, store, table)
			continue
		}
		// get db by store name
		db, ok := dbs[store]
		if !ok || db == nil {
			blog.Warnf("store is not support: %s/%s/%s", index, store, table)
			continue
		}
		blog.Infof("[tspider] add index/table/store ( %s/%s/%s ) to tables", index, table, store)
		pt.Tables[index] = &Public{
			TableName: table,
			DB:        db,
		}
	}

	return pt
}

// GetPowerTradingInfo get operation data ofr power trading
func (pt *ModelPowerTrading) GetPowerTradingInfo(ctx context.Context, request *datamanager.GetPowerTradingDataRequest) ([]*any.Any, int64, error) {
	// validate params
	if err := pt.validate(request); err != nil {
		return nil, 0, err
	}

	public, ok := pt.Tables[request.GetTable()]
	if !ok {
		return nil, 0, fmt.Errorf("do not support table: %s with tspider store", request.GetTable())
	}

	// get query and count builders
	query, count, err := pt.getBuilders(request, public.TableName)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get query builder, err: %s", err)
	}

	// query for data
	response, err := public.QueryxToAny(query)
	if err != nil {
		return nil, 0, err
	}

	// count for data
	total, err := public.Countx(count)
	if err != nil {
		return nil, 0, err
	}

	return response, int64(total), nil
}

func (pt *ModelPowerTrading) validate(request *datamanager.GetPowerTradingDataRequest) error {
	// validate time range params
	startTime, endTime := request.GetStartTime(), request.GetEndTime()
	if startTime == "" || endTime == "" {
		return fmt.Errorf("request params error: empty startTime or endTime")
	}

	// validate page info params
	page, size := request.GetPage(), request.GetSize()
	if (page == 0 && size != 0) || (page != 0 && size == 0) {
		return fmt.Errorf("page and size should greater than 0 or both equal to 0, if both 0 return data will without page info")
	}

	return nil
}

func (pt *ModelPowerTrading) getBuilders(
	request *datamanager.GetPowerTradingDataRequest, tableName string) (sq.SelectBuilder, sq.SelectBuilder, error) {

	// dtEventTime
	startTime, endTime := request.GetStartTime(), request.GetEndTime()

	queryBuilder := sq.Select(SqlSelectAll).
		From(tableName).
		Where(sq.GtOrEq{DtEventTimeKey: startTime}).
		Where(sq.LtOrEq{DtEventTimeKey: endTime})

	countBuilder := sq.Select(SqlSelectCount).
		From(tableName).
		Where(sq.GtOrEq{DtEventTimeKey: startTime}).
		Where(sq.LtOrEq{DtEventTimeKey: endTime})

	// conditions
	params := request.GetParams()
	for k, v := range params {
		queryBuilder = queryBuilder.Where(sq.Eq{k: v})
		countBuilder = countBuilder.Where(sq.Eq{k: v})
	}

	// page info
	page, size := request.GetPage(), request.GetSize()
	if page > 0 && size > 0 {
		limit := uint64(size)
		offset := (uint64(page) - 1) * limit
		queryBuilder = queryBuilder.Limit(limit).Offset(offset)
	}

	// sort by keys, ascending by value
	sortReq := request.GetSort()
	sortParams := []string{}
	for k, v := range sortReq {
		asendingFlag, err := ensureSortAscending(v)
		if err != nil {
			return queryBuilder, queryBuilder, err
		}
		sortParams = append(sortParams, k+" "+asendingFlag)
	}
	if len(sortParams) != 0 {
		queryBuilder = queryBuilder.OrderBy(sortParams...)
		countBuilder = countBuilder.OrderBy(sortParams...)
	}

	return queryBuilder, countBuilder, nil
}
