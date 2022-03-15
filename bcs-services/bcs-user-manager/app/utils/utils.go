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

package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"

	"github.com/asaskevich/govalidator"
	"gopkg.in/go-playground/validator.v9"
)

// Validate local implementation
var Validate = validator.New()

func init() {
	// Use json tag name instead of the real struct field name
	// Source: https://github.com/go-playground/validator/issues/287
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		return name
	})

	_ = Validate.RegisterValidation("apiserver_addresses", ValidateAPIServerAddresses)
}

// ValidateAPIServerAddresses validates if given string is a valid apiserver addresses list.
// A valid addresses should be a list of URL concated with ';'
func ValidateAPIServerAddresses(fl validator.FieldLevel) bool {
	s := fl.Field().String()

	if s == "" {
		return false
	}
	for _, addr := range strings.Split(s, ";") {
		if !govalidator.IsURL(addr) {
			return false
		}
	}
	return true
}

// FormatValidationError turn the original validation errors into error response, it will only use the FIRST
// errorField to construct the error message.
func FormatValidationError(errList error) *ErrorResponse {
	var message string
	for _, err := range errList.(validator.ValidationErrors) {
		if err.Tag() == "required" {
			message = fmt.Sprintf("errcode: %d, ", common.BcsErrApiBadRequest) + fmt.Sprintf(`field '%s' is required`, err.Field())
			break
		}
		message = fmt.Sprintf("errcode: %d, ", common.BcsErrApiBadRequest) + fmt.Sprintf(`'%s' failed on the '%s' tag`, err.Field(), err.Tag())
	}
	return &ErrorResponse{
		Result:  false,
		Code:    common.BcsErrApiBadRequest,
		Message: message,
		Data:    nil,
	}
}

// CreateResponseData common response
func CreateResponseData(err error, msg string, data interface{}) string {
	var rpyErr error
	if err != nil {
		rpyErr = bhttp.InternalError(common.BcsErrMesosSchedCommon, msg)
	} else {
		rpyErr = errors.New(bhttp.GetRespone(common.BcsSuccess, common.BcsSuccessStr, data))
	}

	//blog.V(3).Infof("createRespone: %s", rpyErr.Error())

	return rpyErr.Error()
}
