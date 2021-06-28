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

package v4

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// GetOffer get offer from scheduler
func (bs *bcsScheduler) GetOffer(clusterID string) ([]*types.OfferWithDelta, error) {
	return bs.getOffer(clusterID)
}

func (bs *bcsScheduler) getOffer(clusterID string) ([]*types.OfferWithDelta, error) {
	resp, err := bs.requester.Do(
		fmt.Sprintf(bcsSchedulerOfferURI, bs.bcsAPIAddress),
		http.MethodGet,
		nil,
		getClusterIDHeader(clusterID),
	)

	if err != nil {
		return nil, err
	}

	code, msg, data, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	if code != 0 {
		return nil, fmt.Errorf("get offer failed: %s", msg)
	}

	var result []*types.OfferWithDelta
	err = codec.DecJson(data, &result)
	return result, err
}
