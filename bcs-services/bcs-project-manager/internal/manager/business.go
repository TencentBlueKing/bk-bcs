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

package manager

import (
	"context"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkuser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
)

const (
	defaultBusinessSyncIntervalSec = 600
)

type businessSyncStore interface {
	UpsertBusiness(ctx context.Context, biz *business.Business) error
	DeleteBusinessesNotIn(ctx context.Context, tenantID string, keepIDs []string) (int64, error)
}

// BusinessManager 定时从 CMDB 全量同步业务到本地表
type BusinessManager struct {
	ctx   context.Context
	model store.ProjectModel
}

// NewBusinessManager new business manager
func NewBusinessManager(ctx context.Context, model store.ProjectModel) *BusinessManager {
	return &BusinessManager{
		ctx:   ctx,
		model: model,
	}
}

// Run run business manager
func (m *BusinessManager) Run() {
	logging.Info("start sync business from cmdb")
	m.SyncBusiness()
	interval := time.NewTicker(businessSyncInterval())
	defer interval.Stop()

	for {
		select {
		case <-m.ctx.Done():
			logging.Info("close BusinessManager done")
			return
		case <-interval.C:
			m.SyncBusiness()
		}
	}
}

// SyncBusiness 按租户全量同步一轮；单个租户 CMDB 失败时只打日志，不删除该租户本地数据
func (m *BusinessManager) SyncBusiness() {
	if config.GlobalConf == nil || !config.GlobalConf.BusinessSync.Enable {
		logging.Warn("skip business sync, businessSync.enable is false")
		return
	}
	if config.GlobalConf.CMDB.Host == "" {
		logging.Warn("skip business sync, cmdb host is empty")
		return
	}
	tenantIDs, err := listSyncTenantIDs(m.ctx)
	if err != nil {
		logging.Error("list tenants for business sync failed: %s", err.Error())
		return
	}
	for _, tenantID := range tenantIDs {
		m.syncTenantBusiness(tenantID)
	}
}

func (m *BusinessManager) syncTenantBusiness(tenantID string) {
	ctx := tenant.WithTenantIdFromContext(m.ctx, tenantID)
	businesses, err := cmdb.ListAllBusinesses(ctx)
	if err != nil {
		logging.Error("list all businesses from cmdb failed, tenant=%s: %s", tenantID, err.Error())
		return
	}
	upserted, deleted, err := applyBusinessSnapshot(ctx, m.model, tenantID, businesses)
	if err != nil {
		logging.Error("apply business snapshot failed, tenant=%s: %s", tenantID, err.Error())
		return
	}
	logging.Info("sync business from cmdb success, tenant=%s, upsert=%d, deleted=%d", tenantID, upserted, deleted)
}

func listSyncTenantIDs(ctx context.Context) ([]string, error) {
	if !tenant.IsMultiTenantEnabled() {
		return []string{constant.DefaultTenantId}, nil
	}
	tenants, err := bkuser.ListTenants(ctx, constant.SystemTenantId)
	if err != nil {
		return nil, err
	}
	tenantIDs := make([]string, 0, len(tenants))
	for _, t := range tenants {
		if t.Id == "" {
			continue
		}
		tenantIDs = append(tenantIDs, t.Id)
	}
	return tenantIDs, nil
}

func businessSyncInterval() time.Duration {
	sec := defaultBusinessSyncIntervalSec
	if config.GlobalConf != nil && config.GlobalConf.BusinessSync.Interval > 0 {
		sec = config.GlobalConf.BusinessSync.Interval
	}
	return time.Duration(sec) * time.Second
}

func applyBusinessSnapshot(
	ctx context.Context, model businessSyncStore, tenantID string, businesses []cmdb.BusinessData,
) (int, int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	keepIDs := make([]string, 0, len(businesses))
	for _, src := range businesses {
		biz := transferBusiness(src, tenantID, now)
		if biz == nil {
			continue
		}
		if err := model.UpsertBusiness(ctx, biz); err != nil {
			return 0, 0, err
		}
		keepIDs = append(keepIDs, biz.BusinessID)
	}
	deleted, err := model.DeleteBusinessesNotIn(ctx, tenantID, keepIDs)
	return len(keepIDs), deleted, err
}

func transferBusiness(src cmdb.BusinessData, tenantID, syncTime string) *business.Business {
	if src.BKBizID <= 0 {
		logging.Warn("skip invalid cmdb business, tenant=%s, bk_biz_id=%d", tenantID, src.BKBizID)
		return nil
	}
	return &business.Business{
		TenantID:   tenantID,
		BusinessID: strconv.FormatInt(src.BKBizID, 10),
		Name:       src.BKBizName,
		Default:    src.Default,
		Maintainer: src.BKBizMaintainer,
		Bs2NameID:  src.BS2NameID,
		Productor:  src.BkBizProductor,
		Tester:     src.BkBizTester,
		Developer:  src.BkBizDeveloper,
		UpdateTime: syncTime,
		SyncTime:   syncTime,
	}
}
