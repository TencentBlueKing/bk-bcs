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

package proxy

import (
	"context"
	"crypto/tls"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/common"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned/typed/tkex/v1alpha1"
	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type InstanceProxyDispatcher struct {
	InstanceVarName string
	SubPathVarName  string
	tkexIf          tkexv1alpha1.TkexV1alpha1Interface
}

// NewInstanceProxyDispatcher create a TunnelProxyDispatcher
func NewInstanceProxyDispatcher(
	instanceVarName, subPathVarName string,
	tkexIf tkexv1alpha1.TkexV1alpha1Interface) *InstanceProxyDispatcher {
	return &InstanceProxyDispatcher{
		InstanceVarName: instanceVarName,
		SubPathVarName:  subPathVarName,
		tkexIf:          tkexIf,
	}
}

// ServeHTTP implements http.Handler
func (f *InstanceProxyDispatcher) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	// Get instance id
	instanceID := vars[f.InstanceVarName]

	if f.tkexIf == nil {
		blog.Errorf("[instance-proxy] tkexIf has not init")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	instance, err := f.tkexIf.ArgocdInstances(common.ArgocdManagerNamespace).Get(context.TODO(), instanceID, metav1.GetOptions{})
	if err != nil {
		blog.Error("[instance-proxy] get instance failed, err: %s", err.Error())
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	if instance.Status.ServerHost == "" {
		blog.Error("[instance-proxy] argocd instance server host is empty")
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	fullPath := "https://" + instance.Status.ServerHost + ":443" + "/" + vars[f.SubPathVarName]
	// proxy to argocd server
	u, err := url.Parse(fullPath)
	if nil != err {
		blog.Errorf("[instance-proxy] parse argocd server url failed, err: %s", err.Error())
		return
	}

	proxy := httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL = u
		},
		ErrorHandler: func(rw http.ResponseWriter, req *http.Request, err error) {
			blog.Errorf("[instance-proxy] proxy request failed, err: %s", err.Error())
			rw.WriteHeader(http.StatusInternalServerError)
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	blog.Infof("[instance-proxy] proxy to instance: %s, url: %s", instance.GetName(), u.String())
	proxy.ServeHTTP(rw, req)
}
