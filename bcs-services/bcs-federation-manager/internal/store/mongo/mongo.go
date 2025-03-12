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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
)

var _ store.FederationMangerModel = &server{}

const (
	tableNamePrefix            = "federation_manager_"
	federationClusterTableName = "federation_cluster"
	subClusterTableName        = "sub_cluster"
)

const (
	federationClusterIdUniqueKey     = "federation_cluster_id"
	subClusterUniqueKey              = "uid"
	subClusterFederationClusterIdKey = "federation_cluster_id"
	isDeletedKey                     = "is_deleted"
	projectIdKey                     = "project_id"
)

// Public public model set
type Public struct {
	TableName           string
	Indexes             []drivers.Index
	DB                  drivers.DB
	IsTableEnsured      bool
	IsTableEnsuredMutex sync.RWMutex
}

type server struct {
	*ModelFedCluster
	*ModelSubCluster
}

// NewServer create new server
func NewServer(db drivers.DB) store.FederationMangerModel {
	return &server{
		ModelFedCluster: NewModelFedCluster(db),
		ModelSubCluster: NewModelSubCluster(db),
	}
}
