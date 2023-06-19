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

package sfs

import (
	"testing"

	pbbase "bscp.io/pkg/protocol/core/base"
)

func TestIsVersionMatch(t *testing.T) {
	with := &pbbase.Versioning{
		Major: 3,
		Minor: 4,
		Patch: 5,
	}
	leastAPIVersion = with

	// matched test case.
	test := &pbbase.Versioning{
		Major: 3,
		Minor: 4,
		Patch: 5,
	}

	yes := IsAPIVersionMatch(test)

	if !yes {
		t.Errorf("should matched, but not.")
		return
	}

	test = &pbbase.Versioning{
		Major: 3,
		Minor: 4,
		Patch: 4,
	}

	yes = IsAPIVersionMatch(test)

	if !yes {
		t.Errorf("should matched, but not.")
		return
	}

	test = &pbbase.Versioning{
		Major: 3,
		Minor: 3,
		Patch: 10,
	}

	yes = IsAPIVersionMatch(test)

	if !yes {
		t.Errorf("should matched, but not.")
		return
	}

	test = &pbbase.Versioning{
		Major: 2,
		Minor: 10,
		Patch: 10,
	}

	yes = IsAPIVersionMatch(test)

	if !yes {
		t.Errorf("should matched, but not.")
		return
	}

	// not match test case.
	test = &pbbase.Versioning{
		Major: 3,
		Minor: 5,
		Patch: 0,
	}

	yes = IsAPIVersionMatch(test)

	if yes {
		t.Errorf("should not matched, but not.")
		return
	}

	test = &pbbase.Versioning{
		Major: 3,
		Minor: 4,
		Patch: 6,
	}

	yes = IsAPIVersionMatch(test)

	if yes {
		t.Errorf("should not matched, but not.")
		return
	}

	test = &pbbase.Versioning{
		Major: 4,
		Minor: 0,
		Patch: 0,
	}

	yes = IsAPIVersionMatch(test)

	if yes {
		t.Errorf("should not matched, but not.")
		return
	}

}
