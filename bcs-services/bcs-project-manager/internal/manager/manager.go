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

// Package manager xxx
package manager

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/itsm/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
)

// NamespaceManager manager for namespace
type NamespaceManager struct {
	ctx   context.Context
	model store.ProjectModel
}

// NewNamespaceManager new namespace manager
func NewNamespaceManager(ctx context.Context, model store.ProjectModel) *NamespaceManager {
	mgr := &NamespaceManager{
		ctx:   ctx,
		model: model,
	}
	return mgr
}

// Run run namespace manager
func (n *NamespaceManager) Run() {
	logging.Info("start sync namespace records with itsm")
	interval := time.NewTicker(30 * time.Second)
	defer interval.Stop()

	for {
		select {
		case <-n.ctx.Done():
			logging.Info("close NamespaceManager done")
			return
		case <-interval.C:
			n.SyncNamespaceItsmStatus()
		}
	}
}

// SyncNamespaceItsmStatus task to sync namespace status with itsm tickets
func (n *NamespaceManager) SyncNamespaceItsmStatus() {
	namespaces, err := n.model.ListNamespaces(n.ctx)
	if err != nil {
		logging.Error("namespace manager list namespaces failed: %s", err.Error())
	}
	snList := []string{}
	nsMap := map[string]nsm.Namespace{}
	for _, namespace := range namespaces {
		snList = append(snList, namespace.ItsmTicketSN)
		nsMap[namespace.ItsmTicketSN] = namespace
	}
	if len(snList) == 0 {
		return
	}

	// 多租户暂时不处理itsm，待租户方案确认后续再处理
	if config.GlobalConf.EnableMultiTenant {
		logging.Info("skip sync namespace itsm status for multi tenant mode")
		return
	}

	// TODO: 获取TenantID
	tickets, err := v2.ListTickets(n.ctx, snList)
	if err != nil {
		logging.Error("list namespace itsm tickets %v failed, err: %s", snList, err.Error())
		return
	}
	for _, ticket := range tickets {
		if ticket.CurrentStatus == "FINISHED" ||
			ticket.CurrentStatus == "TERMINATED" ||
			ticket.CurrentStatus == "REVOKED" {
			namespace, ok := nsMap[ticket.SN]
			if !ok {
				logging.Error("namespace ticket %s doesn't exits in db", ticket.SN)
				return
			}
			err := n.model.DeleteNamespace(n.ctx, namespace.ProjectCode, namespace.ClusterID, namespace.Name)
			if err != nil {
				logging.Error("delete namespace %s/%s/%s failed, err: %s",
					namespace.ProjectCode, namespace.ClusterID, namespace.Name, err.Error())
				return
			}
			logging.Info("sync delete namespace %s/%s/%s for %s ticket %s success",
				namespace.ProjectCode, namespace.ClusterID, namespace.Name, ticket.CurrentStatus, namespace.ItsmTicketSN)
		}
	}
}
