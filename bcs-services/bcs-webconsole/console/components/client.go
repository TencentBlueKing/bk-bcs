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

package components

import (
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	req "github.com/imroc/req/v3"
)

const (
	userAgent = "bcs-webconsole"
	timeout   = time.Second * 30
)

var (
	clientOnce   sync.Once
	globalClient *req.Client
)

// GetClient
func GetClient() *req.Client {
	if globalClient == nil {
		clientOnce.Do(func() {
			globalClient = req.C().SetUserAgent(userAgent).SetTimeout(timeout)
			if config.G.Base.RunEnv == config.DevEnv {
				globalClient.EnableDumpAll().EnableDebugLog().EnableTraceAll()
			}
		})
	}
	return globalClient
}
