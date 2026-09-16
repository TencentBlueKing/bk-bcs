/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package iamv4

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	restful "github.com/emicklei/go-restful/v3"
)

// Querier 按资源类型查询实例，便于单测注入
type Querier interface {
	List(ctx context.Context, tenantID, resType string, filter Filter, page Page) (int, []Instance, error)
	Fetch(ctx context.Context, resType string, filter Filter) ([]Instance, error)
}

// Handler IAM V4 资源回调
type Handler struct {
	Querier Querier
}

var defaultHandler = &Handler{}

// SetDefaultQuerier 注册默认查数实现（由 resources 包在启动时注入）
func SetDefaultQuerier(q Querier) {
	defaultHandler = &Handler{Querier: q}
}

// ResourceDispatch 分发 IAM V4 资源回调
func ResourceDispatch(request *restful.Request, response *restful.Response) {
	defaultHandler.Serve(request, response)
}

// Serve 处理回调
func (h *Handler) Serve(request *restful.Request, response *restful.Response) {
	if h == nil || h.Querier == nil {
		writeError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	echoRequestID(request, response)

	var req CallbackRequest
	limited := http.MaxBytesReader(response.ResponseWriter, request.Request.Body, maxBodyBytes)
	if err := json.NewDecoder(limited).Decode(&req); err != nil {
		writeError(response, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}
	if err := validateCallback(req); err != nil {
		var re *requestError
		if errors.As(err, &re) {
			writeError(response, re.status, re.code, re.message)
			return
		}
		writeError(response, http.StatusBadRequest, "INVALID_REQUEST", "invalid request")
		return
	}

	ctx := request.Request.Context()
	tenantID := request.Request.Header.Get("X-Bk-Tenant-Id")
	page := normalizePage(req.Page)

	switch req.Method {
	case methodListInstance:
		count, results, err := h.Querier.List(ctx, tenantID, req.Type, req.Filter, page)
		if err != nil {
			writeQuerierError(response, req, err)
			return
		}
		writeData(response, ListResult{Count: count, Results: results})
	case methodFetchInstanceInfo:
		results, err := h.Querier.Fetch(ctx, req.Type, req.Filter)
		if err != nil {
			writeQuerierError(response, req, err)
			return
		}
		writeData(response, applyRequiresList(results, req.Requires))
	default:
		writeError(response, http.StatusBadRequest, "INVALID_REQUEST", "unsupported method")
	}
}

func writeQuerierError(response *restful.Response, req CallbackRequest, err error) {
	var re *requestError
	if errors.As(err, &re) {
		writeError(response, re.status, re.code, re.message)
		return
	}
	blog.Errorf("iamv4 provider type=%s method=%s failed", req.Type, req.Method)
	writeError(response, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
}

func writeData(response *restful.Response, data interface{}) {
	_ = response.WriteHeaderAndJson(http.StatusOK, SuccessResponse{Data: data}, "application/json")
}

func writeError(response *restful.Response, status int, code, message string) {
	_ = response.WriteHeaderAndJson(status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}},
		"application/json")
}

func echoRequestID(request *restful.Request, response *restful.Response) {
	rid := request.Request.Header.Get("X-Request-Id")
	if rid != "" {
		response.AddHeader("X-Request-Id", rid)
	}
}

func validateCallback(req CallbackRequest) error {
	if !validResourceType(req.Type) {
		return invalidRequest("unsupported type")
	}
	if req.Method != methodListInstance && req.Method != methodFetchInstanceInfo {
		return invalidRequest("unsupported method")
	}
	if req.Page.PageSize > maxPageSize {
		return invalidRequest("page_size exceeds limit")
	}
	if len(req.Filter.Keyword) > maxKeywordLen {
		return invalidRequest("keyword exceeds limit")
	}
	if len(req.Requires) > maxRequires {
		return invalidRequest("requires exceeds limit")
	}
	if req.Filter.Parent.ID != "" {
		if !validParentType(req.Filter.Parent.Type) {
			return invalidRequest("unsupported parent type")
		}
		if len(req.Filter.Parent.ID) > maxIDLen {
			return invalidRequest("parent id exceeds limit")
		}
	}
	for _, a := range req.Filter.Ancestors {
		if !validParentType(a.Type) && a.Type != Namespace {
			return invalidRequest("unsupported ancestor type")
		}
		if len(a.ID) > maxIDLen {
			return invalidRequest("ancestor id exceeds limit")
		}
	}
	if req.Method == methodFetchInstanceInfo {
		if len(req.Filter.IDs) == 0 {
			return invalidRequest("ids is empty")
		}
		if len(req.Filter.IDs) > maxIDs {
			return invalidRequest("ids exceeds limit")
		}
		for _, id := range req.Filter.IDs {
			if id == "" || len(id) > maxIDLen {
				return invalidRequest("invalid instance id")
			}
		}
	}
	return nil
}

func validResourceType(t string) bool {
	switch t {
	case Project, Cluster, Namespace, TemplateSet, CloudAccount:
		return true
	default:
		return false
	}
}

func validParentType(t string) bool {
	switch t {
	case "", Project, Cluster:
		return true
	default:
		return false
	}
}
