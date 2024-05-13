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

// Package filterclb filters invalid cloud loadbalance
package filterclb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
)

const (
	pluginName = "filterclb"

	annotationServiceInternalSubnetID = "service.kubernetes.io/qcloud-loadbalancer-internal-subnetid"
	annotationServiceExistedLbID      = "service.kubernetes.io/tke-existed-lbid"

	annotationIngressSubnetID  = "kubernetes.io/ingress.subnetId"
	annotationIngressExistLbID = "kubernetes.io/ingress.existLbId"
	annotationsIngressClass    = "kubernetes.io/ingress.class"

	annotationSkipDenyFilterCLB = "bkbcs.tencent.com/skip-filter-clb"

	annotationIngressClassDefaultKey = "ingressclass.kubernetes.io/is-default-class"

	ingressClassQcloud = "qcloud"

	l7LbControllerName = "l7-lb-controller"

	constStringTrue = "true"
)

var (
	serviceKind = metav1.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	}
	ingressExtensionsV1Kind = metav1.GroupVersionKind{
		Group:   "extensions",
		Version: "v1beta1",
		Kind:    "Ingress",
	}
	ingressV1Beta1Kind = metav1.GroupVersionKind{
		Group:   "networking.k8s.io",
		Version: "v1beta1",
		Kind:    "Ingress",
	}
	ingressV1Kind = metav1.GroupVersionKind{
		Group:   "networking.k8s.io",
		Version: "v1",
		Kind:    "Ingress",
	}
)

func init() {
	p := &Handler{}
	pluginmanager.Register(pluginName, p)
}

// Handler defines the handler of filter auto create clb
type Handler struct {
	gvkArrayString      []string
	supportIngressClass bool
	kubeClient          clientset.Interface
}

// AnnotationKey return the annotation key of filterclb
func (h *Handler) AnnotationKey() string {
	return ""
}

// Init init the filterclb plugin
func (h *Handler) Init(configFilePath string) error {
	gvkArrayString := make([]string, 0, 4)
	gvkArrayString = append(gvkArrayString, serviceKind.String(), ingressV1Kind.String(),
		ingressExtensionsV1Kind.String(), ingressV1Beta1Kind.String())
	h.gvkArrayString = gvkArrayString

	client, err := getK8sClient()
	if err != nil {
		return err
	}
	h.supportIngressClass = checkSupportIngressClass(client)
	blog.Infof("filterclb: support ingress class %v", h.supportIngressClass)
	h.kubeClient = client
	return nil
}

// Handle handle the request of filterclb
func (h *Handler) Handle(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := review.Request
	// 只拦截service和ingress 创建请求
	findGvk := false
	for _, gvk := range h.gvkArrayString {
		if req.Kind.String() == gvk {
			findGvk = true
			break
		}
	}
	if !findGvk {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	// 只处理创建和更新
	if req.Operation != v1beta1.Create && req.Operation != v1beta1.Update {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	// 含有特定annotations 也过滤
	started := time.Now()
	blog.Infof("filterclb: %s %s/%s", req.Operation, req.Namespace, req.Name)
	if req.Kind.Kind == "Service" {
		svc := &corev1.Service{}
		if err := json.Unmarshal(req.Object.Raw, svc); err != nil {
			blog.Errorf("cannot decode raw object %s to svc, err %s", string(req.Object.Raw), err.Error())
			metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
			return pluginutil.ToAdmissionResponse(err)
		}
		if svc.ObjectMeta.Namespace == "" {
			svc.ObjectMeta.Namespace = req.Namespace
		}
		if h.DenyService(svc) {
			blog.Infof("filterclb: %s %s/%s deny service", req.Operation, req.Namespace, req.Name)
			metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
			return pluginutil.ToAdmissionResponse(fmt.Errorf("service %s/%s is not allowed to create, "+
				"It is forbidden to directly create external network clb", req.Namespace, req.Name))
		}

		// pass
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusSuccess, started)
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	if req.Kind.Kind == "Ingress" {
		ingress := &v1.Ingress{}
		if err := json.Unmarshal(req.Object.Raw, ingress); err != nil {
			blog.Errorf("cannot decode raw object %s to ingress, err %s", string(req.Object.Raw), err.Error())
			metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
			return pluginutil.ToAdmissionResponse(err)
		}
		if ingress.ObjectMeta.Namespace == "" {
			ingress.ObjectMeta.Namespace = req.Namespace
		}
		if h.DenyIngress(ingress) {
			blog.Infof("filterclb: %s %s/%s deny ingress", req.Operation, req.Namespace, req.Name)
			metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
			return pluginutil.ToAdmissionResponse(fmt.Errorf("ingress %s/%s is not allowed to create, "+
				"It is forbidden to directly create external network clb", req.Namespace, req.Name))
		}

		// pass
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusSuccess, started)
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusSuccess, started)
	return &v1beta1.AdmissionResponse{Allowed: true}
}

// DenyService deny the service
func (h *Handler) DenyService(svc *corev1.Service) bool {
	blog.Infof("filterclb: svc %s/%s annotations %v, type=%s", svc.Name, svc.Namespace, svc.Annotations, svc.Spec.Type)
	if svc.Annotations[annotationSkipDenyFilterCLB] == constStringTrue {
		return false
	}
	if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
		return false
	}
	if svc.Annotations[annotationServiceInternalSubnetID] != "" {
		return false
	}
	if svc.Annotations[annotationServiceExistedLbID] != "" {
		return false
	}
	return true
}

// DenyIngress deny the ingress
func (h *Handler) DenyIngress(ingress *v1.Ingress) bool {
	if ingress.Annotations[annotationSkipDenyFilterCLB] == constStringTrue {
		return false
	}
	if ingress.Annotations[annotationIngressSubnetID] != "" {
		return false
	}
	if ingress.Annotations[annotationIngressExistLbID] != "" {
		return false
	}
	if ingress.Annotations[annotationsIngressClass] != "" &&
		ingress.Annotations[annotationsIngressClass] != ingressClassQcloud {
		return false
	}
	if ingress.Spec.IngressClassName != nil && *ingress.Spec.IngressClassName != ingressClassQcloud {
		return false
	}

	if h.supportIngressClass {
		ingressClassList, err := h.kubeClient.NetworkingV1().IngressClasses().List(context.Background(),
			metav1.ListOptions{})
		if err != nil {
			blog.Errorf("get ingress class list failed, err %s", err.Error())
			return true
		}
		// 只有一个，并且不是qcloud，不需要拦截
		if len(ingressClassList.Items) == 1 && ingressClassList.Items[0].Name != ingressClassQcloud {
			return false
		}
		// 默认的是非qcloud，不需要拦截
		for _, ingressClass := range ingressClassList.Items {
			// 也许可能存在多个默认，但是就必须指定ingressClassName了（k8s设定）
			if ingressClass.Annotations[annotationIngressClassDefaultKey] == constStringTrue &&
				ingressClass.Name != ingressClassQcloud {
				return false
			}
		}
	} else {
		// 1.18以下版本，判断是否存在deploy l7-lb-controller
		_, err := h.kubeClient.ExtensionsV1beta1().Deployments("kube-system").Get(context.Background(),
			l7LbControllerName, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				blog.Infof("get deploy %s not found, skip filter clb", l7LbControllerName)
				return false
			}
			blog.Errorf("get deploy %s failed, err %s", l7LbControllerName, err.Error())
			return true
		}
	}

	return true
}

// Close close the handler
func (h *Handler) Close() error {
	return nil
}

func checkSupportIngressClass(client clientset.Interface) bool {
	version118, _ := version.ParseGeneric("v1.18.0")
	serverVersion, err := client.Discovery().ServerVersion()
	if err != nil {
		return false
	}

	runningVersion, err := version.ParseGeneric(serverVersion.String())
	if err != nil {
		blog.Errorf("parse server version %s failed, err %s", serverVersion.String(), err.Error())
		return false
	}
	return runningVersion.AtLeast(version118)
}

func getK8sClient() (*kubernetes.Clientset, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "get incluster rest config failed")
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "init incluster k8s client failed")
	}
	return client, nil
}
