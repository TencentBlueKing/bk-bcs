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
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"

	req "github.com/imroc/req/v3"
	"github.com/pkg/errors"
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

type BKResult struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (r *BKResult) IsOK() error {
	var resultCode int

	switch code := r.Code.(type) {
	case int:
		resultCode = code
	case float64:
		resultCode = int(code)
	case string:
		c, err := strconv.Atoi(code)
		if err != nil {
			return err
		}
		resultCode = c
	default:
		return errors.Errorf("conversion to int from %T not supported", code)
	}

	if resultCode != 0 {
		return errors.Errorf("resp code %d != 0, %s", resultCode, r.Message)
	}
	return nil
}

func (r *BKResult) Unmarshal(v interface{}) error {
	// 再次序列化为bytes
	data, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
