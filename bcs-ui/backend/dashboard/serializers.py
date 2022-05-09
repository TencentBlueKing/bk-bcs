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
from rest_framework import serializers


class ListResourceSLZ(serializers.Serializer):
    """查询 K8S 资源列表"""

    # NOTE：暂时只先支持 label_selector
    label_selector = serializers.CharField(label=_('标签选择算符'), required=False)
    # 过滤子资源（Pod）用参数
    owner_name = serializers.CharField(label=_('所属资源名称'), max_length=256, required=False)
    owner_kind = serializers.CharField(label=_('所属资源类型'), max_length=64, required=False)


class CreateResourceSLZ(serializers.Serializer):
    """创建 K8S 资源对象"""

    manifest = serializers.JSONField(label=_('资源配置信息'))


class UpdateResourceSLZ(CreateResourceSLZ):
    """更新 K8S 资源对象"""

    pass
