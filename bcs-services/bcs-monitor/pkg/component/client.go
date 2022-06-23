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

package component

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	resty "github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

const (
	timeout = time.Second * 30
)

var (
	clientOnce   sync.Once
	globalClient *resty.Client
)

func restyToCurl(c *resty.Client, r *resty.Response) error {
	headers := ""
	for key, values := range r.Request.Header {
		for _, value := range values {
			headers += fmt.Sprintf(" -H %q", fmt.Sprintf("%s: %s", key, value))
		}
	}

	reqMsg := fmt.Sprintf("curl -X %s %s%s", r.Request.Method, r.Request.URL, headers)
	if r.Request.Body != nil {
		switch body := r.Request.Body.(type) {
		case []byte:
			reqMsg += fmt.Sprintf(" -d %q", body)
		case string:
			reqMsg += fmt.Sprintf(" -d %q", body)
		case io.Reader:
			reqMsg += fmt.Sprintf(" -d %q (io.Reader)", body)
		default:
			prtBodyBytes, err := json.Marshal(body)
			if err != nil {
				klog.Errorf("marshal json, %s", err)
			} else {
				reqMsg += fmt.Sprintf(" -d '%s'", prtBodyBytes)
			}
		}
	}

	klog.Infof("REQ: %s", reqMsg)

	respMsg := fmt.Sprintf("[%s] %s %s", r.Status(), r.Time(), r.Body())
	if len(respMsg) > 1024 {
		respMsg = respMsg[:1024] + fmt.Sprintf("...(Total %s)", humanize.Bytes(uint64(len(respMsg))))
	}
	klog.Infof("RESP: %s", respMsg)

	return nil
}

// GetClient
func GetClient() *resty.Client {
	if globalClient == nil {
		clientOnce.Do(func() {
			globalClient = resty.New().SetTimeout(timeout)
			globalClient = globalClient.SetDebug(false) // 更多详情, 可以开启为 true
			globalClient.SetDebugBodyLimit(1024)
			globalClient.OnAfterResponse(restyToCurl)
			globalClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
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
