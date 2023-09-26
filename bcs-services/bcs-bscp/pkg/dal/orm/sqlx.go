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

package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	prm "github.com/prometheus/client_golang/prometheus"
)

var (
	// ErrRecordNotFound returns a "record not found error".
	// Occurs only when attempting to query the database with a struct,
	// querying with a slice won't return this error.
	ErrRecordNotFound = sql.ErrNoRows
	// ErrDeadLock concurrent exec deadlock, error message returned by db.
	ErrDeadLock = "Error 1213: Deadlock found when trying to get lock"
)

var _ DoOrm = new(do)

type do struct {
	db *sqlx.DB
	ro *runtimeOrm
}

// Get one data and decode into dest *struct{}.
func (do *do) Get(ctx context.Context, dest interface{}, expr string, args ...interface{}) error {
	if err := do.ro.tryAccept(); err != nil {
		return err
	}

	start := time.Now()

	err := do.db.GetContext(ctx, dest, expr, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "get"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "get"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Select a collection of data, and decode into dest *[]struct{}.
func (do *do) Select(ctx context.Context, dest interface{}, expr string, args ...interface{}) error {
	if err := do.ro.tryAccept(); err != nil {
		return err
	}

	start := time.Now()

	expr, args, err := sqlx.In(expr, args...)
	if err != nil {
		return err
	}

	err = do.db.SelectContext(ctx, dest, expr, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "select"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "select"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Count the number of the filtered resource.
func (do *do) Count(ctx context.Context, expr string, args ...interface{}) (uint32, error) {
	if err := do.ro.tryAccept(); err != nil {
		return 0, err
	}

	start := time.Now()

	count := uint32(0)
	if err := do.db.GetContext(ctx, &count, expr, args...); err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "count"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "count"}).Observe(float64(time.Since(start).Milliseconds()))

	return count, nil
}

// Delete a collection of data.
func (do *do) Delete(ctx context.Context, expr string, args ...interface{}) (int64, error) {
	if err := do.ro.tryAccept(); err != nil {
		return 0, err
	}

	start := time.Now()

	expr, args, err := sqlx.In(expr, args...)
	if err != nil {
		return 0, err
	}

	result, err := do.db.ExecContext(ctx, expr, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "delete"}).Observe(float64(time.Since(start).Milliseconds()))

	return rowsAffected, nil
}

// Update a collection of data
func (do *do) Update(ctx context.Context, expr string, args interface{}) (int64, error) {
	if args == nil {
		return 0, errors.New("update args is required")
	}

	if err := do.ro.tryAccept(); err != nil {
		return 0, err
	}

	start := time.Now()

	result, err := do.db.NamedExecContext(ctx, expr, args)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "update"}).Observe(float64(time.Since(start).Milliseconds()))

	return rowsAffected, nil
}

// Insert a row data to db
func (do *do) Insert(ctx context.Context, expr string, data interface{}) error {
	if err := do.ro.tryAccept(); err != nil {
		return err
	}

	start := time.Now()

	_, err := do.db.ExecContext(ctx, expr, data)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "insert"}).Inc()
		return fmt.Errorf("insert failed, err: %v", err)
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "insert"}).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

// Exec a command
func (do *do) Exec(ctx context.Context, expr string) (int64, error) {
	if err := do.ro.tryAccept(); err != nil {
		return 0, err
	}

	start := time.Now()

	result, err := do.db.ExecContext(ctx, expr)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "exec"}).Inc()
		return 0, fmt.Errorf("exec failed, err: %v", err)
	}

	effected, err := result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "exec"}).Inc()
		return 0, fmt.Errorf("exec failed, err: %v", err)
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "exec"}).Observe(float64(time.Since(start).Milliseconds()))

	return effected, nil
}

// BulkInsert insert multiple data at one time.
func (do *do) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	if err := do.ro.tryAccept(); err != nil {
		return err
	}

	start := time.Now()

	q, arrayArgs, err := bindArray(sqlx.BindType(do.db.DriverName()), expr, args, do.db.Mapper)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "bulk-insert"}).Inc()
		return err
	}

	_, err = do.db.ExecContext(ctx, q, arrayArgs...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "bulk-insert"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "bulk-insert"}).Observe(float64(time.Since(start).Milliseconds()))

	return err
}

var _ DoOrmWithTransaction = new(doTxn)

type doTxn struct {
	tx *sqlx.Tx
	ro *runtimeOrm
}

// Delete a collection of data with transaction.
func (do *doTxn) Delete(ctx context.Context, expr string, args ...interface{}) error {
	if err := do.ro.tryAccept(); err != nil {
		return err
	}

	start := time.Now()

	expr, args, err := sqlx.In(expr, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return err
	}

	_, err = do.tx.ExecContext(ctx, expr, args...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "delete"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "delete"}).Observe(float64(time.Since(start).Milliseconds()))

	return err
}

// Insert data with transaction
func (do *doTxn) Insert(ctx context.Context, expr string, args interface{}) error {
	if args == nil {
		return errors.New("insert args is required")
	}

	if err := do.ro.tryAccept(); err != nil {
		return err
	}

	start := time.Now()

	_, err := do.tx.NamedExecContext(ctx, expr, args)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "insert"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "insert"}).Observe(float64(time.Since(start).Milliseconds()))

	return err
}

// BulkInsert insert data batch with transaction
func (do *doTxn) BulkInsert(ctx context.Context, expr string, args interface{}) error {
	if err := do.ro.tryAccept(); err != nil {
		return err
	}

	start := time.Now()

	q, arrayArgs, err := bindArray(sqlx.BindType(do.tx.DriverName()), expr, args, do.tx.Mapper)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "bulk-insert"}).Inc()
		return err
	}

	_, err = do.tx.ExecContext(ctx, q, arrayArgs...)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "bulk-insert"}).Inc()
		return err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "bulk-insert"}).Observe(float64(time.Since(start).Milliseconds()))

	return err
}

// Update with transaction
func (do *doTxn) Update(ctx context.Context, expr string, args interface{}) (int64, error) {
	if args == nil {
		return 0, errors.New("update args is required")
	}

	if err := do.ro.tryAccept(); err != nil {
		return 0, err
	}

	start := time.Now()

	result, err := do.tx.NamedExecContext(ctx, expr, args)
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		do.ro.mc.errCounter.With(prm.Labels{"cmd": "update"}).Inc()
		return 0, err
	}

	do.ro.logSlowCmd(ctx, expr, time.Since(start))
	do.ro.mc.cmdLagMS.With(prm.Labels{"cmd": "update"}).Observe(float64(time.Since(start).Milliseconds()))

	return rowsAffected, nil
}
