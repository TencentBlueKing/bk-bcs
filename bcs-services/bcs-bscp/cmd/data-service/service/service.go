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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/vault"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tmplprocess"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Service do all the data service's work
type Service struct {
	dao     dao.Set
	cs      pbcs.CacheClient
	vault   vault.Set
	gateway *gateway
	// esb esb api client.
	esb      client.Client
	repo     repository.Provider
	tmplProc tmplprocess.TmplProcessor
}

// NewService create a service instance.
func NewService(sd serviced.Service, ssd serviced.ServiceDiscover, daoSet dao.Set, vaultSet vault.Set,
	esb client.Client, repo repository.Provider) (*Service, error) {
	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}
	gateway, err := newGateway(state, daoSet)
	if err != nil {
		return nil, fmt.Errorf("new gateway failed, err: %v", err)
	}

	opts := make([]grpc.DialOption, 0)

	// add dial load balancer.
	opts = append(opts, ssd.LBRoundRobin())

	tls := cc.DataService().Network.TLS

	if !tls.Enable() {
		// dial without ssl
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// dial with ssl.
		// nolint
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return nil, fmt.Errorf("init client set tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	csConn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(cc.CacheServiceName), opts...)
	if err != nil {
		logs.Errorf("dial cache service failed, err: %v", err)
		return nil, errf.New(errf.Unknown, fmt.Sprintf("dial cache service failed, err: %v", err))
	}

	svc := &Service{
		dao:      daoSet,
		vault:    vaultSet,
		gateway:  gateway,
		esb:      esb,
		repo:     repo,
		tmplProc: tmplprocess.NewTmplProcessor(),
		cs:       pbcs.NewCacheClient(csConn),
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
