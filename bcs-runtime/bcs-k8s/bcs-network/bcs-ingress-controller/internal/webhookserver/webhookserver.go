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
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/eventer"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	v1 "k8s.io/api/admission/v1"

	"k8s.io/api/admission/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	deserializer = codecs.UniversalDeserializer()
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
	k8sClient    client.Client
	eventWatcher eventer.WatchEventInterface
	poolCache    *portpoolcache.Cache
	podName      string
	podNamespace string
}

// NewHookServer create new hook server object
func NewHookServer(opt *ServerOption, k8sClient client.Client, poolCache *portpoolcache.Cache,
	eventWatcher eventer.WatchEventInterface) (*Server, error) {
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
		k8sClient:    k8sClient,
		eventWatcher: eventWatcher,
		poolCache:    poolCache,
		podName:      os.Getenv(constant.EnvIngressPodName),
		podNamespace: os.Getenv(constant.EnvIngressPodNamespace),
	}, nil
}

// Start start http server
func (s *Server) Start(stop <-chan struct{}) error {
	blog.Infof("start webhook server")
	mux := http.NewServeMux()
	// register handler function
	mux.HandleFunc("/portpool/v1/validate", s.HandleValidatingWebhook)
	mux.HandleFunc("/portpool/v1/mutate", s.HandleMutatingWebhook)
	mux.HandleFunc("/crd/v1/validate", s.HandleValidatingCRD)
	s.server.Handler = mux

	go func() {
		if err := s.server.ListenAndServeTLS("", ""); err != nil {
			blog.Fatalf("failed to listen and serve webhook server, err %s", err.Error())
		}
	}()

	blog.Infof("webhook server started")

	// patch pod label to add leader label
	if err := s.patchPod(s.podName, s.podNamespace, constant.LeaderLabelValueTrue); err != nil {
		blog.Errorf("failed to patch pod %s/%s, err %s", s.podNamespace, s.podName, err.Error())
		return err
	}
	<-stop
	blog.Infof("Got controller stop signal, shutting down webhook server gracefully...")
	s.server.Shutdown(context.Background())
	// patch pod label to remove leader
	if err := s.patchPod(s.podName, s.podNamespace, constant.LeaderLabelValueFalse); err != nil {
		blog.Errorf("failed to patch pod %s/%s, err %s", s.podNamespace, s.podName, err.Error())
		return err
	}

	return nil
}

// NeedLeaderElection return true if need leader election
func (s *Server) NeedLeaderElection() bool {
	return true
}

// HandleValidatingWebhook handle validating webhook request
func (s *Server) HandleValidatingWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, "validate", newDelegateToV1AdmitHandler(s.validatingWebhook))
}

// HandleMutatingWebhook handle mutating webhook request
func (s *Server) HandleMutatingWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, "mutate", newDelegateToV1AdmitHandler(s.mutatingWebhook))
}

// HandleValidatingCRD handle validating CRD delete webhook request
func (s *Server) HandleValidatingCRD(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, "validateCRD", newDelegateToV1AdmitHandler(s.validatingCRDDelete))
}

func (s *Server) handleWebhook(
	w http.ResponseWriter, r *http.Request, handleName string,
	admit admitHandler) {
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

	obj, gvk, err := deserializer.Decode(body, nil, nil)
	if err != nil {
		werr := errors.Wrapf(err, "could not decode body")
		blog.Error(werr.Error())
		http.Error(w, werr.Error(), http.StatusBadRequest)
		return
	}

	var responseObj runtime.Object
	switch *gvk {
	case v1beta1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1beta1.AdmissionReview)
		if !ok {
			blog.Errorf("Expected v1beta1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissionReview := &v1beta1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1beta1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	case v1.SchemeGroupVersion.WithKind("AdmissionReview"):
		requestedAdmissionReview, ok := obj.(*v1.AdmissionReview)
		if !ok {
			blog.Errorf("Expected v1.AdmissionReview but got: %T", obj)
			return
		}
		responseAdmissionReview := &v1.AdmissionReview{}
		responseAdmissionReview.SetGroupVersionKind(*gvk)
		responseAdmissionReview.Response = admit.v1(*requestedAdmissionReview)
		responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
		responseObj = responseAdmissionReview
	default:
		msg := fmt.Sprintf("Unsupported group version kind: %v", gvk)
		blog.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	blog.V(2).Info(fmt.Sprintf("sending response: %v", responseObj))
	respBytes, err := json.Marshal(responseObj)
	if err != nil {
		blog.Error(err.Error())
		http.Error(w, fmt.Sprintf("could encode response: %v", err), http.StatusInternalServerError)
		metrics.ReportAPIRequestMetric(
			handleName, r.Method, strconv.Itoa(http.StatusInternalServerError), startTime)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBytes); err != nil {
		blog.Errorf("Could not write response: %v", err)
		http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
		metrics.ReportAPIRequestMetric(
			handleName, r.Method, strconv.Itoa(http.StatusInternalServerError), startTime)
		return
	}

	metrics.ReportAPIRequestMetric(handleName, r.Method, strconv.Itoa(http.StatusOK), startTime)
}

func (s *Server) validatingWebhook(ar v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request
	// only hook create and update operation
	if req.Operation != v1.Create && req.Operation != v1.Update {
		blog.Warnf("operation is not create or update, ignore")
		return &v1.AdmissionResponse{Allowed: true}
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
			return errResponse(fmt.Errorf("decode %s to port pool failed, err %s", string(req.Object.Raw),
				err.Error()))
		}
		if err := s.validatePortPool(portPool); err != nil {
			blog.Warnf("PortPool %s/%s is invalid, err %s", portPool.GetName(), portPool.GetNamespace(), err.Error())
			return errResponse(fmt.Errorf("PortPool %s/%s is invalid, err %s",
				portPool.GetName(), portPool.GetNamespace(), err.Error()))
		}
	}

	return &v1.AdmissionResponse{Allowed: true}
}

func (s *Server) mutatingWebhook(ar v1.AdmissionReview) (response *v1.AdmissionResponse) {
	defer func() {
		if response == nil || response.Allowed == false {
			metrics.IncreasePodCreateCounter(false)
		} else {
			metrics.IncreasePodCreateCounter(true)
		}
	}()

	req := ar.Request
	if req.Operation != v1.Create {
		blog.Warnf("operation is not create, ignore")
		return &v1.AdmissionResponse{Allowed: true}
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
	if len(pod.Namespace) == 0 {
		pod.Namespace = req.Namespace
	}
	if len(pod.Name) == 0 {
		pod.Name = req.Name
	}
	_, ok := pod.Annotations[constant.AnnotationForPortPool]
	if !ok {
		blog.Warnf("pod %s/%s has no portpool annotation", pod.GetName(), pod.GetNamespace())
		return &v1.AdmissionResponse{Allowed: true}
	}

	blog.Infof("received pod '%s/%s' create event", pod.GetNamespace(), pod.GetName())
	patches, err := s.mutatingPod(pod)
	if err != nil {
		blog.Errorf("mutating pod '%s/%s' got an error: %s", pod.GetNamespace(), pod.GetName(), err.Error())
		return errResponse(errors.Wrapf(err, "mutating pod '%s/%s' failed",
			pod.GetNamespace(), pod.GetNamespace()))
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("marshal pod '%s/%s' patches failed: %s", pod.GetNamespace(), pod.GetName(), err.Error())
		return errResponse(errors.Wrapf(err, "encoding patches for '%s/%s' failed",
			pod.GetNamespace(), pod.GetNamespace()))
	}
	return &v1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1.PatchType {
			pt := v1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func (s *Server) validatingCRDDelete(ar v1.AdmissionReview) *v1.AdmissionResponse {
	allowResp := &v1.AdmissionResponse{Allowed: true}

	req := ar.Request
	if req.Operation != v1.Delete {
		blog.Warnf("operation is not delete, ignore")
		return allowResp
	}
	// only hook delete operation of CRD
	if req.Kind.Kind != constant.KindCRD {
		blog.Warnf("kind %s is not CRD", req.Kind.Kind)
		return errResponse(fmt.Errorf("kind %s is not CRD", req.Kind.Kind))
	}
	labels, err := s.getCRDLabelFromAR(ar)
	if err != nil {
		blog.Warnf("get CRD from admissionReview failed, err: %s", err.Error())
		return errResponse(err)
	}
	if err := s.validateCRDDeletion(labels); err != nil {
		return errResponse(err)
	}

	return allowResp
}

// convert error to admission response
func errResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
}

func (s *Server) patchPod(name, namespace, isLeader string) error {
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		patchStruct := map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]string{
					constant.LeaderLabel: isLeader,
				},
			},
		}
		patchData, err := json.Marshal(patchStruct)
		if err != nil {
			return err
		}
		updatePod := &k8scorev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		return s.k8sClient.Patch(context.TODO(), updatePod, client.RawPatch(types.MergePatchType, patchData))
	})
	if err != nil {
		return err
	}
	return nil
}
