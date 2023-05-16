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

package release

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"helm.sh/helm/v3/pkg/storage/driver"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetReleasePodsAction return a new GetGetReleaseStatusAction instance
func NewGetReleasePodsAction(releaseHandler release.Handler) *GetReleasePodsAction {
	return &GetReleasePodsAction{
		releaseHandler: releaseHandler,
	}
}

// GetReleasePodsAction provides the action to do get release pods
type GetReleasePodsAction struct {
	ctx context.Context

	releaseHandler release.Handler

	req  *helmmanager.GetReleasePodsReq
	resp *helmmanager.CommonListResp
}

// Handle the release pods getting process
func (g *GetReleasePodsAction) Handle(ctx context.Context,
	req *helmmanager.GetReleasePodsReq, resp *helmmanager.CommonListResp) error {
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get release pods failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	result, err := g.getResourcePods()
	if err != nil {
		blog.Errorf("get release pods failed, %s, clusterID: %s namespace: %s, name: %s", err.Error(),
			g.req.GetClusterID(), g.req.GetNamespace(), g.req.GetName())
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}
	g.setResp(common.ErrHelmManagerSuccess, "ok", result)
	blog.Infof("get release pods successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s",
		g.req.GetProjectCode(), g.req.GetClusterID(), g.req.GetNamespace(), g.req.GetName())
	return nil
}

func (g *GetReleasePodsAction) getResourcePods() (*_struct.ListValue, error) {
	clusterID := g.req.GetClusterID()
	namespace := g.req.GetNamespace()
	name := g.req.GetName()

	rl, err := g.releaseHandler.Cluster(clusterID).Get(g.ctx, release.GetOption{
		Namespace: namespace,
		Name:      name,
		GetObject: true,
	})
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		return &_struct.ListValue{}, nil
	}
	if err != nil {
		return nil, err
	}
	if rl.Infos == nil {
		return &_struct.ListValue{}, nil
	}
	allPods, err := storage.GetPods(clusterID, namespace)
	if err != nil {
		blog.Errorf("get release pods failed, %s, clusterID: %s namespace: %s, name: %s", err.Error(), clusterID,
			namespace, name)
		return nil, err
	}
	pods := make([]*corev1.Pod, 0)
	for _, pod := range allPods {
		if g.podIsOwnRelease(pod.Data, rl.Infos) &&
			pod.Data.CreationTimestamp.Local().Unix() > int64(g.req.GetAfter()) {
			pod.Data.APIVersion = "v1"
			pod.Data.Kind = "Pod"
			pods = append(pods, pod.Data)
		}
	}

	return common.MarshalInterfacesToListValue(pods)
}

func (g *GetReleasePodsAction) setResp(err common.HelmManagerError, message string,
	r *_struct.ListValue) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}

func (g *GetReleasePodsAction) podIsOwnRelease(pod *corev1.Pod, infos []*resource.Info) bool {
	for _, info := range infos {
		if info.Name == "" || info.Object == nil {
			continue
		}
		if !stringx.StringInSlice(info.Object.GetObjectKind().GroupVersionKind().Kind, availableKind) {
			continue
		}
		blog.V(6).Infof("check pod %s with info %s(%s)", pod.GetName(), info.Name,
			info.Object.GetObjectKind().GroupVersionKind().Kind)
		for _, or := range pod.OwnerReferences {
			if !stringx.StringInSlice(or.Kind, availableKind) {
				continue
			}
			if or.Kind == info.Object.GetObjectKind().GroupVersionKind().Kind &&
				or.Name == info.Name {
				return true
			}
			if pod.Labels[labelControllerType] == info.Object.GetObjectKind().GroupVersionKind().Kind &&
				pod.Labels[labelControllerName] == info.Name {
				return true
			}
		}
	}
	return false
}

var availableKind = []string{
	"ReplicaSet",
	"Deployment",
	"StatefulSet",
	"GameDeployment",
	"GameStatefulSet",
}

const (
	labelControllerType = "io.tencent.bcs.controller.type"
	labelControllerName = "io.tencent.bcs.controller.name"
)
