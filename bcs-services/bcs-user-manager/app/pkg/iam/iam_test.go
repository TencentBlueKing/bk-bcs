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

package iam

import "testing"

func TestIam_RegisterSystem(t *testing.T) {
	config := &AuthConfig{
		Server:      "http://xxxx:8080",
		SystemID:    SystemIDBKBCS,
		AppCode:     "xxxx",
		AppSecret:   "xxxx",
		ServerDebug: false,
	}
	iam, err := NewIam(config)
	if err != nil {
		t.Fatalf("NewIam failed; %v", err)
	}

	sysConfig := &SysConfig{
		Host:    "http://demo_callback_host:9000",
		Auth:    "basic",
		Healthz: "/healthz",
	}
	err = iam.RegisterSystem(sysConfig)
	if err != nil {
		t.Fatalf("iam RegisterSystem failed: %v", err)
	}

	t.Logf("RegisterSystem successful")

	sysInfo, err := iam.client.GetSystemInfo(defaultTimeout)
	if err != nil {
		t.Fatalf("iam GetSystemInfo failed: %v", err)
	}

	t.Logf("%+v", sysInfo)
}
