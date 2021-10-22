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

package alertmanager

import (
	"testing"
	"time"
)

func newAlertManager() AlertManageInterface {
	options := Options{
		Server:     "https://xxx:xxx/xx/v4",
		ClientAuth: true,
		Debug:      true,
		Token:      "dFYn6pFOouFePmpKlfBPoBaNbFbnoSJX",
	}

	alert, _ := NewAlertManager(options)
	return alert
}

func TestAlertManager_CreateAlertInfo(t *testing.T) {
	alertClient := newAlertManager()

	req := &CreateBusinessAlertInfoReq{
		Starttime:    time.Now().Unix(),
		Endtime:      time.Now().Add(24 * time.Hour).Unix(),
		Generatorurl: "http://123456",
		AlarmType:    "module",
		ClusterID:    "bcs-2048",
		AlertAnnotation: &AlertAnnotation{
			Message: "cpu test",
		},
		ModuleAlertLabel: &ModuleAlertLabel{
			ModuleName: "bcs-scheduler",
			ModuleIP:   "1.1.1.1",
			AlarmName:  "cpu负载过高",
			AlarmLevel: "warning",
		},
	}
	err := alertClient.CreateAlertInfoToAlertManager(req, time.Second*10)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}
