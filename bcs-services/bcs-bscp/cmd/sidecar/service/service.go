/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package service NOTES
package service

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"bscp.io/pkg/runtime/handler"
)

// InitService initial the service instance
func InitService() (*Service, error) {
	return &Service{}, nil

}

// Service defines the sidecar's services.
type Service struct {
}

// Handler returns all the http handler supported by the sidecar's service.
func (s *Service) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(handler.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 公共方法
	r.Get("/healthz", s.Healthz)
	r.Mount("/", handler.RegisterCommonToolHandler())

	r.Post("/api/v1/sidecar/ping", s.Ping)

	return r
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK"))
}
