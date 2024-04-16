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

package alert

import (
	"os"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/config"
)

var defaultOptions = &config.AlertServerOptions{
	Server:      "http://xxx/prod",
	AppCode:     "xxx",
	AppSecret:   "xxx",
	ServerDebug: true,
}

func TestNewAlertServer(t *testing.T) {
	client := NewAlertServer(defaultOptions, WithTestDebug(true))

	data := []AlarmReqData{
		{
			StartsTime:   time.Now(),
			EndsTime:     time.Now().Add(60 * time.Hour),
			GeneratorURL: "http://xxx",
			Annotations: map[string]string{
				"uuid":    "cee84faf-7ee3-11ea-xxx",
				"message": "0.gseagent.gse.30012.1586932748085923931()status changed:Staging->Failed",
			},
			Labels: map[string]string{
				"alertname":       "测试cee84faf",
				"project_id":      "5805f1b824134fa39318fb0cf59f694b",
				"cluster_id":      "BCS-K8S-40185",
				"namespace":       "gse",
				"ip":              "xx.xx.xx.xx",
				"module_name":     "scheduler",
				"app_alarm_level": "important",
				"reason":          "scheduler",
			},
		},
	}

	err := client.SendAlarmInfoToAlertServer(data, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("call SendAlarmInfoToAlertServer successful")
}
