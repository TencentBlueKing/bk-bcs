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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"bk-bscp/internal/database"
	"bk-bscp/pkg/logger"
)

// PatchData return data about patch
type PatchData struct {
	// CurrentVersion current patch version
	CurrentVersion string `json:"current_version"`

	// Operator operator who executes the patch
	Operator string `json:"operator"`

	// Kind version type, such as oa, ee, ce
	Kind string `json:"kind"`
}

// writeResult return the result of function execution
func writeResult(w http.ResponseWriter, curVersionRecord *database.System) {
	data := &PatchData{
		CurrentVersion: curVersionRecord.CurrentVersion,
		Operator:       curVersionRecord.Operator,
		Kind:           curVersionRecord.Kind,
	}

	response, err := json.Marshal(&data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, fmt.Sprintf("marshal patch response error, %+v", err))
		return
	}

	if _, err := w.Write(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, fmt.Sprintf("write patch response error, %+v", err))
	}
}

// setupFilters setups all api filters here. All request would cross here, and we filter
// request base on URL.
func (p *Patcher) setupFilters(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("Patcher| method: %s, uri: %s, remote_addr: %s", r.Method, r.RequestURI, r.RemoteAddr)
		mux.ServeHTTP(w, r)
	})
}

// setupRouters setups all routers here.
func (p *Patcher) setupRouters(rtr *mux.Router) {
	// put all patchs.
	rtr.HandleFunc("/api/v2/patch/{operator}", func(w http.ResponseWriter, req *http.Request) {
		operator := mux.Vars(req)["operator"]

		curVersionRecord, err := p.hpm.PutPatchs(operator)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		writeResult(w, curVersionRecord)
		return
	}).Methods("POST")

	// put patch until limit.
	rtr.HandleFunc("/api/v2/patch/{limit_version}/{operator}", func(w http.ResponseWriter, req *http.Request) {
		operator := mux.Vars(req)["operator"]
		limitVersion := mux.Vars(req)["limit_version"]

		curVersionRecord, err := p.hpm.PutPatch(operator, limitVersion)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		writeResult(w, curVersionRecord)
		return
	}).Methods("POST")

	// get the current version.
	rtr.HandleFunc("/api/v2/patch", func(w http.ResponseWriter, req *http.Request) {
		curVersionRecord, err := p.hpm.GetCurrentVersion()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		writeResult(w, curVersionRecord)
		return
	}).Methods("GET")
}
