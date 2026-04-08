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

package common

import (
	"os"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
)

func TestSetCustomResourceTypesFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     map[string][]string
		wantErr  bool
		setup    func()
		teardown func()
	}{
		{
			name:     "empty env value",
			envValue: "",
			want:     nil,
			wantErr:  false,
			setup:    func() {},
			teardown: func() {},
		},
		{
			name:     "single cluster single type",
			envValue: `{"cluster-id-1":["BkApp"]}`,
			want:     map[string][]string{"cluster-id-1": {"BkApp"}},
			wantErr:  false,
			setup:    func() {},
			teardown: func() {},
		},
		{
			name:     "multiple clusters multiple types",
			envValue: `{"cluster-id-1":["BkApp","CronJob"],"cluster-id-2":["BkApp"]}`,
			want: map[string][]string{
				"cluster-id-1": {"BkApp", "CronJob"},
				"cluster-id-2": {"BkApp"},
			},
			wantErr:  false,
			setup:    func() {},
			teardown: func() {},
		},
		{
			name:     "empty map",
			envValue: "{}",
			want:     map[string][]string{},
			wantErr:  false,
			setup:    func() {},
			teardown: func() {},
		},
		{
			name:     "invalid json",
			envValue: `{invalid}`,
			want:     nil,
			wantErr:  true,
			setup:    func() {},
			teardown: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.teardown()

			os.Setenv("synchronizer_customResourceTypes", tt.envValue)
			defer os.Unsetenv("synchronizer_customResourceTypes")

			opt := &option.BkcmdbSynchronizerOption{}
			err := SetCustomResourceTypesFromEnv(opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCustomResourceTypesFromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !compareMapStringSlices(opt.Synchronizer.CustomResourceTypes, tt.want) {
				t.Errorf("SetCustomResourceTypesFromEnv() = %v, want %v", opt.Synchronizer.CustomResourceTypes, tt.want)
			}
		})
	}
}

func compareMapStringSlices(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, aVals := range a {
		bVals, ok := b[k]
		if !ok {
			return false
		}
		if len(aVals) != len(bVals) {
			return false
		}
		for i := range aVals {
			if aVals[i] != bVals[i] {
				return false
			}
		}
	}
	return true
}
