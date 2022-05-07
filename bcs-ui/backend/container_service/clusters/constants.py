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
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum
from backend.utils.basic import ChoicesEnum

# default node count
DEFAULT_NODE_LIMIT = 10000

# no specific resource flag
NO_RES = '**'

# 主机key的映射，便于前端进行展示
CCHostKeyMappings = {
    'bak_operator': 'bk_bak_operator',
    'classify_level_name': 'classify_level_name',
    'device_class': 'svr_device_class',
    'device_type_id': 'bk_svr_type_id',
    'device_type_name': 'svr_type_name',
    'hard_memo': 'hard_memo',
    'host_id': 'bk_host_id',
    'host_name': 'bk_host_name',
    'idc': 'idc_name',
    'idc_area': 'bk_idc_area',
    'idc_area_id': 'bk_idc_area_id',
    'idc_id': 'idc_id',
    'idcunit': 'idc_unit_name',
    'idcunit_id': 'idc_unit_id',
    'inner_ip': 'bk_host_innerip',
    'memo': 'bk_comment',
    'module_name': 'module_name',
    'operator': 'operator',
    'osname': 'bk_os_name',
    'osversion': 'bk_os_version',
    'outer_ip': 'bk_host_outerip',
    'server_rack': 'rack',
    'project_name': 'project_name',
    'cluster_name': 'cluster_name',
    'cluster_id': 'cluster_id',
    'is_used': 'is_used',
    'bk_cloud_id': 'bk_cloud_id',
    "is_valid": "is_valid",
}

# skip namespace
K8S_SKIP_NS_LIST = ['kube-system', 'thanos', 'web-console']

# K8S 系统预留标签的key
# Kubernetes 预留关键字 kubernetes.io, 用于系统的标签和注解
# https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
K8S_RESERVED_KEY_WORDS = ["kubernetes.io"]


class ClusterManagerNodeStatus(str, StructuredEnum):
    """cluster manager 中节点的状态"""

    RUNNING = EnumField("RUNNING", label="正常状态")
    INITIALIZATION = EnumField("INITIALIZATION", label="初始化中")
    DELETING = EnumField("DELETING", label="删除中")
    ADDFAILURE = EnumField("ADD-FAILURE", label="添加节点失败")
    REMOVEFAILURE = EnumField("REMOVE-FAILURE", label="下架节点失败")
    REMOVABLE = EnumField("REMOVABLE", label="可移除状态")
    NOTREADY = EnumField("NOTREADY", label="非正常状态")
    UNKNOWN = EnumField("UNKNOWN", label="未知状态")


# docker状态排序
# default will be 100
DockerStatusDefaultOrder = 100
DockerStatusOrdering = {"running": 0, "waiting": 1, "lost": 8, "terminated": 9}


class ClusterType(str, StructuredEnum):
    """集群类型"""

    SINGLE = EnumField('SINGLE', label="独立集群")
    SHARED = EnumField('SHARED', label="共享集群")
    FEDERATION = EnumField('FEDERATION', label="联邦集群")
    FEDERATION_SHARED = EnumField('FEDERATION_SHARED', label="共享联邦集群")


# TODO: 待前端整理接口后，清理掉下面内容
IP_LIST_RESERVED_LENGTH = 200


class ClusterNetworkType(ChoicesEnum):
    """集群网络类型"""

    OVERLAY = "overlay"
    UNDERLAY = "underlay"

    _choices_labels = ((OVERLAY, "overlay"), (UNDERLAY, "underlay"))


# BK Agent 默认状态，默认为不在线
DEFAULT_BK_AGENT_ALIVE = 0
