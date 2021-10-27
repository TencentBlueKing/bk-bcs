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

from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.utils.exceptions import ResNotFoundError

from .. import models

RE_SHOW_NAME = re.compile(r"^[a-zA-Z0-9-_.]{1,45}$")


class ShowVersionNameSLZ(serializers.Serializer):
    name = serializers.RegexField(
        RE_SHOW_NAME, max_length=45, required=True, error_messages={"invalid": "请填写1至45个字符（字母、数字、下划线以及 - 或 .）"}
    )
    comment = serializers.CharField(default="", allow_blank=True)


class ShowVersionCreateSLZ(ShowVersionNameSLZ):
    project_id = serializers.CharField(required=True)
    template_id = serializers.CharField(required=True)


class ShowVersionWithEntitySLZ(ShowVersionCreateSLZ):
    real_version_id = serializers.IntegerField(required=True)
    show_version_id = serializers.IntegerField(required=True)

    def validate(self, data):
        real_version_id = data["real_version_id"]
        if real_version_id <= 0:
            raise ValidationError(_("请先填写模板内容，再保存"))

        template_id = data["template_id"]
        try:
            models.VersionedEntity.objects.get(id=real_version_id, template_id=template_id)
        except models.VersionedEntity.DoesNotExist:
            raise ValidationError(_("模板集版本(id:{})不属于该模板(id:{})").format(real_version_id, template_id))

        template = models.get_template_by_project_and_id(data["project_id"], template_id)
        data["template"] = template
        return data


class GetShowVersionSLZ(serializers.Serializer):
    show_version_id = serializers.CharField(required=True)
    template_id = serializers.CharField(required=True)
    project_id = serializers.CharField(required=True)

    def validate(self, data):
        try:
            data["show_version_id"] = int(data["show_version_id"])
        except Exception as e:
            raise ValidationError(e)

        template_id = data["template_id"]
        template = models.get_template_by_project_and_id(data["project_id"], template_id)

        data["template"] = template

        show_version_id = data["show_version_id"]

        if show_version_id == -1:
            data["show_version"] = None
            return data

        try:
            data["show_version"] = models.ShowVersion.objects.get(id=show_version_id, template_id=template_id)
        except models.ShowVersion.DoesNotExist:
            raise ValidationError(
                f"show version(id:{show_version_id}) does not exist or not belong to template(id:{template_id})"
            )
        else:
            return data


class GetLatestShowVersionSLZ(serializers.Serializer):
    template_id = serializers.CharField(required=True)
    project_id = serializers.CharField(required=True)

    def validate(self, data):
        template = models.get_template_by_project_and_id(data["project_id"], data["template_id"])
        data["template"] = template
        data["show_version"] = models.ShowVersion.objects.get_latest_by_template(template.id)
        data["show_version_id"] = data["show_version"].id
        return data


class ResourceConfigSLZ(serializers.Serializer):
    show_version_id = serializers.IntegerField(required=True)
    config = serializers.SerializerMethodField()

    def get_config(self, obj):
        show_version_id = obj["show_version_id"]
        template = obj["template"]
        config = {"show_version_id": show_version_id}
        if show_version_id == -1:
            config["version"] = template.draft_version
            config.update(template.get_draft())
            return config

        show_version = obj["show_version"]
        real_version_id = show_version.real_version_id
        # ugly! real_version_id may be integer(-1, 0, ...) or None
        if real_version_id is None:
            config["version"] = real_version_id
            return config

        # version_id 为 -1 则查看草稿
        if real_version_id == -1:
            config["version"] = real_version_id
            config.update(template.get_draft())
            return config

        # real_version_id 为 0 查看最新版本
        if real_version_id == 0:
            ventity = models.VersionedEntity.objects.get_latest_by_template(template.id)
        else:
            try:
                ventity = models.VersionedEntity.objects.get(id=real_version_id)
            except models.VersionedEntity.DoesNotExist:
                raise ResNotFoundError(_("模板集版本(id:{})不存在").format(real_version_id))

        if ventity:
            config["version"] = ventity.id
            config.update(ventity.get_resource_config())

        return config

    def to_representation(self, instance):
        instance = super().to_representation(instance)
        config = instance["config"]
        del instance["config"]
        instance.update(config)
        return instance


class ListShowVersionSLZ(serializers.ModelSerializer):
    show_version_id = serializers.IntegerField(required=False, source="id")

    class Meta:
        model = models.ShowVersion
        fields = ("show_version_id", "real_version_id", "name", "updator", "updated", "comment")


class ListShowVersionISLZ(serializers.ModelSerializer):
    id = serializers.IntegerField(source="real_version_id")
    show_version_id = serializers.IntegerField(source="id")
    show_version_name = serializers.CharField(source="name")
    version = serializers.CharField(source="name")

    class Meta:
        model = models.ShowVersion
        fields = ("id", "show_version_id", "show_version_name", "version", "comment")
