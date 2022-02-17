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
 *
 */

package cloudvpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	tableName = "cloudvpc"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	cloudKey                  = "cloudid"
	vpcIDKey                  = "vpcid"
	defaultCloudVPCListLength = 1000
)

var (
	cloudIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: cloudKey, Value: 1},
				bson.E{Key: vpcIDKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelCloudVPC database operation for cloudVPC
type ModelCloudVPC struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create cloudVPC model
func New(db drivers.DB) *ModelCloudVPC {
	return &ModelCloudVPC{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   cloudIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelCloudVPC) ensureTable(ctx context.Context) error {
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

// CreateCloudVPC insert cloudVPC
func (m *ModelCloudVPC) CreateCloudVPC(ctx context.Context, vpc *types.CloudVPC) error {
	if vpc == nil {
		return fmt.Errorf("cloudVPC to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{vpc}); err != nil {
		return err
	}
	return nil
}

// UpdateCloudVPC update cloudVPC with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelCloudVPC) UpdateCloudVPC(ctx context.Context, vpc *types.CloudVPC) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		cloudKey: vpc.CloudID,
		vpcIDKey: vpc.VpcID,
	})

	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": vpc})
}

// DeleteCloudVPC delete cloudVPC
func (m *ModelCloudVPC) DeleteCloudVPC(ctx context.Context, cloudID string, vpcID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		cloudKey: cloudID,
		vpcIDKey: vpcID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetCloudVPC get cloudVPC
func (m *ModelCloudVPC) GetCloudVPC(ctx context.Context, cloudID, vpcID string) (*types.CloudVPC, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		cloudKey: cloudID,
		vpcIDKey: vpcID,
	})
	cloudVPC := &types.CloudVPC{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, cloudVPC); err != nil {
		return nil, err
	}

	return cloudVPC, nil
}

// ListCloudVPC list cloudVPC
func (m *ModelCloudVPC) ListCloudVPC(ctx context.Context, vpc *operator.Condition, opt *options.ListOption) (
	[]types.CloudVPC, error) {
	cloudVPCList := make([]types.CloudVPC, 0)
	finder := m.db.Table(m.tableName).Find(vpc)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultCloudVPCListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &cloudVPCList); err != nil {
		return nil, err
	}

	return cloudVPCList, nil
}
