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

package client

import (
	"context"
	"fmt"

	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	urlAvailable = "/available"
)

// Available check if helm manager service is available, nil for yes
func (c *Client) Available(ctx context.Context) error {
	resp, err := c.get(ctx, urlPrefix+urlAvailable, nil, nil)
	if err != nil {
		return err
	}

	var r helmmanager.AvailableResp
	if err = unmarshalPB(resp.Reply, &r); err != nil {
		return err
	}

	if r.GetCode() != resultCodeSuccess {
		return fmt.Errorf("check available get result code %d, message: %s", r.GetCode(), r.GetMessage())
	}

	return nil
}
