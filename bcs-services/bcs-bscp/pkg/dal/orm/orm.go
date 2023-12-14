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

// Package orm NOTES
package orm

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	prm "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// DoOrm defines all the orm method.
type DoOrm interface {
	Get(ctx context.Context, dest interface{}, expr string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, expr string, args ...interface{}) error
	Count(ctx context.Context, expr string, args ...interface{}) (uint32, error)
	Delete(ctx context.Context, expr string, args ...interface{}) (int64, error)
	Update(ctx context.Context, expr string, args interface{}) (int64, error)
	Insert(ctx context.Context, expr string, data interface{}) error
	BulkInsert(ctx context.Context, expr string, args interface{}) error
	Exec(ctx context.Context, expr string) (int64, error)
}

// DoOrmWithTransaction defines all the orm method with transaction.
type DoOrmWithTransaction interface {
	Insert(ctx context.Context, expr string, args interface{}) error
	BulkInsert(ctx context.Context, expr string, args interface{}) error
	Delete(ctx context.Context, expr string, args ...interface{}) error
	Update(ctx context.Context, expr string, args interface{}) (int64, error)
}

// Interface defines all the orm related operations.
type Interface interface {
	Do(db *sqlx.DB) DoOrm
	Txn(tx *sqlx.Tx) DoOrmWithTransaction
}

// Do return orm operations.
func Do(opt cc.Sharding) Interface {
	return &runtimeOrm{
		mc:             initMetric(),
		ingressLimiter: rate.NewLimiter(rate.Limit(opt.Limiter.QPS), int(opt.Limiter.Burst)),
		logLimiter:     rate.NewLimiter(50, 25),
		slowLogMS:      time.Duration(opt.MaxSlowLogLatencyMS) * time.Millisecond,
	}
}

type runtimeOrm struct {
	// ingressLimiter the limiter to limit the incoming request frequency.
	// Note: test the accept for each sharding, but not for all the sharding with one limiter.
	ingressLimiter *rate.Limiter
	logLimiter     *rate.Limiter
	mc             *metric
	slowLogMS      time.Duration
}

func (o *runtimeOrm) logSlowCmd(ctx context.Context, sql string, latency time.Duration) {

	if latency < o.slowLogMS {
		return
	}

	if !o.logLimiter.Allow() {
		// if the log rate have already exceeded the limit, then skip the log.
		// we do this to avoid write lots of log to file and slow down the request.
		return
	}

	rid := ctx.Value(constant.RidKey)
	logs.InfoDepthf(2, "[orm slow log], sql: %s, latency: %d ms, rid: %v", sql, latency.Milliseconds(), rid)
}

// tryAccept is used to test if the incoming orm request can be accepted.
// Note: test the accept for each sharding, but not for all the sharding with one limiter.
func (o *runtimeOrm) tryAccept() error {
	if o.ingressLimiter.Allow() {
		return nil
	}

	o.mc.errCounter.With(prm.Labels{"cmd": "limiter"}).Inc()

	// have already oversize the limit
	return errf.New(errf.TooManyRequest, "orm too many requests")
}

// Do create a new orm do instance.
func (o *runtimeOrm) Do(db *sqlx.DB) DoOrm {
	return &do{
		db: db,
		ro: o,
	}
}

// Txn create a new transaction orm instance.
func (o *runtimeOrm) Txn(tx *sqlx.Tx) DoOrmWithTransaction {
	return &doTxn{
		tx: tx,
		ro: o,
	}
}
