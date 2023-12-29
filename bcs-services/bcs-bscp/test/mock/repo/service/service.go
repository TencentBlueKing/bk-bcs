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

// Package service NOTES
package service

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// Service do all the data service's work
type Service struct {
	Workspace string
}

// NewService new service.
func NewService(ws string) (*Service, error) {
	return &Service{
		Workspace: ws,
	}, nil
}

// Handler return service handler.
func (s *Service) Handler() http.Handler {

	r := chi.NewRouter()
	r.Use(setupFilters)
	r.Post("/repository/api/repo/create", s.createRepo)
	r.Get("/repository/api/metadata/{project}/{repo}/file/{sign}", s.queryMetadataInfo)
	r.Put("/generic/{project}/{repo}/file/{sign}", s.uploadNode)
	r.Head("/generic/{project}/{repo}/file/{sign}", s.getNodeInfo)
	r.Get("/generic/{project}/{repo}/file/{sign}", s.downloadNode)

	return r
}

func setupFilters(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(constant.RidKey)

		// request and response details landing log for monitoring and troubleshooting problem.
		logs.Infof("uri: %s, method: %s, rid: %s", r.RequestURI, r.Method, rid)

		// handler request.
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
