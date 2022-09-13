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
from django.conf import settings

from backend.resources.constants import K8sResourceKind

# 资源模板相关配置 目录
RESOURCE_EXAMPLE_DIR = f'{settings.BASE_DIR}/backend/resources/examples'
# 模板配置信息 目录
EXAMPLE_CONFIG_DIR = f'{RESOURCE_EXAMPLE_DIR}/configs'
# Demo 配置文件 目录
DEMO_RESOURCE_MANIFEST_DIR = f'{RESOURCE_EXAMPLE_DIR}/manifests'
# 参考资料 目录
RESOURCE_REFERENCES_DIR = f'{RESOURCE_EXAMPLE_DIR}/references'

# 资源名称后缀长度
RANDOM_SUFFIX_LENGTH = 8
# 后缀可选字符集（小写 + 数字）
SUFFIX_ALLOWED_CHARS = 'abcdefghijklmnopqrstuvwxyz0123456789'

RES_KIND_WITH_DEMO_MANIFEST = [
    # workload
    K8sResourceKind.Deployment.value,
    K8sResourceKind.StatefulSet.value,
    K8sResourceKind.DaemonSet.value,
    K8sResourceKind.CronJob.value,
    K8sResourceKind.Job.value,
    K8sResourceKind.Pod.value,
    # network
    K8sResourceKind.Ingress.value,
    K8sResourceKind.Service.value,
    K8sResourceKind.Endpoints.value,
    # configuration
    K8sResourceKind.ConfigMap.value,
    K8sResourceKind.Secret.value,
    # storage
    K8sResourceKind.PersistentVolume.value,
    K8sResourceKind.PersistentVolumeClaim.value,
    K8sResourceKind.StorageClass.value,
    # hpa
    K8sResourceKind.HorizontalPodAutoscaler.value,
    # rbac
    K8sResourceKind.ServiceAccount.value,
    # CustomResource
    K8sResourceKind.CustomObject.value,
    K8sResourceKind.GameDeployment.value,
    K8sResourceKind.GameStatefulSet.value,
]
