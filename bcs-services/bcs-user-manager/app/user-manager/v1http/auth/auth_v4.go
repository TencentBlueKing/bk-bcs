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

package auth

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	restful "github.com/emicklei/go-restful/v3"

	iamv4 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/iamv4"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

var errIAMV4EmptyToken = errors.New("iam v4 auth token is empty")

func init() {
	iamv4.FetchToken = defaultFetchIAMV4Token
}

// BKIAMV4AuthFunc IAM V4 provider 回调鉴权
func BKIAMV4AuthFunc(rb *restful.RouteBuilder) *restful.RouteBuilder {
	return iamv4.AuthFunc(rb)
}

func defaultFetchIAMV4Token(ctx context.Context) (string, error) {
	if config.GlobalIAMV4Client == nil {
		return "", iamv4.ErrNotConfigured
	}
	data, err := config.GlobalIAMV4Client.RetrieveSystemAuthToken(ctx, "")
	if err != nil {
		blog.Errorf("iam v4 provider auth: get token failed")
		return "", err
	}
	if data == nil || data.AuthToken == "" {
		return "", errIAMV4EmptyToken
	}
	return data.AuthToken, nil
}
