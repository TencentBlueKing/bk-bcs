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
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

var (
	modelPublicIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: CreateTimeKey, Value: 1},
			},
			Name: CreateTimeKey + "_1",
		},
		{
			Name: types.PublicTableName + "_idx",
			Key: bson.D{
				bson.E{Key: ObjectTypeKey, Value: 1},
				bson.E{Key: ProjectIDKey, Value: 1},
				bson.E{Key: ClusterIDKey, Value: 1},
				bson.E{Key: NamespaceKey, Value: 1},
				bson.E{Key: WorkloadTypeKey, Value: 1},
				bson.E{Key: WorkloadNameKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelPublic public model
type ModelPublic struct {
	Public
}

// NewModelPublic new public model
func NewModelPublic(db drivers.DB) *ModelPublic {
	return &ModelPublic{Public: Public{
		TableName: types.DataTableNamePrefix + types.PublicTableName,
		Indexes:   modelPublicIndexes,
		DB:        db,
	}}
}

// InsertPublicInfo insert public info
func (m *ModelPublic) InsertPublicInfo(ctx context.Context, metrics *types.PublicData,
	opts *types.JobCommonOpts) error {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}

	cond, err := m.generateCond(opts)
	if err != nil {
		return err
	}
	retPublic := &types.PublicData{}
	err = m.DB.Table(m.TableName).Find(cond).One(ctx, retPublic)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Infof("public info not found, create a new data")
			metrics.CreateTime = primitive.NewDateTimeFromTime(time.Now())
			metrics.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{metrics})
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	retPublic.UpdateTime = primitive.NewDateTimeFromTime(time.Now())
	return m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": retPublic})
}

// GetRawPublicInfo get raw public info data
func (m *ModelPublic) GetRawPublicInfo(ctx context.Context, opts *types.JobCommonOpts) ([]*types.PublicData, error) {
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond, err := m.generateCond(opts)
	if err != nil {
		return nil, err
	}
	retPublic := make([]*types.PublicData, 0)
	err = m.DB.Table(m.TableName).Find(cond).All(ctx, &retPublic)
	if err != nil {
		return nil, err
	}
	return retPublic, nil
}

func (m *ModelPublic) generateCond(opts *types.JobCommonOpts) (*operator.Condition, error) {
	switch opts.ObjectType {
	case types.ProjectType:
		return operator.NewLeafCondition(operator.Eq, operator.M{
			ProjectIDKey:  opts.ProjectID,
			ObjectTypeKey: types.ProjectType,
		}), nil
	case types.ClusterType:
		return operator.NewLeafCondition(operator.Eq, operator.M{
			ProjectIDKey:  opts.ProjectID,
			ClusterIDKey:  opts.ClusterID,
			ObjectTypeKey: types.ClusterType,
		}), nil
	case types.NamespaceType:
		return operator.NewLeafCondition(operator.Eq, operator.M{
			ProjectIDKey:  opts.ProjectID,
			ClusterIDKey:  opts.ClusterID,
			NamespaceKey:  opts.Namespace,
			ObjectTypeKey: types.NamespaceType,
		}), nil
	case types.WorkloadType:
		return operator.NewLeafCondition(operator.Eq, operator.M{
			ProjectIDKey:    opts.ProjectID,
			ClusterIDKey:    opts.ClusterID,
			NamespaceKey:    opts.Namespace,
			WorkloadTypeKey: opts.WorkloadType,
			WorkloadNameKey: opts.WorkloadName,
			ObjectTypeKey:   types.WorkloadType,
		}), nil
	default:
		return nil, fmt.Errorf("wrong object type: %s", opts.ObjectType)
	}
}
