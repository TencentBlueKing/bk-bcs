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

package dynamicquery

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
)

func TestGetQueryJson(t *testing.T) {
	data := url.Values{
		"foo": []string{"bar1", "bar2", "bar3"},
		"hi":  []string{"ha", "ho"},
	}
	expect := map[string]interface{}{
		"foo": "bar1",
		"hi":  "ha",
	}

	var p map[string]interface{}
	err := codec.DecJson(getQueryJSON(data), &p)
	if err != nil {
		t.Errorf("getQueryJson() failed! err: %v", err)
	}

	if !reflect.DeepEqual(p, expect) {
		t.Errorf("getQueryJson() failed! \nresult:\n%v\nexpect:\n%v\n", p, expect)
	}
}
