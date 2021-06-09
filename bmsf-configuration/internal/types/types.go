/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"errors"
	"time"
)

var (
	// ErrorTimeout is normal timeout error.
	ErrorTimeout = errors.New("timeout")
)

const (
	// AppInstanceOfflineMaxTimeout is the max timeout that means the app instance
	// offline without database session flush.
	AppInstanceOfflineMaxTimeout = 30 * time.Minute

	// AppInstanceFlushDBSessionInterval is app instance flush session in database interval.
	AppInstanceFlushDBSessionInterval = 10 * time.Minute
)

const (
	// RPCLargeTimeout is a very long timeout, used for inner rpc call.
	RPCLargeTimeout = 30 * time.Minute

	// RPCLongTimeout is a long timeout, used for inner rpc call.
	RPCLongTimeout = 10 * time.Minute

	// RPCMiddleTimeout is a middle long timeout, used for inner rpc call.
	RPCMiddleTimeout = 3 * time.Minute

	// RPCNormalTimeout is a normal timeout, used for inner rpc call.
	RPCNormalTimeout = 60 * time.Second

	// RPCShortTimeout is a short timeout, used for inner rpc call.
	RPCShortTimeout = 10 * time.Second

	// RPCTinyTimeout is a tiny timeout, used for inner rpc call.
	RPCTinyTimeout = 3 * time.Second
)

const (
	// EffectCodePending is pending effect code.
	EffectCodePending = 0

	// EffectCodeSuccess is success effect code.
	EffectCodeSuccess = 1

	// EffectCodeFailed is failed effect code.
	EffectCodeFailed = -1

	// EffectCodeTimeout is timeout effect code.
	EffectCodeTimeout = -2

	// EffectCodeOffline is offline effect code.
	EffectCodeOffline = -3

	// EffectMsgPending is effect message pending.
	EffectMsgPending = "PENDING"

	// EffectMsgSuccess is effect message success.
	EffectMsgSuccess = "SUCCESS"

	// EffectMsgTimeout is effect message timeout.
	EffectMsgTimeout = "TIMEOUT"

	// EffectMsgOffline is effect message offline.
	EffectMsgOffline = "OFFLINE"
)

const (
	// ReloadCodePending is pending reload code.
	ReloadCodePending = 0

	// ReloadCodeSuccess is success reload code.
	ReloadCodeSuccess = 1

	// ReloadCodeRollbackSuccess is rollback success reload code.
	ReloadCodeRollbackSuccess = 2

	// ReloadCodeFailed is failed reload code.
	ReloadCodeFailed = -1

	// ReloadCodeTimeout is timeout reload code.
	ReloadCodeTimeout = -2

	// ReloadCodeOffline is offline reload code.
	ReloadCodeOffline = -3

	// ReloadMsgPending is reload message pending.
	ReloadMsgPending = "PENDING"

	// ReloadMsgSuccess is reload message success.
	ReloadMsgSuccess = "SUCCESS"

	// ReloadMsgRollbackSuccess is reload message success.
	ReloadMsgRollbackSuccess = "ROLLBACK SUCCESS"

	// ReloadMsgTimeout is reload message timeout.
	ReloadMsgTimeout = "TIMEOUT"

	// ReloadMsgOffline is reload message offline.
	ReloadMsgOffline = "OFFLINE"
)
