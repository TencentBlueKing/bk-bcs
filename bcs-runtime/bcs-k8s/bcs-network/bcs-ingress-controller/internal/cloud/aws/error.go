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

package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go"
)

// SDKError is a wrapper for the SDK's Error type.
type SDKError struct {
	isOperationError       *bool
	isExceededAttemptError *bool
	err                    error
}

// Error returns the error message.
func (e *SDKError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

// Unwrap returns the underlying error.
func (e *SDKError) Unwrap() error {
	return e.err
}

// IsOperationError returns true if the error is an aws operation error.
func (e *SDKError) IsOperationError() bool {
	if e.isOperationError == nil || !*e.isOperationError {
		return false
	}
	return true
}

// IsExceededAttemptError returns true if the error is an aws api retry error.
func (e *SDKError) IsExceededAttemptError() bool {
	if e.isExceededAttemptError == nil || !*e.isExceededAttemptError {
		return false
	}
	return true
}

// ResolveError is a wrapper for the SDK's ResolveError function.
// It check the error type.
func ResolveError(err error) *SDKError {
	retErr := SDKError{}
	if err == nil {
		return nil
	}
	retErr.err = err
	if _, ok := err.(*smithy.OperationError); ok {
		retErr.isOperationError = aws.Bool(true)
	}
	if _, ok := err.(*retry.MaxAttemptsError); ok {
		retErr.isExceededAttemptError = aws.Bool(true)
	}
	return &retErr
}
