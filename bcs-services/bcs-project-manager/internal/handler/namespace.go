/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"context"
	"sort"

	na "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
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
	resp.Data = retData
	return nil
}

func sortNamespaces(list []*proto.NamespaceData) []*proto.NamespaceData {
	creating := []*proto.NamespaceData{}
	for _, ns := range list {
		if ns.GetItsmTicketType() == nsm.ItsmTicketTypeCreate {
			creating = append(creating, ns)
		}
	}
	updating := []*proto.NamespaceData{}
	for _, ns := range list {
		if ns.GetItsmTicketType() == nsm.ItsmTicketTypeUpdate {
			updating = append(updating, ns)
		}
	}
	deleting := []*proto.NamespaceData{}
	for _, ns := range list {
		if ns.GetItsmTicketType() == nsm.ItsmTicketTypeDelete {
			deleting = append(deleting, ns)
		}
	}
	exists := []*proto.NamespaceData{}
	for _, ns := range list {
		if ns.GetItsmTicketType() == "" {
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
