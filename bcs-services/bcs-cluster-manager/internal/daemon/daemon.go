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

// Package daemon for daemon
package daemon

import (
	"context"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// DaemonInterface for run daemon
type DaemonInterface interface { // nolint
	InitDaemon(ctx context.Context)
	Stop()
}

// DaemonOptions options
type DaemonOptions struct {
	// EnableDaemon enable
	EnableDaemon bool
}

// DoFunc func() type
type DoFunc func()

// Daemon for realize daemon
type Daemon struct {
	ctx      context.Context
	cancel   context.CancelFunc
	interval int
	model    store.ClusterManagerModel
	options  DaemonOptions
}

// NewDaemon init daemon
func NewDaemon(interval int, model store.ClusterManagerModel, options DaemonOptions) DaemonInterface {
	ctx, cancel := context.WithCancel(context.Background())

	if interval <= 0 {
		interval = 30
	}

	return &Daemon{
		ctx:      ctx,
		cancel:   cancel,
		model:    model,
		interval: interval,
		options:  options,
	}
}

// InitDaemon init task and run daemon
func (d *Daemon) InitDaemon(ctx context.Context) {
	if !d.options.EnableDaemon {
		blog.Infof("cluster-manager InitDaemon %s", d.options.EnableDaemon)
		return
	}

	wg := sync.WaitGroup{}

	errChan := make(chan error)
	go func() {
		for err := range errChan {
			blog.Infof("InitDaemon error: %v", err)
		}
	}()

	go d.simpleDaemon(ctx, &wg, func() {
		d.reportVpcAvailableIPCount(errChan)
	}, 180)

	go d.simpleDaemon(ctx, &wg, func() {
		d.reportClusterHealthStatus(errChan)
	}, 30)

	go d.simpleDaemon(ctx, &wg, func() {
		d.reportClusterGroupNodeNum(errChan)
	}, 60)

	go d.simpleDaemon(ctx, &wg, func() {
		d.reportMachineryTaskNum(errChan)
	}, 60)

	go d.simpleDaemon(ctx, &wg, func() {
		d.reportClusterCaUsageRatio(errChan)
	}, 300)

	go d.simpleDaemon(ctx, &wg, func() {
		d.reportRegionInsTypeUsage(errChan)
	}, 300)

	wg.Wait()
}

func (d *Daemon) simpleDaemon(ctx context.Context, wg *sync.WaitGroup, exec DoFunc, intervalSecs ...int) {
	t := 10
	if len(intervalSecs) > 0 {
		t = intervalSecs[0]
	}
	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(time.Second * time.Duration(t))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			exec()
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		}
	}
}

// Stop quit all daemon
func (d *Daemon) Stop() {
	if !d.options.EnableDaemon {
		return
	}
	d.cancel()
}
