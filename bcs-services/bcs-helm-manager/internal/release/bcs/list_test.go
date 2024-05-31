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

package bcs

import (
	"context"
	"os"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release/bcs/sdk"
)

func BenchmarkList(b *testing.B) {
	cluster := cluster{
		clusterID: "BCS-K8S-00000",
		handler: &handler{
			sdkClientGroup: sdk.NewGroup(sdk.Config{}),
		},
	}
	option := release.ListOption{
		Namespace: "bcs-system",
	}
	options.GlobalOptions = &options.HelmManagerOptions{
		Release: options.ReleaseConfig{
			APIServer: os.Getenv("APIServer"),
			Token:     os.Getenv("BearerToken"),
		},
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cluster.listV2(context.Background(), option)
	}
	b.StopTimer()
}
