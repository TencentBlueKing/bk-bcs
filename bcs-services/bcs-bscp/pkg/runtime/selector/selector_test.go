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

package selector

import (
	"testing"
)

func TestMarshal(t *testing.T) {
	ft := Selector{
		MatchAll: false,
		LabelsOr: []Element{
			{Key: "biz", Op: new(EqualType), Value: "2001"},
			{Key: "set", Op: new(InType), Value: []string{"1", "2", "3"}},
			{Key: "module", Op: new(GreaterThanType), Value: 1},
			{Key: "game", Op: new(NotEqualType), Value: "stress"},
		},
	}

	pb, err := ft.MarshalPB()
	if err != nil {
		t.Errorf("selector marshal pb failed, err: %v", err)
		return
	}

	json, err := pb.MarshalJSON()
	if err != nil {
		t.Errorf("pb selector marshal json failed, err: %v", err)
		return
	}

	result := `{"labels_or":[{"key":"biz", "op":"eq", "value":"2001"}, {"key":"set", "op":"in", "value":` +
		`["1", "2", "3"]}, {"key":"module", "op":"gt", "value":1}, {"key":"game", "op":"ne", "value":"stress"}]}`
	if string(json) != result {
		t.Errorf("selector marshal json not right")
		return
	}
}
