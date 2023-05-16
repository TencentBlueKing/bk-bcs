/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package validator

import (
	"fmt"
	"testing"

	"bscp.io/pkg/criteria/constant"
)

func TestUnixFilePath(t *testing.T) {
	unixPath := "/root/code/go/src/bscp.io/test/benchmark/tools/gen-data"
	if err := ValidateUnixFilePath(unixPath); err != nil {
		t.Log(err)
		return
	}

	winPath := `C:\Documents\Newsletters`
	if err := ValidateUnixFilePath(winPath); err == nil {
		t.Log("unix file path validate failed")
		return
	}
}

func TestWinFilePath(t *testing.T) {
	winPath := `C:\Documents\Newsletters/test`
	if err := ValidateWinFilePath(winPath); err != nil {
		t.Log(err)
		return
	}

	unixPath := "/root/code/go/src/bscp.io/test/benchmark/tools/gen-data"
	if err := ValidateWinFilePath(unixPath); err == nil {
		t.Log("win file path validate failed")
		return
	}
}

func TestReloadFilePath(t *testing.T) {
	path := "/root/reload/reload.json"
	if err := ValidateReloadFilePath(path); err != nil {
		t.Error(err)
		return
	}

	path = fmt.Sprintf("/root/%s/reload/reload.json", constant.SideWorkspaceDir)
	if err := ValidateReloadFilePath(path); err == nil {
		t.Errorf("validate reload file path failed, case result not except")
		return
	}
}
