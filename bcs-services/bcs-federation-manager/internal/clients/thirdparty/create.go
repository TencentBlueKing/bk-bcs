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

// CreateNamespaceForTaijiV3 request thirdparty service
func (t *thirdpartyClient) CreateNamespaceForTaijiV3(req *third.CreateNamespaceForTaijiV3Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.CreateNamespaceForTaijiV3(t.getMetadataCtx(ctx), req)
	if err != nil {
		blog.Errorf("CreateNamespaceForTaijiV3 failed err: %s", err.Error())
		return err
	}

	blog.Infof("CreateNamespaceForTaijiV3 resp: %+v", resp)
	if resp.Error != nil && resp.Error.Code != ResultSuccessKey {
		return fmt.Errorf("CreateNamespaceForTaijiV3 failed resp: %+v", resp)
	}

	return nil
}

// CreateModule request thirdparty service
func (t *thirdpartyClient) CreateModule(moduleName string) (*third.CreateModuleResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.CreateModule(t.getMetadataCtx(ctx),
		&third.CreateModuleRequest{BkModuleName: moduleName})
	if err != nil {
		return nil, err
	}

	blog.Infof("CreateModule resp: %+v", resp)
	if resp.Code != "0" {
		blog.Errorf("CreateModule failed resp: %+v", resp)
		return nil, fmt.Errorf("CreateModule failed resp: %+v", resp)
	}

	return resp, nil
}

// CreateNamespaceForSuanli create suanli namespace
func (t *thirdpartyClient) CreateNamespaceForSuanli(req *third.CreateNamespaceForSuanliRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.CreateNamespaceForSuanli(t.getMetadataCtx(ctx), req)
	if err != nil {
		blog.Errorf("CreateNamespaceForSuanli failed err: %s", err.Error())
		return err
	}

	blog.Infof("CreateNamespaceForSuanli resp: %+v", resp)
	if resp == nil || resp.Message != ResultSuccessKey {
		return fmt.Errorf("CreateNamespaceForSuanli failed resp: %+v", resp)
	}

	return nil
}
