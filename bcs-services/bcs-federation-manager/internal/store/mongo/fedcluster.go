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
	modelFederationClusterIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: "_id", Value: 1},
				bson.E{Key: federationClusterIdUniqueKey, Value: 1},
			},
			Name: federationClusterTableName + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: federationClusterIdUniqueKey, Value: 1},
			},
			Name: federationClusterIdUniqueKey + "_1",
		},
	}
)

// ModelFedCluster model for federation cluster
type ModelFedCluster struct {
	Public
}

// NewModelFedCluster create model for federation cluster
func NewModelFedCluster(db drivers.DB) *ModelFedCluster {
	return &ModelFedCluster{
		Public: Public{
			TableName: tableNamePrefix + federationClusterTableName,
			Indexes:   modelFederationClusterIndexes,
			DB:        db,
		}}
}

// ListFederationClusters list all the federation clusters with options
func (m *ModelFedCluster) ListFederationClusters(ctx context.Context, opt *store.FederationListOptions) ([]*store.FederationCluster, error) {
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

	if opt.Conditions != nil && len(opt.Conditions) != 0 {
		for key, value := range opt.Conditions {
			cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
				key: value,
			}))
		}
	}

	// find
	federationClusterList := make([]*store.FederationCluster, 0)
	err = m.DB.Table(m.TableName).
		Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{
			federationClusterIdUniqueKey: 1,
		}).
		All(ctx, &federationClusterList)
	if err != nil {
		return nil, fmt.Errorf("list federation clusters err: %s", err.Error())
	}
	return federationClusterList, nil
}

// GetFederationCluster get a federation cluster by clusterID
func (m *ModelFedCluster) GetFederationCluster(ctx context.Context, clusterID string) (*store.FederationCluster, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("clusterID is empty")
	}

	if err := ensureTable(ctx, &m.Public); err != nil {
		return nil, err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		federationClusterIdUniqueKey: clusterID,
		isDeletedKey:                 false,
	})

	cluster := &store.FederationCluster{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, cluster); err != nil {
		return nil, fmt.Errorf("get federation cluster err: %s", err.Error())
	}

	return cluster, nil
}

// CreateFederationCluster create a federation cluster
func (m *ModelFedCluster) CreateFederationCluster(ctx context.Context, cluster *store.FederationCluster) error {
	if cluster == nil {
		return fmt.Errorf("federation cluster is nil")
	}
	if cluster.FederationClusterID == "" || cluster.HostClusterID == "" {
		return fmt.Errorf("federation cluster id or host cluster id is empty")
	}

	cluster.IsDeleted = false

	// set default status
	if cluster.Status == "" {
		cluster.Status = store.RunningStatus
		cluster.StatusMessage = "federation cluster is created"
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
	if cluster.Extras == nil {
		cluster.Extras = make(map[string]string, 0)
	}

	// insert
	if _, err := m.DB.Table(m.TableName).Insert(ctx, []interface{}{cluster}); err != nil {
		return fmt.Errorf("create federation cluster err: %s", err.Error())
	}

	return nil
}

// DeleteFederationCluster soft delete a fed cluster
func (m *ModelFedCluster) DeleteFederationCluster(ctx context.Context, opt *store.FederationClusterDeleteOptions) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return err
	}
	if opt == nil {
		return fmt.Errorf("fed cluster delete options is nil")
	}

	fedClusterId, updater := opt.FederationClusterID, opt.Updater
	if fedClusterId == "" {
		return fmt.Errorf("fedClusterId or fedClusterId is empty")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		federationClusterIdUniqueKey: fedClusterId,
		isDeletedKey:                 false,
	})

	fedCluster := &store.FederationCluster{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, fedCluster); err != nil {
		return fmt.Errorf("get fed cluster err: %s", err.Error())
	}

	// soft delete
	fedCluster.IsDeleted = true
	fedCluster.UpdatedTime = time.Now()
	fedCluster.DeletedTime = time.Now()
	fedCluster.Updater = updater
	fedCluster.Status = store.DeletedStatus

	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": fedCluster}); err != nil {
		return fmt.Errorf("delete fed cluster err: %s", err.Error())
	}

	return nil
}

// UpdateFederationCluster update federation cluster
func (m *ModelFedCluster) UpdateFederationCluster(ctx context.Context, cluster *store.FederationCluster, updater string) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		federationClusterIdUniqueKey: cluster.FederationClusterID,
		isDeletedKey:                 false,
	})

	// make sure cluster exists
	fedCluster := &store.FederationCluster{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, fedCluster); err != nil {
		return fmt.Errorf("can not find the cluster to be update, cluster: %s, err: %s", cluster.FederationClusterID, err.Error())
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
