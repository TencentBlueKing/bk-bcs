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

// Package audit xxx
package audit

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/TencentBlueKing/bk-audit-go-sdk/bkaudit"
	"github.com/google/uuid"
	blog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

var auditClient *bkaudit.EventClient

var (
	bkAppCode   = utils.GetEnvWithDefault("BkAppCode", iam.SystemIDBKBCS)
	bkAppSecret = utils.GetEnvWithDefault("BkAppSecret", "")
)

func init() {
	// init formatter
	var formatter = &bkaudit.EventFormatter{}
	var exporters = []bkaudit.Exporter{&bkaudit.LoggerExporter{Logger: blog.V(0)}}
	// init client
	var err error
	auditClient, err = bkaudit.InitEventClient(bkAppCode, bkAppSecret, formatter, exporters, 0, nil)
	if err != nil {
		blog.Errorf("init auditClient client failed, %s", err.Error())
		return
	}
}

// AddEvent 添加审计日志，instanceData 可以选填
func AddEvent(actionID, resourceTypeID, instanceID, username string, allow bool,
	instanceData map[string]interface{}) {
	// resource
	action := &bkaudit.AuditAction{ActionID: actionID}
	resourceType := &bkaudit.AuditResource{ResourceTypeID: resourceTypeID}
	instance := &bkaudit.AuditInstance{
		InstanceID:         instanceID,
		InstanceData:       instanceData,
		InstanceOriginData: instanceData,
	}

	// generate event id
	ctx := &bkaudit.AuditContext{Username: username}
	eventID := utils.GenerateEventID(bkAppCode, uuid.New().String())
	startTime := time.Now().UnixMilli()

	// message
	resultCode := int64(1)
	if allow {
		resultCode = 0
	}
	extendData := make(map[string]interface{}, 0)
	eventContent := ""
	resultContenct := ""
	auditClient.AddEvent(action, resourceType, instance, ctx, eventID, eventContent, startTime, 0,
		resultCode, resultContenct, extendData)
}
