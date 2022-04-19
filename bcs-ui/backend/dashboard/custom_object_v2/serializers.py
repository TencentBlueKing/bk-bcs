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
from rest_framework.exceptions import ValidationError

from backend.dashboard.utils.resources import get_crd_info
from backend.resources.constants import ResourceScope


class OptionalNamespaceSLZ(serializers.Serializer):
    # 部分自定义资源对象是可以没有命名空间维度的
    namespace = serializers.CharField(label=_('命名空间'), required=False)

    def validate(self, attrs):
        """ 若没有指定命名空间，则检查，若资源有命名空间维度，抛出异常 """
        if get_crd_info(**self.context).get('scope') == ResourceScope.Cluster:
            attrs['namespace'] = None
        elif not attrs.get('namespace'):
            raise ValidationError(_('查看/操作自定义资源 {} 需指定命名空间').format(self.context['crd_name']))
        return attrs


class ListCustomObjectSLZ(OptionalNamespaceSLZ):
    """ 获取自定义资源列表 """


class FetchCustomObjectSLZ(OptionalNamespaceSLZ):
    """ 获取单个自定义对象 """


class CreateCustomObjectSLZ(serializers.Serializer):
    """ 创建自定义对象 """

    manifest = serializers.JSONField(label=_('资源配置信息'))


class UpdateCustomObjectSLZ(CreateCustomObjectSLZ, OptionalNamespaceSLZ):
    """ 更新（replace）某个自定义对象 """


class DestroyCustomObjectSLZ(OptionalNamespaceSLZ):
    """ 删除单个自定义对象 """
