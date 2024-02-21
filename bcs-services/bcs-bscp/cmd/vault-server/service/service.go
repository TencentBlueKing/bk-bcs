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

// Package service provides the functionality for the service layer
package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	pbvs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/vault-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

// Service do all the data service's work
type Service struct {
	gateway *gateway
}

// NewService create a service instance.
func NewService(sd serviced.Discover) (*Service, error) {

	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}
	gateway, err := newGateway(state)
	if err != nil {
		return nil, fmt.Errorf("new gateway failed, err: %v", err)
	}

	s := &Service{
		gateway: gateway,
	}

	return s, nil
}

// Handler return service's handler.
func (s *Service) Handler() (http.Handler, error) {
	if s.gateway == nil {
		return nil, errors.New("gateway is nil")
	}

	return s.gateway.handler(), nil
}

// Ping .
func (s *Service) Ping(ctx context.Context, in *pbvs.PingMsg) (*pbvs.PingMsg, error) {
	return nil, nil
}
