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

package utils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

// ValidateResourceLimit 检查limit和request是否合法，并且limit >= request
func ValidateResourceLimit(request string, limit string) error {
	var (
		requestQuantity resource.Quantity
		limitQuantity   resource.Quantity
		err             error
	)

	if limit != "" {
		limitQuantity, err = resource.ParseQuantity(limit)
		if err != nil {
			return fmt.Errorf("limit %s is invalid, err: %s", limit, err)
		}
	}
	if request != "" {
		requestQuantity, err = resource.ParseQuantity(request)
		if err != nil {
			return fmt.Errorf("request %s is invalid, err: %s", request, err)
		}
	}
	if limit != "" && request != "" && limitQuantity.Cmp(requestQuantity) < 0 {
		return fmt.Errorf("limit %s must be greater than or equal to request %s", limit, request)
	}

	return nil
}
