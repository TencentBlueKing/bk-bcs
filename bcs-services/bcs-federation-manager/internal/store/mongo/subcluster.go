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

// Package mongo sub cluster store
package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
)

var (
	modelSubClusterIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: "_id", Value: 1},
				bson.E{Key: subClusterUniqueKey, Value: 1},
			},
			Name: subClusterTableName + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: subClusterUniqueKey, Value: 1},
			},
			Name: subClusterUniqueKey + "_1",
		},
	}
)

// ModelSubCluster create model for sub cluster
type ModelSubCluster struct {
	Public
}

// NewModelSubCluster create model for sub cluster
func NewModelSubCluster(db drivers.DB) *ModelSubCluster {
	return &ModelSubCluster{
		Public: Public{
			TableName: tableNamePrefix + subClusterTableName,
			Indexes:   modelSubClusterIndexes,
			DB:        db,
		}}
}

// ListSubClusters list all the sub clusters with options
func (m *ModelSubCluster) ListSubClusters(ctx context.Context, opt *store.SubClusterListOptions) ([]*store.SubCluster, error) {
	// params check
	if opt == nil {
		return nil, fmt.Errorf("ListOption is nil")
	}

	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}

	// conditions
	cond := make([]*operator.Condition, 0)
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		isDeletedKey: false,
	}))
	if opt.FederationClusterID != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			subClusterFederationClusterIdKey: opt.FederationClusterID,
			isDeletedKey:                     false,
		}))
	}
	if opt.Conditions != nil && len(opt.Conditions) != 0 {
		for key, value := range opt.Conditions {
			cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
				key: value,
			}))
		}
	}

	// find
	subClusterList := make([]*store.SubCluster, 0)
	err = m.DB.Table(m.TableName).
		Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{
			subClusterUniqueKey: 1,
		}).
		All(ctx, &subClusterList)
	if err != nil {
		return nil, fmt.Errorf("list sub clusters err: %s", err.Error())
	}
	return subClusterList, nil
}

// GetSubCluster get a sub cluster by subClusterId and federation cluster clusterId
func (m *ModelSubCluster) GetSubCluster(ctx context.Context, fedClusterId, subClusterId string) (*store.SubCluster, error) {
	if fedClusterId == "" || subClusterId == "" {
		return nil, fmt.Errorf("fedClusterId or subClusterId is empty")
	}

	if err := ensureTable(ctx, &m.Public); err != nil {
		return nil, err
	}
	uid := formatSubclusterUID(fedClusterId, subClusterId)
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		subClusterUniqueKey: uid,
		isDeletedKey:        false,
	})

	subCluster := &store.SubCluster{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, subCluster); err != nil {
		return nil, fmt.Errorf("get sub cluster err: %s", err.Error())
	}

	return subCluster, nil
}

// CreateSubCluster create a sub cluster
func (m *ModelSubCluster) CreateSubCluster(ctx context.Context, cluster *store.SubCluster) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return err
	}

	if cluster == nil {
		return fmt.Errorf("sub cluster is nil")
	}
	if cluster.SubClusterID == "" || cluster.HostClusterID == "" || cluster.FederationClusterID == "" {
		return fmt.Errorf("sub cluster id or host cluster id or federation cluster id is empty")
	}

	cluster.IsDeleted = false

	if cluster.UID == "" {
		cluster.UID = formatSubclusterUID(cluster.FederationClusterID, cluster.SubClusterID)
	}

	// set default status
	if cluster.Status == "" {
		cluster.Status = store.RunningStatus
		cluster.StatusMessage = "sub cluster is created"
	}

	// set default time
	now := time.Now()
	cluster.CreatedTime = now
	cluster.UpdatedTime = now

	// set default creator and updater
	if cluster.Creator == "" {
		cluster.Creator = "admin"
	}
	if cluster.Updater == "" {
		cluster.Updater = cluster.Creator
	}

	// set default extra
	if cluster.Labels == nil {
		cluster.Labels = make(map[string]string, 0)
	}

	// insert
	if _, err := m.DB.Table(m.TableName).Insert(ctx, []interface{}{cluster}); err != nil {
		return fmt.Errorf("create sub cluster err: %s", err.Error())
	}

	return nil
}

// DeleteSubCluster soft delete a sub cluster
func (m *ModelSubCluster) DeleteSubCluster(ctx context.Context, opt *store.SubClusterDeleteOptions) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return err
	}
	if opt == nil {
		return fmt.Errorf("sub cluster delete options is nil")
	}

	fedClusterId, subClusterId, updater := opt.FederationClusterID, opt.SubClusterID, opt.Updater
	if fedClusterId == "" || subClusterId == "" {
		return fmt.Errorf("fedClusterId or subClusterId is empty")
	}

	uid := formatSubclusterUID(fedClusterId, subClusterId)
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		subClusterUniqueKey: uid,
		isDeletedKey:        false,
	})

	subCluster := &store.SubCluster{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, subCluster); err != nil {
		return fmt.Errorf("get sub cluster err: %s", err.Error())
	}

	// soft delete
	subCluster.IsDeleted = true
	subCluster.UpdatedTime = time.Now()
	subCluster.DeletedTime = time.Now()
	subCluster.Updater = updater
	subCluster.Status = store.DeletedStatus

	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": subCluster}); err != nil {
		return fmt.Errorf("delete sub cluster err: %s", err.Error())
	}

	return nil
}

// UpdateSubCluster update sub cluster
func (m *ModelSubCluster) UpdateSubCluster(ctx context.Context, cluster *store.SubCluster, updater string) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		subClusterUniqueKey: cluster.UID,
		isDeletedKey:        false,
	})

	// make sure cluster exists
	subCluster := &store.SubCluster{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, subCluster); err != nil {
		return fmt.Errorf("can not find the cluster to be update, cluster: %s, err: %s", cluster.SubClusterID, err.Error())
	}

	// set update time and updater
	cluster.UpdatedTime = time.Now()
	cluster.Updater = updater

	// update cluster
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": cluster}); err != nil {
		return fmt.Errorf("update fed cluster err: %s", err.Error())
	}

	return nil
}
