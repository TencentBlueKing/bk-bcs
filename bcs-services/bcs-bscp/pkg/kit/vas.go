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

package kit

import (
	"context"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
)

// NewVas create a vas instance.
func NewVas() *Vas {
	return &Vas{
		Rid: uuid.UUID(),
		Ctx: context.TODO(),
		Wg:  sync.WaitGroup{},
	}
}

// OutgoingVas create a vas instance with pairs for grpc outgoing request.
// these pairs will be set to the context under this vas used by the grpc
// outgoing request.
func OutgoingVas(pairs ...map[string]string) *Vas {
	rid := uuid.UUID()
	md := metadata.Pairs(constant.SideRidKey, rid)

	if len(pairs) != 0 {
		for _, one := range pairs {
			for k, v := range one {
				if k == constant.SideRidKey {
					continue
				}
				md.Set(k, v)
			}
		}
	}

	ctx := metadata.NewOutgoingContext(context.TODO(), md)

	return &Vas{
		Rid: rid,
		Ctx: ctx,
		Wg:  sync.WaitGroup{},
	}
}

// Vas is a simple container to store the basic information for a request.
// It is similar with Kit, but Vas is more lightweight for the bscp system
// 'inner' information delivery.
type Vas struct {
	// Rid is request id.
	Rid string
	// Ctx is request context.
	Ctx context.Context
	// Wg is wait group.
	Wg sync.WaitGroup
}

// Validate the vas is valid or not.
func (v *Vas) Validate() error {
	if v.Ctx == nil {
		return errors.New("vas context is nil")
	}

	if len(v.Rid) == 0 {
		return errors.New("vas rid is empty")
	}

	return nil
}

// WithTimeout return child vas with timeout.
func (v *Vas) WithTimeout(timeout time.Duration) (*Vas, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(v.Ctx, timeout)

	child := &Vas{
		Rid: v.Rid,
		Ctx: ctx,
		Wg:  sync.WaitGroup{},
	}

	return child, cancel
}
