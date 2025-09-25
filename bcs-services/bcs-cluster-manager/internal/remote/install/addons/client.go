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

// Package addons xxx
package addons

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/install/types"
)

// addonsClient addons client
var addonsClient = &AddonsClient{}

// GetAddonsClient get addon client
func GetAddonsClient() *AddonsClient {
	return addonsClient
}

// AddonsClient client for addons
type AddonsClient struct { // nolint
}

// GetAddonsClient get addons client
func (ac *AddonsClient) GetAddonsClient() (helmmanager.ClusterAddonsClient, func(), error) {
	if ac == nil {
		return nil, nil, types.ErrNotInited
	}

	cli, conn, err := helmmanager.GetClient(common.ClusterManager)
	if err != nil {
		return nil, nil, err
	}
	return cli.ClusterAddonsClient, conn, nil
}
