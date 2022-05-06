/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package component

import (
	"errors"
	"time"

	goReq "github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/stringx"
)

// Request request third api client
func Request(req goReq.SuperAgent, timeout int, proxy string, headers map[string]string) (string, error) {
	client := goReq.New().Timeout(time.Duration(timeout) * time.Second)
	// request by method
	client = client.CustomMethod(req.Method, req.Url)
	// set proxy
	if proxy != "" {
		client = client.Proxy(proxy)
	}
	// set headers
	for key, val := range headers {
		client = client.Set(key, val)
	}
	// request data
	client = client.Send(req.QueryData).Send(req.Data)
	client = client.SetDebug(req.Debug)
	_, body, errs := client.End()

	if len(errs) > 0 {
		logging.Error(
			"request api error, url: %s, method: %s, params: %s, data: %s, err: %v",
			req.Url, req.Method, req.QueryData, req.Data, errs,
		)
		return "", errors.New(stringx.Errs2String(errs))
	}
	return body, nil
}
