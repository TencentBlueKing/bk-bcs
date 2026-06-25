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

package quota

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcsstorage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// newFakeQuotaStorage 构造一个带有指定 used 资源的伪 MultiClusterResourceQuota。
// used 的 key 为资源名（如 "requests.nvidia.com/gpu"），value 为可被 resource.MustParse 解析的字符串（如 "3"）。
func newFakeQuotaStorage(used map[string]string) *bcsstorage.MultiClusterResourceQuota {
	rl := corev1.ResourceList{}
	for k, v := range used {
		rl[corev1.ResourceName(k)] = resource.MustParse(v)
	}
	return &bcsstorage.MultiClusterResourceQuota{
		Status: bcsstorage.MultiClusterResourceQuotaStatus{
			TotalQuota: bcsstorage.MultiClusterResourceQuotaTotalQuotaStatus{
				Used: rl,
			},
		},
	}
}

// newProjectQuotaWithGpu 构造一个含 Gpu 字段的 proto.ProjectQuota，便于 setGPUUsageForProto 直接写入
func newProjectQuotaWithGpu() *proto.ProjectQuota {
	return &proto.ProjectQuota{
		Quota: &proto.QuotaResource{
			Gpu: &proto.DeviceInfo{},
		},
	}
}

// resetGlobalConf 在每个用例开始时重置全局配置，避免用例间相互影响
func resetGlobalConf(extra []string) {
	config.GlobalConf = &config.ProjectConfig{
		SystemConfig: config.SystemCommonConfig{
			GPUResourceKeys: extra,
		},
	}
}

// TestGetExtraGPUKeys_NilConf 当 GlobalConf 为 nil 时应返回 nil，不应 panic
func TestGetExtraGPUKeys_NilConf(t *testing.T) {
	config.GlobalConf = nil
	if got := getExtraGPUResourceKeys(); got != nil {
		t.Fatalf("want nil when GlobalConf is nil, got %v", got)
	}
}

// TestGetExtraGPUKeys_Empty 未配置 GPUResourceKeys 时应返回空（nil 或长度为 0）
func TestGetExtraGPUKeys_Empty(t *testing.T) {
	resetGlobalConf(nil)
	if got := getExtraGPUResourceKeys(); len(got) != 0 {
		t.Fatalf("want empty, got %v", got)
	}
}

// TestGetExtraGPUKeys_WithValues 配置后应原样返回
func TestGetExtraGPUKeys_WithValues(t *testing.T) {
	resetGlobalConf([]string{"requests.amd.com/gpu", "requests.intel.com/gpu"})
	got := getExtraGPUResourceKeys()
	if len(got) != 2 || got[0] != "requests.amd.com/gpu" || got[1] != "requests.intel.com/gpu" {
		t.Fatalf("unexpected result: %v", got)
	}
}

// TestSetGPUUsage_OnlyBuiltin 不配置 GPUResourceKeys 时，仅硬编码段生效（NVIDIA = 3）
func TestSetGPUUsage_OnlyBuiltin(t *testing.T) {
	resetGlobalConf(nil)
	storage := newFakeQuotaStorage(map[string]string{
		"requests.nvidia.com/gpu": "3",
	})
	q := newProjectQuotaWithGpu()

	setGPUUsageForProto(q, storage)

	if got := q.Quota.Gpu.DeviceQuotaUsed; got != "3.00" {
		t.Fatalf("want 3.00, got %s", got)
	}
}

// TestSetGPUUsage_BuiltinHuawei 验证华为 Ascend910 这类硬编码 key 仍然生效
func TestSetGPUUsage_BuiltinHuawei(t *testing.T) {
	resetGlobalConf(nil)
	storage := newFakeQuotaStorage(map[string]string{
		"requests.huawei.com/Ascend910": "2",
	})
	q := newProjectQuotaWithGpu()

	setGPUUsageForProto(q, storage)

	if got := q.Quota.Gpu.DeviceQuotaUsed; got != "2.00" {
		t.Fatalf("want 2.00, got %s", got)
	}
}

// TestSetGPUUsage_ExtraKey 配置中的额外 key 在 quotaStorage 中存在时，应被读取并覆盖到 DeviceQuotaUsed
func TestSetGPUUsage_ExtraKey(t *testing.T) {
	resetGlobalConf([]string{"requests.amd.com/gpu"})
	storage := newFakeQuotaStorage(map[string]string{
		"requests.amd.com/gpu": "5",
	})
	q := newProjectQuotaWithGpu()

	setGPUUsageForProto(q, storage)

	if got := q.Quota.Gpu.DeviceQuotaUsed; got != "5.00" {
		t.Fatalf("want 5.00, got %s", got)
	}
}

// TestSetGPUUsage_ExtraKeyMissing 配置了额外 key 但 quotaStorage 中无对应资源：
// 不应崩溃，且不应清掉硬编码已写入的值
func TestSetGPUUsage_ExtraKeyMissing(t *testing.T) {
	resetGlobalConf([]string{"requests.intel.com/gpu"}) // storage 里没这个 key
	storage := newFakeQuotaStorage(map[string]string{
		"requests.huawei.com/Ascend910": "2",
	})
	q := newProjectQuotaWithGpu()

	setGPUUsageForProto(q, storage)

	if got := q.Quota.Gpu.DeviceQuotaUsed; got != "2.00" {
		t.Fatalf("want 2.00 (huawei kept), got %s", got)
	}
}

// TestSetGPUUsage_OverrideOrder 处理顺序：硬编码段先执行，配置中的额外 key 后执行；
// 当两者都命中时，最终值应为配置 key 的值（后写覆盖前写）
func TestSetGPUUsage_OverrideOrder(t *testing.T) {
	resetGlobalConf([]string{"requests.amd.com/gpu"})
	storage := newFakeQuotaStorage(map[string]string{
		"requests.nvidia.com/gpu": "3", // 硬编码命中
		"requests.amd.com/gpu":    "5", // 额外命中
	})
	q := newProjectQuotaWithGpu()

	setGPUUsageForProto(q, storage)

	// 硬编码先把 NVIDIA=3 写入；额外段把 amd=5 覆盖；最终 5
	if got := q.Quota.Gpu.DeviceQuotaUsed; got != "5.00" {
		t.Fatalf("want 5.00 (overridden by extra), got %s", got)
	}
}
