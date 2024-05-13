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

package azure

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

var multiplePortInOneTargetGroupError = fmt.Errorf("multiple port in one target group")
var unknownFrontIPConfiguration = fmt.Errorf("unknown front IP configuration")

// isNotFoundError returns true if the error is
// caused by a not found error.
func isNotFoundError(err error) bool {
	var respErr *azcore.ResponseError
	if !errors.As(err, &respErr) {
		return false
	}
	return respErr.StatusCode == http.StatusNotFound || respErr.ErrorCode == "NotFound"
}

// checkRetryableError return true if the error is a retryable error
func checkRetryableError(err error) bool {
	var respErr *azcore.ResponseError
	if !errors.As(err, &respErr) {
		return false
	}
	if respErr.ErrorCode == "RetryableError" {
		blog.Warnf("resource is busy, have a rest for %d second", waitPeriodLBDealing)
		time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
		return true
	}
	return false
}
