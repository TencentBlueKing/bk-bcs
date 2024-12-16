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

package handler

import (
	"context"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	authnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	spb "google.golang.org/protobuf/types/known/structpb"

	na "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/convert"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// NamespaceHandler ...
type NamespaceHandler struct {
	model store.ProjectModel
}

// NewNamespace return a variable service hander
func NewNamespace(model store.ProjectModel) *NamespaceHandler {
	return &NamespaceHandler{
		model: model,
	}
}

// SyncNamespace implement for SyncNamespace interface
func (p *NamespaceHandler) SyncNamespace(ctx context.Context,
	req *proto.SyncNamespaceRequest, resp *proto.SyncNamespaceResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.SyncNamespace(ctx, req, resp)
}

// WithdrawNamespace implement for WithdrawNamespace interface
func (p *NamespaceHandler) WithdrawNamespace(ctx context.Context,
	req *proto.WithdrawNamespaceRequest, resp *proto.WithdrawNamespaceResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.WithdrawNamespace(ctx, req, resp)
}

// CreateNamespace implement for CreateNamespace interface
func (p *NamespaceHandler) CreateNamespace(ctx context.Context,
	req *proto.CreateNamespaceRequest, resp *proto.CreateNamespaceResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.CreateNamespace(ctx, req, resp)
}

// CreateNamespaceCallback implement for CreateNamespaceCallback interface
func (p *NamespaceHandler) CreateNamespaceCallback(ctx context.Context,
	req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.CreateNamespaceCallback(ctx, req, resp)
}

// UpdateNamespace implement for UpdateNamespace interface
func (p *NamespaceHandler) UpdateNamespace(ctx context.Context,
	req *proto.UpdateNamespaceRequest, resp *proto.UpdateNamespaceResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.UpdateNamespace(ctx, req, resp)
}

// UpdateNamespaceCallback implement for UpdateNamespaceCallback interface
func (p *NamespaceHandler) UpdateNamespaceCallback(ctx context.Context,
	req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.UpdateNamespaceCallback(ctx, req, resp)
}

// GetNamespace implement for GetNamespace interface
func (p *NamespaceHandler) GetNamespace(ctx context.Context,
	req *proto.GetNamespaceRequest, resp *proto.GetNamespaceResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.GetNamespace(ctx, req, resp)
}

// ListNamespaces implement for ListNamespaces interface
func (p *NamespaceHandler) ListNamespaces(ctx context.Context,
	req *proto.ListNamespacesRequest, resp *proto.ListNamespacesResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	err = action.ListNamespaces(ctx, req, resp)
	if err != nil {
		return err
	}
	retData := sortNamespaces(resp.GetData())
	authUser, err := middleware.GetUserFromContext(ctx)
	if err == nil && authUser.Username != "" {
		p, err := p.model.GetProject(ctx, req.GetProjectCode())
		if err != nil {
			logging.Error("get project %s failed, err: %s", req.GetProjectCode(), err.Error())
			resp.Data = retData
			return nil
		}
		namespaces := []authnamespace.ProjectNamespaceData{}
		for _, ns := range retData {
			namespaces = append(namespaces, authnamespace.ProjectNamespaceData{
				Project:   p.ProjectID,
				Cluster:   req.GetClusterID(),
				Namespace: ns.GetName(),
			})
		}
		perms, err := auth.NamespaceIamClient.GetMultiNamespaceMultiActionPerm(
			authUser.Username, namespaces,
			[]string{auth.NamespaceCreate, auth.NamespaceView,
				auth.NamespaceUpdate, auth.NamespaceDelete,
				auth.NamespaceScopedCreate, auth.NamespaceScopedView,
				auth.NamespaceScopedUpdate, auth.NamespaceScopedDelete},
		)
		newPerms := map[string]map[string]bool{}
		for _, ns := range retData {
			newPerms[ns.GetName()] = perms[authutils.CalcIAMNsID(req.GetClusterID(), ns.GetName())]
			if ns.GetItsmTicketType() == nsm.ItsmTicketTypeCreate {
				newPerms[ns.GetName()]["namespace_view"] = true
			}
		}
		if err != nil {
			logging.Error("get multi namespaces multi action permission failed, err: %s", err.Error())
			resp.Data = retData
			return nil
		}
		resp.WebAnnotations = &proto.Perms{Perms: convert.MapBool2pbStruct(newPerms)}
	}
	resp.Data = retData
	return nil
}

// ListNativeNamespaces implement for ListNativeNamespaces interface
func (p *NamespaceHandler) ListNativeNamespaces(ctx context.Context,
	req *proto.ListNativeNamespacesRequest, resp *proto.ListNativeNamespacesResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectIDOrCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.ListNativeNamespaces(ctx, req, resp)
}

func sortNamespaces(list []*proto.NamespaceData) []*proto.NamespaceData {
	creating := []*proto.NamespaceData{}
	updating := []*proto.NamespaceData{}
	deleting := []*proto.NamespaceData{}
	exists := []*proto.NamespaceData{}
	for _, ns := range list {
		switch ns.GetItsmTicketType() {
		case nsm.ItsmTicketTypeCreate:
			creating = append(creating, ns)
		case nsm.ItsmTicketTypeUpdate:
			updating = append(updating, ns)
		case nsm.ItsmTicketTypeDelete:
			deleting = append(deleting, ns)
		default:
			exists = append(exists, ns)
		}
	}
	sort.SliceStable(creating, func(i, j int) bool {
		return creating[i].GetName() < creating[j].GetName()
	})
	sort.SliceStable(updating, func(i, j int) bool {
		return updating[i].GetName() < updating[j].GetName()
	})
	sort.SliceStable(deleting, func(i, j int) bool {
		return deleting[i].GetName() < deleting[j].GetName()
	})
	sort.SliceStable(exists, func(i, j int) bool {
		return exists[i].GetName() < exists[j].GetName()
	})
	retData := []*proto.NamespaceData{}
	retData = append(retData, creating...)
	retData = append(retData, updating...)
	retData = append(retData, deleting...)
	retData = append(retData, exists...)
	return retData
}

// ListNativeNamespacesContent implement for ListNativeNamespacesContent interface
func (p *NamespaceHandler) ListNativeNamespacesContent(ctx context.Context,
	req *proto.ListNativeNamespacesContentRequest, resp *spb.Struct) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.ProjectIDOrCode)
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}

	return action.ListNativeNamespacesContent(ctx, req, resp)
}

// DeleteNamespace implement for DeleteNamespace interface
func (p *NamespaceHandler) DeleteNamespace(ctx context.Context,
	req *proto.DeleteNamespaceRequest, resp *proto.DeleteNamespaceResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.DeleteNamespace(ctx, req, resp)
}

// DeleteNamespaceCallback implement for DeleteNamespaceCallback interface
func (p *NamespaceHandler) DeleteNamespaceCallback(ctx context.Context,
	req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error {
	action, err := na.NewNamespaceFactory(p.model).Action(req.GetClusterID(), req.GetProjectCode())
	if err != nil {
		logging.Error("get namespace client for cluster %s from client factory failed, err: %s",
			req.GetClusterID(), err.Error())
		return err
	}
	return action.DeleteNamespaceCallback(ctx, req, resp)
}
