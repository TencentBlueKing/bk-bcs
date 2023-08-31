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
from enum import Enum

from django.utils.translation import ugettext_lazy as _

from backend.utils.basic import ChoicesEnum

# 变量的格式
VARIABLE_PATTERN = "[A-Za-z][A-Za-z0-9-_]"
# 填写数字的地方，可以填写变量
NUM_VAR_PATTERN = "^{{%s*}}$" % VARIABLE_PATTERN
# 需要与 backend.templatesets.var_mgmt.serializers.py 中的说明保持一致
NUM_VAR_ERROR_MSG = _("只能包含字母、数字、中划线和下划线，且以字母开头")
# 文件目录正则
FILE_DIR_PATTERN = "^((?!\.{\$)[\w\d\-\.\/~{}\$]+)+$"

KEY_PATTERN = re.compile(r'{{([^}]*)}}')

REAL_NUM_VAR_PATTERN = re.compile(r"^%s*$" % VARIABLE_PATTERN)

# configmap/secret key 名称限制
KEY_NAME_PATTERN = "^[a-zA-Z{]{1}[a-zA-Z0-9-_.{}]{0,254}$"


class TemplateCategory(ChoicesEnum):
    SYSTEM = 'sys'
    CUSTOM = 'custom'

    _choices_labels = ((SYSTEM, _("系统")), (CUSTOM, _("自定义")))


class TemplateEditMode(ChoicesEnum):
    PageForm = 'page_form'
    YAML = 'yaml'

    _choices_labels = ((PageForm, "PageForm"), (YAML, "YAML"))


class FileAction(ChoicesEnum):
    CREATE = 'create'
    UPDATE = 'update'
    DELETE = 'delete'
    UNCHANGE = 'unchange'

    _choices_labels = ((CREATE, 'create'), (UPDATE, 'update'), (DELETE, 'delete'), (UNCHANGE, 'unchange'))


# TODO mark refactor 考虑用起来
class K8sResourceName(ChoicesEnum):
    K8sDeployment = 'K8sDeployment'
    K8sDaemonSet = 'K8sDaemonSet'
    K8sJob = 'K8sJob'
    K8sStatefulSet = 'K8sStatefulSet'
    K8sService = 'K8sService'
    K8sConfigMap = 'K8sConfigMap'
    K8sSecret = 'K8sSecret'
    K8sIngress = 'K8sIngress'
    K8sHPA = 'K8sHPA'

    _choices_labels = (
        (K8sDeployment, 'K8sDeployment'),
        (K8sDaemonSet, 'K8sDaemonSet'),
        (K8sJob, 'K8sJob'),
        (K8sStatefulSet, 'K8sStatefulSet'),
        (K8sService, 'K8sService'),
        (K8sConfigMap, 'K8sConfigMap'),
        (K8sSecret, 'K8sSecret'),
        (K8sIngress, 'K8sIngress'),
        (K8sHPA, 'K8sHPA'),
    )


class FileResourceName(ChoicesEnum):
    Deployment = 'Deployment'
    Service = 'Service'
    ConfigMap = 'ConfigMap'
    Secret = 'Secret'
    Ingress = 'Ingress'
    StatefulSet = 'StatefulSet'
    DaemonSet = 'DaemonSet'
    Job = 'Job'
    HPA = 'HPA'

    ServiceAccount = 'ServiceAccount'
    ClusterRole = 'ClusterRole'
    ClusterRoleBinding = 'ClusterRoleBinding'
    PodDisruptionBudget = 'PodDisruptionBudget'
    StorageClass = 'StorageClass'
    PersistentVolume = 'PersistentVolume'
    PersistentVolumeClaim = 'PersistentVolumeClaim'

    CustomManifest = 'CustomManifest'

    _choices_labels = (
        (Deployment, 'Deployment'),
        (Service, 'Service'),
        (ConfigMap, 'ConfigMap'),
        (Secret, 'Secret'),
        (Ingress, 'Ingress'),
        (StatefulSet, 'StatefulSet'),
        (DaemonSet, 'DaemonSet'),
        (Job, 'Job'),
        (HPA, 'HPA'),
        (ServiceAccount, 'ServiceAccount'),
        (ClusterRole, 'ClusterRole'),
        (ClusterRoleBinding, 'ClusterRoleBinding'),
        (PodDisruptionBudget, 'PodDisruptionBudget'),
        (StorageClass, 'StorageClass'),
        (PersistentVolume, 'PersistentVolume'),
        (PersistentVolumeClaim, 'PersistentVolumeClaim'),
        (CustomManifest, 'CustomManifest'),
    )


KRESOURCE_NAMES = K8sResourceName.choice_values()
RESOURCE_NAMES = KRESOURCE_NAMES
RESOURCES_WITH_POD = [
    FileResourceName.Deployment.value,
    FileResourceName.StatefulSet.value,
    FileResourceName.DaemonSet.value,
    FileResourceName.Job.value,
]


# env 环境, 现在是给namespace使用
class EnvType(Enum):
    DEV = "dev"
    TEST = "test"
    PROD = "prod"
