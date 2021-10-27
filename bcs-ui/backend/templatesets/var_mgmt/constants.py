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

ALL_PROJECTS = 0


class VariableScope(ChoicesEnum):
    GLOBAL = 'global'
    CLUSTER = 'cluster'
    NAMESPACE = 'namespace'

    _choices_labels = (
        (GLOBAL, _("全局变量")),
        (CLUSTER, _("集群变量")),
        (NAMESPACE, _("命名空间变量")),
    )


class VariableCategory(ChoicesEnum):
    SYSTEM = 'sys'
    CUSTOM = 'custom'

    _choices_labels = ((SYSTEM, _("系统内置")), (CUSTOM, _("自定义")))
