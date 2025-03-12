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

// Package thirdparty
package thirdparty

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/bcsapi/thirdparty-service"
)

// UpdateQuotaInfoForTaiji request thirdparty service
func (t *thirdpartyClient) UpdateQuotaInfoForTaiji(req *third.UpdateQuotaInfoForTaijiRequest) error {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.UpdateQuotaInfoForTaiji(t.getMetadataCtx(ctx), req)
	if err != nil {
		blog.Errorf(fmt.Sprintf("UpdateQuotaInfoForTaiji failed req: %+v, err: %s,", req, err.Error()))
		return err
	}

	blog.Infof(fmt.Sprintf("UpdateQuotaInfoForTaiji req: %+v, resp: %+v,", req, resp))
	if resp.Error != nil && resp.Error.Code != ResultSuccessKey {
		return fmt.Errorf("UpdateQuotaInfoForTaiji failed, err: %+v", resp.Error)
	}

	return nil
}

// UpdateQuotaInfoForSuanli request thirdparty service
func (t *thirdpartyClient) UpdateQuotaInfoForSuanli(req *third.UpdateNamespaceForSuanliRequest) error {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.UpdateNamespaceForSuanli(t.getMetadataCtx(ctx), req)
	if err != nil {
		blog.Errorf(fmt.Sprintf("UpdateQuotaInfoForSuanli failed req: %+v, err: %s,", req, err.Error()))
		return err
	}

	blog.Infof(fmt.Sprintf("UpdateQuotaInfoForSuanli req: %+v, resp: %+v,", req, resp))
	if resp.Code != "0" && resp.Message != ResultSuccessKey {
		return fmt.Errorf("UpdateQuotaInfoForSuanli failed, err: %+v", resp.Message)
	}

	return nil
}
