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
 *
 */

package audit

import (
	"github.com/TencentBlueKing/bk-audit-go-sdk/bkaudit"
	"github.com/google/uuid"
)

var auditClient *bkaudit.EventClient

var (
	bkAppCode = GetEnvWithDefault("BkAppCode", "bk_bcs_app")
)

// AddEvent 添加审计日志，instanceData 可以选填
func AddEvent(data AuditData) {
	// resource
	action := &bkaudit.AuditAction{ActionID: data.ActionID}
	resourceType := &bkaudit.AuditResource{ResourceTypeID: string(data.ResourceType)}
	instance := &bkaudit.AuditInstance{
		InstanceID:         data.InstanceID,
		InstanceName:       data.InstanceName,
		InstanceData:       data.InstanceData,
		InstanceOriginData: data.InstanceData,
	}

	// generate event id
	ctx := &bkaudit.AuditContext{
		Username:        data.Username,
		RequestID:       data.RequestID,
		AccessSourceIp:  data.SourceIP,
		AccessUserAgent: data.UserAgent,
	}
	eventID := GenerateEventID(bkAppCode, uuid.New().String())

	auditClient.AddEvent(action, resourceType, instance, ctx, eventID, data.EventContent, data.StartTime.UnixMilli(),
		data.EndTime.UnixMilli(), int64(data.ResultCode), data.ResultContent, data.ExtendData)
}
