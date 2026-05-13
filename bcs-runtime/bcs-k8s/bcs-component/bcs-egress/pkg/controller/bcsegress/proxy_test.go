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

package bcsegress

import (
	"os"
	"path/filepath"
	"testing"
)

// NOCC:tosa/fn_length(设计如此)
func TestConfigValidationDoesNotInvokeShell(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "shell-injection")
	nginx := filepath.Join(dir, "nginx")
	if err := os.WriteFile(nginx, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write fake nginx: %v", err)
	}

	ngx := &Nginx{
		option: &EgressOption{ProxyExecutable: nginx},
	}
	filename := filepath.Join(dir, "nginx.conf; touch "+marker)
	if err := ngx.configValidation(filename); err != nil {
		t.Fatalf("configValidation() error = %v", err)
	}

	if _, err := os.Stat(marker); err == nil {
		t.Fatalf("configValidation() executed shell metacharacters in filename")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat marker: %v", err)
	}
}
