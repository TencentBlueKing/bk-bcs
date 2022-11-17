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

package rest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/gwparser"

	"github.com/emicklei/go-restful/v3"
	prm "github.com/prometheus/client_golang/prometheus"
)

var once sync.Once

// NewHandler create a new restfull handler
func NewHandler() *Handler {
	once.Do(func() {
		initMetric()
	})

	return &Handler{
		actions: make([]*action, 0),
	}
}

// action defines a http request action
type action struct {
	Verb    string
	Path    string
	Alias   string
	Handler func(contexts *Contexts) (reply interface{}, err error)
}

// Handler contains all the restfull http handler actions
type Handler struct {
	actions []*action
}

// Add add a http handler
func (r *Handler) Add(alias, verb, path string, handler func(contexts *Contexts) (reply interface{}, err error)) {

	switch verb {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete:
	default:
		panic(fmt.Sprintf("add http handler failed, inavlid http verb: %s.", verb))
	}

	if len(path) == 0 {
		panic("add http handler, but got empty http path.")
	}

	if handler == nil {
		panic("add http handler, but got nil http handler")
	}

	r.actions = append(r.actions, &action{Verb: verb, Path: path, Alias: alias, Handler: handler})
}

// Load add actions to the restful webservice, and add to the rest container.
func (r *Handler) Load(c *restful.Container) {
	if len(r.actions) == 0 {
		panic("no actions has been added, can not load the handler")
	}

	ws := new(restful.WebService)
	ws.Produces(restful.MIME_JSON)

	for _, action := range r.actions {
		switch action.Verb {
		case http.MethodPost:
			ws.Route(ws.POST(action.Path).To(r.wrapperAction(action)))
		case http.MethodDelete:
			ws.Route(ws.DELETE(action.Path).To(r.wrapperAction(action)))
		case http.MethodPut:
			ws.Route(ws.PUT(action.Path).To(r.wrapperAction(action)))
		case http.MethodGet:
			ws.Route(ws.GET(action.Path).To(r.wrapperAction(action)))
		default:
			panic(fmt.Sprintf("add handler to webservice, but got unsupport verb: %s .", action.Verb))
		}
	}

	c.Add(ws)
}

func (r *Handler) wrapperAction(action *action) func(req *restful.Request, resp *restful.Response) {
	return func(req *restful.Request, resp *restful.Response) {
		cts := new(Contexts)
		cts.Request = req
		cts.resp = resp

		kt, err := gwparser.Parse(req.Request.Context(), req.Request.Header)
		if err != nil {
			rid := req.Request.Header.Get(constant.RidKey)
			logs.Errorf("invalid request for %s, err: %v, rid: %s", action.Alias, err, rid)
			cts.WithStatusCode(http.StatusBadRequest)
			cts.respError(err)
			restMetric.errCounter.With(prm.Labels{"alias": action.Alias, "biz": cts.bizID, "app": cts.appID}).Inc()
			return
		}

		cts.Kit = kt

		if logs.V(4) && req.Request.Body != nil {

			byt, err := ioutil.ReadAll(req.Request.Body)
			if err != nil {
				logs.Errorf("restful request %s peek failed, err: %v, rid: %s", action.Alias, err, cts.Kit.Rid)

				cts.WithStatusCode(http.StatusBadRequest)
				cts.respError(errf.New(errf.InvalidParameter, err.Error()))
				restMetric.errCounter.With(prm.Labels{"alias": action.Alias, "biz": cts.bizID, "app": cts.appID}).Inc()
				return
			}

			req.Request.Body = ioutil.NopCloser(bytes.NewBuffer(byt))

			logs.Infof("%s received restful request, body: %s, rid: %s", action.Alias, string(byt), kt.Rid)
		}

		start := time.Now()
		reply, err := action.Handler(cts)
		if err != nil {
			if logs.V(2) {
				logs.Errorf("do restful request %s failed, err: %v, rid: %s", action.Alias, err, cts.Kit.Rid)
			}

			cts.respError(err)
			restMetric.errCounter.With(prm.Labels{"alias": action.Alias, "biz": cts.bizID, "app": cts.appID}).Inc()
			return
		}

		cts.respEntity(reply)

		restMetric.lagMS.With(prm.Labels{"alias": action.Alias, "biz": cts.bizID, "app": cts.appID}).
			Observe(float64(time.Since(start).Milliseconds()))
	}
}
