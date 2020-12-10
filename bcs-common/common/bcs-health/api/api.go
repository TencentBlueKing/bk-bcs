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

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"text/template"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/utils"
)

// MessageKind type for message
type MessageKind string

const (
	// ErrorKind message type for error
	ErrorKind MessageKind = "Error"
	// WarnKind message type for warn
	WarnKind MessageKind = "Warn"
	// InfoKind message type for info
	InfoKind MessageKind = "Info"
)

// TLSConfig config for tls
type TLSConfig struct {
	CaFile   string `json:"ca-file"`
	CertFile string `json:"cert-file"`
	KeyFile  string `json:"key-file"`
}

// HealthInfo report health information
type HealthInfo struct {
	Module    string      `json:"module"`
	Kind      MessageKind `json:"kind"`
	AlarmName string      `json:"alarmName"`
	// which this event affiliation is, should be one of user, platform and both.
	// both means that this event shoud be cared by both user and platform.
	Affiliation types.AffiliationType `json:"affiliation"`
	// user defined application alarm level, which is value of
	// label with the key "io.tencent.bcs.monitor.level", should be one of important,
	// unimportant and general.
	AppAlarmLevel      string          `json:"alarm_level"`
	AlarmID            string          `json:"-"`
	AlarmType          utils.AlarmType `json:"-"`
	ConvergenceSeconds *uint16         `json:"-"`
	VoiceMessage       string          `json:"-"`
	IP                 string          `json:"ip"`
	ClusterID          string          `json:"clusterid"`
	Namespace          string          `json:"namespace,omitempty"`
	Message            string          `json:"message"`
	Version            string          `json:"version"`
	ReportTime         string          `json:"reporttime"`
	ResourceType       string          `json:"resource_type"`
	ResourceName       string          `json:"resource_name"`
}

var statusController *Status

var healthinfoTemplate *template.Template

// NewBcsHealth zkSvrs eg: 127.0.0.1:2181,127.0.0.2:3181
func NewBcsHealth(zkSvr string, tls TLSConfig) error {
	status, err := newStatusController(zkSvr, tls)
	if nil != err {
		return err
	}
	t, err := template.New("healthinfo").Parse(healthTemplate)
	if nil != err {
		return fmt.Errorf("new healthinfo template failed. err: %v", err)
	}
	healthinfoTemplate = t
	statusController = status
	statusController.run()
	return nil
}

// SendHealthInfo sending health info to server
func SendHealthInfo(health *HealthInfo) error {
	if nil == statusController || nil == healthinfoTemplate {
		return errors.New("no status controller or healthinfo template can be used. please call function NewBcsHealth first. ")
	}

	w := Writer{}
	if err := healthinfoTemplate.Execute(&w, health); nil != err {
		return err
	}
	var typer utils.AlarmType
	switch health.Kind {
	case InfoKind:
		typer = utils.INFO_ALARM
	case WarnKind:
		typer = utils.WARN_ALARM
	case ErrorKind:
		typer = utils.ERROR_ALARM
	default:
		return fmt.Errorf("unknown health kind: %s", health.Kind)
	}

	if 0 != int32(health.AlarmType) {
		typer = health.AlarmType
	}
	voiceMsg := health.VoiceMessage
	if health.AlarmType.IsVoice() && len(health.VoiceMessage) == 0 {
		voiceMsg = fmt.Sprintf("BCS 异常告警, 模块: %s, IP: %s", health.Module, health.IP)
	}

	var level string
	if len(health.AppAlarmLevel) == 0 {
		level = "important"
	} else {
		level = health.AppAlarmLevel
	}

	alarm := utils.AlarmOptions{
		Namespace:          health.Namespace,
		Module:             health.Module,
		AlarmKind:          typer,
		AlarmName:          health.AlarmName,
		ClusterID:          health.ClusterID,
		AlarmID:            health.AlarmID,
		ConvergenceSeconds: health.ConvergenceSeconds,
		VoiceReadMsg:       voiceMsg,
		AlarmMsg:           w.String(),
		ResourceType:       health.ResourceType,
		ResourceName:       health.ResourceName,

		EventMessage:  health.Message,
		ModuleVersion: health.Version,
		ModuleIP:      health.IP,
		AtTime:        time.Now().Unix(),
		AppAlarmLevel: level,
		Affiliation:   health.Affiliation,
	}

	if typer.IsSMS() {
		alarm.SmsMsg = fmt.Sprintf("BCS Alarm\n%s\n%s\n%s\n%s\nMsg:%s", health.Module, health.Kind, health.IP, health.ClusterID, health.Message)
	}
	alarmjs, err := json.Marshal(alarm)
	if nil != err {
		return fmt.Errorf("json marshal alarm info failed. err: %v", err)
	}

	return statusController.TryDoRequest(string(alarmjs))
}

// SendKubeEventAlarm interface of sending kubernetes event alarm
func SendKubeEventAlarm(namespace string, msg string) error {
	alarm := utils.AlarmOptions{
		Namespace: namespace,
		AlarmMsg:  msg,
	}
	alarmjs, err := json.Marshal(alarm)
	if nil != err {
		return fmt.Errorf("json marshal alarm info failed. err: %v", err)
	}
	return statusController.TryDoAlarmRequest(string(alarmjs))
}
