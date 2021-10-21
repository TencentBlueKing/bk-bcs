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

package lib

import (
	"reflect"
	"testing"
)

func TestExtraField(t *testing.T) {
	origin := "{\"foo\":\"bar\",\"hello\":\"world\"}"
	code := "eyJmb28iOiJiYXIiLCJoZWxsbyI6IndvcmxkIn0="
	extra := NewExtra(code)
	if extra.raw != code {
		t.Errorf("NewExtra() failed!")
	}

	if s, err := extra.GetStr(); err != nil || s != origin {
		t.Errorf("ExtraField GetStr failed! \nresult:\n%v\nexpect:\n%v\nerr:\n%v\n", s, origin, err)
	}

	var r map[string]interface{}
	expect := map[string]string{"foo": "bar", "hello": "world"}
	if err := extra.Unmarshal(&r); err != nil || reflect.DeepEqual(r, expect) {
		t.Errorf("ExtraField Unmarshal failed! \nresult:\n%v\nexpect:\n%v\nerr:\n%v\n", r, expect, err)
	}
}
