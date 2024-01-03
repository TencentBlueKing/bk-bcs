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

package observer

import (
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Interface defines all the observer support operations.
type Interface interface {
	// IsReady returns 'true' if the observer is already initialized,
	// which means the start cursor is loaded from db successfully.
	IsReady() bool

	// Next return a channel which is only used by caller, and it blocks
	// until a batch of events occurs.
	Next() <-chan []*types.EventMeta

	// CurrentCursor returns the latest consumed event's cursor id which is
	// consumed by the local cache.
	CurrentCursor() uint32

	// LoopInterval return the observer's loop duration to watch the events.
	LoopInterval() time.Duration
}
