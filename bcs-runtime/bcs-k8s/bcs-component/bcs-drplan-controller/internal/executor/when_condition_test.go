/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import "testing"

func TestShouldExecuteActionByWhen(t *testing.T) {
	tests := []struct {
		name      string
		when      string
		params    map[string]interface{}
		wantExec  bool
		wantError bool
	}{
		{
			name:     "empty when always executes",
			when:     "",
			params:   nil,
			wantExec: true,
		},
		{
			name:     "when executes in compatibility mode without mode",
			when:     `mode == "install"`,
			params:   map[string]interface{}{},
			wantExec: true,
		},
		{
			name:     "when executes when mode matches",
			when:     `mode == "upgrade"`,
			params:   map[string]interface{}{"mode": "upgrade"},
			wantExec: true,
		},
		{
			name:     "delete mode is supported",
			when:     `mode == "delete"`,
			params:   map[string]interface{}{"mode": "delete"},
			wantExec: true,
		},
		{
			name:     "rollback mode is supported",
			when:     `mode == "rollback"`,
			params:   map[string]interface{}{"mode": "rollback"},
			wantExec: true,
		},
		{
			name:     "when does not execute when mode mismatches",
			when:     `mode == "upgrade"`,
			params:   map[string]interface{}{"mode": "install"},
			wantExec: false,
		},
		{
			name:      "invalid expression returns error",
			when:      `mode in ["install"]`,
			params:    map[string]interface{}{"mode": "install"},
			wantError: true,
		},
		{
			name:     "legacy operation expression is still supported",
			when:     `operation == "upgrade"`,
			params:   map[string]interface{}{"mode": "upgrade"},
			wantExec: true,
		},
		{
			name:     "multiple OR conditions are supported",
			when:     `mode == "install" || mode == "upgrade"`,
			params:   map[string]interface{}{"mode": "install"},
			wantExec: true,
		},
		{
			name:     "multiple OR conditions skip when none match",
			when:     `mode == "install" || mode == "upgrade"`,
			params:   map[string]interface{}{"mode": "delete"},
			wantExec: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := shouldExecuteActionByWhen(tt.when, tt.params)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantExec {
				t.Fatalf("got execute=%v, want %v", got, tt.wantExec)
			}
		})
	}
}
