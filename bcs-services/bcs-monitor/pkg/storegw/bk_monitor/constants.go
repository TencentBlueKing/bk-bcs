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

package bkmonitor

// AvailableNodeMetrics 蓝鲸监控节点的metrics
var AvailableNodeMetrics = []string{
	"bkmonitor:system:cpu_detail:usage",
	"bkmonitor:system:mem:total",
	"bkmonitor:system:mem:used",
	"bkmonitor:system:disk:total",
	"bkmonitor:system:disk:used",
}

// AvailableFuncNames 允许传递到数据源的函数
var AvailableFuncNames = []string{
	"avg_over_time",
}
