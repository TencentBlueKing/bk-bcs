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

package processor

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/util"
	"github.com/pborman/uuid"
)

type PolicyCenterInterf interface {
	GetAlarm(jCache map[string]map[string]*util.Job) []*utils.AlarmOptions
}

func NewPolicyCenter() *PolicyCenter {
	return &PolicyCenter{}
}

type PolicyCenter struct {
}

func (p *PolicyCenter) GetAlarm(jCache map[string]map[string]*util.Job) []*utils.AlarmOptions {
	now := time.Now().Unix()
	aTasks := make([]*AlarmTask, 0)
	for _, slave := range jCache {
		var success, fail []*util.Job
		for _, j := range slave {
			if now-j.Status.FinishedAt >= job_expire_seconds {
				continue
			}
			if j.Status.Success {
				success = append(success, j)
			} else {
				fail = append(fail, j)
			}
		}

		if len(fail) > 0 {
			aTasks = append(aTasks, &AlarmTask{
				Success: success,
				Fail:    fail,
			})
		}
	}
	return p.generateAlarm(aTasks)
}

type AlarmTask struct {
	Success []*util.Job
	Fail    []*util.Job
}

func (p *PolicyCenter) generateAlarm(aTasks []*AlarmTask) []*utils.AlarmOptions {
	var alarms []*utils.AlarmOptions
	delay := uint16(alarm_convergence_seconds)
	for _, at := range aTasks {
		var as *utils.AlarmOptions

		if len(at.Success) == 0 && len(at.Fail) != 0 {
			job := at.Fail[0]
			info := fmt.Sprintf("check %s status failed. err: %s", job.Url, job.Status.Message)
			msg := formatNormalMsg(job.Module, "Health Check", string(job.Zone), info)
			sms := formatSmsMsg(job.Module, string(job.Zone), info)
			as = &utils.AlarmOptions{
				AlarmID:            at.Fail[0].Name(),
				ConvergenceSeconds: &delay,
				AlarmName:          fmt.Sprintf("%s-%s", job.Zone, job.Module),
				ClusterID:          string(job.Zone),
				AlarmKind:          utils.ERROR_ALARM,
				AlarmMsg:           msg,
				Namespace:          "",
				VoiceReadMsg:       fmt.Sprintf("check url:%s status failed.", job.Url),
				SmsMsg:             sms,

				Module:        "bcs-health-master",
				EventMessage:  msg,
				ModuleIP:      job.Url,
				ModuleVersion: version.GetVersion(),
				UUID:          uuid.NewUUID().String(),
			}
			labels := make(map[string]string)
			labels[utils.EndpointsNameLabelKey] = utils.LBEndpoints
			labels[utils.VoiceAlarmLabelKey] = "true"
			labels[utils.VoiceMsgLabelKey] = fmt.Sprintf("check url:%s status failed.", job.Url)
			as.Labels = labels

			alarms = append(alarms, as)
		}

		blog.V(5).Infof("generate alarm with job: %s, success: %d, failed: %d.",
			at.Fail[0].Name(), len(at.Success), len(at.Fail))
	}
	return alarms
}
