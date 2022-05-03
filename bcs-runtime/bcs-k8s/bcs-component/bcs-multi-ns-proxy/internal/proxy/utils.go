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

package proxy

import (
	"fmt"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

const (
	// DefaultLegacyAPIPrefix is where the legacy APIs will be located.
	DefaultLegacyAPIPrefix = "/api"
	// APIGroupPrefix is where non-legacy API group will be located.
	APIGroupPrefix = "/apis"
)

func getNamespaceFromRequest(req *http.Request) (string, error) {
	apiPrefixes := sets.NewString(strings.Trim(APIGroupPrefix, "/"))
	legacyAPIPrefixes := sets.String{}
	apiPrefixes.Insert(strings.Trim(DefaultLegacyAPIPrefix, "/"))
	legacyAPIPrefixes.Insert(strings.Trim(DefaultLegacyAPIPrefix, "/"))

	requestInfoFactory := &apirequest.RequestInfoFactory{
		APIPrefixes:          apiPrefixes,
		GrouplessAPIPrefixes: legacyAPIPrefixes,
	}

	requestInfo, err := requestInfoFactory.NewRequestInfo(req)
	if err != nil {
		return "", fmt.Errorf("create info from request %s %s failed, err %s",
			req.RemoteAddr, req.URL.String(), err.Error())
	}
	return requestInfo.Namespace, nil
}
