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

// Package bcsstorage xxx
package bcsstorage

import (
	"os"
	"testing"

	svcConfig "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"

	"k8s.io/apimachinery/pkg/api/resource"
)

// TestGetMultiClusterResourceQuota test get multi cluster resource quota
func TestGetMultiClusterResourceQuota(t *testing.T) {
	type args struct {
		clusterID string
		name      string
	}
	tests := []struct {
		name    string
		args    args
		want    *MultiClusterResourceQuota
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				clusterID: "BCS-K8S-xxxx",
				name:      "11111",
			},
		},
	}
	for _, tt := range tests {
		svcConfig.GlobalConf = &svcConfig.ProjectConfig{}
		svcConfig.GlobalConf.BcsGateway.Host = os.Getenv("TEST_BCSGATEWAY_HOST")
		svcConfig.GlobalConf.BcsGateway.Token = os.Getenv("TEST_BCSGATEWAY_TOKEN")
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMultiClusterResourceQuota(tt.args.clusterID, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMultiClusterResourceQuota() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("got info = %v", got)
			t.Logf("gpu %v", got.Status.TotalQuota.Used.Name("gpu", resource.DecimalSI))
			t.Logf("cpu %v", got.Status.TotalQuota.Used.Name("cpu", resource.DecimalSI))
			t.Logf("mem %v", got.Status.TotalQuota.Used.Name("memory", resource.DecimalSI))
			memory := got.Status.TotalQuota.Used["memory"]
			t.Logf("memory %v", memory.AsApproximateFloat64()/1024/1024/1024)
			cpu := got.Status.TotalQuota.Used["cpu"]
			t.Logf("cpu %v, %t", cpu.AsApproximateFloat64(), cpu.AsApproximateFloat64() > 0)
			gpu := got.Status.TotalQuota.Used["requests.huawei.com/Ascend910"]
			t.Logf("requests.huawei.com/Ascend910 %v", gpu.AsApproximateFloat64())
			gpu = got.Status.TotalQuota.Used["requests.nvidia.com/gpu"]
			t.Logf("requests.nvidia.com/gpu %v, %t", gpu.AsApproximateFloat64(), gpu.AsApproximateFloat64() != 0)
		})
	}
}
