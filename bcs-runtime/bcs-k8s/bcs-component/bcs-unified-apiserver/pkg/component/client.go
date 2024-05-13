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

// Package component xxx
package component

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/config"
)

const (
	timeout = time.Second * 30
)

var (
	clientOnce   sync.Once
	globalClient *resty.Client
)

// GetClient xxx
func GetClient() *resty.Client {
	if globalClient == nil {
		clientOnce.Do(func() {
			globalClient = resty.New().SetTimeout(timeout)
			if config.G.Base.RunEnv == config.DevEnv {
				globalClient = globalClient.SetDebug(true)
			}
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

// Pagination 分页配置
type Pagination struct {
	Total    int64 `json:"total"`
	PageSize int64 `json:"pageSize"`
	Offset   int64 `json:"offset"`
}

// BCSStorResult BCS storage 返回, 添加了分页字段
type BCSStorResult struct {
	BKResult   `json:",inline"`
	Pagination `json:",inline"`
}

// UnmarshalBKResult 反序列化为蓝鲸返回规范
func UnmarshalBKResult(resp *resty.Response, data interface{}) error {
	if resp.StatusCode() != http.StatusOK {
		return errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	// 部分接口，如 usermanager 返回的content-type不是json, 需要手动Unmarshal
	bkResult := &BKResult{Data: data}
	if err := json.Unmarshal(resp.Body(), bkResult); err != nil {
		return err
	}

	if err := bkResult.ValidateCode(); err != nil {
		return err
	}

	return nil
}

// UnmarshalBCSStorResult 反序列化为BCS Stor返回规范
func UnmarshalBCSStorResult(resp *resty.Response, data interface{}) (*Pagination, error) {
	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200", resp.StatusCode())
	}

	// 部分接口，如 usermanager 返回的content-type不是json, 需要手动Unmarshal
	bkResult := BKResult{Data: data}
	bcsResult := &BCSStorResult{BKResult: bkResult}
	if err := json.Unmarshal(resp.Body(), bcsResult); err != nil {
		return nil, err
	}

	if err := bcsResult.ValidateCode(); err != nil {
		return nil, err
	}

	return &bcsResult.Pagination, nil
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
