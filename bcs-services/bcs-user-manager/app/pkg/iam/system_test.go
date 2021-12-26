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

import (
	"testing"
	"time"
)

var auth = &AuthConfig{
	Server:      "http://xxxx/dev",
	SystemID:    "xxxx",
	AppCode:     "xxxx",
	AppSecret:   "xxxx",
	ServerDebug: true,
}

var iamSystem = newIamModelServer(auth)

func TestIamModelServer_GetSystemInfo(t *testing.T) {
	sysInfo, err := iamSystem.GetSystemInfo(10 * time.Second)
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	t.Logf("%+v", sysInfo)
}

func TestIamModelServer_UpdateSystemConfig(t *testing.T) {
	config := &SysConfig{
		Host:    "http://demo_callback_host:8080",
		Auth:    "basic",
		Healthz: "/healthz",
	}
	err := iamSystem.UpdateSystemConfig(time.Second*10, config)
	if err != nil {
		t.Fatalf("UpdateSystemConfig failed: %v", err)
	}

	t.Log("UpdateSystemConfig successful")
}

func TestIamModelServer_RegisterSystem(t *testing.T) {
	sys := System{
		ID:                 "xxx",
		Name:               "yyy平台",
		EnglishName:        "",
		Description:        "A bcs SaaS for quick start",
		EnglishDescription: "A bcs SaaS for quick start",
		Clients:            "xxx",
		ProviderConfig: &SysConfig{
			Host:    "http://demo_callback_host:8080",
			Auth:    "basic",
			Healthz: "/healthz",
		},
	}

	err := iamSystem.RegisterSystem(time.Second*10, sys)
	if err != nil {
		t.Fatalf("RegisterSystem failed: %v", err)
	}

	t.Log("successful")
}

func TestIamModelServer_GetSystemToken(t *testing.T) {
	token, err := iamSystem.GetSystemToken(time.Second * 10)
	if err != nil {
		t.Fatalf("GetSystemInfo failed: %v", err)
	}

	t.Logf("token %s", token)
}
