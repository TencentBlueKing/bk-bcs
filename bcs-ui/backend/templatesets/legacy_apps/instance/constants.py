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

实例化需要的后台添加的信息
使用 Mako 模板做变量替换
"""
import copy

from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.utils.basic import ChoicesEnum

K8S_MODULE_NAME = 'k8s'

# 所有配置文件的公共 label
# TODO: 后续可以去掉下面几个label相关
LABLE_TEMPLATE_ID = "io.tencent.paas.templateid"
LABLE_VERSION = "io.tencent.paas.version"
LABLE_INSTANCE_ID = "io.tencent.paas.instanceid"
LABLE_PROJECT_ID = "io.tencent.paas.projectid"

ANNOTATIONS_CREATOR = 'io.tencent.paas.creator'
ANNOTATIONS_UPDATOR = 'io.tencent.paas.updator'
ANNOTATIONS_CREATE_TIME = 'io.tencent.paas.createTime'
ANNOTATIONS_UPDATE_TIME = 'io.tencent.paas.updateTime'
ANNOTATIONS_WEB_CACHE = 'io.tencent.paas.webCache'
ANNOTATIONS_VERSION_ID = 'io.tencent.paas.versionid'
ANNOTATIONS_VERSION = "io.tencent.paas.version"
ANNOTATIONS_INSTANCE_ID = "io.tencent.paas.instanceid"
ANNOTATIONS_PROJECT_ID = "io.tencent.paas.projectid"
ANNOTATIONS_TEMPLATE_ID = "io.tencent.paas.templateid"

# Application中用于关联service的label
LABEL_SERVICE_NAME = "io.tencent.paas.application"
# 有container的资源都需要打一个唯一的label,以便其他资源 selector
LABLE_CONTAINER_SELECTOR_LABEL = "io.tencent.paas"
# metric 资源添加
LABLE_METRIC_SELECTOR_LABEL = "io.tencent.paas.metric"

# 监控需要的重要级别label
LABEL_MONITOR_LEVEL = "io.tencent.bcs.monitor.level"
LABEL_MONITOR_LEVEL_DEFAULT = "general"

# 数据平台相关 label，k8s 下需要打在 pod 级别
BKDATA_LABEL = {
    # 按 项目-》业务ID 通过API获取
    "io.tencent.bkdata.container.stdlog.dataid": "{{SYS_STANDARD_DATA_ID}}",
    "io.tencent.bkdata.baseall.dataid": "6566",
}

BCS_LABELS = {
    "io.tencent.bcs.clusterid": "{{SYS_CLUSTER_ID}}",
    "io.tencent.bcs.cluster": "{{SYS_CLUSTER_ID}}",
    "io.tencent.bcs.namespace": "{{SYS_NAMESPACE}}",
    "io.tencent.bcs.app.appid": "{{SYS_CC_APP_ID}}",
    "io.tencent.bcs.kind": "{{SYS_PROJECT_KIND}}",
}
BCS_LABELS.update(BKDATA_LABEL)

K8S_LOG_ENV = [
    {"name": "io_tencent_bcs_cluster", "value": "{{SYS_CLUSTER_ID}}"},
    {"name": "io_tencent_bcs_namespace", "value": "{{SYS_NAMESPACE}}"},
    {"name": "io_tencent_bcs_app_appid", "value": "{{SYS_CC_APP_ID}}"},
    {"name": "io_tencent_bkdata_baseall_dataid", "value": "6566"},
    {"name": "io_tencent_bkdata_container_stdlog_dataid", "value": "{{SYS_STANDARD_DATA_ID}}"},
]

# 资源来源标签
SOURCE_TYPE_LABEL_KEY = 'io.tencent.paas.source_type'

# 自定义日志标签
K8S_CUSTOM_LOG_ENV_KEY = 'io_tencent_bcs_custom_labels'

PAAS_LABLES = {
    LABLE_TEMPLATE_ID: "{{SYS_TEMPLATE_ID}}",
    LABLE_VERSION: "{{SYS_VERSION}}",
    LABLE_INSTANCE_ID: "{{SYS_INSTANCE_ID}}",
    LABLE_PROJECT_ID: "{{SYS_PROJECT_ID}}",
    SOURCE_TYPE_LABEL_KEY: "template",
}
# 公共的labels
PUBLIC_LABELS = BCS_LABELS.copy()
PUBLIC_LABELS.update(PAAS_LABLES)

# 公共的备注信息 annotations
PUBLIC_ANNOTATIONS = {
    ANNOTATIONS_CREATOR: "{{SYS_CREATOR}}",
    ANNOTATIONS_UPDATOR: "{{SYS_UPDATOR}}",
    ANNOTATIONS_CREATE_TIME: "{{SYS_CREATE_TIME}}",
    ANNOTATIONS_UPDATE_TIME: "{{SYS_UPDATE_TIME}}",
    ANNOTATIONS_VERSION_ID: "{{SYS_VERSION_ID}}",
    ANNOTATIONS_VERSION: "{{SYS_VERSION}}",
    ANNOTATIONS_INSTANCE_ID: "{{SYS_INSTANCE_ID}}",
    ANNOTATIONS_PROJECT_ID: "{{SYS_PROJECT_ID}}",
    ANNOTATIONS_TEMPLATE_ID: "{{SYS_TEMPLATE_ID}}",
    SOURCE_TYPE_LABEL_KEY: "template",
}

# TODO: 先添加上，不影响先前使用
BCS_ANNOTATIONS = BCS_LABELS.copy()
PUBLIC_ANNOTATIONS.update(BCS_ANNOTATIONS)


# ############################## k8s 相关资源

# 模板集相关的实例化资源
K8S_SECRET_SYS_CONFIG = {
    "kind": "Secret",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
}

K8S_CONFIGMAP_SYS_CONFIG = {
    "kind": "ConfigMap",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
}

K8S_INGRESS_SYS_CONFIG = {
    "kind": "Ingress",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
}

K8S_SEVICE_SYS_CONFIG = {
    "kind": "Service",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
}

POD_SYS_CONFIG = {"template": {"metadata": {"labels": PUBLIC_LABELS}}}

K8S_DEPLPYMENT_SYS_CONFIG = {
    "kind": "Deployment",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
    "spec": POD_SYS_CONFIG,
}

K8S_DAEMONSET_SYS_CONFIG = {
    "kind": "DaemonSet",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
    "spec": POD_SYS_CONFIG,
}

K8S_JOB_SYS_CONFIG = {
    "kind": "Job",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
    "spec": POD_SYS_CONFIG,
}

K8S_STATEFULSET_SYS_CONFIG = {
    "kind": "StatefulSet",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
    "spec": POD_SYS_CONFIG,
}

K8S_HPA_SYS_CONFIG = {
    "kind": "HorizontalPodAutoscaler",
    "metadata": {
        "namespace": "{{SYS_NAMESPACE}}",
        "labels": PUBLIC_LABELS,
        "annotations": PUBLIC_ANNOTATIONS,
    },
}

# k8s 资源限制单位
K8S_RESOURCE_UNIT = {'cpu': 'm', 'memory': 'Mi'}
# k8s 环境变量key
K8S_ENV_KEY = {
    'configmapKey': 'configMapKeyRef',
    'secretKey': 'secretKeyRef',
    'configmapFile': 'configMapRef',
    'secretFile': 'secretRef',
}

# k8s 镜像仓库Secret名称前缀
# https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
K8S_IMAGE_SECRET_PRFIX = 'paas.image.registry.'

# 负载均衡的配置文件
"""
service 中添加:
"labels": {
    "BCSGROUP": "hn8-loadbalance"
}
"""
LB_LABLES = copy.deepcopy(BCS_LABELS)
LB_LABLES["loadbalance"] = "{{SYS_BCSGROUP}}"
LB_LABLES[LABLE_PROJECT_ID] = "{{SYS_PROJECT_ID}}"

METRIC_SYS_CONFIG = [
    {
        "version": "",
        "name": "",
        "namespace": "",
        "networkMode": "",
        "networkType": "",
        "clusterID": "",
        "dataID": 0,
        "port": 0,
        "uri": "/xxx/xxx/xxx",
        "method": "POST|GET",
        "frequency": 0,
        "head": {"key": "val"},
        "selector": {"key": "val"},
        "parameters": {"key": "val"},
    }
]

# 非标准日志采集生成的configmap的name后缀
LOG_CONFIG_MAP_SUFFIX = 'non-standard-configmap'
LOG_CONFIG_MAP_KEY_SUFFIX = '.conf'
LOG_CONFIG_MAP_PATH_PRFIX = '/etc/'
LOG_CONFIG_MAP_APP_LABEL = 'io.tencent.bkdata.container.log.cfgfile'
APPLICATION_ID_SEPARATOR = '_APPLICATIONID_'
INGRESS_ID_SEPARATOR = '_INGRESS-'


class EventType(ChoicesEnum):
    """InstanceEvent类型定义"""

    REQ_FAILED = 1
    REQ_ERROR = 2
    ROLLBACK_FAILED = 3
    ROLLBACK_ERROR = 4

    _choices_labels = (
        (REQ_FAILED, _("请求失败")),
        (REQ_ERROR, _("请求异常")),  # 这个现在应该无法记录
        (ROLLBACK_FAILED, _("回滚失败")),
        (ROLLBACK_ERROR, _("回滚异常")),  # 这个现在应该无法记录
    )


class InsState(ChoicesEnum):
    """实例化状态"""

    # 未实例化
    NO_INS = 0

    # 已实例化，但是调用API返回失败
    INS_FAILED = 1

    # 已实例化，调用API成功
    INS_SUCCESS = 2

    UPDATE_FAILED = 3
    UPDATE_SUCCESS = 4

    # 绑定的appliation, deployment已经删除, metric使用
    INS_DELETED = 10

    # Metric 手动删除
    METRIC_DELETED = 11
    METRIC_UPDATED = 12

    _choices_labels = (
        (NO_INS, _("未实例化")),
        (INS_FAILED, _("实例化失败")),
        (INS_SUCCESS, _("实例化成功")),
        (UPDATE_FAILED, _("更新实例失败")),
        (UPDATE_SUCCESS, _("更新实例成功")),
        (INS_DELETED, _("实例已经删除")),
        (METRIC_DELETED, _("Metric配置已经删除")),
        (METRIC_UPDATED, _("Metric配置更新")),
    )


# 各个版本对应的apiVersion
API_VERSION = {
    '1.8.3': {
        'DaemonSet': 'apps/v1beta1',
        'Deployment': 'apps/v1beta1',
        'Job': 'batch/v1',
        'StatefulSet': 'apps/v1beta1',
        'Ingress': 'extensions/v1beta1',
        'Service': 'v1',
        'ConfigMap': 'v1',
        'Secret': 'v1',
        'CronJob': 'batch/v1',
        'Pod': 'v1',
        'Endpoints': 'v1',
        'ReplicaSets': 'apps/v1beta1',
        'ReplicationController': 'v1',
        'PersistentVolumeClaim': 'v1',
        'StorageClass': 'storage.k8s.io/v1beta1',
        'Volume': 'v1',
        'VolumeAttachment': 'storage.k8s.io/v1beta1',
        'Namespace': 'v1',
        'Node': 'v1',
        'PersistentVolume': 'v1',
        'Role': 'rbac.authorization.k8s.io/v1alpha1',
        'HorizontalPodAutoscaler': '',
    },
    '1.12.3': {
        'DaemonSet': 'apps/v1',
        'Deployment': 'apps/v1',
        'Job': 'batch/v1',
        'StatefulSet': 'apps/v1',
        'Ingress': 'extensions/v1beta1',
        'Service': 'v1',
        'ConfigMap': 'v1',
        'Secret': 'v1',
        'CronJob': 'batch/v1',
        'Pod': 'v1',
        'Endpoints': 'v1',
        'ReplicaSets': 'apps/v1',
        'ReplicationController': 'v1',
        'PersistentVolumeClaim': 'v1',
        'StorageClass': 'storage.k8s.io/v1',
        'Volume': 'v1',
        'VolumeAttachment': 'storage.k8s.io/v1',
        'Namespace': 'v1',
        'Node': 'v1',
        'PersistentVolume': 'v1',
        'Role': 'rbac.authorization.k8s.io/v1',
        'HorizontalPodAutoscaler': 'autoscaling/v2beta2',
    },
}
