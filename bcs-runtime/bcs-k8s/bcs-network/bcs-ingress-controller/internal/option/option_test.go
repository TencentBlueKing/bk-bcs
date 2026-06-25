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

package option

import (
	"flag"
	"testing"
)

func TestCertificateCheckEnabledDefault(t *testing.T) {
	var op ControllerOption
	if op.CertificateCheckEnabled {
		t.Fatal("CertificateCheckEnabled should default to false")
	}
}

func TestCertificateCheckEnabledFlag(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	var enabled bool
	fs.BoolVar(&enabled, "certificate_check_enabled", false,
		"if true, register certificate expiry checker for tencentcloud")

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("parse empty args failed: %v", err)
	}
	if enabled {
		t.Fatal("certificate_check_enabled should default to false")
	}

	if err := fs.Parse([]string{"--certificate_check_enabled=true"}); err != nil {
		t.Fatalf("parse enabled flag failed: %v", err)
	}
	if !enabled {
		t.Fatal("certificate_check_enabled=true should set enabled to true")
	}
}
