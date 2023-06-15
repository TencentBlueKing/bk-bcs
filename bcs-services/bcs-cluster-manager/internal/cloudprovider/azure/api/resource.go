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

package api

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

/*
	资源
*/

// ListResourceByLocation 从区域获取资源
func (aks *AksServiceImpl) ListResourceByLocation(ctx context.Context, location string) ([]*armcompute.ResourceSKU,
	error) {
	resp := make([]*armcompute.ResourceSKU, 0)
	pager := aks.resourceClient.NewListPager(&armcompute.ResourceSKUsClientListOptions{
		Filter: to.Ptr(fmt.Sprintf("location eq '%s'", location)),
	})
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		resp = append(resp, nextResult.Value...)
	}
	return resp, nil
}
