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
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	bcsapiClusterManager "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcsstorage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
)

func checkProjectValidate(model store.ProjectModel, projectId, projectCode, name string) (*pm.Project, error) {
	if len(projectId) == 0 && len(projectCode) == 0 && len(strings.TrimSpace(name)) == 0 {
		return nil, fmt.Errorf("project id/code/name field all empty")
	}

	p, err := model.GetProjectByField(context.Background(), &pm.ProjectField{ProjectID: projectId,
		ProjectCode: projectCode, Name: name})
	if err != nil {
		return nil, fmt.Errorf("projectId(%s) projectCode(%s) projectName(%s) is invalid",
			projectId, projectCode, name)
	}

	return p, nil
}

func checkClusterValidate(ctx context.Context, clusterId string) (*bcsapiClusterManager.Cluster, error) {
	if len(strings.TrimSpace(clusterId)) == 0 {
		return nil, nil
	}

	cls, err := clustermanager.GetCluster(ctx, clusterId, true)
	if err != nil {
		return nil, err
	}

	return cls, nil
}

// getTaskWithSN 根据任务ID获取带有SN号的任务
func getTaskWithSN(taskID string) *types.Task {
	timeoutDuration := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	for {
		logging.Info("TaskID: %s", taskID)
		t, err := manager.GetTaskServer().GetTaskWithID(ctx, taskID)
		if err != nil {
			logging.Error("GetTaskWithID error: %v", err)
		}
		// 检查任务是否包含SN号
		sn, ok := t.GetCommonParams(utils.ItsmSnKey.String())
		if ok {
			logging.Info("sn: %s", sn)
			return t
		}

		// 检查是否超时
		select {
		case <-ctx.Done():
			logging.Error("Get sn timeout.")
			return nil
		default:
			// 未超时继续循环
		}
	}
}

// getQuotaUsage 获取项目配额使用情况
func getQuotaUsage(q *quota.ProjectQuota) {
	// 初始化配额使用量为0
	if q.Quota.Cpu != nil {
		q.Quota.Cpu.DeviceQuotaUsed = "0"
	}
	if q.Quota.Mem != nil {
		q.Quota.Mem.DeviceQuotaUsed = "0"
	}
	if q.Quota.Gpu != nil {
		q.Quota.Gpu.DeviceQuotaUsed = "0"
	}

	// 获取集群ID
	var clusterID = q.Labels["federation.bkbcs.tencent.com/host-cluster-id"]
	if clusterID == "" {
		return
	}
	quotaStorage, quotaErr := bcsstorage.GetMultiClusterResourceQuota(clusterID, q.QuotaName)

	if quotaErr != nil {
		logging.Error("bcsstorage.GetMultiClusterResourceQuota err: %v", quotaErr)
	}

	if quotaStorage != nil {
		// 获取CPU使用量
		cpu := quotaStorage.Status.TotalQuota.Used["cpu"]
		if cpu.AsApproximateFloat64() != 0 {
			q.Quota.Cpu.DeviceQuotaUsed = strconv.FormatFloat(cpu.AsApproximateFloat64(), 'f', 2, 64)
		}
		// 获取内存使用量,单位转换为GB
		memory := quotaStorage.Status.TotalQuota.Used["memory"]
		if memory.AsApproximateFloat64() != 0 {
			q.Quota.Mem.DeviceQuotaUsed = strconv.FormatFloat(memory.AsApproximateFloat64()/1024/1024/1024, 'f', 2, 64)
		}
		// 获取华为GPU使用量
		gpuHuawei := quotaStorage.Status.TotalQuota.Used["requests.huawei.com/Ascend910"]
		if gpuHuawei.AsApproximateFloat64() != 0 {
			q.Quota.Gpu.DeviceQuotaUsed = strconv.FormatFloat(gpuHuawei.AsApproximateFloat64(), 'f', 2, 64)
		}
		// 获取NVIDIA GPU使用量
		gpuNvdia := quotaStorage.Status.TotalQuota.Used["requests.nvidia.com/gpu"]
		if gpuNvdia.AsApproximateFloat64() != 0 {
			q.Quota.Gpu.DeviceQuotaUsed = strconv.FormatFloat(gpuNvdia.AsApproximateFloat64(), 'f', 2, 64)
		}
		// 获取通用GPU使用量
		gpu := quotaStorage.Status.TotalQuota.Used["gpu"]
		if gpu.AsApproximateFloat64() != 0 {
			q.Quota.Gpu.DeviceQuotaUsed = strconv.FormatFloat(gpu.AsApproximateFloat64(), 'f', 2, 64)
		}
	}
}

// GetCpuMemFromInstanceType 获取CPU和内存
func GetCpuMemFromInstanceType(instanceType string) (cpu, mem uint32) {
	if tmp := strings.Split(instanceType, "."); len(tmp) == 2 {
		regNum := regexp.MustCompile(`[0-9]+`)
		regStr := regexp.MustCompile(`[A-Z]+`)

		numMatch := regNum.FindAllString(tmp[1], 2)
		strMatch := regStr.FindAllString(tmp[1], 1)

		if len(numMatch) == 1 {
			// 格式1: 单数字表示内存大小
			m, err := strconv.ParseUint(numMatch[0], 10, 32)
			if err != nil {
				return 0, 0
			}
			mem = uint32(m)
			switch strMatch[0] {
			case "SMALL":
				cpu = 1
			case "MEDIUM":
				cpu = 2
			case "LARGE":
				cpu = 4
			default:
				return 0, 0
			}
		} else if len(numMatch) == 2 {
			// 格式2: 两个数字分别表示CPU倍数和内存大小
			m, err := strconv.ParseUint(numMatch[1], 10, 32)
			if err != nil {
				return 0, 0
			}
			c, err := strconv.ParseUint(numMatch[0], 10, 32)
			if err != nil {
				return 0, 0
			}
			mem = uint32(m)
			if strings.Contains(strMatch[0], "LARGE") {
				if c == 22 {
					cpu = 90 // 特殊规格
				} else {
					cpu = uint32(c * 4)
				}
			} else if strings.Contains(strMatch[0], "MEDIUM") {
				cpu = uint32(c * 2)
			} else if strings.Contains(strMatch[0], "SMALL") {
				cpu = uint32(c * 1)
			}
		}
	}

	return cpu, mem
}
