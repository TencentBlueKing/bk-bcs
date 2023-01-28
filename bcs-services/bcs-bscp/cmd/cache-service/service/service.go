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
	"errors"
	"fmt"
	"net/http"

	"bscp.io/cmd/cache-service/service/cache/client"
	"bscp.io/cmd/cache-service/service/cache/event"
	"bscp.io/pkg/dal/bedis"
	"bscp.io/pkg/dal/dao"
	"bscp.io/pkg/serviced"
)

// Service do all the cache service's work
type Service struct {
	// dao only use base compress test.
	dao     dao.Set
	op      client.Interface
	gateway *gateway
	state   serviced.State
}

// NewService initial the service instance.
func NewService(sd serviced.State, daoSet dao.Set, bs bedis.Client, op client.Interface) (*Service, error) {
	err := event.Run(daoSet, sd, bs)
	if err != nil {
		return nil, fmt.Errorf("run event handling task failed, err: %v", err)
	}

	gateway, err := newGateway(sd, daoSet, bs)
	if err != nil {
		return nil, fmt.Errorf("new gateway failed, err: %v", err)
	}

	s := &Service{
		dao:     daoSet,
		op:      op,
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
