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

// Package audit audit
package audit

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	bkbase "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_base"
	bklog "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_log"
	bkmonitor "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
)

// EnableAuditReq params
type EnableAuditReq struct {
	ProjectId string `json:"projectId" in:"path=projectId" validate:"required"`
	ClusterId string `json:"clusterId" in:"path=clusterId" validate:"required"`
}

// GenCollectorConfigName generate collector config name
func GenCollectorConfigName(clusterID string) string {
	clusterID = strings.ToLower(clusterID)
	clusterID = strings.ReplaceAll(clusterID, "-", "_")
	return fmt.Sprintf("bkbcs_audit_%s", clusterID)
}

// GenBKLogConfigName generate bk log config name
func GenBKLogConfigName(clusterID string) string {
	return fmt.Sprintf("bkbcs-audit-%s", strings.ToLower(clusterID))
}

// EnableAudit enable audit
// 1. ensure bklog data id is created
// 2. ensure BKLogConfig is created
// 3. ensure bkbase data id is created
// 4. ensure bkbase databus is created
// nolint
func EnableAudit(c context.Context, req *EnableAuditReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}

	project, err := bcs.GetProject(c, config.G.BCS, rctx.ProjectCode)
	if err != nil {
		return nil, err
	}
	bizID, err := strconv.Atoi(project.CcBizID)
	if err != nil {
		return nil, err
	}

	audit, err := storage.GlobalStorage.FirstAuditOrCreate(c, &entity.Audit{
		ProjectCode: rctx.ProjectCode,
		ClusterID:   rctx.ClusterId,
	})
	if err != nil {
		return nil, err
	}

	// ensure bklog data id is created
	if audit.CollectorConfigID == 0 {
		databusCustomResp, derr := bklog.DatabusCustomCreate(c, &bklog.DatabusCustomCreateReq{
			BkBizID:               bizID,
			CollectorConfigName:   GenCollectorConfigName(audit.ClusterID),
			CollectorConfigNameEN: GenCollectorConfigName(audit.ClusterID),
			Description:           "create by bcs",
			CustomType:            "log",
			CategoryID:            "host_process",
			Retention:             7,
			EsShards:              1,
			StorageReplies:        0,
			AllocationMinDays:     0,
			DataLinkID:            config.G.BKBase.AuditDataLinkID,
		})
		if derr != nil {
			return nil, derr
		}
		audit.DataID = databusCustomResp.BkDataID
		audit.CollectorConfigID = databusCustomResp.CollectorConfigID
		audit.CollectorConfigName = GenCollectorConfigName(audit.ClusterID)
		err = storage.GlobalStorage.UpdateAudit(c, audit.ID.Hex(), entity.M{
			"collectorConfigID":   audit.CollectorConfigID,
			"collectorConfigName": audit.CollectorConfigName,
			"dataID":              audit.DataID,
		})
		if err != nil {
			return nil, err
		}
	}

	// get topic from bkmonitor
	topic, err := bkmonitor.MetadataQueryDataSource(c, config.G.BKMonitor.MetadataURL, audit.DataID)
	if err != nil {
		return nil, err
	}

	// ensure BKLogConfig is created
	bkLogConfigName := GenBKLogConfigName(rctx.ClusterId)
	_, err = k8sclient.GetBkLogConfig(c, rctx.ClusterId, "default", bkLogConfigName)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		_, err = k8sclient.CreateBkLogConfig(c, rctx.ClusterId, "default", &k8sclient.BkLogConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: bkLogConfigName,
			},
			Spec: &k8sclient.BkLogConfigSpec{
				DataID:        audit.DataID,
				LogConfigType: "container_log_config",
				Namespace:     "kube-system",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"k8s-app": "kube-apiserver",
					},
				},
				Path: []string{"/etc/kubernetes/*.audit"},
				ExtMeta: map[string]string{
					"bk_bcs_cluster_id": rctx.ClusterId,
				},
			},
		})
		if err != nil {
			return nil, err
		}
		err = storage.GlobalStorage.UpdateAudit(c, audit.ID.Hex(), entity.M{
			"bkLogConfigName": bkLogConfigName,
		})
		if err != nil {
			return nil, err
		}
	}

	// ensure bkbase data id
	err = bkbase.ApplyDataID(c, bkbase.GenDataIDName(audit.ClusterID), bizID, audit.DataID,
		topic.MQConfig.StorageConfig.Topic)
	if err != nil {
		return nil, err
	}

	// ensure bkbase databus
	err = bkbase.ApplyDatabus(c, bkbase.GenDatabusName(audit.ClusterID), bkbase.GenDataIDName(audit.ClusterID))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// DisableAuditReq params
type DisableAuditReq struct {
	ProjectId string `json:"projectId" in:"path=projectId" validate:"required"`
	ClusterId string `json:"clusterId" in:"path=clusterId" validate:"required"`
}

// DisableAudit disable audit
func DisableAudit(c context.Context, req *DisableAuditReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}

	bkLogConfigName := GenBKLogConfigName(rctx.ClusterId)
	_, err = k8sclient.GetBkLogConfig(c, rctx.ClusterId, "default", bkLogConfigName)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	err = k8sclient.DeleteBkLogConfig(c, rctx.ClusterId, "default", bkLogConfigName)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GetAuditStatusReq params
type GetAuditStatusReq struct {
	ProjectId string `json:"projectId" in:"path=projectId" validate:"required"`
	ClusterId string `json:"clusterId" in:"path=clusterId" validate:"required"`
}

// GetAuditStatusResp audit status resp
type GetAuditStatusResp struct {
	Enabled bool `json:"enabled"`
}

// GetAuditStatus get audit status
func GetAuditStatus(c context.Context, req *GetAuditStatusReq) (*GetAuditStatusResp, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}

	bkLogConfigName := GenBKLogConfigName(rctx.ClusterId)
	_, err = k8sclient.GetBkLogConfig(c, rctx.ClusterId, "default", bkLogConfigName)
	if err == nil {
		return &GetAuditStatusResp{Enabled: true}, nil
	}

	return &GetAuditStatusResp{Enabled: false}, nil
}

// EnableAuditESReq params
type EnableAuditESReq struct {
	ProjectId        string `json:"projectId" in:"path=projectId" validate:"required"`
	ClusterId        string `json:"clusterId" in:"path=clusterId" validate:"required"`
	StorageClusterID int    `json:"storage_cluster_id"`
}

// EnableAuditES enable audit es
func EnableAuditES(c context.Context, req *EnableAuditESReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}

	audit, err := storage.GlobalStorage.GetAudit(c, rctx.ProjectCode, rctx.ClusterId)
	if err != nil {
		return nil, err
	}

	err = bklog.DatabusCustomUpdate(c, audit.CollectorConfigID, &bklog.DatabusCustomUpdateReq{
		CollectorConfigName: GenCollectorConfigName(rctx.ClusterId),
		Description:         "create by bcs",
		CustomType:          "log",
		CategoryID:          "host_process",
		StorageClusterID:    req.StorageClusterID,
		Retention:           7,
		EsShards:            1,
		StorageReplies:      0,
		AllocationMinDays:   0,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}
