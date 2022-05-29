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
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thanos-io/thanos/pkg/component"
	"github.com/thanos-io/thanos/pkg/prober"
	grpcserver "github.com/thanos-io/thanos/pkg/server/grpc"
	"github.com/thanos-io/thanos/pkg/store"
	"github.com/thanos-io/thanos/pkg/store/storepb"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	bkmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storegw/bk_monitor"
)

const (
	BKMONITOR config.StoreProvider = "BK_MONITOR"
)

// StoreGW Store 基类
type Store struct {
	*config.StoreConf
	Address string
	Server  *grpcserver.Server
	cancel  func()
	mtx     sync.Mutex
	ctx     context.Context
	logger  log.Logger
}

func GetStoreSvr(logger log.Logger, reg *prometheus.Registry, conf *config.StoreConf) (storepb.StoreServer, error) {
	level.Info(logger).Log("msg", "loading store configuration")

	config, err := yaml.Marshal(conf.Config)
	if err != nil {
		return nil, errors.Wrap(err, "marshal content of bucket configuration")
	}

	switch strings.ToUpper(string(conf.Type)) {
	case string(BKMONITOR):
		return bkmonitor.NewBKMonitorStore(config)
	default:
		return nil, errors.Errorf("store with type %s is not supported", conf.Type)
	}
}

func NewStore(ctx context.Context, logger log.Logger, reg *prometheus.Registry, gprcAdvertiseIP string, conf *config.StoreConf) (*Store, error) {
	storeSvr, err := GetStoreSvr(logger, reg, conf)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	// 重新初始化, 解决 duplicate metrics collector registration attempted
	_reg := prometheus.NewRegistry()

	grpcProbe := prober.NewGRPC()
	address := fmt.Sprintf("%s:%d", gprcAdvertiseIP, 1998)

	g := grpcserver.New(logger, _reg, nil, nil, nil, component.Store, grpcProbe,
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
