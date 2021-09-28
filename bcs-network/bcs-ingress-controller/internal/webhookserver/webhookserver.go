/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
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
	"strconv"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/portpoolcache"

	"k8s.io/api/admission/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
	defaulter     = runtime.ObjectDefaulter(runtimeScheme)
)

// ServerOption option of server
type ServerOption struct {
	Addr           string
	Port           int
	ServerCertFile string
	ServerKeyFile  string
}

// Server webhook server
type Server struct {
	server *http.Server
	// k8s client
	k8sClient client.Client
	poolCache *portpoolcache.Cache
}

// NewHookServer create new hook server object
func NewHookServer(opt *ServerOption, k8sClient client.Client, poolCache *portpoolcache.Cache) (*Server, error) {
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
		k8sClient: k8sClient,
		poolCache: poolCache,
	}, nil
}

// Start start http server
func (s *Server) Start() {
	mux := http.NewServeMux()
	// register handler function
	mux.HandleFunc("/portpool/v1/validate", s.HandleValidatingWebhook)
	mux.HandleFunc("/portpool/v1/mutate", s.HandleMutatingWebhook)
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
	return
}

// HandleValidatingWebhook handle validating webhook request
func (s *Server) HandleValidatingWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, "validate", s.validatingWebhook)
}

// HandleMutatingWebhook handle mutating webhook request
func (s *Server) HandleMutatingWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, "mutate", s.mutatingWebhook)
}

func (s *Server) handleWebhook(
	w http.ResponseWriter, r *http.Request, handleName string,
	handleFunc func(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse) {
	startTime := time.Now()
	var body []byte
	if r.Body != nil {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			blog.Errorf("read body failed, err %s", err.Error())
			http.Error(w, "read body failed", http.StatusBadRequest)
			metrics.ReportAPIRequestMetric(handleName, r.Method, strconv.Itoa(http.StatusBadRequest), startTime)
			return
		}
		body = data
	}
	if len(body) == 0 {
		blog.Errorf("body missing")
		http.Error(w, "body missing", http.StatusBadRequest)
		metrics.ReportAPIRequestMetric(handleName, r.Method, strconv.Itoa(http.StatusBadRequest), startTime)
		return
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		blog.Errorf("Could not decode body: %s", err.Error())
		reviewResponse = errResponse(err)
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
		metrics.ReportAPIRequestMetric(
			handleName, r.Method, strconv.Itoa(http.StatusInternalServerError), startTime)
		return
	}
	if _, err := w.Write(resp); err != nil {
		blog.Errorf("Could not write response: %v", err)
		http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
		metrics.ReportAPIRequestMetric(
			handleName, r.Method, strconv.Itoa(http.StatusInternalServerError), startTime)
		return
	}

	metrics.ReportAPIRequestMetric(handleName, r.Method, strconv.Itoa(http.StatusOK), startTime)
}

func (s *Server) validatingWebhook(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	// only hook create and update operation
	if req.Operation != v1beta1.Create && req.Operation != v1beta1.Update {
		blog.Warnf("operation is not create or update, ignore")
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	// only hook portpool and ingress
	if req.Kind.Kind != "PortPool" && req.Kind.Kind != "Ingress" {
		blog.Warnf("kind %s is not PortPool or Ingress", req.Kind.Kind)
		return errResponse(fmt.Errorf("kind %s is not PortPool or Ingress", req.Kind.Kind))
	}
	if req.Kind.Group != "networkextension.bkbcs.tencent.com" {
		blog.Warnf("group %s is not networkextension.bkbcs.tencent.com", req.Kind.Group)
		return errResponse(fmt.Errorf("group %s is not networkextension.bkbcs.tencent.com", req.Kind.Group))
	}
	// validate port pool
	if req.Kind.Kind == "PortPool" {
		portPool := &networkextensionv1.PortPool{}
		if err := json.Unmarshal(req.Object.Raw, portPool); err != nil {
			blog.Warnf("decode %s to port pool failed, err %s", string(req.Object.Raw), err.Error)
			return errResponse(fmt.Errorf("decode %s to port pool failed, err %s", string(req.Object.Raw), err.Error()))
		}
		if err := s.validatePortPool(portPool); err != nil {
			blog.Warnf("PortPool %s/%s is invalid, err %s", portPool.GetName(), portPool.GetNamespace(), err.Error())
			return errResponse(fmt.Errorf("PortPool %s/%s is invalid, err %s",
				portPool.GetName(), portPool.GetNamespace(), err.Error()))
		}
	}

	return &v1beta1.AdmissionResponse{Allowed: true}
}

func (s *Server) mutatingWebhook(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	if req.Operation != v1beta1.Create {
		blog.Warnf("operation is not create, ignore")
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	// only hook create operation of pod
	if req.Kind.Kind != "Pod" {
		blog.Warnf("kind %s is not Pod", req.Kind.Kind)
		return errResponse(fmt.Errorf("kind %s is not Pod", req.Kind.Kind))
	}
	pod := &k8scorev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Warnf("decode %s to pod failed, err %s", string(req.Object.Raw), err.Error)
		return errResponse(fmt.Errorf("decode %s to pod failed, err %s", string(req.Object.Raw), err.Error()))
	}
	pod.Namespace = req.Namespace
	pod.Name = req.Name
	_, ok := pod.Annotations[constant.AnnotationForPortPool]
	if !ok {
		blog.Warnf("pod %s/%s has no portpool annotation", pod.GetName(), pod.GetNamespace())
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	// do mutate
	blog.Infof("pod %s/%s do port inject", pod.GetName(), pod.GetNamespace())
	patches, err := s.mutatingPod(pod)
	if err != nil {
		blog.Warnf("mutating pod failed, err %s", err.Error())
		return errResponse(fmt.Errorf("mutating pod failed, err %s", err.Error()))
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Warnf("encoding patches faile, err %s", err.Error())
		return errResponse(fmt.Errorf("encoding patches faile, err %s", err.Error()))
	}
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// convert error to admission response
func errResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
}
