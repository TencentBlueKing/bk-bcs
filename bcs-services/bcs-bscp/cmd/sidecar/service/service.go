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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/runtime/ctl"

	"github.com/emicklei/go-restful/v3"
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
	root := http.NewServeMux()
	root.HandleFunc("/", s.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	root.HandleFunc("/debug/", http.DefaultServeMux.ServeHTTP)
	root.HandleFunc("/metrics", metrics.Handler().ServeHTTP)
	root.HandleFunc("/ctl", ctl.Handler().ServeHTTP)

	return root
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, req *http.Request) {
	// TODO: implement this
	rest.WriteResp(w, rest.NewBaseResp(errf.OK, ""))
	return
}

func (s *Service) apiSet() *restful.Container {

	handler := rest.NewHandler()

	handler.Add("ping", "POST", "/api/v1/sidecar/ping", s.Ping)

	c := restful.NewContainer()
	handler.Load(c)
	return c
}
