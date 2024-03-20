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
 *
 */

package extender

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/cachemanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/migrator"
)

// HttpExtenderInterface defines the interface of scheduler extender.
type HttpExtenderInterface interface {
	Init() error
	Run(ctx context.Context) error
}

// HTTPExtender will create a http server and receive the pod request from
// scheduler. Extender returns the filter result.
type HTTPExtender struct {
	op *options.DeSchedulerOption

	cacheManager    cachemanager.CacheInterface
	migratorManager migrator.DescheduleMigratorInterface

	httpServer *http.Server
	tlsConfig  *tls.Config
}

// NewHTTPExtender creates the instance of HTTPExtender
func NewHTTPExtender() HttpExtenderInterface {
	return &HTTPExtender{
		op:           options.GlobalConfigHandler().GetOptions(),
		cacheManager: cachemanager.NewCacheManager(),
	}
}

const (
	filterType     = "filter"
	prioritizeType = "prioritize"
	preemptType    = "preempt"
	bindType       = "bind"
)

// Init will init the http_server
func (h *HTTPExtender) Init() error {
	h.migratorManager = migrator.GlobalMigratorManager()
	if len(h.op.ServerCert) != 0 && len(h.op.ServerKey) != 0 && len(h.op.ServerCa) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(h.op.ServerCa, h.op.ServerCert,
			h.op.ServerKey, static.ServerCertPwd)
		if err != nil {
			return errors.Wrapf(err, "load cluster manager server tls config failed")
		}

		h.tlsConfig = tlsConfig
		blog.Infof("load cluster manager server tls config successfully")
	}

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/"+filterType, func(w http.ResponseWriter, r *http.Request) {
		h.handleFunc(w, r, filterType, new(extenderv1.ExtenderArgs))
	})
	httpMux.HandleFunc("/"+prioritizeType, func(w http.ResponseWriter, r *http.Request) {
		h.handleFunc(w, r, prioritizeType, new(extenderv1.ExtenderArgs))
	})
	httpMux.HandleFunc("/"+bindType, func(w http.ResponseWriter, r *http.Request) {
		h.handleFunc(w, r, bindType, new(extenderv1.ExtenderBindingArgs))
	})
	httpMux.HandleFunc("/"+preemptType, func(w http.ResponseWriter, r *http.Request) {
		h.handleFunc(w, r, preemptType, new(extenderv1.ExtenderPreemptionArgs))
	})
	httpServerEndpoint := "0.0.0.0:" + strconv.Itoa(int(h.op.ExtenderPort))
	h.httpServer = &http.Server{
		Addr:    httpServerEndpoint,
		Handler: httpMux,
	}
	return nil
}

// Run will run the http server. It will stop with context done
func (h *HTTPExtender) Run(ctx context.Context) (err error) {
	go func() {
		<-ctx.Done()
		_ = h.httpServer.Shutdown(context.Background())
	}()

	if h.tlsConfig != nil {
		h.httpServer.TLSConfig = h.tlsConfig
		err = h.httpServer.ListenAndServeTLS("", "")
	} else {
		err = h.httpServer.ListenAndServe()
	}
	if err != nil {
		return errors.Wrapf(err, "http_server exit with err")
	}
	return nil
}

func (h *HTTPExtender) handleFunc(writer http.ResponseWriter, request *http.Request,
	handlerType string, args interface{}) {
	msgId := uuid.New().String()
	if err := handleFuncReadRequestBody(request, args); err != nil {
		blog.Errorf("MsgID[%s] scheduler extender '%s' read request failed: %s",
			msgId, handlerType, err.Error())
		handleFuncReturn500(writer, errors.Wrapf(err, "msgID="+msgId))
		return
	}
	ctx := request.Context()
	var result interface{}
	var err error
	switch handlerType {
	case filterType:
		result, err = h.Filter(ctx, msgId, args.(*extenderv1.ExtenderArgs))
	case prioritizeType:
		result, err = h.Prioritize(ctx, msgId, args.(*extenderv1.ExtenderArgs))
	case preemptType:
		result, err = h.ProcessPreemption(ctx, msgId, args.(*extenderv1.ExtenderPreemptionArgs))
	case bindType:
		result, err = h.Bind(ctx, msgId, args.(*extenderv1.ExtenderBindingArgs))
	default:
		handleFuncReturn500(writer, fmt.Errorf("unknown type '%s'", handlerType))
		return
	}
	if err != nil {
		blog.Errorf("MsgID[%s] scheduler extender '%s' occurred an error: %s",
			msgId, handlerType, err.Error())
		handleFuncReturn500(writer, errors.Wrapf(err, "handler=%s, msgID=%s", handlerType, msgId))
		return
	}
	handleFuncReturn200(writer, result)
}

type errorResult struct {
	Error string `json:"error"`
}

func handleFuncReturn500(writer http.ResponseWriter, err error) {
	bs, _ := json.Marshal(&errorResult{
		Error: err.Error(),
	})
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusInternalServerError)
	_, _ = writer.Write(bs)
}

func handleFuncReturn200(writer http.ResponseWriter, data interface{}) {
	bs, _ := json.Marshal(data)
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(bs)
}

func handleFuncReadRequestBody(request *http.Request, body interface{}) error {
	bs, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return errors.Wrapf(err, "read request body failed")
	}
	if err := json.Unmarshal(bs, body); err != nil {
		return errors.Wrapf(err, "unmarshal body failed")
	}
	return nil
}
