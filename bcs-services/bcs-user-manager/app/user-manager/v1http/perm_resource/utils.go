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

package perm_resource

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/iam"

	"github.com/emicklei/go-restful"
	"github.com/rs/xid"
)

var (
	// ErrServerNotInit err server not init
	ErrServerNotInit = errors.New("server not init")
)

const (
	// IamPathKey show resource common attr "_bk_iam_path_"
	IamPathKey = "_bk_iam_path_"
	// idField show instance id
	idField = "id"
	// UnusedPrefix prefix
	UnusedPrefix = "0000"
)

// GenerateRequestID build request-id
func GenerateRequestID() string {
	id := xid.New()
	return fmt.Sprintf("%s%s", UnusedPrefix, id.String())
}

// GetUserManagerRequestID get usermanager request-id
func GetUserManagerRequestID(header http.Header) string {
	return header.Get(BCSHTTPUserManagerRequestID)
}

// CreateResEntity return response data
func CreateResEntity(response *restful.Response, data interface{}) {
	rsp := iam.SystemResponse{
		BaseResponse: iam.BaseResponse{
			Code:    SuccessCode,
			Message: SuccessMessage,
		},
		Data: data,
	}

	_ = response.WriteAsJson(rsp)
	return
}

// CreateResError return response err
func CreateResError(response *restful.Response, code int, message string) {
	rsp := iam.BaseResponse{
		Code:    code,
		Message: message,
	}
	_ = response.WriteAsJson(rsp)
	return
}

// IsStringExistSlice check from in slice dst
func IsStringExistSlice(from string, dst []string) bool {
	if len(dst) == 0 {
		return false
	}

	if len(from) == 0 {
		return false
	}

	for i := range dst {
		if strings.EqualFold(from, dst[i]) {
			return true
		}
	}

	return false
}

// RemoveDuplicateElement remove duplicate data
func RemoveDuplicateElement(data []string) []string {
	result := make([]string, 0, len(data))
	temp := map[string]struct{}{}
	for _, item := range data {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

