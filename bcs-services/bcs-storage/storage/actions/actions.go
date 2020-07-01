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

package actions

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
)

const (
	prefix       = "/bcsstorage/"
	APIVersionV1 = "v1"

	PathV1 = prefix + APIVersionV1
)

type Action httpserver.Action

var apiV1actions = make([]*httpserver.Action, 0, 100)

// Register a handler into v1 actions
// means all the URL of these handlers are start with PathV1
func RegisterV1Action(action Action) {
	apiV1actions = append(apiV1actions, httpserver.NewAction(action.Verb, action.Path, action.Params, action.Handler))
}

// Get V1 actions
func GetApiV1Action() []*httpserver.Action {
	return apiV1actions
}

var daemonFunc = make([]func(), 0, 10)

// called by actions for registering some daemon functions
// and these functions will be called after flag-init and server-start
func RegisterDaemonFunc(f func()) {
	daemonFunc = append(daemonFunc, f)
}

// one by one start daemon goroutine
func StartActionDaemon() {
	for _, f := range daemonFunc {
		go f()
	}
}
