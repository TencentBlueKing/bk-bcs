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
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

const (
	// EventAdd add event
	EventAdd = "add"
	// EventUpdate update event
	EventUpdate = "update"
	// EventDelete delete event
	EventDelete = "delete"
	// EventClose close event
	EventClose = "close"
	// EventError err event
	EventError = "error"
)

// EventType event type for store
type EventType string

// Event event for store watch
type Event struct {
	Type EventType
	Obj  *types.RawObject
}

// Store interface for store object
type Store interface {
	Get(ctx context.Context, t types.ObjectType, key types.ObjectKey, opt *GetOptions) (*types.RawObject, error)
	Create(ctx context.Context, obj *types.RawObject, opt *CreateOptions) error
	Update(ctx context.Context, obj *types.RawObject, opt *UpdateOptions) error
	Delete(ctx context.Context, obj *types.RawObject, opt *DeleteOptions) error
	List(ctx context.Context, objectType types.ObjectType, opts *ListOptions) ([]*types.RawObject, error)
	Watch(ctx context.Context, resourceType types.ObjectType, opts *WatchOptions) (chan *Event, error)
}
