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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
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
	PatchType     types.PatchType
	ListOptions   *metav1.ListOptions
	DeleteOptions *metav1.DeleteOptions
	GetOptions    *metav1.GetOptions
	CreateOptions *metav1.CreateOptions
	UpdateOptions *metav1.UpdateOptions
	PatchOptions  *metav1.PatchOptions
	PodLogOptions *v1.PodLogOptions
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

	options, err := ParseOptions(req, requestInfo, requestInfo.Verb)
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
func ParseOptions(req *http.Request, reqInfo *apirequest.RequestInfo, rawVerb string) (*Options, error) {
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
		if reqInfo.Subresource == "log" {
			podLogOptions, err := clientutil.MakePodLogOptions(req.URL.Query())
			if err != nil {
				return nil, err
			}
			options.PodLogOptions = podLogOptions
		}
	case "delete":
		deleteOptions, err := clientutil.GetDeleteOptionsFromReq(req)
		if err != nil {
			return nil, err
		}
		options.DeleteOptions = deleteOptions
	case "create":
		createOptions, err := clientutil.MakeCreateOptions(req.URL.Query())
		if err != nil {
			return nil, err
		}
		options.CreateOptions = createOptions
	case "update":
		options.UpdateOptions = &metav1.UpdateOptions{}
	case "patch":
		patchOptions, err := clientutil.MakePatchOptions(req.URL.Query())
		if err != nil {
			return nil, err
		}
		options.PatchOptions = patchOptions
		// PatchType 从头部获取
		options.PatchType = types.PatchType(req.Header.Get("Content-Type"))
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
		if reqInfo.Subresource == "log" {
			options.Verb = GetLogsVerb
		} else if options.AcceptHeader != "" {
			options.Verb = GetAsTableVerb
		} else {
			options.Verb = GetVerb
		}
	case "watch":
		options.Verb = WatchVerb
	case "delete":
		options.Verb = DeleteVerb
	case "create":
		if reqInfo.Subresource == "exec" {
			options.GetOptions = &metav1.GetOptions{}
			options.Verb = ExecVerb
		} else {
			options.Verb = CreateVerb
		}
	case "update":
		options.Verb = UpdateVerb
	// clusternet patch 有问题, 先过滤
	// case "patch":
	// 	options.Verb = PatchVerb
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
