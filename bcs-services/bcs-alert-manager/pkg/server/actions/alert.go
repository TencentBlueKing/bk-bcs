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

package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/proto/alertmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/remote/alert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/server/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/types"

	"github.com/google/uuid"
)

// Console interface for business logic
type Console interface {
	CreateRawAlertInfo(ctx context.Context, req *alertmanager.CreateRawAlertInfoReq, resp *alertmanager.CreateRawAlertInfoResp)
	CreateBusinessAlertInfo(ctx context.Context, req *alertmanager.CreateBusinessAlertInfoReq, resp *alertmanager.CreateBusinessAlertInfoResp)
}

// AlertAction object implement Console
type AlertAction struct {
	alertClient alert.BcsAlarmInterface
}

// NewAlertAction create AlertAction object
func NewAlertAction(alertClient alert.BcsAlarmInterface) Console {
	return &AlertAction{
		alertClient: alertClient,
	}
}

// CreateRawAlertInfo create raw alert info
func (ac *AlertAction) CreateRawAlertInfo(ctx context.Context, req *alertmanager.CreateRawAlertInfoReq, resp *alertmanager.CreateRawAlertInfoResp) {
	tracer := utils.GetTraceFromContext(ctx)

	if req == nil || resp == nil {
		errMsg := fmt.Sprintf("CreateRawAlertInfo req or resp is nil")
		tracer.Error(errMsg)
		resp.ErrCode = types.BcsErrAlertManagerInvalidParameter
		resp.ErrMsg = errMsg
		return
	}

	err := req.Validate()
	if err != nil {
		tracer.Errorf("req parameter invalid: %v", err.Error())
		resp.ErrCode = types.BcsErrAlertManagerInvalidParameter
		resp.ErrMsg = err.Error()
		return
	}

	alertData := alert.AlarmReqData{
		StartsTime:   timeUnixToTime(req.Starttime),
		GeneratorURL: req.Generatorurl,
		Annotations:  req.Annotations,
		Labels:       req.Labels,
	}

	endsTime := getEndsTime(req.Starttime, req.Endtime)
	if endsTime > 0 {
		alertData.EndsTime = timeUnixToTime(endsTime)
	}

	err = ac.alertClient.SendAlarmInfoToAlertServer([]alert.AlarmReqData{alertData}, time.Second*10)
	if err != nil {
		tracer.Errorf("SendAlarmInfoToAlertServer failed: %v", err)
		resp.ErrCode = types.BcsErrAlertManagerAlertClientOperationFailed
		resp.ErrMsg = err.Error()
		return
	}

	resp.ErrCode = types.BcsErrAlertManagerSuccess
	resp.ErrMsg = types.BcsErrAlertManagerSuccessStr

	return
}

// CreateBusinessAlertInfo create business alert info
func (ac *AlertAction) CreateBusinessAlertInfo(ctx context.Context, req *alertmanager.CreateBusinessAlertInfoReq, resp *alertmanager.CreateBusinessAlertInfoResp) {
	tracer := utils.GetTraceFromContext(ctx)

	if req == nil || resp == nil {
		errMsg := fmt.Sprintf("CreateRawAlertInfo req or resp is nil")
		tracer.Error(errMsg)
		resp.ErrCode = types.BcsErrAlertManagerInvalidParameter
		resp.ErrMsg = errMsg
		return
	}

	err := req.Validate()
	if err != nil {
		tracer.Errorf("req parameter invalid: %v", err.Error())
		resp.ErrCode = types.BcsErrAlertManagerInvalidParameter
		resp.ErrMsg = err.Error()
		return
	}

	alertData := alert.AlarmReqData{
		StartsTime:   timeUnixToTime(req.Starttime),
		GeneratorURL: req.Generatorurl,
		Annotations: map[string]string{
			string(alert.AlarmAnnotationsUUID):    uuid.New().String(),
			string(alert.AlarmAnnotationsBody):    req.AlertAnnotation.Message,
			string(alert.AlarmAnnotationsComment): req.AlertAnnotation.Comment,
		},
	}
	endsTime := getEndsTime(req.Starttime, req.Endtime)
	if endsTime > 0 {
		alertData.EndsTime = timeUnixToTime(endsTime)
	}

	switch req.AlarmType {
	case alert.Resource:
		alertData.Labels = map[string]string{
			string(alert.AlarmLabelsAlertType):         req.AlarmType,
			string(alert.AlarmLabelsClusterID):         req.ClusterID,
			string(alert.AlarmLabelsClusterNameSpace):  req.ResourceAlertLabel.NameSpace,
			string(alert.AlarmLabelsAlarmResourceType): req.ResourceAlertLabel.AlarmResourceType,
			string(alert.AlarmLabelsAlarmResourceName): req.ResourceAlertLabel.AlarmResourceName,
			string(alert.AlarmLabelsAlarmLevel):        req.ResourceAlertLabel.AlarmLevel,
		}
	case alert.Module:
		alertData.Labels = map[string]string{
			string(alert.AlarmLabelsAlertType):  req.AlarmType,
			string(alert.AlarmLabelsAlarmName):  req.ModuleAlertLabel.AlarmName,
			string(alert.AlarmLabelsClusterID):  req.ClusterID,
			string(alert.AlarmLabelsModuleName): req.ModuleAlertLabel.ModuleName,
			string(alert.AlarmLabelsModuleIP):   req.ModuleAlertLabel.ModuleIP,
			string(alert.AlarmLabelsAlarmLevel): req.ModuleAlertLabel.AlarmLevel,
		}
	default:
		tracer.Errorf("invalid alarmType, please input[resource|module]")
		resp.ErrCode = types.BcsErrAlertManagerInvalidParameter
		resp.ErrMsg = fmt.Sprintf("invalid alarmType, please input[resource|module]")
		return
	}

	err = ac.alertClient.SendAlarmInfoToAlertServer([]alert.AlarmReqData{alertData}, time.Second*10)
	if err != nil {
		tracer.Errorf("SendAlarmInfoToAlertServer failed: %v", err)
		resp.ErrCode = types.BcsErrAlertManagerAlertClientOperationFailed
		resp.ErrMsg = err.Error()
		return
	}

	resp.ErrCode = types.BcsErrAlertManagerSuccess
	resp.ErrMsg = types.BcsErrAlertManagerSuccessStr

	return
}

func timeUnixToTime(timeUnix int64) time.Time {
	return time.Unix(timeUnix, 0)
}

func getEndsTime(startTime, endTime int64) int64 {
	if endTime <= 0 || endTime <= startTime {
		return -1
	}

	return endTime
}
