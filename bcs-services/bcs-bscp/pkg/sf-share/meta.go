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

package sfs

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/metadata"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

// SidecarMetaHeader defines the metadata stored at the request header
// from sidecar to feed server.
type SidecarMetaHeader struct {
	BizID       uint32 `json:"bid"`
	Fingerprint string `json:"fpt"`
}

// Validate the sidecar meta header is valid or not.
func (sm SidecarMetaHeader) Validate() error {
	if sm.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if len(sm.Fingerprint) == 0 {
		return errors.New("invalid fingerprint")
	}

	return nil
}

// IncomingMeta defines metadata parsed from incoming request from sidecar.
type IncomingMeta struct {
	Kit  *kit.Kit
	Meta *SidecarMetaHeader
}

// ParseFeedIncomingContext parse metadata from the feed server's incoming request context which is fired from sidecar.
func ParseFeedIncomingContext(ctx context.Context) (*IncomingMeta, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("invalid incoming context from sidecar")
	}

	var rid string
	if sr := md.Get(constant.SideRidKey); len(sr) != 0 {
		rid = sr[0]
	}

	if len(rid) == 0 {
		return nil, errors.New("invalid request without 'rid' header from sidecar")
	}

	var metaHeader string
	if sm := md.Get(constant.SidecarMetaKey); len(sm) != 0 {
		metaHeader = sm[0]
	}

	if len(metaHeader) == 0 {
		return nil, errors.New("invalid request without 'metadata' header from sidecar")
	}

	sm := new(SidecarMetaHeader)
	if err := jsoni.UnmarshalFromString(metaHeader, sm); err != nil {
		return nil, fmt.Errorf("parse sidecar meta failed, err: %v", err)
	}

	if err := sm.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sidecar meta, err: %v", err)
	}

	return &IncomingMeta{
		Kit: &kit.Kit{
			Ctx: ctx,
			Rid: rid,
		},
		Meta: sm,
	}, nil
}
