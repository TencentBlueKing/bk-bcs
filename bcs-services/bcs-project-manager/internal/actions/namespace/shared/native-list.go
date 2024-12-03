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

package shared

import (
	"context"
	"encoding/json"
	"strings"

	spb "google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	nsutils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/namespace"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListNativeNamespaces implement for ListNativeNamespaces interface
func (a *SharedNamespaceAction) ListNativeNamespaces(ctx context.Context,
	req *proto.ListNativeNamespacesRequest, resp *proto.ListNativeNamespacesResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return errorx.NewClusterErr(err.Error())
	}
	namespaces := nsList.Items
	if req.GetProjectIDOrCode() != "-" {
		project, err := a.model.GetProject(context.TODO(), req.GetProjectIDOrCode())
		if err != nil {
			logging.Error("get project from db failed, err: %s", err.Error())
			return errorx.NewDBErr(err.Error())
		}
		namespaces = nsutils.FilterNamespaces(nsList, true, project.ProjectCode)
	}
	retDatas := []*proto.NativeNamespaceData{}
	for _, namespace := range namespaces {
		projectCode, ok := namespace.Annotations[constant.AnnotationKeyProjectCode]
		if !ok {
			continue
		}
		p, err := a.model.GetProject(ctx, projectCode)
		if err != nil {
			logging.Error("get project %s from db failed, err: %s", projectCode, err.Error())
			return errorx.NewDBErr(err.Error())
		}
		retData := &proto.NativeNamespaceData{
			Uid:         string(namespace.GetUID()),
			Name:        namespace.GetName(),
			Status:      string(namespace.Status.Phase),
			CreateTime:  namespace.GetCreationTimestamp().Format(constant.TimeLayout),
			ProjectID:   p.ProjectID,
			ProjectCode: p.ProjectCode,
		}
		retDatas = append(retDatas, retData)
	}
	resp.Data = retDatas
	return nil
}

// ListNativeNamespacesContent implement for ListNativeNamespacesContent interface
func (a *SharedNamespaceAction) ListNativeNamespacesContent(ctx context.Context,
	req *proto.ListNativeNamespacesContentRequest, resp *spb.Struct) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	accept := common.GetAcceptType(ctx)
	nsList, err := client.CoreV1().RESTClient().Get().Resource("namespaces").SetHeader("Accept", accept).DoRaw(ctx)
	if err != nil {
		return errorx.NewClusterErr(err.Error())
	}

	if req.GetProjectIDOrCode() != "-" {
		project, errr := a.model.GetProject(context.TODO(), req.GetProjectIDOrCode())
		if errr != nil {
			logging.Error("get project from db failed, err: %s", errr.Error())
			return errorx.NewDBErr(errr.Error())
		}
		// Table和json返回的结果不同
		if strings.Contains(accept, "Table") {
			mt := &metav1.Table{}
			err = json.Unmarshal(nsList, mt)
			if err != nil {
				return errorx.NewInnerErr(err.Error())
			}
			mt, err = nsutils.FilterTableNamespaces(mt, true, project.ProjectCode)
			if err != nil {
				return errorx.NewInnerErr(err.Error())
			}
			nsList, err = json.Marshal(mt)
			if err != nil {
				return errorx.NewInnerErr(err.Error())
			}

		} else {
			mt := &v1.NamespaceList{}
			err = json.Unmarshal(nsList, mt)
			if err != nil {
				return errorx.NewInnerErr(err.Error())
			}
			mt.Items = nsutils.FilterNamespaces(mt, true, project.ProjectCode)
			nsList, err = json.Marshal(mt)
			if err != nil {
				return errorx.NewInnerErr(err.Error())
			}
		}
	}

	err = resp.UnmarshalJSON(nsList)
	if err != nil {
		return errorx.NewInnerErr(err.Error())
	}
	return nil
}
