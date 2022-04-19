/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tke

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	tkeCidrTableName = "tkecidrs"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	tkeCidrKeyVPC            = "vpc"
	tkeCidrKeyCIDR           = "cidr"
	defaultTkeCidrListLength = 3000
)

var (
	tkeCidrIndexes = []drivers.Index{
		{
			Name: tkeCidrTableName + "_idx",
			Key: bson.D{
				bson.E{Key: tkeCidrKeyVPC, Value: 1},
				bson.E{Key: tkeCidrKeyCIDR, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelTkeCidr database operation for tke cidr
type ModelTkeCidr struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create tke cidr model
func New(db drivers.DB) *ModelTkeCidr {
	return &ModelTkeCidr{
		tableName: util.DataTableNamePrefix + tkeCidrTableName,
		indexes:   tkeCidrIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelTkeCidr) ensureTable(ctx context.Context) error {
	m.isTableEnsuredMutex.RLock()
	if m.isTableEnsured {
		m.isTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := util.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
		m.isTableEnsuredMutex.RUnlock()
		return err
	}
	m.isTableEnsuredMutex.RUnlock()

	m.isTableEnsuredMutex.Lock()
	m.isTableEnsured = true
	m.isTableEnsuredMutex.Unlock()
	return nil
}

// CreateTkeCidr create tke cidr
func (m *ModelTkeCidr) CreateTkeCidr(ctx context.Context, cidr *types.TkeCidr) error {
	if cidr == nil {
		return fmt.Errorf("cidr to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{cidr}); err != nil {
		return err
	}
	return nil
}

// UpdateTkeCidr update tke cidr
func (m *ModelTkeCidr) UpdateTkeCidr(ctx context.Context, cidr *types.TkeCidr) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tkeCidrKeyVPC:  cidr.VPC,
		tkeCidrKeyCIDR: cidr.CIDR,
	})
	oldTkeCidr := &types.TkeCidr{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, oldTkeCidr); err != nil {
		return err
	}
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": cidr})
}

// DeleteTkeCidr delete tke cidr
func (m *ModelTkeCidr) DeleteTkeCidr(ctx context.Context, vpc string, cidr string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tkeCidrKeyVPC:  vpc,
		tkeCidrKeyCIDR: cidr,
	})
	deleteCounter, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	if deleteCounter == 0 {
		blog.Warnf("no tke cidr %s of vpc %s delete ", vpc, cidr)
	}
	return nil
}

// GetTkeCidr get tke cidr
func (m *ModelTkeCidr) GetTkeCidr(ctx context.Context, vpc string, cidr string) (*types.TkeCidr, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tkeCidrKeyVPC:  vpc,
		tkeCidrKeyCIDR: cidr,
	})
	retTkeCidr := &types.TkeCidr{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retTkeCidr); err != nil {
		return nil, err
	}
	return retTkeCidr, nil
}

// ListTkeCidr list tke cidr
func (m *ModelTkeCidr) ListTkeCidr(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.TkeCidr, error) {
	retTkeCidrList := make([]types.TkeCidr, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultTkeCidrListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &retTkeCidrList); err != nil {
		return nil, err
	}
	return retTkeCidrList, nil
}

// ListTkeCidrCount list tke cidr count
func (m *ModelTkeCidr) ListTkeCidrCount(ctx context.Context, opt *options.ListOption) ([]types.TkeCidrCount, error) {
	retTkeCidrCountList := make([]types.TkeCidrCount, 0)
	pipeline := []map[string]interface{}{
		{
			"$group": map[string]interface{}{
				"_id": map[string]interface{}{
					"vpc":      "$vpc",
					"ipnumber": "$ipnumber",
					"status":   "$status",
				},
				"vpc":      map[string]interface{}{"$first": "$vpc"},
				"ipnumber": map[string]interface{}{"$first": "$ipnumber"},
				"status":   map[string]interface{}{"$first": "$status"},
				"count": map[string]interface{}{
					"$sum": 1,
				},
			},
		},
	}
	if len(opt.Sort) != 0 {
		pipeline = append(pipeline, map[string]interface{}{
			"$sort": util.MapInt2MapIf(opt.Sort),
		})
	}
	if opt.Limit == 0 {
		pipeline = append(pipeline, map[string]interface{}{
			"$limit": defaultTkeCidrListLength,
		})
	} else {
		pipeline = append(pipeline, map[string]interface{}{
			"$limit": opt.Limit,
		})
	}
	if err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &retTkeCidrCountList); err != nil {
		return nil, err
	}
	return retTkeCidrCountList, nil
}
