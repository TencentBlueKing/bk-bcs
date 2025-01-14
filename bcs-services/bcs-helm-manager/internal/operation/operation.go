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

// Package operation xxx
package operation

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/metrics"
)

// Operation ensure release install/upgrade/uninstall/rollback action to be executed
type Operation interface {
	// Action operation action name
	Action() string
	// Name operation name
	Name() string
	// Validate check param is valid
	Validate(ctx context.Context) error
	// Prepare do something to prepare release execute, like download chart content
	Prepare(ctx context.Context) error
	// Execute execute release install/upgrade/uninstall/rollback action
	Execute(ctx context.Context) error
	// Done do something for done operation
	Done(err error)
}

const (
	operateInit    = "init"
	operateSuccess = "success"
	operateFail    = "fail"
)

// GlobalOperator global operator
var GlobalOperator = &operator{
	operationCount: common.GetInt32P(0),
	terminate:      make(chan struct{}),
	pause:          common.GetUint32P(0),
	once:           sync.Once{},
}

// Operator operator
type operator struct {
	operationCount *int32
	// quit all operation when got terminate signal
	terminate chan struct{}
	// pause when operator can't add operation, 0 means unpause, 1 means pause, default is 0
	pause *uint32
	once  sync.Once
}

func (o *operator) inc() {
	atomic.AddInt32(o.operationCount, 1)
}

func (o *operator) dec() {
	atomic.AddInt32(o.operationCount, -1)
}

func (o *operator) isPause() bool {
	return atomic.LoadUint32(o.pause) == 1
}

// GetOperationCount get operation count
func (o *operator) GetOperationCount() int32 {
	return atomic.LoadInt32(o.operationCount)
}

// TerminateOperation terminate operation
func (o *operator) TerminateOperation() {
	// pause operator
	atomic.AddUint32(o.pause, 1)
	for i := 0; i < int(o.GetOperationCount()); i++ {
		o.terminate <- struct{}{}
	}
}

// WaitTerminate wait all operation to exit
func (o *operator) WaitTerminate(ctx context.Context, ttl time.Duration) {
	ticker := time.NewTicker(ttl)
	for o.GetOperationCount() != 0 {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			continue
		}
	}
	blog.Info("all operations exit")
}

// ReportOperatorCount report operator count loop
func (o *operator) ReportOperatorCount() {
	o.once.Do(
		func() {
			go func() {
				for range time.Tick(time.Second) {
					metrics.ReportOperationCountMetric(o.GetOperationCount())
				}
			}()
		},
	)
}

func (o *operator) Dispatch(op Operation, timeout time.Duration) (<-chan struct{}, error) {
	if o.isPause() {
		return nil, fmt.Errorf("can't operate release, program is exiting")
	}
	done := make(chan struct{}, 1)
	go o.dispatch(op, timeout, done)
	return done, nil
}

func (o *operator) dispatch(op Operation, timeout time.Duration, done chan struct{}) {
	start := time.Now()
	o.inc()

	metrics.ReportOperationMetric(op.Action(), operateInit, start)
	blog.Infof("operation %s dispatched, timeout: %s", op.Name(), timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// run operation
	go func() {
		defer func() {
			// 防止部署过程 panic 导致整个程序都挂掉，同时 panic 后把对应 release 设置为失败状态
			if r := recover(); r != nil {
				op.Done(fmt.Errorf("operation error, %v", r))
			}
			o.dec()
			cancel()
			done <- struct{}{}
		}()
		if err := op.Prepare(ctx); err != nil {
			metrics.ReportOperationMetric(op.Action(), operateFail, start)
			blog.Errorf("operation %s prepare error, %s", op.Name(), err.Error())
			op.Done(fmt.Errorf("prepare error, %s", err.Error()))
			return
		}
		if err := op.Validate(ctx); err != nil {
			metrics.ReportOperationMetric(op.Action(), operateFail, start)
			blog.Errorf("operation %s validate error, %s", op.Name(), err.Error())
			op.Done(fmt.Errorf("validate error, %s", err))
			return
		}
		if err := op.Execute(ctx); err != nil {
			metrics.ReportOperationMetric(op.Action(), operateFail, start)
			blog.Errorf("operation %s execute error %s", op.Name(), err.Error())
			op.Done(fmt.Errorf("execute error, %s", err.Error()))
			return
		}
		op.Done(nil)
		metrics.ReportOperationMetric(op.Action(), operateSuccess, start)
		blog.Infof("operation %s done", op.Name())
	}()

	// wait whole operation terminate signal or timeout signal
	select {
	case <-o.terminate:
		blog.Infof("operation %s got terminate signal, terminating...", op.Name())
		return
	case <-ctx.Done():
		if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
			metrics.ReportOperationMetric(op.Action(), operateFail, start)
			blog.Errorf("operation %s timeout, %s", op.Name(), err.Error())
			op.Done(err)
			return
		}
		blog.Infof("operation dispatch %s exit", op.Name())
		return
	}
}
