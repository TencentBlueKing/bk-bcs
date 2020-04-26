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

package v1

import (
	"fmt"
	"github.com/bitly/go-simplejson"
)

func parseResponse(resp []byte) (int, string, []byte, error) {
	var code int
	var msg string
	var data []byte
	var err error
	var js *simplejson.Json
	js, err = simplejson.NewJson(resp)
	if err != nil {
		return -1, fmt.Sprintf("decode response failed, raw resp: %s", string(resp)), nil, err
	}

	msg, _ = js.Get("message").String()
	code, err = js.Get("code").Int()
	if err != nil {
		return -1, fmt.Sprintf("decode response failed, raw resp: %s", string(resp)), nil, err
	}

	data, err = js.Get("data").Encode()
	if err != nil {
		return -1, fmt.Sprintf("decode response failed, raw resp: %s", string(resp)), nil, err
	}

	return code, msg, data, err
}
