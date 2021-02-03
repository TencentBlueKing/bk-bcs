/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"encoding/json"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

/*
 * In order to be compatible with the bcs-user-manager interface,
 * we keep the way of that the server returns error from bcs-user-manager
 */

const (
	// ErrorStatusCreateTunnel error status for create tunnel
	ErrorStatusCreateTunnel = "CREATE_TUNNEL_ERROR"
)

var (
	// GroupResourceCluster group resource of cluster
	GroupResourceCluster = schema.GroupResource{
		Group:    "bkbcs.tencent.com",
		Resource: "Clusters",
	}
)

// WriteKubeAPIError writes a standard error response
func WriteKubeAPIError(rw http.ResponseWriter, err *errors.StatusError) {
	payload, _ := json.Marshal(err.ErrStatus)

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(int(err.ErrStatus.Code))
	rw.Write(payload)
}

// NewNotFoundError returns a new error which indicates that the resource of the kind and the name was not found.
func NewNotFoundError(qualifiedResource schema.GroupResource, name string, message string) *errors.StatusError {
	status := errors.NewNotFound(qualifiedResource, name)
	status.ErrStatus.Message = message
	return status
}

// NewForbiddenError return forbidden error for the qualifiedResource
func NewForbiddenError(qualifiedResource schema.GroupResource, name string, err error) *errors.StatusError {
	status := errors.NewForbidden(qualifiedResource, name, err)
	return status
}

// NewInternalError return internal error for the qualifiedResource
func NewInternalError(err error) *errors.StatusError {
	status := errors.NewInternalError(err)
	return status
}

// NewInvalidError returns an error indicating the item is invalid and cannot be processed.
func NewInvalidError(qualifiedKind schema.GroupKind, name string, fieldName string, err error) *errors.StatusError {
	fieldError := field.Error{
		Field:  fieldName,
		Detail: err.Error(),
	}
	errs := field.ErrorList{&fieldError}
	status := errors.NewInvalid(qualifiedKind, name, errs)
	return status
}

// NewUnauthorizedError returns an error indicating the client is not authorized to perform the requested action.
func NewUnauthorizedError(reason string) *errors.StatusError {
	return errors.NewUnauthorized(reason)
}
