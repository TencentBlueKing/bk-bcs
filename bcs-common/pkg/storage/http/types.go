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

package http

import (
	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

//Response basic response from http api
type Response struct {
	Code    int             `json:"code"`           //operation code
	Message string          `json:"message"`        //response message
	Data    json.RawMessage `json:"data,omitempty"` //response data
}

//WatchResponse basic response from http api
type WatchResponse struct {
	Code    int    `json:"code"`           //operation code
	Message string `json:"message"`        //response message
	Data    *Event `json:"data,omitempty"` //response data
}

//Event for http watch event
type Event struct {
	Type watch.EventType `json:"type"`
	Data json.RawMessage `json:"data"`
}
