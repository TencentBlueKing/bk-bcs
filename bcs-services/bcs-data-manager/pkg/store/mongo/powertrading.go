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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

// ModelPowerTrading model for power trading operation data
type ModelPowerTrading struct {
	// Tables map[tableIndex]Public
	Tables map[string]*Public
}

// NewModelPowerTrading new powerTrading model
func NewModelPowerTrading(db drivers.DB, bkbaseConf *types.BkbaseConfig) *ModelPowerTrading {
	pt := ModelPowerTrading{
		Tables: map[string]*Public{},
	}

	// create tables for powertrading
	for _, item := range bkbaseConf.PowerTrading {
		if item.MongoTable == "" {
			continue
		}
		p := Public{
			TableName: item.MongoTable,
			Indexes:   make([]drivers.Index, 0),
			DB:        db,
		}
		blog.Infof("[mongo] add index/table ( %s/%s ) to tables", item.Index, item.MongoTable)
		pt.Tables[item.Index] = &p
	}
	return &pt
}

// GetPowerTradingInfo get operation data for power trading
func (pt *ModelPowerTrading) GetPowerTradingInfo(
	ctx context.Context, request *datamanager.GetPowerTradingDataRequest) ([]*any.Any, int64, error) {
	var total int64
	tableIndex := request.GetTable()
	public, ok := pt.Tables[tableIndex]
	if !ok {
		return nil, total, fmt.Errorf("do not support table: %s with mongo store", tableIndex)
	}

	// time conditions
	startTime := request.GetStartTime()
	startTimeCond := operator.NewLeafCondition(operator.Gte, operator.M{
		DtEventTimeKey: startTime,
	})
	endTime := request.GetEndTime()
	endTimeCond := operator.NewLeafCondition(operator.Lte, operator.M{
		DtEventTimeKey: endTime,
	})

	// sortParams conditions
	sortParams := make(map[string]interface{}, 0)
	sortReq := request.GetSort()
	for k, v := range sortReq {
		ascending, err := ensureSortAscending(v)
		if err != nil {
			return nil, total, err
		}
		sortParams[k] = ascending
	}

	// other conditions
	params := request.GetParams()
	operatorMs := make(operator.M, 0)
	for k, v := range params {
		operatorMs[k] = v
	}
	paramsCond := operator.NewLeafCondition(operator.Eq, operatorMs)

	// packing all conditions
	cond := make([]*operator.Condition, 0)
	cond = append(cond, endTimeCond, startTimeCond, paramsCond)
	conds := operator.NewBranchCondition(operator.And, cond...)

	// find all results
	// because of return type is Any, must disable primitive type like primitive.DateTime、primitive.ObjectId
	result := make([]map[string]interface{}, 0)
	err := public.DB.Table(public.TableName).Find(conds).WithProjection(map[string]int{
		"_id":       0,
		"create_at": 0,
	}).WithSort(sortParams).All(ctx, &result)
	if err != nil {
		return nil, total, err
	}

	// packing data
	response := make([]*any.Any, 0)
	for _, r := range result {
		structData, err := structpb.NewStruct(r)
		if err != nil {
			return nil, total, err
		}
		anyData, err := anypb.New(structData)
		if err != nil {
			return nil, total, err
		}

		response = append(response, anyData)
	}

	return response, int64(len(response)), nil
}
