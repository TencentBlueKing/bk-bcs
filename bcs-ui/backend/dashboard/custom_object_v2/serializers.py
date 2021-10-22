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


class OptionalNamespaceSLZ(serializers.Serializer):
    # 部分资源对象是可以没有命名空间维度的
    namespace = serializers.CharField(label=_('命名空间'), required=False)


class FetchCustomObjectSLZ(OptionalNamespaceSLZ):
    """ 获取单个自定义对象 """


class CreateCustomObjectSLZ(serializers.Serializer):
    """ 创建自定义对象 """

    manifest = serializers.JSONField(label=_('资源配置信息'))


class UpdateCustomObjectSLZ(CreateCustomObjectSLZ):
    """ 更新（replace）某个自定义对象 """


class DestroyCustomObjectSLZ(OptionalNamespaceSLZ):
    """ 删除单个自定义对象 """
