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

from backend.dashboard.utils.resources import get_crd_scope
from backend.resources.constants import ResourceScope

from .constants import CLUSTER_SCOPE_RESOURCE_KINDS
from .utils import is_cobj_kind


class FetchResourceWatchResultSLZ(serializers.Serializer):
    """获取单类资源一段时间内的变更记录"""

    resource_version = serializers.CharField(label=_('资源版本号'), max_length=32)
    kind = serializers.CharField(label=_('资源类型'), max_length=128)
    crd_name = serializers.CharField(label=_('CRD 名称'), max_length=64, required=False)
    api_version = serializers.CharField(label=_('API 版本'), max_length=128, required=False)
    # 部分资源如 PersistentVolume，StorageClass 等不存在命名空间维度
    namespace = serializers.CharField(label=_('命名空间'), max_length=64, required=False)

    def validate(self, attrs):
        """ 若不是确定支持的资源类型（如自定义资源），则需要提供 apiVersion，以免需先查询 CRD """
        kind, namespace, crd_name = attrs['kind'], attrs.get('namespace'), attrs.get('crd_name')
        if is_cobj_kind(kind):
            if not (attrs.get('api_version') and crd_name):
                raise ValidationError(_('当资源类型为自定义对象时，需要指定 ApiVersion & CRDName'))
            # 自定义资源 & 没有指定命名空间则查询 CRD 检查配置
            if not namespace and get_crd_scope(crd_name, **self.context) == ResourceScope.Namespaced:
                raise ValidationError(_('查询当前自定义资源事件需要指定 Namespace'))
        # K8S 原生资源固定数种不需要指定命名空间
        elif not (kind in CLUSTER_SCOPE_RESOURCE_KINDS or namespace):
            raise ValidationError(_('查询当前资源事件需要指定 Namespace'))
        return attrs
