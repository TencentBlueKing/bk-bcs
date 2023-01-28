/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package service NOTES
package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/dao"
	"bscp.io/pkg/metrics"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/thirdparty/esb/client"
)

// Service do all the data service's work
type Service struct {
	dao     dao.Set
	gateway *gateway
	// esb esb api client.
	esb client.Client
}

// NewService create a service instance.
func NewService(sd serviced.Service, daoSet dao.Set) (*Service, error) {
	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}
	gateway, err := newGateway(state, daoSet)
	if err != nil {
		return nil, fmt.Errorf("new gateway failed, err: %v", err)
	}

	// initialize esb client
	settings := cc.DataService().Esb
	esbCli, err := client.NewClient(&settings, metrics.Register())
	if err != nil {
		return nil, err
	}

	svc := &Service{
		dao:     daoSet,
		gateway: gateway,
		esb:     esbCli,
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
