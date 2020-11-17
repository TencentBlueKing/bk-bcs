/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package store

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

// GetOptions options for get operation
type GetOptions struct {
	Env string
}

// CreateOptions options for create operation
type CreateOptions struct {
	// UpdateExists if do update when exists
	UpdateExists bool
	Env          string
}

// UpdateOptions options for update operation
type UpdateOptions struct {
	// CreateNotExists if do creation when not exists
	CreateNotExists bool
	Env             string
}

// DeleteOptions options for delete operation
type DeleteOptions struct {
	// IgnoreNotFound if return err when data not found
	IgnoreNotFound bool
	Env            string
}

// ListOptions options for list operation
type ListOptions struct {
	Selector  *types.ValueSelector
	Cluster   string
	Namespace string
	Offset    int64
	Limit     int64
	Env       string
}

// WatchStartTimeStamp start time stamp for watch
type WatchStartTimeStamp struct {
	T uint32
	I uint32
}

// WatchOptions options for watch operation
type WatchOptions struct {
	Selector     *types.ValueSelector
	BatchSize    int32
	MaxAwaitTime time.Duration
	StartTime    *WatchStartTimeStamp
	Env          string
}
