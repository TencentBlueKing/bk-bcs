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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/pkg/bcsapi/thirdparty-service"
)

// GetKubeConfigForTaiji request thirdparty service
func (t *thirdpartyClient) GetKubeConfigForTaiji(nameSpace string) (*third.GetKubeConfigForTaijiResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.GetKubeConfigForTaiji(t.getMetadataCtx(ctx),
		&third.GetKubeConfigForTaijiRequest{NameSpace: nameSpace})
	if err != nil {
		return nil, err
	}

	blog.Infof("GetKubeConfigForTaiji resp: %+v", resp)
	return resp, nil
}

// GetKubeConfigForSuanli request thirdparty service
func (t *thirdpartyClient) GetKubeConfigForSuanli(namespace string) (*third.GetKubeConfigForSuanliResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	resp, err := t.thirdpartySvc.GetKubeConfigForSuanli(t.getMetadataCtx(ctx),
		&third.GetKubeConfigForSuanliRequest{
			NameSpace: namespace,
		})
	if err != nil {
		return nil, err
	}

	blog.Infof("GetKubeConfigForTaiji resp: %+v", resp)
	return resp, nil
}
