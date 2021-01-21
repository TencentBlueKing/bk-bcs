/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// setupRouters setups all api routers here.
func (s *APIServer) setupRouters(rtr *mux.Router) {
	// handle swagger files.
	if s.viper.GetBool("server.api.open") {
		rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/",
			http.FileServer(http.Dir(s.viper.GetString("server.api.dir")))))
	}

	// healthz.
	rtr.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		if err := s.healthz(w, req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
	}).Methods("GET")

	// bkrepo http proxy.
	rtr.HandleFunc("/api/v2/file/content/biz/{biz_id}", func(w http.ResponseWriter, req *http.Request) {
		if err := s.bkRepoProxy.Verify(req); err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, err.Error())
			return
		}
		s.bkRepoProxy.ServeHTTP(w, req)
	}).Methods("PUT", "GET", "HEAD")

	// register configserver interfaces.
	rtr.PathPrefix("/api/v2/config/").Handler(s.cfgGWMux)

	// register templateserver interfaces.
	rtr.PathPrefix("/api/v2/template/").Handler(s.tplGWMux)
}
