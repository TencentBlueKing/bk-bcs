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

package v1http

import (
	"testing"

	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

func TestInitExtraTokenRouters(t *testing.T) {
	origin := config.GetGlobalConfig()
	defer config.SetGlobalConfig(origin)

	tests := []struct {
		name       string
		enable     bool
		wantRoutes int
	}{
		{name: "disabled by default", enable: false, wantRoutes: 0},
		{name: "enabled", enable: true, wantRoutes: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.NewUserMgrConfig()
			conf.EnableExtraToken = tt.enable
			config.SetGlobalConfig(conf)

			ws := new(restful.WebService).Path("/usermanager")
			initExtraTokenRouters(ws)
			if got := len(ws.Routes()); got != tt.wantRoutes {
				t.Errorf("routes = %d, want %d", got, tt.wantRoutes)
			}
		})
	}
}
