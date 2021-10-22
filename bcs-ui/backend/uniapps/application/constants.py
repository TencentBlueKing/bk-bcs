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
from django.utils.translation import ugettext_lazy as _

from backend.utils.basic import ChoicesEnum

# Application status
"""
staging：资源不足等导致处于这个状态
deploying: 处于一个中间状态，正在创建
running: 正常状态
finish: 针对短任务的状态，用完即销毁
error: 不会重新调度，因为配置文件直接出错
operating: 操作状态，是中间状态
rollingupdate: 滚动更新中
unnormal: 分正常状态，主要是没有正常销毁等
"""
# Application非正常状态,因为deployment关联application，所以检测application
UNNORMAL_STATUS = [
    "Error",
    "Unnormal",
    "Failed",
    "Lost",
    "UpdateSuspend",
    "BackendError",
    "Abnormal",
    "Unready",
]  # noqa
NORMAL_STATUS = ["Running", "Finish", "UpdatePaused"]

# deployment 状态
DEPLOYMENT_NORMAL_STATUS = ["Running"]

BACKEND_APPLICATION_ERROR = BACKEND_ERROR = "Error"
BACKEND_APPLICATION_NORMAL = BACKEND_NORMAL = "Normal"
BACKEND_APPLICATION_RUNNING = BACKEND_RUNNING = "Running"

# 分割images的分隔符
SPLIT_IMAGE = ["prod", "dev", "test"]

# 操作类型
CREATE_INSTANCE = "create"
SCALE_INSTANCE = "scale"
ROLLING_UPDATE_INSTANCE = "rollingupdate"
DELETE_INSTANCE = "delete"
REBUILD_INSTANCE = "rebuild"
PAUSE_INSTANCE = "pause"
RESUME_INSTANCE = "resume"
CANCEL_INSTANCE = "cancel"

CATEGORY_MAP = {
    "deployment": "K8sDeployment",
    "job": "K8sJob",
    "daemonset": "K8sDaemonSet",
    "statefulset": "K8sStatefulSet",
}

FUNC_MAP = {
    "deployment": "%s_deployment",
    "daemonset": "%s_daemonset",
    "job": "%s_job",
    "statefulset": "%s_statefulset",
    "K8sDeployment": "%s_deployment",
    "K8sJob": "%s_job",
    "K8sDaemonSet": "%s_daemonset",
    "K8sStatefulSet": "%s_statefulset",
    "K8sSecret": "%s_secret",
    "K8sConfigMap": "%s_configmap",
    "K8sService": "%s_service",
    "K8sIngress": "%s_ingress",
}

REVERSE_CATEGORY_MAP = {
    "K8sDeployment": "deployment",
    "K8sJob": "job",
    "K8sDaemonSet": "daemonset",
    "K8sStatefulSet": "statefulset",
}

# resource and replicas key匹配
RESOURCE_REPLICAS_KEYS = {
    'deployment': {
        'desired_replicas_keys': ['data', 'spec', 'replicas'],
        'ready_replicas_keys': ['data', 'status', 'availableReplicas'],
    },
    'daemonset': {
        'desired_replicas_keys': ['data', 'status', 'desiredNumberScheduled'],
        'ready_replicas_keys': ['data', 'status', 'numberReady'],
    },
    'job': {
        'desired_replicas_keys': ['data', 'spec', 'parallelism'],
        'ready_replicas_keys': ['data', 'status', 'succeeded'],
    },
    'statefulset': {
        'desired_replicas_keys': ['data', 'spec', 'replicas'],
        'ready_replicas_keys': ['data', 'status', 'readyReplicas'],
    },
}

ALL_CATEGORY_LIST = ["application", "deployment", "K8sDeployment", "K8sDaemonSet", "K8sJob", "K8sStatefulSet"]

# k8s
K8S_KIND = 1

# 集群类型及状态
CLUSTER_TYPE = [1, 2, "1", "2"]
APP_STATUS = [1, 2, "1", "2"]

# MESOS Application Type
MESOS_APPLICATION_TYPE = ["application", "deployment"]

# SOURCE TYPE MAP
SOURCE_TYPE_MAP = {"template": _("模板集"), "helm": _("Helm模板"), "other": _("Client")}

NOT_TMPL_SOURCE_TYPE = _("非模板集")

# instance not from template
NOT_TMPL_IDENTIFICATION = "0"


# source type
class SourceType(ChoicesEnum):
    TEMPLATE = 'template'
    HELM = 'helm'
    OTHER = 'other'

    _choices_labels = (('template', _("模板集")), ('helm', _("Helm模板")), ('other', _('Client')))


# instance config labels
LABEL_CLUSTER_ID = 'io.tencent.bcs.clusterid'
LABEL_TEMPLATE_ID = 'io.tencent.paas.templateid'

# reference resource label
REFERENCE_RESOURCE_LABEL = 'data.metadata.ownerReferences.name'

# query storage field
STORAGE_FIELD_LIST = [
    'resourceName',
    'createTime',
    'data.status.podIP',
    'data.status.hostIP',
    'data.status.phase',
    'data.status.containerStatuses',
]


class ResourceStatus(ChoicesEnum):
    Running = 'Running'
    Unready = 'Unready'
    Completed = 'Completed'


RESOURCE_STATUS_FIELD_LIST = [
    'data.status',
    'resourceName',
    'namespace',
    'data.spec.parallelism',
    'data.spec.paused',
    'createTime',
    'updateTime',
    'data.metadata.labels',
    'data.spec.replicas',
    'data.spec.completions',
]

# 兼容前端传递参数内容
OWENER_REFERENCE_MAP = {
    # NOTE: 针对deployment，ownerreference是ReplicaSet
    "deployment": "ReplicaSet",
    "K8sDeployment": "ReplicaSet",
    "daemonset": "DaemonSet",
    "K8sDaemonSet": "DaemonSet",
    "statefulset": "StatefulSet",
    "K8sStatefulSet": "StatefulSet",
    "job": "Job",
    "K8sJob": "Job",
}
