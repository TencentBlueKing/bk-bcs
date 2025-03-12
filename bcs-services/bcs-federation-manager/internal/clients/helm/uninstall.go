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

// Package helm xxx
package helm

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"github.com/avast/retry-go"
)

// UninstallRelease uninstall release
func (h *helmClient) UninstallRelease(opt *HelmOptions) error {

	req := &helmmanager.UninstallReleaseV1Req{
		ProjectCode: opt.ProjectID,
		ClusterID:   opt.ClusterID,
		Namespace:   opt.Namespace,
		Name:        opt.ReleaseName,
	}

	resp := &helmmanager.UninstallReleaseV1Req{}
	err := retry.Do(func() error {
		resp, err := h.helmSvc.UninstallReleaseV1(h.getMetadataCtx(context.Background()), req)
		if err != nil {
			blog.Errorf("[HelmManager] UninstallRelease failed, err: %s", err.Error())
			return err
		}
		if resp == nil {
			blog.Errorf("[HelmManager] UninstallRelease failed, resp is empty")
			return fmt.Errorf("UninstallRelease failed, resp is empty")
		}

		if resp.Code != 0 || !resp.Result {
			blog.Errorf("[HelmManager] UninstallRelease failed, code: %d, message: %s", resp.Code, resp.Message)
			return fmt.Errorf("UninstallRelease failed, code: %d, message: %s", resp.Code, resp.Message)
		}
		return nil
	}, retry.Attempts(DefaultRetryCount), retry.Delay(DefaultTimeout), retry.DelayType(retry.FixedDelay))
	if err != nil {
		return fmt.Errorf("call api HelmManager UninstallRelease failed: %v, resp: %v", err, resp)
	}

	return nil
}
