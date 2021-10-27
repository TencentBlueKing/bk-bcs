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
from typing import Dict, List

from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.templatesets.legacy_apps.configuration import models as cfg_models

from .. import models


class GetReleaseResourcesSLZ(serializers.Serializer):
    name = serializers.CharField(max_length=200)
    project_id = serializers.CharField()
    cluster_id = serializers.CharField()
    namespace = serializers.CharField()
    namespace_id = serializers.IntegerField(required=False)
    template_id = serializers.IntegerField()
    show_version_id = serializers.IntegerField()
    instance_entity = serializers.JSONField()
    template_variables = serializers.JSONField(required=False)
    is_preview = serializers.BooleanField()

    def validate_instance_entity(self, instance_entity: Dict[str, List[dict]]) -> Dict[str, List[int]]:
        """清洗 instance_entity: 保留资源id，去除name等其他信息"""
        if not instance_entity:
            raise ValidationError("empty instance_entity")

        try:
            entity = {}
            for res_kind in instance_entity:
                entity[res_kind] = [int(res_data['id']) for res_data in instance_entity[res_kind]]
            return entity
        except (KeyError, ValueError) as e:
            raise ValidationError(f"invalid instance_entity: {e}")

    def validate(self, data):
        """
        - 校验 template_id 和 show_version_id
        - 增加 template/show_version 字段
        """
        project_id = data['project_id']
        template_id = data['template_id']
        try:
            data['template'] = cfg_models.Template.objects.get(id=template_id, project_id=project_id)
        except cfg_models.Template.DoesNotExist:
            raise ValidationError(_("项目(project_id:{})下的模板集(template_id:{})不存在").format(project_id, template_id))

        try:
            data['show_version'] = cfg_models.ShowVersion.objects.get(
                id=data['show_version_id'], template_id=template_id
            )
        except cfg_models.ShowVersion.DoesNotExist:
            raise ValidationError(
                _("模板集版本(show_version_id:{})不存在或不属于模板集(template_id:{})").format(data['show_version_id'], template_id)
            )

        return data


class ListReleaseSLZ(serializers.Serializer):
    name = serializers.CharField(required=False)
    cluster_id = serializers.CharField(required=False)
    namespace = serializers.CharField(required=False)


class ReleaseSLZ(serializers.ModelSerializer):
    class Meta:
        model = models.AppRelease
        exclude = ('is_deleted', 'deleted_time')
