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

package pkgs

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/remote/alert"
)

var (
	alertClientOnce sync.Once
	alertClient     alert.BcsAlarmInterface
)

// GetAlertClient for init alert system client
func GetAlertClient(options *config.AlertManagerOptions) alert.BcsAlarmInterface {
	alertClientOnce.Do(func() {
		alertClient = alert.NewAlertServer(&options.AlertServerOptions)
		if alertClient == nil {
			panic("init alertClient failed")
		}
	})

	return alertClient
}
