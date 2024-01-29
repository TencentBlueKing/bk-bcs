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
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	en_translations "gopkg.in/go-playground/validator.v9/translations/en"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

// TokenStatus is a enum for token status.
// nolint
type TokenStatus uint8

const (
	// TokenStatusExpired mean that token is expired.
	TokenStatusExpired TokenStatus = iota
	// TokenStatusActive mean that token is active.
	TokenStatusActive
)

// Validate local implementation
var Validate = validator.New()
var trans ut.Translator

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
	en := en.New()
	uni := ut.New(en, en)

	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	trans, _ = uni.GetTranslator("en")

	_ = en_translations.RegisterDefaultTranslations(Validate, trans)
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
			message = fmt.Sprintf("errcode: %d, ", common.BcsErrApiBadRequest) + fmt.Sprintf(`field '%s' is required`,
				err.Field())
			break
		}
		message = fmt.Sprintf("errcode: %d, ", common.BcsErrApiBadRequest) + fmt.Sprintf(`'%s' failed on the '%s' tag`,
			err.Field(), err.Tag())
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

	// blog.V(3).Infof("createRespone: %s", rpyErr.Error())

	return rpyErr.Error()
}

// StringInSlice returns true if given string in slice
func StringInSlice(s string, l []string) bool {
	for _, objStr := range l {
		if s == objStr {
			return true
		}
	}
	return false
}

// GetProjectFromAttribute get project from attribute
func GetProjectFromAttribute(request *restful.Request) *component.Project {
	project := request.Attribute(constant.ProjectAttr)
	if p, ok := project.(*component.Project); ok {
		return p
	}
	return nil
}

// GetUserFromAttribute get user from attribute
func GetUserFromAttribute(request *restful.Request) *models.BcsUser {
	user := request.Attribute(constant.CurrentUserAttr)
	if p, ok := user.(*models.BcsUser); ok {
		return p
	}
	return nil
}

// ParamsErrorData is the error data for params error
type ParamsErrorData struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ParseValidationError parse validation error
func ParseValidationError(errList error) []ParamsErrorData {
	results := make([]ParamsErrorData, 0)
	// nolint
	switch errs := errList.(type) {
	case validator.ValidationErrors:
		for _, err := range errs {
			results = append(results, ParamsErrorData{
				Field:   err.Field(),
				Message: err.Translate(trans),
			})
		}
	}
	return results
}
