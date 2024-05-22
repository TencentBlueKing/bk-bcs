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

package http

import (
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/handler"
)

func ProviderProxyHandler() http.Handler {
	upstream := cc.FeedProxy().Upstream
	var targetHost string
	switch upstream.StorageType {
	case cc.BkRepo:
		targetHost = upstream.BkRepoHost
	case cc.S3:
		targetHost = upstream.CosHost
	}
	return handler.ReverseProxyHandler("proxy_download", "/proxy/download", targetHost)
}