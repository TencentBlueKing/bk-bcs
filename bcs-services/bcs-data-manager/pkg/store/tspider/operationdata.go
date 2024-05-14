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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
)

const (
	// ProjectRequestType  type of operation data supportted
	ProjectRequestType = "project"
	// ClusterRequestType type of operation data supportted
	ClusterRequestType = "cluster"
	// NamespaceRequestType type of operation data supportted
	NamespaceRequestType = "namespace"

	// ProjectCodeColumnKey column key for project type data
	ProjectCodeColumnKey = "project_code"
	// ClusterIdColumnKey column key for cluster type data
	ClusterIdColumnKey = "clusterId"
	// NamespaceColumnKey column key for namespace type data
	NamespaceColumnKey = "namespace"
)

// ModelUserOperationData model for bcs user operation data
type ModelUserOperationData struct {
	// Tables map[tableIndex]Public
	Tables map[string]*Public
}

// NewModelUserOperationData new bcs user operation data model
func NewModelUserOperationData(dbs map[string]*sqlx.DB, bkbaseConf *types.BkbaseConfig) *ModelUserOperationData {
	od := &ModelUserOperationData{
		Tables: map[string]*Public{},
	}

	for _, item := range bkbaseConf.BcsOperationData {
		index, table, store := item.Index, item.TspiderTable, item.TspiderStore
		if index == "" || table == "" || store == "" {
			blog.Warnf("tspider set error when init operation data model: %s/%s/%s", index, store, table)
			continue
		}
		// get db by store name
		db, ok := dbs[store]
		if !ok || db == nil {
			blog.Warnf("store is not support: %s/%s/%s", index, store, table)
			continue
		}
		blog.Infof("[tspider] add index/table/store ( %s/%s/%s ) to tables", index, table, store)
		od.Tables[index] = &Public{
			TableName: table,
			DB:        db,
		}
	}

	return od
}

// GetUserOperationDataList get operation data for bcs user
func (od *ModelUserOperationData) GetUserOperationDataList(ctx context.Context,
	request *datamanager.GetUserOperationDataListRequest) ([]*structpb.Struct, int64, error) {
	// validate params
	if err := od.validate(request); err != nil {
		return nil, 0, err
	}

	// check db exist
	public, ok := od.Tables[request.GetType()]
	if !ok {
		return nil, 0, fmt.Errorf("request params error: not support request type %s", request.GetType())
	}

	// get builders
	query, count := od.getBuilders(request, public.TableName)

	// query for data
	response, err := public.QueryxToStructpb(query)
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

func (od *ModelUserOperationData) validate(request *datamanager.GetUserOperationDataListRequest) error {
	// 所有类型都需要projectCode
	if request.GetProjectCode() == "" {
		return fmt.Errorf("request params error: empty projectCode")
	}

	// cluster和namespace需要有clusterId
	if request.GetClusterId() == "" &&
		(request.GetType() == ClusterRequestType || request.GetType() == NamespaceRequestType) {
		return fmt.Errorf("request params error: empty clusterId")
	}

	// namespace需要有namespace
	if request.GetNamespace() == "" && request.GetType() == NamespaceRequestType {
		return fmt.Errorf("request params error: empty namespace")
	}

	// page and size should not less than 1
	if request.GetPage() < 1 || request.GetSize() < 1 {
		return fmt.Errorf("request params error: page and size should not less than 1")
	}

	// startTime and endTime are required
	if request.GetStartTime() == 0 || request.GetEndTime() == 0 {
		return fmt.Errorf("request params error: both startTime and endTime are required")
	}

	return nil
}

func (od *ModelUserOperationData) getBuilders(
	request *datamanager.GetUserOperationDataListRequest, tableName string) (
	queryBuilder sq.SelectBuilder, countBuilder sq.SelectBuilder) {

	// bkbase dtEventTimeStamp is accurate to millseconds
	startTime := request.GetStartTime() * 1000
	endTime := request.GetEndTime() * 1000

	size := request.GetSize()
	offset := (request.GetPage() - 1) * size

	// query builder
	queryBuilder = sq.Select(SqlSelectAll).
		From(tableName).
		Where(sq.GtOrEq{DtEventTimeStampKey: startTime}).
		Where(sq.LtOrEq{DtEventTimeStampKey: endTime}).
		Where(sq.Eq{ProjectCodeColumnKey: request.GetProjectCode()}).
		OrderBy(DtEventTimeStampKey + " " + DescendingFlag).
		Limit(uint64(size)).
		Offset(uint64(offset))

	// count builder
	countBuilder = sq.Select(SqlSelectCount).
		From(tableName).
		Where(sq.GtOrEq{DtEventTimeStampKey: startTime}).
		Where(sq.LtOrEq{DtEventTimeStampKey: endTime}).
		Where(sq.Eq{ProjectCodeColumnKey: request.GetProjectCode()}).
		OrderBy(DtEventTimeStampKey + " " + DescendingFlag)

	if request.GetType() == ClusterRequestType || request.GetType() == NamespaceRequestType {
		queryBuilder = queryBuilder.Where(sq.Eq{ClusterIdColumnKey: request.GetClusterId()})
		countBuilder = countBuilder.Where(sq.Eq{ClusterIdColumnKey: request.GetClusterId()})
	}

	if request.GetType() == NamespaceRequestType {
		queryBuilder = queryBuilder.Where(sq.Eq{NamespaceColumnKey: request.GetNamespace()})
		countBuilder = countBuilder.Where(sq.Eq{NamespaceColumnKey: request.GetNamespace()})
	}

	return queryBuilder, countBuilder
}
