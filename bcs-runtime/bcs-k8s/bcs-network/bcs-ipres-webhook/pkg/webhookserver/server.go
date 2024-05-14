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

package webhookserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
	defaulter     = runtime.ObjectDefaulter(runtimeScheme)
)

// ServerOption option for server
type ServerOption struct {
	Addr           string
	Port           int
	ValidatingPath string
	MutatingPath   string
	ServerCertFile string
	ServerKeyFile  string
}

// Server webhook server object
type Server struct {
	server  *http.Server
	opt     *ServerOption
	handler WebhookHandler
}

// NewHookServer create new webhook server
func NewHookServer(opt *ServerOption) (*Server, error) {
	if opt == nil {
		return nil, fmt.Errorf("opt cannot be empty")
	}
	pair, err := tls.LoadX509KeyPair(opt.ServerCertFile, opt.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load x509 key pair cert %s, key %s failed, err %s",
			opt.ServerCertFile, opt.ServerKeyFile, err.Error())
	}

	return &Server{
		server: &http.Server{
			Addr:      fmt.Sprintf("%s:%v", opt.Addr, opt.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
		opt: opt,
	}, nil
}

// RegisterWebhookHandler register handler to webhook server
func (s *Server) RegisterWebhookHandler(handler WebhookHandler) {
	s.handler = handler
}

// Start start http server
func (s *Server) Start() error {
	if s.handler == nil {
		return fmt.Errorf("handler cannot be empty")
	}
	mux := http.NewServeMux()
	// register handler
	mux.HandleFunc(s.opt.ValidatingPath, s.HandleValidatingWebhook)
	mux.HandleFunc(s.opt.MutatingPath, s.HandleMutatingWebhook)
	s.server.Handler = mux

	go func() {
		if err := s.server.ListenAndServeTLS("", ""); err != nil {
			blog.Fatalf("failed to listen and serve webhook server, err %s", err.Error())
		}
	}()

	blog.Infof("webhook server started")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	blog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	s.server.Shutdown(context.Background())
	return nil
}

// HandleValidatingWebhook handle validating webhook request
func (s *Server) HandleValidatingWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, s.opt.ValidatingPath, s.handler.HandleValidatingWebhook)
}

// HandleMutatingWebhook handle mutating webhook request
func (s *Server) HandleMutatingWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, s.opt.MutatingPath, s.handler.HandleMutatingWebhook)
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request, handleName string,
	handleFunc func(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse) {
	startTime := time.Now()
	var body []byte
	if r.Body != nil {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			blog.Errorf("read body failed, err %s", err.Error())
			http.Error(w, "read body failed", http.StatusBadRequest)
			ReportMetric(handleName, ResultFail, startTime)
			return
		}
		body = data
	}
	if len(body) == 0 {
		blog.Errorf("body missing")
		http.Error(w, "body missing", http.StatusBadRequest)
		ReportMetric(handleName, ResultFail, startTime)
		return
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		blog.Errorf("Could not decode body: %s", err.Error())
		reviewResponse = &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
	} else {
		reviewResponse = handleFunc(ar)
	}

	response := v1beta1.AdmissionReview{}
	if reviewResponse != nil {
		response.Response = reviewResponse
		if ar.Request != nil {
			response.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(response)
	if err != nil {
		blog.Errorf("Could not encode response: %v", err)
		http.Error(w, fmt.Sprintf("could encode response: %v", err), http.StatusInternalServerError)
		ReportMetric(handleName, ResultFail, startTime)
		return
	}
	if _, err := w.Write(resp); err != nil {
		blog.Errorf("Could not write response: %v", err)
		http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
		ReportMetric(handleName, ResultFail, startTime)
		return
	}

	ReportMetric(handleName, ResultSuccess, startTime)
}
