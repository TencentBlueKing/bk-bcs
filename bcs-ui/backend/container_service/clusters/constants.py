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
from collections import OrderedDict

from django.utils.translation import ugettext_lazy as _

from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum
from backend.utils.basic import ChoicesEnum

# default node count
DEFAULT_NODE_LIMIT = 10000

# filter removed node
FILTER_NODE_STATUS = ['removed']

# cluster status
COMMON_FAILED_STATUS = [
    "initial_failed",
    "failed",
    "check_failed",
    "remove_failed",
    "so_init_failed",
    "upgrade_failed",
]  # noqa
COMMON_RUNNING_STATUS = ["initializing", "running", "initial_checking", "removing", "so_initializing", "upgrading"]
CLUSTER_FAILED_STATUS = COMMON_FAILED_STATUS
CLUSTER_RUNNING_STATUS = COMMON_RUNNING_STATUS
NODE_FAILED_STATUS = ['ScheduleFailed', 'bke_failed']
NODE_FAILED_STATUS.extend(COMMON_FAILED_STATUS)
NODE_RUNNING_STATUS = ['Scheduling', None, 'bke_installing']
NODE_RUNNING_STATUS.extend(COMMON_RUNNING_STATUS)
DEFAULT_PAGE_LIMIT = 5
DEFAULT_MIX_VALUE = "*****-----$$$$$"

# no specific resource flag
NO_RES = '**'

# project all cluster flag
PROJECT_ALL_CLUSTER = 'all'

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

# 节点默认标签
DEFAULT_SYSTEM_LABEL_KEYS = [
    "beta.kubernetes.io/arch",
    "beta.kubernetes.io/os",
    "kubernetes.io/hostname",
    "node-role.kubernetes.io/node",
]


# TODO: 第一版只创建两个module: master和node
CC_MODULE_INFO = {
    "mesos": {
        "stag": "test",
        "prod": "pro",
        "debug": "debug",
        "module_suffix_name": [
            "master",
            "node",
            # "zk"
        ],
    },
    "k8s": {
        "stag": "test",
        "prod": "pro",
        "debug": "debug",
        "module_suffix_name": [
            "master",
            "node",
            # "etcd",
            # "bcs",
        ],
    },
}


class OpType(ChoicesEnum):
    ADD_NODE = 'add_node'
    DELETE_NODE = 'delete_node'


# skip namespace
K8S_SKIP_NS_LIST = ['kube-system', 'thanos', 'web-console']

# mesos类型跳过的命名空间列表
MESOS_SKIP_NS_LIST = ["bcs-system"]

# 调用接口异常的消息，记录到db中，可以直接转换
BCS_OPS_ERROR_INFO = {"state": "FAILURE", "node_tasks": [{"state": "FAILURE", "name": _("- 调用初始化接口失败")}]}


# 状态映射
class ClusterStatusName(ChoicesEnum):
    normal = _("正常")
    initial_checking = _("前置检查中")
    check_failed = _("前置检查失败")
    so_initializing = _("SO初始化中")
    so_init_failed = _("SO初始化失败")
    initializing = _("初始化中")
    initial_failed = _("初始化失败")
    removing = _("删除中")
    remove_failed = _("删除失败")
    removed = _("已删除")


class ClusterState(ChoicesEnum):
    BCSNew = "bcs_new"
    Existing = "existing"

    _choices_labels = ((BCSNew, "bcs_new"), (Existing, "existing"))


class ClusterNetworkType(ChoicesEnum):
    """集群网络类型"""

    OVERLAY = "overlay"
    UNDERLAY = "underlay"

    _choices_labels = ((OVERLAY, "overlay"), (UNDERLAY, "underlay"))


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


# Kube-proxy代理模式
class KubeProxy(str, StructuredEnum):
    IPTABLES = EnumField("iptables")
    IPVS = EnumField("ipvs")


# k8s cluster master role
# 参考rancher中定义nodeRoleMaster="node-role.kubernetes.io/master"
K8S_NODE_ROLE_MASTER = "node-role.kubernetes.io/master"


# docker状态排序
# default will be 100
DockerStatusDefaultOrder = 100
DockerStatusOrdering = {"running": 0, "waiting": 1, "lost": 8, "terminated": 9}


# 集群升级版本
CLUSTER_UPGRADE_VERSION = OrderedDict(
    {re.compile(r'^\S+[vV]?1\.8\.\S+$'): ["v1.12.6"], re.compile(r'^\S+[vV]?1\.12\.\S+$'): ["v1.14.3-tk8s-v1.1-1"]}
)

# TODO: 先放到前端传递，后续gcloud版本统一后，支持分支判断再去掉
UPGRADE_TYPE = {"v1.12.6": "update8to12", "v1.14.3-tk8s-v1.1-1": "update12to14"}


class ClusterType(str, StructuredEnum):
    """ 集群类型 """

    SINGLE = EnumField('SINGLE', label="独立集群")
    SHARED = EnumField('SHARED', label="共享集群")
    FEDERATION = EnumField('FEDERATION', label="联邦集群")
    FEDERATION_SHARED = EnumField('FEDERATION_SHARED', label="共享联邦集群")


# TODO: 待前端整理接口后，清理掉下面内容
IP_LIST_RESERVED_LENGTH = 200


# BK Agent 默认状态，默认为不在线
DEFAULT_BK_AGENT_ALIVE = 0
