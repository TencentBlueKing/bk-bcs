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

// Package handler NOTES
package handler

import (
	"net/http"
	// import pprof.
	_ "net/http/pprof"

	"bscp.io/pkg/metrics"
	"bscp.io/pkg/runtime/ctl"
)

// Handler defines http handlers to add to http service
type Handler struct {
	Pattern string
	Handler http.Handler
}

// HTTPMiddleware is the http middleware that adds common handler to root http handler
func HTTPMiddleware(root http.Handler, handlers ...Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", root.ServeHTTP)

	// add metrics handler
	mux.HandleFunc("/metrics", metrics.Handler().ServeHTTP)

	// add pprof handler
	mux.HandleFunc("/debug/", http.DefaultServeMux.ServeHTTP)

	// add tools handler
	mux.HandleFunc("/ctl", ctl.Handler().ServeHTTP)

	for _, handler := range handlers {
		mux.HandleFunc(handler.Pattern, handler.Handler.ServeHTTP)
	}

	return mux
}
