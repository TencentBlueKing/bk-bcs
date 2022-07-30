# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
from prometheus_client import Counter

from backend.packages.blue_krill.data_types.enum import StructuredEnum

ResultLabelName = 'result'


class Result(str, StructuredEnum):
    """执行结果"""

    Success = 'success'
    Failure = 'failure'


# 命名空间
namespace_create_total = Counter('namespace_create_total', 'Count of create namespace', labelnames=[ResultLabelName])

# 模板集
templateset_create_total = Counter(
    'templateset_create_total', 'Count of create templateset', labelnames=[ResultLabelName]
)
templateset_instantiate_total = Counter(
    'templateset_instantiate_total', 'Count of instantiate templateset', labelnames=[ResultLabelName]
)

# Helm 管理
helm_install_total = Counter('helm_install_total', 'Count of install helm chart', labelnames=[ResultLabelName])
helm_upgrade_total = Counter('helm_upgrade_total', 'Count of upgrade helm release', labelnames=[ResultLabelName])
helm_rollback_total = Counter('helm_rollback_total', 'Count of rollback helm release', labelnames=[ResultLabelName])

# 应用管理
workload_upgrade_total = Counter('workload_upgrade_total', 'Count of upgrade workload', labelnames=[ResultLabelName])
workload_scale_total = Counter('workload_scale_total', 'Count of scale up/down workload', labelnames=[ResultLabelName])
workload_recreate_total = Counter(
    'workload_recreate_total', 'Count of recreate workload', labelnames=[ResultLabelName]
)
workload_delete_total = Counter('workload_delete_total', 'Count of delete workload', labelnames=[ResultLabelName])
# 由于"重新调度"按钮对应的服务下沉了, 指标需要由 cluster-resources 模块提供
workload_reschedule_total = Counter(
    'workload_reschedule_total', 'Count of reschedule workload', labelnames=[ResultLabelName]
)

# 组件库管理
cluster_tools_install_total = Counter(
    'cluster_tools_install_total', 'Count of install cluster tools', labelnames=[ResultLabelName]
)
cluster_tools_upgrade_total = Counter(
    'cluster_tools_upgrade_total', 'Count of upgrade cluster tools', labelnames=[ResultLabelName]
)


resource_counters = {
    'namespace add': namespace_create_total,
    'template add': templateset_create_total,
    'template instantiate': templateset_instantiate_total,
    'instance modify': workload_upgrade_total,
    'instance scale': workload_scale_total,
    'instance delete': workload_delete_total,
    'cluster_tools install': cluster_tools_install_total,
    'cluster_tools upgrade': cluster_tools_upgrade_total,
}


def counter_inc(resource_type: str, activity_type: str, result: str):
    """根据资源和操作, 匹配对应的 Counter 指标, 统计 +1

    :param: resource_type: 对应审计表中的 resource_type 字段, 表明操作的资源类型
    :param: activity_type: 对应审计表中的 activity_type 字段, 表明具体的操作类型
    :param: result: 操作成功或者失败, 成功: Success, 失败: Failure

    """
    counter = resource_counters.get(f'{resource_type} {activity_type}')
    if not counter:
        return

    if result in Result.get_values():
        counter.labels(result).inc()
