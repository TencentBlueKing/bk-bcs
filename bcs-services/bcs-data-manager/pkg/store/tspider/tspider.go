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

// Package tspider xxx
package tspider

import (
	"github.com/jmoiron/sqlx"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

type server struct {
	*ModelPowerTrading
	*ModelUserOperationData
	*ModelInterface
	*ModelCloudNative
}

// NewServer new db server
// nolint
func NewServer(dbs map[string]*sqlx.DB, bkbaseConf *types.BkbaseConfig) *server {
	return &server{
		ModelCloudNative:       NewModelCloudNative(dbs, bkbaseConf),
		ModelPowerTrading:      NewModelPowerTrading(dbs, bkbaseConf),
		ModelUserOperationData: NewModelUserOperationData(dbs, bkbaseConf),
		ModelInterface:         NewModelInterface(),
	}
}
