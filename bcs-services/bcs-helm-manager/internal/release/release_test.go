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

package release

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
)

func TestTransfer2ProtoDefaultUser(t *testing.T) {
	origin := options.GlobalOptions
	defer func() { options.GlobalOptions = origin }()

	tests := []struct {
		name   string
		global *options.HelmManagerOptions
		want   string
	}{
		{name: "nil options", global: nil, want: defaultHelmUser},
		{name: "empty default user", global: &options.HelmManagerOptions{}, want: defaultHelmUser},
		{
			name: "configured default user",
			global: &options.HelmManagerOptions{
				Release: options.ReleaseConfig{DefaultUser: "bk_admin_x1y2"},
			},
			want: "bk_admin_x1y2",
		},
	}

	rl := &Release{Name: "demo", Namespace: "default"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options.GlobalOptions = tt.global

			p := rl.Transfer2Proto("blueking", "BCS-K8S-00000")
			if p.GetCreateBy() != tt.want || p.GetUpdateBy() != tt.want {
				t.Errorf("Transfer2Proto createBy/updateBy = %s/%s, want %s", p.GetCreateBy(), p.GetUpdateBy(), tt.want)
			}

			d := rl.Transfer2DetailProto()
			if d.GetCreateBy() != tt.want || d.GetUpdateBy() != tt.want {
				t.Errorf("Transfer2DetailProto createBy/updateBy = %s/%s, want %s", d.GetCreateBy(), d.GetUpdateBy(), tt.want)
			}
		})
	}
}
