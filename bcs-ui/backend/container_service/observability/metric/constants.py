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
import re

from django.utils.translation import ugettext_lazy as _

from backend.components import bcs_monitor as prom
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum

# 没有指定时间范围的情况下，默认获取一小时的数据
METRICS_DEFAULT_TIMEDELTA = 3600

# 默认查询的命名空间（所有）
METRICS_DEFAULT_NAMESPACE = '.*'

# 查询容器指标可不指定特定的 Pod（不推荐）
METRICS_DEFAULT_POD_NAME = '.*'

# 默认查询 POD 下所有的容器
METRICS_DEFAULT_CONTAINER_LIST = ['.*']


class MetricDimension(str, StructuredEnum):
    """指标维度"""

    CpuUsage = EnumField('cpu_usage', label=_('CPU 使用率'))
    MemoryUsage = EnumField('memory_usage', label=_('内存使用率'))
    DiskUsage = EnumField('disk_usage', label=_('磁盘使用率'))
    DiskIOUsage = EnumField('diskio_usage', label=_('磁盘 IO 使用率'))


# 节点各指标维度获取方法
NODE_DIMENSIONS_FUNC = {
    MetricDimension.CpuUsage: prom.get_node_cpu_usage,
    MetricDimension.MemoryUsage: prom.get_node_memory_usage,
    MetricDimension.DiskUsage: prom.get_node_disk_usage,
    MetricDimension.DiskIOUsage: prom.get_node_diskio_usage,
}

# 集群各指标维度获取方法
CLUSTER_DIMENSIONS_FUNC = {
    MetricDimension.CpuUsage: prom.get_cluster_cpu_usage,
    MetricDimension.MemoryUsage: prom.get_cluster_memory_usage,
    MetricDimension.DiskUsage: prom.get_cluster_disk_usage,
}

# 节点普通指标
NODE_UNAME_METRIC = [
    'dockerVersion',
    'osVersion',  # from cadvisor
    'domainname',
    'machine',
    'nodename',
    'release',
    'sysname',
    'version',  # from node-exporter
]

# 节点使用率类指标
NODE_USAGE_METRIC = ['cpu_count', 'memory', 'disk']

# 需要被过滤的注解 匹配器
FILTERED_ANNOTATION_PATTERN = re.compile(r'__meta_kubernetes_\w+_annotation')

# Job 名称 匹配器
JOB_PATTERN = re.compile(r'^(?P<namespace>[\w-]+)/(?P<name>[\w-]+)/(?P<port_idx>\d+)$')

# Service 不返回给前端的字段
INNER_USE_SERVICE_METADATA_FIELDS = [
    'annotations',
    'selfLink',
    'uid',
    'resourceVersion',
    'initializers',
    'generation',
    'deletionTimestamp',
    'deletionGracePeriodSeconds',
    'clusterName',
]

# 不展示给前端的 Label（符合前缀的）
INNER_USE_LABEL_PREFIX = [
    'io_tencent_bcs_',
    'io.tencent.paas.',
    'io.tencent.bcs.',
    'io.tencent.bkdata.',
    'io.tencent.paas.',
]

# 默认 Endpoint 路径
DEFAULT_ENDPOINT_PATH = '/metrics'

# 默认 Endpoint 时间间隔（单位：s）
DEFAULT_ENDPOINT_INTERVAL = 30

# Service Monitor 存放 Service Name 的 Label 键名
SM_SERVICE_NAME_LABEL = 'io.tencent.bcs.service_name'

# Service Monitor 无权限的命名空间 及对应的权限结构
SM_NO_PERM_NAMESPACE = ['thanos']

SM_NO_PERM_MAP = {
    'view': True,
    'use': False,
    'edit': False,
    'delete': False,
    'view_msg': '',
    'edit_msg': _('不允许操作系统命名空间'),
    'use_msg': _('不允许操作系统命名空间'),
    'delete_msg': _('不允许操作系统命名空间'),
}

# Service Monitor 名称格式
SM_NAME_PATTERN = re.compile(r'^[a-z][-a-z0-9]*$')

# 可选的 Service Monitor 时间间隔
ALLOW_SM_INTERVAL = [30, 60, 120]

# 样本行数限制最大最小值
SM_SAMPLE_LIMIT_MAX = 100000
SM_SAMPLE_LIMIT_MIN = 1
