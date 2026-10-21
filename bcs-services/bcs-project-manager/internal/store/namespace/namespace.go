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

// Package namespace xxx
package namespace

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
)

const (
	// table name
	tableName = "namespace"
	// FieldKeyName name
	FieldKeyName = "name"
	// FieldKeyProjectCode projectCode
	FieldKeyProjectCode = "projectCode"
	// FieldKeyClusterID clusterID
	FieldKeyClusterID = "clusterID"
	// FieldKeyItsmTicketType itsmTicketType
	FieldKeyItsmTicketType = "itsmTicketType"
	// FieldKeyIsDeleted isDeleted
	FieldKeyIsDeleted = "isDeleted"
)

var (
	namespaceIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyProjectCode, Value: 1},
				bson.E{Key: FieldKeyClusterID, Value: 1},
				bson.E{Key: FieldKeyName, Value: 1},
			},
			Unique: false,
		},
	}
)

var (
	// ItsmTicketStatusCreated enum string for created status
	ItsmTicketStatusCreated = "CREATED"

	// ItsmTicketTypeCreate enum string for itsm ticket type create
	ItsmTicketTypeCreate = "CREATE"
	// ItsmTicketTypeUpdate enum string for itsm ticket type update
	ItsmTicketTypeUpdate = "UPDATE"
	// ItsmTicketTypeDelete enum string for itsm ticket type delete
	ItsmTicketTypeDelete = "DELETE"
)

// Namespace staging namespace entity
type Namespace struct {
	ProjectCode      string      `json:"projectCode" bson:"projectCode"`
	ClusterID        string      `json:"clusterID" bson:"clusterID"`
	Name             string      `json:"name" bson:"name"`
	CreateTime       string      `json:"createTime" bson:"createTime"`
	Status           string      `json:"status" bson:"status"`
	Creator          string      `json:"creator" bson:"creator"`
	Updater          string      `json:"updater" bson:"updater"`
	Managers         string      `json:"managers" bson:"managers"`
	ResourceQuota    *Quota      `json:"resourceQuota" bson:"resourceQuota"`
	UsedQuota        *Quota      `json:"usedQuota" bson:"usedQuota"`
	Variables        []*Variable `json:"variables" bson:"variables"`
	IsDeleted        bool        `json:"isDeleted" bson:"isDeleted"`
	ItsmTicketType   string      `json:"itsmTicketType" bson:"itsmTicketType"`
	ItsmTicketURL    string      `json:"itsmTicketURL" bson:"itsmTicketURL"`
	ItsmTicketSN     string      `json:"itsmTicketSN" bson:"itsmTicketSN"`
	ItsmTicketStatus string      `json:"itsmTicketStatus" bson:"itsmTicketStatus"`
}

// Quota staging quota entity
type Quota struct {
	CPURequests    string `json:"cpuRequests" bson:"cpuRequests"`
	MemoryRequests string `json:"memoryRequests" bson:"memoryRequests"`
	CPULimits      string `json:"cpuLimits" bson:"cpuLimits"`
	MemoryLimits   string `json:"memoryLimits" bson:"memoryLimits"`
}

// Variable staging variable entity
type Variable struct {
	VariableID string `json:"variableID" bson:"variableID"`
	ClusterID  string `json:"clusterID" bson:"clusterID"`
	Namespace  string `json:"namespace" bson:"namespace"`
	Key        string `json:"key" bson:"key"`
	Value      string `json:"value" bson:"value"`
}

// ModelNamespace provide namespace db
type ModelNamespace struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New return a new namespace model instance
func New(db drivers.DB) *ModelNamespace {
	return &ModelNamespace{
		tableName: dbtable.DataTableNamePrefix + tableName,
		indexes:   namespaceIndexes,
		db:        db,
	}
}

// ensureTable xxx
func (m *ModelNamespace) ensureTable(ctx context.Context) error {
	m.isTableEnsuredMutex.RLock()
	if m.isTableEnsured {
		m.isTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := dbtable.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
		m.isTableEnsuredMutex.RUnlock()
		return err
	}
	m.isTableEnsuredMutex.RUnlock()

	m.isTableEnsuredMutex.Lock()
	m.isTableEnsured = true
	m.isTableEnsuredMutex.Unlock()
	return nil
}

// CreateNamespace create namespace
func (m *ModelNamespace) CreateNamespace(ctx context.Context, ns *Namespace) error {
	if ns == nil {
		return fmt.Errorf("namespace cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	ns.CreateTime = time.Now().UTC().Format(time.RFC3339)
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{ns}); err != nil {
		return err
	}
	return nil
}

// UpdateNamespace update namespace
func (m *ModelNamespace) UpdateNamespace(ctx context.Context, ns entity.M) (*Namespace, error) {
	if ns == nil {
		return nil, fmt.Errorf("can not update empty namespace")
	}

	if ns.GetString(FieldKeyProjectCode) == "" ||
		ns.GetString(FieldKeyClusterID) == "" ||
		ns.GetString(FieldKeyName) == "" {
		return nil, fmt.Errorf("can not update namespace, no enough arguments")
	}

	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyProjectCode: ns.GetString(FieldKeyProjectCode),
		FieldKeyClusterID:   ns.GetString(FieldKeyClusterID),
		FieldKeyName:        ns.GetString(FieldKeyName),
		FieldKeyIsDeleted:   false,
	})

	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": ns}); err != nil {
		return nil, err
	}
	namespace := &Namespace{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, namespace); err != nil {
		return nil, err
	}
	return namespace, nil
}

// DeleteNamespace delete namespace
func (m *ModelNamespace) DeleteNamespace(ctx context.Context, projectCode, clusterID, namespace string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyProjectCode: projectCode,
		FieldKeyClusterID:   clusterID,
		FieldKeyName:        namespace,
		FieldKeyIsDeleted:   false,
	})
	return m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": operator.M{
		FieldKeyIsDeleted: true,
	}})
}

// GetNamespace get namespace
func (m *ModelNamespace) GetNamespace(ctx context.Context,
	projectCode, clusterID, name string) (*Namespace, error) {
	if projectCode == "" || clusterID == "" || name == "" {
		return nil, fmt.Errorf("can not get namespace, no enough arguments")
	}
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyProjectCode: projectCode,
		FieldKeyClusterID:   clusterID,
		FieldKeyName:        name,
		FieldKeyIsDeleted:   false,
	})

	namespace := &Namespace{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, namespace); err != nil {
		return nil, err
	}
	return namespace, nil
}

// GetNamespaceByItsmTicketType get namespace by ITSM ticket type
func (m *ModelNamespace) GetNamespaceByItsmTicketType(ctx context.Context,
	projectCode, clusterID, name, itsmTicketType string) (*Namespace, error) {
	if projectCode == "" || clusterID == "" || name == "" || itsmTicketType == "" {
		return nil, fmt.Errorf("can not get namespace, no enough arguments")
	}
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyProjectCode:    projectCode,
		FieldKeyClusterID:      clusterID,
		FieldKeyName:           name,
		FieldKeyItsmTicketType: itsmTicketType,
		FieldKeyIsDeleted:      false,
	})

	namespace := &Namespace{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, namespace); err != nil {
		return nil, err
	}
	return namespace, nil
}

// ListNamespaces list all staging namespaces
func (m *ModelNamespace) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	nsList := []Namespace{}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{FieldKeyIsDeleted: false})
	err := m.db.Table(m.tableName).Find(cond).All(ctx, &nsList)
	return nsList, err
}

// ListNamespacesByItsmTicketType list namespaces by staging type
func (m *ModelNamespace) ListNamespacesByItsmTicketType(ctx context.Context,
	projectCode, clusterID string, types []string) ([]Namespace, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	nsList := []Namespace{}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyProjectCode: projectCode,
		FieldKeyClusterID:   clusterID,
		FieldKeyIsDeleted:   false,
	})
	cond = operator.NewBranchCondition(operator.And, cond, operator.NewLeafCondition(
		operator.In, operator.M{FieldKeyItsmTicketType: types}))
	err := m.db.Table(m.tableName).Find(cond).All(ctx, &nsList)
	return nsList, err
}
