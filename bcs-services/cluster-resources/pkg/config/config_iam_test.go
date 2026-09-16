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

package config

import "testing"

func TestInitIAMV4RequiresHost(t *testing.T) {
	c := &ClusterResourcesConf{}
	c.Global.IAM.SystemID = "bk_bcs_app"
	c.Global.Basic.AppCode = "bk_bcs_app"
	c.Global.Basic.AppSecret = "secret"
	c.Global.IAM.EnableV4 = true
	c.Global.IAM.V4GateWayHost = ""
	if err := c.initIAM(); err == nil {
		t.Fatal("enable_v4 without v4_gateway_host should fail")
	}
}

func TestInitIAMV4WithHost(t *testing.T) {
	c := &ClusterResourcesConf{}
	c.Global.IAM.SystemID = "bk_bcs_app"
	c.Global.Basic.AppCode = "bk_bcs_app"
	c.Global.Basic.AppSecret = "secret"
	c.Global.IAM.EnableV4 = true
	c.Global.IAM.V4GateWayHost = "https://bkiam-v4.example"
	if err := c.initIAM(); err != nil {
		t.Fatalf("initIAM v4: %v", err)
	}
	if c.Global.IAM.Cli == nil {
		t.Fatal("Cli factory not set")
	}
	if cli := c.Global.IAM.Cli("tenant-a"); cli == nil {
		t.Fatal("expected perm client")
	}
}

func TestInitIAMMissingSecret(t *testing.T) {
	c := &ClusterResourcesConf{}
	c.Global.IAM.SystemID = "bk_bcs_app"
	if err := c.initIAM(); err == nil {
		t.Fatal("missing app secret should fail")
	}
}
