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

package release

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	_struct "github.com/golang/protobuf/ptypes/struct"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetReleaseExtendAction return a new GetGetReleaseExtendAction instance
func NewGetReleaseExtendAction(releaseHandler release.Handler) *GetReleaseExtendAction {
	return &GetReleaseExtendAction{
		releaseHandler: releaseHandler,
	}
}

// GetReleaseExtendAction provides the action to do get release expend
type GetReleaseExtendAction struct {
	ctx context.Context

	releaseHandler release.Handler

	req  *helmmanager.GetReleaseDetailExtendReq
	resp *helmmanager.CommonResp
}

// Handle the release expend getting process
func (g *GetReleaseExtendAction) Handle(ctx context.Context,
	req *helmmanager.GetReleaseDetailExtendReq, resp *helmmanager.CommonResp) error {
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get release expend failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return g.getResource()
}

func (g *GetReleaseExtendAction) getResource() error {
	projectCode := contextx.GetProjectCodeFromCtx(g.ctx)
	clusterID := g.req.GetClusterID()
	namespace := g.req.GetNamespace()
	name := g.req.GetName()

	labelSelector := "app.kubernetes.io/instance=" + name
	// get k8s client set
	clientSet, err := component.GetK8SClientByClusterID(clusterID)
	if err != nil {
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}
	servicesList, err := clientSet.CoreV1().Services(namespace).List(g.ctx,
		metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	ingressestList, err := clientSet.NetworkingV1().Ingresses(namespace).List(g.ctx,
		metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	secretsList, err := clientSet.CoreV1().Secrets(namespace).List(g.ctx,
		metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	rsp := map[string]interface{}{
		"service": getService(servicesList),
		"ingress": getIngress(ingressestList),
		"secret":  getSecret(secretsList),
	}

	result, err := common.MarshalInterfaceToValue(rsp)
	if err != nil {
		blog.Errorf("marshal rsp err, %s", err.Error())
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	g.setResp(common.ErrHelmManagerSuccess, "ok", result)
	blog.Infof("get release expend successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s",
		projectCode, clusterID, namespace, name)
	return nil
}

func getService(serviceList *corev1.ServiceList) []map[string]interface{} {
	rsps := make([]map[string]interface{}, 0)
	for _, v := range serviceList.Items {
		rsp := map[string]interface{}{
			"name":       v.ObjectMeta.Name,
			"clusterIPs": v.Spec.ClusterIPs,
			"port":       v.Spec.Ports,
		}
		rsps = append(rsps, rsp)
	}
	return rsps
}

func getIngress(ingressList *networkingv1.IngressList) []map[string]interface{} {
	rsps := make([]map[string]interface{}, 0)
	for _, v := range ingressList.Items {
		rsp := map[string]interface{}{
			"name":  v.ObjectMeta.Name,
			"rules": v.Spec.Rules,
		}
		rsps = append(rsps, rsp)
	}
	return rsps
}

func getSecret(secretList *corev1.SecretList) []map[string]interface{} {
	rsps := make([]map[string]interface{}, 0)
	for _, v := range secretList.Items {
		rsp := map[string]interface{}{
			"name": v.ObjectMeta.Name,
			"data": v.Data,
		}
		rsps = append(rsps, rsp)
	}
	return rsps
}

func (g *GetReleaseExtendAction) setResp(err common.HelmManagerError, message string,
	r *_struct.Struct) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}
