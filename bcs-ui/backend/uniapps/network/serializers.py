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

from django.core.validators import MaxValueValidator, MinValueValidator
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.templatesets.legacy_apps.configuration.serializers import RE_NAME
from backend.uniapps.network.constants import K8S_LB_NAMESPACE
from backend.uniapps.network.models import K8SLoadBlance
from backend.utils.error_codes import error_codes


class BatchResourceSLZ(serializers.Serializer):
    data = serializers.JSONField(required=True)

    def validate_data(self, data):
        if not isinstance(data, list):
            raise ValidationError(_("数据格式必须为数组"))
        for _d in data:
            if not _d.get('cluster_id'):
                raise ValidationError(_("cluster_id 必填"))
            if not _d.get('namespace'):
                raise ValidationError(_("namespace 必填"))
            if not _d.get('name'):
                raise ValidationError(_("name 必填"))
        return data


class NginxIngressSLZ(serializers.ModelSerializer):
    project_id = serializers.CharField(max_length=32, required=True)
    cluster_id = serializers.CharField(max_length=32, required=True)
    namespace_id = serializers.IntegerField(required=True)
    name = serializers.CharField(max_length=32, required=True)
    protocol_type = serializers.CharField(max_length=32, required=False)
    ip_info = serializers.JSONField(required=True)
    detail = serializers.JSONField(required=False)
    namespace = serializers.CharField()

    class Meta:
        model = K8SLoadBlance
        fields = (
            "id",
            "project_id",
            "cluster_id",
            "namespace_id",
            "name",
            "protocol_type",
            "ip_info",
            "detail",
            "creator",
            "updator",
            "is_deleted",
            "namespace",
        )


class UpdateK8SLoadBalancerSLZ(serializers.Serializer):
    protocol_type = serializers.CharField(default="")
    ip_info = serializers.JSONField(required=False)
    version = serializers.CharField()
    values_content = serializers.JSONField()


class LoadBalancesSLZ(serializers.Serializer):
    """创建LB"""

    name = serializers.RegexField(
        RE_NAME,
        max_length=256,
        required=True,
        error_messages={'invalid': _('名称格式错误，只能包含：小写字母、数字、中划线(-)，首字母必须是字母，长度小于256个字符')},
    )
    cluster_id = serializers.CharField(required=True)
    instance = serializers.IntegerField(required=False, default=1)
    ip_list = serializers.JSONField(required=False, default=[])
    constraints = serializers.JSONField(required=True)
    # type = serializers.ChoiceField(choices=['cover', 'append'], required=True)
    namespace = serializers.CharField(required=False, default="")
    namespace_id = serializers.IntegerField(required=False, default=-1)
    network_type = serializers.CharField(required=True)
    network_mode = serializers.CharField(required=True)
    custom_value = serializers.CharField(required=False, allow_blank=True)
    image_url = serializers.CharField(required=True)
    image_version = serializers.CharField(required=True)
    resources = serializers.JSONField(required=True)
    forward_mode = serializers.CharField(required=True)
    eth_value = serializers.CharField(required=True)
    host_port = serializers.IntegerField(
        required=False, default=31000, validators=[MaxValueValidator(32000), MinValueValidator(31000)]
    )
    use_custom_image_url = serializers.BooleanField(required=False, default=False)

    def validate_constraints(self, constraints):
        """测试数据"""
        return constraints

    def validate_instance(self, instance):
        """当instance和ip_list都存在时，要校验两者数量相同"""
        data = self._kwargs.get('data', {})
        ip_list = data.get('ip_list')
        if not ip_list:
            return instance
        if len(ip_list) != instance:
            raise error_codes.CheckFailed(_("参数[ip_list]和[instance]必须相同"))
        return instance


class UpdateLoadBalancesSLZ(LoadBalancesSLZ):
    """更新值"""

    def validate(self, data):
        if not data:
            raise ValidationError(_("参数不能全部为空"))
        return data


class GetLoadBalanceSLZ(serializers.Serializer):
    limit = serializers.IntegerField(required=True)
    offset = serializers.IntegerField(required=True)


class ServiceListSLZ(serializers.Serializer):
    limit = serializers.IntegerField(default=10)
    offset = serializers.IntegerField(default=0)
    search_name = serializers.CharField(default='')
    cluster_id = serializers.CharField(default='ALL')


class ChartVersionSLZ(serializers.Serializer):
    version = serializers.CharField()
    cluster_id = serializers.CharField(required=False)
    namespace = serializers.CharField(required=False)


class CreateK8SLoadBalancerSLZ(serializers.Serializer):
    version = serializers.CharField()
    cluster_id = serializers.CharField()
    namespace = serializers.CharField(default=K8S_LB_NAMESPACE)
    values_content = serializers.JSONField()
    ip_info = serializers.JSONField()
