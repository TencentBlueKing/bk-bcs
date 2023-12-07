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

package storegw

import (
	"context"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/prober"
	grpcserver "github.com/thanos-io/thanos/pkg/server/grpc"
	"github.com/thanos-io/thanos/pkg/store"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	bcssystem "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bcs_system"
	bkmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bk_monitor"
	prom "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/prometheus"
)

// Store GW Store 基类
type Store struct {
	*config.StoreConf
	Address string
	Server  *grpcserver.Server
	cancel  func()
	ctx     context.Context
	logger  log.Logger
}

// GetStoreSvr 工厂模式
func GetStoreSvr(logger log.Logger, reg *prometheus.Registry, conf *config.StoreConf) (storepb.StoreServer, error) {
	_ = level.Info(logger).Log("msg", "loading store configuration")

	c, err := yaml.Marshal(conf.Config)
	if err != nil {
		return nil, errors.Wrap(err, "marshal content of store configuration")
	}

	switch strings.ToUpper(string(conf.Type)) {
	case string(config.BKMONITOR):
		return bkmonitor.NewBKMonitorStore(c)
	case string(config.BCS_SYSTEM):
		return bcssystem.NewBCSSystemStore(c)
	case string(config.PROMETHEUS):
		return prom.NewPromStore(c)
	default:
		return nil, errors.Errorf("store with type %s is not supported", conf.Type)
	}
}

// NilLogger grpc log, 不打印无效日志
type NilLogger struct{}

// Log :
func (l *NilLogger) Log(keyvals ...interface{}) error {
	return nil
}

// NewStore :
func NewStore(ctx context.Context, logger log.Logger, reg *prometheus.Registry, address string, conf *config.StoreConf,
	storeSvr storepb.StoreServer) (*Store, error) {
	ctx, cancel := context.WithCancel(ctx)

	// 重新初始化, 解决 duplicate metrics collector registration attempted
	_reg := prometheus.NewRegistry()

	grpcProbe := prober.NewGRPC()

	nilLogger := &NilLogger{}

	g := grpcserver.New(nilLogger, _reg, nil, nil, nil, component.Store, grpcProbe,
		grpcserver.WithServer(store.RegisterStoreServer(storeSvr)),
		grpcserver.WithListen(address),
		grpcserver.WithGracePeriod(time.Duration(0)),
	)

	store := &Store{
		StoreConf: conf,
		Server:    g,
		ctx:       ctx,
		cancel:    cancel,
		Address:   address,
		logger:    logger,
	}
	return store, nil
}

// ListenAndServe 启动服务
func (s *Store) ListenAndServe() error {
	return s.Server.ListenAndServe()
}

// Shutdown 关闭服务
func (s *Store) Shutdown(err error) {
	s.cancel()
	s.Server.Shutdown(err)
}
