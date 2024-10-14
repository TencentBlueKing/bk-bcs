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

// Package service NOTES
package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/vault"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tmplprocess"
)

// Service do all the data service's work
type Service struct {
	dao     dao.Set
	vault   vault.Set
	gateway *gateway
	// esb esb api client.
	esb      client.Client
	repo     repository.Provider
	tmplProc tmplprocess.TmplProcessor
}

// NewService create a service instance.
func NewService(sd serviced.Service, daoSet dao.Set, vaultSet vault.Set, esb client.Client, repo repository.Provider) (
	*Service, error) {
	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}
	gateway, err := newGateway(state, daoSet)
	if err != nil {
		return nil, fmt.Errorf("new gateway failed, err: %v", err)
	}

	svc := &Service{
		dao:      daoSet,
		vault:    vaultSet,
		gateway:  gateway,
		esb:      esb,
		repo:     repo,
		tmplProc: tmplprocess.NewTmplProcessor(),
	}

	return svc, nil
}

// Handler return service's handler.
func (s *Service) Handler() (http.Handler, error) {
	if s.gateway == nil {
		return nil, errors.New("gateway is nil")
	}

	return s.gateway.handler(), nil
}

// Ping data service.
func (s *Service) Ping(ctx context.Context, msg *pbds.PingMsg) (*pbds.PingMsg, error) {
	return &pbds.PingMsg{Data: "pong"}, nil
}
