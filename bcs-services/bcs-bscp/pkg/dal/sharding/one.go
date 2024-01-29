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

package sharding

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// One represent one mysql sharding instance.
// Note:
// You must call Err() to test the error BEFORE you use
// the 'One' to do anything.
type One struct {
	// hitErr is generated when the 'One' instance is get
	// from sharding.
	// Attention:
	// 1. THIS ERROR SHOULD BE TESTED BEFORE YOU USE 'One'
	//    TO DO ANYTHING.
	// 2. If hitErr != nil, then shardingUid will be "" and
	//    db is nil.
	hitErr error

	// this sharding db instance's unique id, which is
	// generated when the process is launched.
	shardingUid string
	db          *sqlx.DB
}

// ShardingUid return uid
func (o *One) ShardingUid() string {
	return o.shardingUid
}

// Err return am error if something wrong happens, then
// DO NOT USE the 'One' to do anything.
func (o *One) Err() error {
	return o.hitErr
}

// DB return DB instance
func (o *One) DB() *sqlx.DB {
	return o.db
}

// BeginTx return DB instance's transaction.
func (o *One) BeginTx(kit *kit.Kit) (*Tx, error) {
	txn, err := o.db.BeginTxx(kit.Ctx, new(sql.TxOptions))
	if err != nil {
		return nil, err
	}

	tx := &Tx{
		tx:          txn,
		shardingUid: o.shardingUid,
	}
	return tx, nil
}

// TxnFunc is a callback function to process logic tasks
// between a transaction.
type TxnFunc func(txn *sqlx.Tx, opt *TxnOption) error

// TxnOption defines all the options to do distributed
// transaction in the AutoTxn processes.
type TxnOption struct {
	// ShardingUid is the unique id of a mysql sharding instance.
	// which means a same sharding instance have the same uid.
	// It is used to test if the mysql instance is the same instance
	// in a distributed transaction, such as create an app and save
	// its audit log.
	ShardingUid string
}

// Validate transaction option
func (t TxnOption) Validate() error {
	if len(t.ShardingUid) == 0 {
		return errors.New("invalid txn option sharding uid")
	}

	return nil
}

// ErrRetryTransaction defines errors that need to retry transaction, like deadlock error in upsert scenario
var ErrRetryTransaction = errors.New("RETRY TRANSACTION ERROR")

// AutoTxn is a wrapper to do all the transaction operations as follows:
// 1. auto launch the transaction
// 2. process the logics, which is a callback run function
// 3. rollback the transaction if 'run' hit an error automatically.
// 4. commit the transaction if no error happens.
func (o *One) AutoTxn(kit *kit.Kit, run TxnFunc) error {
	if o.hitErr != nil {
		return o.hitErr
	}

	if run == nil {
		return errors.New("transaction function is nil")
	}

	retry, err := o.autoTxn(kit, run)
	if err != nil {
		return err
	}

	if !retry {
		return nil
	}

	// if the operation need to retry, retry for at most 3 times, each wait for 50~500ms
	for retryCount := 1; retryCount <= 3; retryCount++ {
		logs.Warnf("retry transaction, retry count: %d, rid: %s", retryCount, kit.Rid)

		max := big.NewInt(450)
		var randomNumber *big.Int
		randomNumber, err = rand.Int(rand.Reader, max)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * time.Duration(randomNumber.Int64()+50))

		retry, err = o.autoTxn(kit, run)
		if err != nil {
			return err
		}

		if !retry {
			return nil
		}

		// do next retry
	}

	logs.Warnf("retry transaction exceeds maximum count, **skip**, rid: %s", kit.Rid)
	return err
}

func (o *One) autoTxn(kit *kit.Kit, run TxnFunc) (bool, error) {
	if o.hitErr != nil {
		return false, o.hitErr
	}

	if run == nil {
		return false, errors.New("transaction function is nil")
	}

	txn, err := o.db.BeginTxx(kit.Ctx, new(sql.TxOptions))
	if err != nil {
		return false, fmt.Errorf("auto txn, but begin txn failed, err: %v", err)
	}

	opt := &TxnOption{
		ShardingUid: o.shardingUid,
	}
	if err := run(txn, opt); err != nil {
		if rollErr := txn.Rollback(); rollErr != nil {
			logs.ErrorDepthf(1, "run sharding one transaction rollback failed, err: %v, rid: %v", rollErr, kit.Rid)
			// do not return error. the transaction will be aborted automatically after timeout.
			// mysql transaction's default timeout is 50s.
		}

		if err == ErrRetryTransaction {
			return true, err
		}

		return false, err
	}

	if err := txn.Commit(); err != nil {
		return false, fmt.Errorf("commit sharding transaction failed, err: %v", err)
	}

	return false, nil
}
