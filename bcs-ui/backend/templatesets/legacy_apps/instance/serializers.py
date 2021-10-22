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
from rest_framework import serializers

from ..configuration.utils import to_bcs_res_name
from .models import VersionInstance


class InstanceNamespaceSLZ(serializers.Serializer):
    """"""

    instance_entity = serializers.JSONField(required=False)

    def validate_instance_entity(self, instance_entity):
        new_entity = {}
        project_kind = self.context['project_kind']
        for cate in instance_entity:
            # 将前台的模板类型转换为db存储的类型
            real_cate = to_bcs_res_name(project_kind, cate)
            new_entity[real_cate] = instance_entity[cate]
        return new_entity


# TODO mark refactor
class VariableNamespaceSLZ(InstanceNamespaceSLZ):
    namespaces = serializers.CharField(required=True)


class VersionInstanceSLZ(serializers.ModelSerializer):
    entity = serializers.JSONField(source='get_entity', read_only=True)
    template_id = serializers.JSONField(source='get_template_id', read_only=True)
    project_id = serializers.JSONField(source='get_project_id', read_only=True)

    class Meta:
        model = VersionInstance
        fields = ("id", "entity", "template_id", "project_id", "namespaces", "is_start", "version_id")


class VersionInstanceCreateOrUpdateSLZ(InstanceNamespaceSLZ):
    version_id = serializers.IntegerField(required=True)
    namespaces = serializers.CharField(required=True)
    is_start = serializers.BooleanField(required=True)
    lb_info = serializers.JSONField(required=False)

    show_version_id = serializers.IntegerField(required=True)
    show_version_name = serializers.CharField(required=False)
    variable_info = serializers.JSONField(required=False)

    def validate_is_start(self, is_start):
        """后台默认为True"""
        return True


class PreviewInstanceSLZ(InstanceNamespaceSLZ):
    version_id = serializers.IntegerField(required=True)
    namespace = serializers.CharField(required=True)
    lb_info = serializers.JSONField(required=False)

    show_version_id = serializers.IntegerField(required=True)
    show_version_name = serializers.CharField(required=False)
    variable_info = serializers.JSONField(required=False)


class SingleInstanceSLZ(serializers.Serializer):
    version_id = serializers.IntegerField(required=True)
    namespaces = serializers.CharField(required=True)
    is_start = serializers.BooleanField(required=True)

    category = serializers.CharField(required=True)
    tmpl_app_id = serializers.IntegerField(required=True)
    tmpl_app_name = serializers.CharField(required=True)

    show_version_id = serializers.IntegerField(required=True)
    show_version_name = serializers.CharField(required=False)

    def validate_is_start(self, is_start):
        """后台默认为True"""
        return True


class PreviwSingleInstanceSLZ(serializers.Serializer):
    version_id = serializers.IntegerField(required=True)
    namespace = serializers.CharField(required=True)

    category = serializers.CharField(required=True)
    tmpl_app_id = serializers.IntegerField(required=True)
    tmpl_app_name = serializers.CharField(required=True)

    show_version_id = serializers.IntegerField(required=True)
    show_version_name = serializers.CharField(required=False)
