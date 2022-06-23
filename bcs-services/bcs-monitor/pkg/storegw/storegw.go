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
 *
 */

package storegw

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// StoreGW Store 管理结构
type StoreGW struct {
	confs           []*config.StoreConf
	stores          map[string]*Store
	logger          log.Logger
	ctx             context.Context
	reg             *prometheus.Registry
	stop            func()
	GRPCAdvertiseIP string
	portRange       *PortRange
}

// NewStoreGW
func NewStoreGW(ctx context.Context, logger log.Logger, reg *prometheus.Registry, gprcAdvertiseIP string, grpcAdvertisePortRangeStr string, confs []*config.StoreConf) (*StoreGW, error) {
	portRange, err := NewPortRange(grpcAdvertisePortRangeStr)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	gw := &StoreGW{
		confs:           confs,
		ctx:             ctx,
		stop:            cancel,
		logger:          logger,
		reg:             reg,
		GRPCAdvertiseIP: gprcAdvertiseIP,
		portRange:       portRange,
		stores:          map[string]*Store{},
	}

	return gw, nil
}

// Run 启动服务
func (s *StoreGW) Run() error {
	for idx, conf := range s.confs {
		logger := log.With(s.logger, "provider", conf.Type, "id", idx)
		port, err := s.portRange.AllocatePort(int64(idx))
		if err != nil {
			return err
		}

		address := fmt.Sprintf("%s:%d", s.GRPCAdvertiseIP, port)

		store, err := NewStore(s.ctx, logger, s.reg, address, conf)
		if err != nil {
			return err
		}

		id := strconv.Itoa(idx)
		s.stores[id] = store
		go func() {
			// 因为阻塞, 另外启动，同时打印日志
			err := store.ListenAndServe()
			if err != nil {
				level.Error(logger).Log("msg", "ListenAndServe grpc server done", "err", err)
				return
			}
			level.Info(logger).Log("msg", "ListenAndServe grpc server done")
		}()
	}

	<-s.ctx.Done()

	return nil
}

// Shutdown
func (s *StoreGW) Shutdown(err error) {
	s.stop()
}
