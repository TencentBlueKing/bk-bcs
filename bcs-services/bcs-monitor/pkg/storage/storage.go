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

package storage

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/logcollector"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// Storage 提供了数据库操作的接口
type Storage interface {
	// LogCollector operation
	CreateLogCollector(ctx context.Context, lc *entity.LogCollector) error
	UpdateLogCollector(ctx context.Context, id string, lc entity.M) error
	DeleteLogCollector(ctx context.Context, id string) error
	ListLogCollectors(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (
		int64, []*entity.LogCollector, error)
	GetLogCollector(ctx context.Context, id string) (*entity.LogCollector, error)
	// GetIndexSetID return stdIndexSetID and fileIndexSetID
	GetIndexSetID(ctx context.Context, projectID, clusterID string) (int, int, error)
	CreateOldIndexSetID(ctx context.Context, logIndex *entity.LogIndex) error
	GetOldIndexSetID(ctx context.Context, projectID string) (*entity.LogIndex, error)
}

type modelSet struct {
	*logcollector.ModelLogCollector
}

// New return a new ResourceManagerModel instance
func New(db drivers.DB) Storage {
	return &modelSet{
		ModelLogCollector: logcollector.New(db),
	}
}
