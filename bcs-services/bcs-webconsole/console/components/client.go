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
	"net/http"
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
			globalClient = req.C().SetTimeout(timeout)
			if config.G.Base.RunEnv == config.DevEnv {
				globalClient = globalClient.DevMode()
			}
			// DevMode() 会设置 UserAgent 为浏览器行为, 在 APISix 会被校验登入态, 这里需要覆盖
			globalClient.SetUserAgent(userAgent)
		})
	}
	return globalClient
}

// BKResult 蓝鲸返回规范的结构体
type BKResult struct {
	Code    interface{} `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UnmarshalBKResult 反序列化为蓝鲸返回规范
func UnmarshalBKResult(resp *req.Response, data interface{}) error {
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("http code %d != 200", resp.StatusCode)
	}

	// 部分接口，如 usermanager 返回的content-type不是json, 需要手动Unmarshal
	bkResult := &BKResult{Data: data}
	if err := resp.UnmarshalJson(bkResult); err != nil {
		return err
	}

	if err := bkResult.ValidateCode(); err != nil {
		return err
	}

	return nil
}

// ValidateCode 返回结果是否OK
func (r *BKResult) ValidateCode() error {
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
