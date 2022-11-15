/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package service NOTES
package service

import (
	"net/http"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/logs"

	"github.com/gorilla/mux"
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

	router := mux.NewRouter()
	router.HandleFunc("/repository/api/repo/create", s.createRepo).Methods(http.MethodPost)
	router.HandleFunc("/repository/api/metadata/{project}/{repo}/file/{sign}", s.queryMetadataInfo).
		Methods(http.MethodGet)
	router.HandleFunc("/generic/{project}/{repo}/file/{sign}", s.uploadNode).Methods(http.MethodPut)
	router.HandleFunc("/generic/{project}/{repo}/file/{sign}", s.getNodeInfo).Methods(http.MethodHead)
	router.HandleFunc("/generic/{project}/{repo}/file/{sign}", s.downloadNode).Methods(http.MethodGet)

	apiMux := http.NewServeMux()
	apiMux.Handle("/", router)

	return setupFilters(apiMux)
}

func setupFilters(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(constant.RidKey)

		// request and response details landing log for monitoring and troubleshooting problem.
		logs.Infof("uri: %s, method: %s, rid: %s", r.RequestURI, r.Method, rid)

		// handler request.
		mux.ServeHTTP(w, r)
	})
}
