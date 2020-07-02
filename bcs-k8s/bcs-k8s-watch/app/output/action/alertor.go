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

package action

import (
	"fmt"
	"time"

	"k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/bcs-health/api"
	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
)

type Alertor struct {
	ClusterID string
	Module    string
	ModuleIP  string
}

func NewAlertor(clusterID, moduleIP string, zkHosts string, tls options.TLS) (*Alertor, error) {
	var alertor = &Alertor{
		ClusterID: clusterID,
		Module:    "k8s-watch",
		ModuleIP:  moduleIP,
	}

	var err error

	var tlsCfg api.TLSConfig
	tlsCfg.CaFile = tls.CAFile
	tlsCfg.CertFile = tls.CertFile
	tlsCfg.KeyFile = tls.KeyFile

	err = nil
	if err = api.NewBcsHealth(zkHosts, tlsCfg); err != nil {
		glog.Errorf("NewBcsHealth failed:%s", err.Error())
		err = fmt.Errorf("NewBcsHealth failed:%s", err.Error())
	}
	return alertor, err
}

func (alertor *Alertor) DoAlarm(syncData *SyncData) {
	healthInfo := alertor.genHealthInfo(syncData)
	// do alarm
	if healthInfo != nil {
		go alertor.sendAlarm(healthInfo)
	}
}

func (alertor *Alertor) genHealthInfo(syncData *SyncData) *api.HealthInfo {
	data := syncData.Data
	event, ok := data.(*v1.Event)
	if !ok {
		glog.Errorf("Event Convert object to v1.Event fail! object is %v", data)
		return nil
	}

	// 2018-07-11: change IP from event source IP to module IP
	//IP:        event.Source.Host,
	message := fmt.Sprintf("[%s %s]%s:%s", event.InvolvedObject.Kind,
		event.InvolvedObject.Name, event.Reason, event.Message)
	seconds := uint16(60)
	var healthInfo = &api.HealthInfo{
		AlarmName: "podEventWarnning",
		Kind:      api.WarnKind,
		Message:   message,
		//AlarmID:            string(event.InvolvedObject.UID),
		AlarmID:            syncData.OwnerUID,
		ConvergenceSeconds: &seconds,
		ResourceType:       event.InvolvedObject.Kind,
		ResourceName:       event.InvolvedObject.Name,
	}
	return healthInfo

}

func (alertor *Alertor) sendAlarm(healthInfo *api.HealthInfo) bool {
	healthInfo.Module = alertor.Module
	healthInfo.IP = alertor.ModuleIP
	healthInfo.ClusterID = alertor.ClusterID
	healthInfo.Version = version.GetVersion()
	healthInfo.ReportTime = time.Now().Format("2017-01-01 12:00:00")

	// NOTE: 目前bcs-health & bcs-alarm 只能根据namespace去配置告警接收人
	healthInfo.Namespace = alertor.ClusterID

	glog.Errorf("Add Event Pod Warnning: %v", healthInfo)

	//return true

	if err := api.SendHealthInfo(healthInfo); err != nil {
		glog.Errorf("SendHealthInfo failed:%s", err.Error())
		return false
	}
	return true
}
