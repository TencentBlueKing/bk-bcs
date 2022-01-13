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

# k8s 中系统的命名空间，不允许用户创建，也不能操作上面的资源 kube-system, kube-public
K8S_SYS_NAMESPACE = ["kube-system", "kube-public"]

# k8s 平台服务用的命名空间
# TODO: bcs-system命名空间后续处理
K8S_PLAT_NAMESPACE = ["web-console", "gitlab-ci", "thanos"]

# 平台和系统使用的命名空间
K8S_SYS_PLAT_NAMESPACES = K8S_SYS_NAMESPACE + K8S_PLAT_NAMESPACE

# BCS 服务保留的命名空间（不允许直接创建，更新，删除等操作）
BCS_RESERVED_NAMESPACES = K8S_SYS_PLAT_NAMESPACES + ['bcs-system']

PROJ_CODE_ANNO_KEY = 'io.tencent.bcs.projectcode'
