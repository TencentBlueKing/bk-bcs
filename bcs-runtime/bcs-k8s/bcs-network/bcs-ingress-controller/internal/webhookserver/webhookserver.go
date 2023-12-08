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

// Package webhookserver contains webhook logic
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
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/conflicthandler"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"

	v1 "k8s.io/api/admission/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/eventer"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

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
	Addrs          []string
	Port           int
	ServerCertFile string
	ServerKeyFile  string
}

// Server webhook server
type Server struct {
	ipv6Server *ipv6server.IPv6Server
	// k8s client
	k8sClient        client.Client
	lbClient         cloud.LoadBalance
	eventWatcher     eventer.WatchEventInterface
	poolCache        *portpoolcache.Cache
	podName          string
	podNamespace     string
	ingressValidater cloud.Validater
	ingressConverter *generator.IngressConverter
	defaultRegion    string
	conflictHandler  *conflicthandler.ConflictHandler
	// if node's annotation have not related portpool namespace, will use NodePortBindingNs as default
	nodePortBindingNs string
	eventer           record.EventRecorder
}

// NewHookServer create new hook server object
func NewHookServer(opt *ServerOption, k8sClient client.Client, lbClient cloud.LoadBalance, poolCache *portpoolcache.Cache,
	eventWatcher eventer.WatchEventInterface, validater cloud.Validater, converter *generator.IngressConverter,
	conflictHandler *conflicthandler.ConflictHandler, nodePortBindingNs string,
	eventer record.EventRecorder) (*Server, error) {
	pair, err := tls.LoadX509KeyPair(opt.ServerCertFile, opt.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("load x509 key pair cert %s, key %s failed, err %s",
			opt.ServerCertFile, opt.ServerKeyFile, err.Error())
	}

	return &Server{
		ipv6Server: ipv6server.NewTlsIPv6Server(opt.Addrs, strconv.Itoa(opt.Port), "",
			&tls.Config{Certificates: []tls.Certificate{pair}}, nil),
		k8sClient:         k8sClient,
		lbClient:          lbClient,
		eventWatcher:      eventWatcher,
		poolCache:         poolCache,
		podName:           os.Getenv(constant.EnvIngressPodName),
		podNamespace:      os.Getenv(constant.EnvIngressPodNamespace),
		ingressValidater:  validater,
		ingressConverter:  converter,
		conflictHandler:   conflictHandler,
		nodePortBindingNs: nodePortBindingNs,
		eventer:           eventer,
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
	mux.HandleFunc("/ingress/v1/mutate", s.HandleValidatingIngress)
	mux.HandleFunc("/node/v1/mutate", s.HandleMutatingNodeWebhook)
	// 兼容IPV6
	s.ipv6Server.Server.Handler = mux

	go func() {
		if err := s.ipv6Server.ListenAndServeTLS("", ""); err != nil {
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
	s.ipv6Server.Shutdown(context.Background())
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

// HandleValidatingIngress handler validating ingress
func (s *Server) HandleValidatingIngress(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, "validateIngress", newDelegateToV1AdmitHandler(s.mutatingIngress))
}

// HandleMutatingNodeWebhook handle mutating node webhook request
func (s *Server) HandleMutatingNodeWebhook(w http.ResponseWriter, r *http.Request) {
	s.handleWebhook(w, r, "mutateNode", newDelegateToV1AdmitHandler(s.mutatingNodeWebhook))
}

// handleWebhook 新旧版本K8S Webhook的返回结构体不一致， 这里需要自动适配
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

	blog.V(5).Info(fmt.Sprintf("sending response: %v", responseObj))
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

// validatingWebhook 校验portpool更新，避免端口冲突
func (s *Server) validatingWebhook(ar v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request
	// only hook create and update operation
	if req.Operation != v1.Create && req.Operation != v1.Update {
		blog.Warnf("operation is not create or update, ignore")
		return &v1.AdmissionResponse{Allowed: true}
	}
	// only hook portpool and ingress
	if req.Kind.Kind != "PortPool" {
		blog.Warnf("kind %s is not PortPool", req.Kind.Kind)
		return errResponse(fmt.Errorf("kind %s is not PortPool or Ingress", req.Kind.Kind))
	}
	if req.Kind.Group != "networkextension.bkbcs.tencent.com" {
		blog.Warnf("group %s is not networkextension.bkbcs.tencent.com", req.Kind.Group)
		return errResponse(fmt.Errorf("group %s is not networkextension.bkbcs.tencent.com", req.Kind.Group))
	}
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

	return &v1.AdmissionResponse{Allowed: true}
}

// mutatingIngress 校验用户对ingress的更新，包括端口冲突和配置， warning级别的错误会patch到ingress的注解上
func (s *Server) mutatingIngress(ar v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request
	// only hook create and update operation
	if req.Operation != v1.Create && req.Operation != v1.Update {
		blog.Warnf("operation is not create or update, ignore")
		return &v1.AdmissionResponse{Allowed: true}
	}
	// only hook portpool and ingress
	if req.Kind.Kind != "Ingress" {
		blog.Warnf("kind %s is not Ingress", req.Kind.Kind)
		return errResponse(fmt.Errorf("kind %s is not PortPool or Ingress", req.Kind.Kind))
	}
	if req.Kind.Group != "networkextension.bkbcs.tencent.com" {
		blog.Warnf("group %s is not networkextension.bkbcs.tencent.com", req.Kind.Group)
		return errResponse(fmt.Errorf("group %s is not networkextension.bkbcs.tencent.com", req.Kind.Group))
	}

	ingress := &networkextensionv1.Ingress{}
	if err := json.Unmarshal(req.Object.Raw, ingress); err != nil {
		blog.Warnf("decode %s to ingress failed, err %s", string(req.Object.Raw), err.Error)
		return errResponse(fmt.Errorf("decode %s to ingress failed, err %s", string(req.Object.Raw),
			err.Error()))
	}
	patches, err := s.mutateIngress(ingress, req.Operation)
	if err != nil {
		blog.Warnf("mutate ingress failed, err: %v", err)
		return errResponse(err)
	}

	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		err = errors.Wrapf(err, "marshal ingress '%s/%s' patches failed", ingress.GetNamespace(), ingress.GetName())
		blog.Warnf(err.Error())
		return errResponse(err)
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

// mutatingWebhook 根据用户注解，分配端口池端口到Pod
func (s *Server) mutatingWebhook(ar v1.AdmissionReview) (response *v1.AdmissionResponse) {
	// 统计webhook执行成功/失败次数
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
		blog.Infof("pod %s/%s has no portpool annotation", pod.GetName(), pod.GetNamespace())
		return &v1.AdmissionResponse{Allowed: true}
	}

	blog.Infof("received pod '%s/%s' create event", pod.GetNamespace(), pod.GetName())
	patches, err := s.mutatingPod(pod)
	if err != nil {
		blog.Errorf("mutating pod '%s/%s' got an error: %s", pod.GetNamespace(), pod.GetName(), err.Error())
		s.eventer.Eventf(pod, k8scorev1.EventTypeWarning, "AllocatePortFailed",
			fmt.Sprintf("pod '%s/%s' allocate port failed: %s", pod.GetNamespace(), pod.GetName(), err.Error()))
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

// validatingCRDDelete 根据删除策略，避免用户误删除正在使用中的CRD，导致相关资源被销毁
func (s *Server) validatingCRDDelete(ar v1.AdmissionReview) *v1.AdmissionResponse {
	allowResp := &v1.AdmissionResponse{Allowed: true}
	req := ar.Request
	if req.Operation != v1.Delete && req.Kind.Kind != constant.KindCRD {
		return allowResp
	}
	if !strings.Contains(req.Name, networkextensionv1.GroupVersion.Group) {
		return allowResp
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

func (s *Server) mutatingNodeWebhook(ar v1.AdmissionReview) (response *v1.AdmissionResponse) {
	req := ar.Request
	if req.Operation != v1.Create && req.Operation != v1.Update {
		blog.Warnf("operation is not create, ignore")
		return &v1.AdmissionResponse{Allowed: true}
	}
	// only hook create operation of pod
	if req.Kind.Kind != "Node" {
		blog.Warnf("kind %s is not Node", req.Kind.Kind)
		return &v1.AdmissionResponse{Allowed: true}
	}
	node := &k8scorev1.Node{}
	if err := json.Unmarshal(req.Object.Raw, node); err != nil {
		blog.Errorf("decode %s to node failed, err %s", string(req.Object.Raw), err.Error)
		return &v1.AdmissionResponse{Allowed: true}
	}
	if len(node.Namespace) == 0 {
		node.Namespace = req.Namespace
	}
	if len(node.Name) == 0 {
		node.Name = req.Name
	}
	_, ok := node.Annotations[constant.AnnotationForPortPool]
	if !ok {
		blog.Infof("node %s/%s has no portpool annotation", node.GetName(), node.GetNamespace())
		return &v1.AdmissionResponse{Allowed: true}
	}

	blog.Infof("received node '%s/%s' create/update event", node.GetNamespace(), node.GetName())
	patches, err := s.mutatingNode(node)
	if err != nil {
		blog.Errorf("mutating node '%s/%s' got an error: %s", node.GetNamespace(), node.GetName(), err.Error())
		s.eventer.Eventf(node, k8scorev1.EventTypeWarning, constant.EventReasonAllocatePortFailed,
			fmt.Sprintf("node '%s' allocate port failed: %s", node.GetName(), err.Error()))
		return &v1.AdmissionResponse{Allowed: true}
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("marshal node'%s/%s' patches failed: %s", node.GetNamespace(), node.GetName(), err.Error())
		s.eventer.Eventf(node, k8scorev1.EventTypeWarning, constant.EventReasonAllocatePortFailed,
			fmt.Sprintf("node '%s' patch port info failed: %s", node.GetName(), err.Error()))
		return &v1.AdmissionResponse{Allowed: true}
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
