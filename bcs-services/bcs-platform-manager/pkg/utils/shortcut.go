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

// Package utils xxx
package utils

import (
	"net/http"

	"github.com/go-chi/render"
)

// SuccessResp is the success response for restful response
type SuccessResp struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

// Render chi render interface implementation
func (s *SuccessResp) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

// ErrResp is the error response for restful response
type ErrResp struct {
	ErrObj   ErrObj `json:"error"`
	HttpCode int    `json:"-"` // http response status code
}

// ErrObj is the error response for restful response
type ErrObj struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Data    interface{}   `json:"data"`
	Details []ErrorDetail `json:"details"`
}

// ErrorDetail is the error detail response for restful response
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ErrResp) Error() string {
	return ""
}

// Render chi render interface implementation
func (e *ErrResp) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HttpCode)
	return nil
}

// ParamsError return params error response
func ParamsError(err error) *ErrResp {
	return &ErrResp{
		HttpCode: 400,
		ErrObj: ErrObj{
			Code:    "INVALID_ARGUMENT",
			Message: "invalid argument",
			Data:    nil,
			Details: []ErrorDetail{
				{Code: "INTERNAL", Message: err.Error()},
			},
		},
	}
}

// DBError return db error response
func DBError(err error) *ErrResp {
	return &ErrResp{
		HttpCode: 500,
		ErrObj: ErrObj{
			Code:    "INTERNAL",
			Message: "database error",
			Data:    nil,
			Details: []ErrorDetail{
				{Code: "INTERNAL", Message: err.Error()},
			},
		},
	}
}

// SystemError return system error response
func SystemError(err error) *ErrResp {
	return &ErrResp{
		HttpCode: 500,
		ErrObj: ErrObj{
			Code:    "INTERNAL",
			Message: "api system error",
			Data:    nil,
			Details: []ErrorDetail{
				{Code: "INTERNAL", Message: err.Error()},
			},
		},
	}
}
