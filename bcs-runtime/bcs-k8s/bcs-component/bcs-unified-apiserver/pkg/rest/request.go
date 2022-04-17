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

package rest

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

type TableRequest struct {
	IsTable      bool
	AcceptHeader string
}

type RequestInfo struct {
	*apirequest.RequestInfo
	TableReq *TableRequest
	Writer   http.ResponseWriter
	Request  *http.Request
}

func NewRequestInfo(req *http.Request) (*RequestInfo, error) {
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
		return nil, fmt.Errorf("create info from request %s %s failed, err %s",
			req.RemoteAddr, req.URL.String(), err.Error())
	}

	tableReq := &TableRequest{}

	acceptHeader := req.Header.Get("Accept")
	if strings.Contains(acceptHeader, "as=Table") {
		tableReq.AcceptHeader = acceptHeader
		tableReq.IsTable = true
	}
	reqInfo := &RequestInfo{RequestInfo: requestInfo, TableReq: tableReq}
	return reqInfo, nil
}

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(*RequestInfo)

type Handler interface {
	Serve(*RequestInfo)
}
