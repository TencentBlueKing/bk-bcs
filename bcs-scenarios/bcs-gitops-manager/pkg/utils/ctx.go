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
	"strings"
)

// IsUnauthenticated unauthenticated
func IsUnauthenticated(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Unauthenticated")
}

// IsClusterAskCredentials cluster not ready
func IsClusterAskCredentials(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "the server has asked for the client to provide credentials")
}

// IsClusterNotFound check the cluster not found
func IsClusterNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "cluster") && strings.Contains(err.Error(), "not found")
}

// IsClusterRequestTimeout check save cluster timeout
func IsClusterRequestTimeout(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "client.Timeout exceeded while awaiting headers")
}

// IsPermissionDenied permission
func IsPermissionDenied(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "PermissionDenied")
}

// IsContextCanceled confirm the error whether context canceled
func IsContextCanceled(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "context canceled")
}

// IsContextDeadlineExceeded context exceeded
func IsContextDeadlineExceeded(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "context deadline exceeded")
}

// IsAuthenticationFailed auth failed
func IsAuthenticationFailed(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Authentication failed")
}

// IsSecretAlreadyExist defines the secret whether exist
func IsSecretAlreadyExist(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "already exists")
}

// IsArgoResourceNotFound defines the argo resource not found
func IsArgoResourceNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "code = NotFound")
}

// IsNotFound defines is not found
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found")
}

// IsArgoNotFoundAsPartOf defines the resource not found as port of application
func IsArgoNotFoundAsPartOf(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found as part of")
}

// IsUnexpectedEndOfJSON check the error is unexpected json
func IsUnexpectedEndOfJSON(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "unexpected end of JSON input")
}

// NeedRetry defines which error need retry
func NeedRetry(err error) bool {
	return IsContextDeadlineExceeded(err) || IsUnexpectedEndOfJSON(err)
}
