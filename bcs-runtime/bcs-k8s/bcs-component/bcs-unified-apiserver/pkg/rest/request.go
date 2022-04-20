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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
)

const (
	// DefaultLegacyAPIPrefix is where the legacy APIs will be located.
	DefaultLegacyAPIPrefix = "/api"
	// APIGroupPrefix is where non-legacy API group will be located.
	APIGroupPrefix = "/apis"
)

// 未实现的方法Err, 抛到上层处理
var (
	ErrNotImplemented = errors.New("NotImplementedError")
	ErrInit           = errors.New("InitError")
)

// Options K8S Rest Reqeust Options
type Options struct {
	Verb          Verb
	AcceptHeader  string
	ListOptions   *metav1.ListOptions
	DeleteOptions *metav1.DeleteOptions
	GetOptions    *metav1.GetOptions
}

// RequestContext K8S Rest Request Context
type RequestContext struct {
	*apirequest.RequestInfo
	Writer  http.ResponseWriter
	Request *http.Request
	Options *Options
}

// NewRequestContext Make RequestContext from http.Request
func NewRequestContext(rw http.ResponseWriter, req *http.Request) (*RequestContext, error) {
	requestInfo, err := ParseRequestInfo(req)
	if err != nil {
		return nil, err
	}

	options, err := ParseOptions(req, requestInfo.Verb)
	if err != nil {
		return nil, err
	}

	reqInfo := &RequestContext{
		RequestInfo: requestInfo,
		Options:     options,
		Request:     req,
		Writer:      rw,
	}
	return reqInfo, nil
}

// ParserOptions 解析request头部操作, header等
func ParseOptions(req *http.Request, rawVerb string) (*Options, error) {
	options := new(Options)

	// 解析参数
	switch rawVerb {
	case "list", "watch":
		listOptions, err := clientutil.GetListOptionsFromQueryParam(req.URL.Query())
		if err != nil {
			return nil, err
		}
		options.ListOptions = listOptions
	case "get":
		// Get 没有参数可解析
		options.GetOptions = &metav1.GetOptions{}
	case "delete":
		deleteOptions, err := clientutil.GetDeleteOptionsFromReq(req)
		if err != nil {
			return nil, err
		}
		options.DeleteOptions = deleteOptions
	}

	acceptHeader := req.Header.Get("Accept")
	if strings.Contains(acceptHeader, "as=Table") {
		options.AcceptHeader = acceptHeader
	}

	// 解析类型
	switch rawVerb {
	case "list":
		if options.AcceptHeader != "" {
			options.Verb = ListAsTableVerb
		} else {
			options.Verb = ListVerb
		}
	case "get":
		// Get 没有参数可解析
		options.GetOptions = &metav1.GetOptions{}
		if options.AcceptHeader != "" {
			options.Verb = GetAsTableVerb
		} else {
			options.Verb = GetVerb
		}
	case "watch":
		options.Verb = WatchVerb
	case "delete":
		options.Verb = DeleteVerb
	default:
		return nil, ErrNotImplemented
	}
	return options, nil
}

// ParseRequestInfo 解析url等
func ParseRequestInfo(req *http.Request) (*apirequest.RequestInfo, error) {
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
		return nil, fmt.Errorf("parse info from request %s %s failed, err %s", req.RemoteAddr, req.URL.String(), err.Error())
	}
	return requestInfo, nil
}

// HandlerFunc defines the handler used by rest middleware as return value.
type HandlerFunc func(*RequestContext)

// Handler Rest Handle Interface
type Handler interface {
	Serve(*RequestContext)
}
