/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package version

import "testing"

func TestCurrentVersion(t *testing.T) {
	VERSION = "server-v1.2.3"
	ver, err := parseVersion()
	if err != nil {
		t.Errorf("parse version failed, err: %v", err)
		return
	}

	if ver[0] != 1 {
		t.Errorf("invalid major version: %d", ver[0])
		return
	}

	if ver[1] != 2 {
		t.Errorf("invalid minor version: %d", ver[0])
		return
	}

	if ver[2] != 3 {
		t.Errorf("invalid patch version: %d", ver[0])
		return
	}

}

func TestIncorrectVersion(t *testing.T) {
	VERSION = "server-1.2.3"
	if _, err := parseVersion(); err == nil {
		t.Errorf("expect parse version failed, but not")
		return
	}

	VERSION = "server-v1.2.3.4"
	if _, err := parseVersion(); err == nil {
		t.Errorf("expect parse version failed, but not")
		return
	}

	VERSION = "server-v1.2"
	if _, err := parseVersion(); err == nil {
		t.Errorf("expect parse version failed, but not")
		return
	}

	VERSION = "server-v1"
	if _, err := parseVersion(); err == nil {
		t.Errorf("expect parse version failed, but not")
		return
	}

}
