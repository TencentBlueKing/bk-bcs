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

// Package clustermanager xxx
package clustermanager

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// GetCloud get cloud from cluster manager
func GetCloud(ctx context.Context,
	req *clustermanager.GetCloudRequest) (*clustermanager.Cloud, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.GetCloud(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetCloud error: %s", err)
	}
	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("GetCloud error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p.Data, nil
}
