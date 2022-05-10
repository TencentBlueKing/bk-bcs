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

package i18n

import "testing"

func TestGetMessage(t *testing.T) {
	Localize()

	type args struct {
		messageID string
		values    []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty", args: args{messageID: "empty", values: nil}, want: "empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMessage(tt.args.messageID, tt.args.values...); got != tt.want {
				t.Errorf("GetMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
