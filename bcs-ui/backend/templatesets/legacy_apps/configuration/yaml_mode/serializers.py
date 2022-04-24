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
from collections import OrderedDict

from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.utils.error_codes import error_codes

from .. import utils
from ..constants import FileAction, FileResourceName, TemplateEditMode
from ..showversion.serializers import GetShowVersionSLZ, ShowVersionNameSLZ
from . import files2res, res2files


class ResourceFileSLZ(serializers.Serializer):
    name = serializers.CharField()
    content = serializers.CharField(allow_blank=True)
    id = serializers.CharField(required=False)
    action = serializers.ChoiceField(choices=FileAction.get_choices())

    def validate_id(self, id):
        try:
            assert int(id) > 0, f"file id {id} must be positive integer"
        except Exception as e:
            raise ValidationError(f"file id {id} error: {e}")
        return id


class TemplateFileSLZ(serializers.Serializer):
    resource_name = serializers.ChoiceField(choices=FileResourceName.get_choices())
    files = serializers.ListField(child=ResourceFileSLZ(), allow_empty=False)


class CreateTemplateFileSLZ(TemplateFileSLZ):
    def validate_files(self, files):
        for f in files:
            if f["action"] != FileAction.CREATE.value:
                raise ValidationError(f"file {f['name']} action must be {FileAction.CREATE.value}")
            if not f["content"]:
                raise ValidationError(f"file {f['name']} content cannot be blank")

        name_list = [f["name"] for f in files]
        if len(name_list) != len(set(name_list)):
            raise ValidationError("file name is duplicated")

        return files


class UpdateTemplateFileSLZ(TemplateFileSLZ):
    def validate_files(self, files):
        for f in files:
            if f["action"] not in FileAction.choice_values():
                raise ValidationError(f"file {f['name']} action {f['action']} is invalid")
            if "id" not in f and f["action"] != FileAction.CREATE.value:
                raise ValidationError(f"file {f['name']} miss file id")
            if not f["content"] and f["action"] != FileAction.DELETE.value:
                raise ValidationError(f"file {f['name']} content cannot be blank")

        name_list = [f["name"] for f in files]
        if len(name_list) != len(set(name_list)):
            raise ValidationError("file name is duplicated")

        return files


class YamlTemplateSLZ(serializers.Serializer):
    name = serializers.CharField(max_length=30)
    desc = serializers.CharField(max_length=50, required=False, allow_blank=True)
    project_id = serializers.CharField(max_length=64)


class CreateTemplateSLZ(YamlTemplateSLZ):
    show_version = ShowVersionNameSLZ()
    template_files = serializers.ListField(child=CreateTemplateFileSLZ(), allow_empty=False)

    def create(self, validated_data):
        request = self.context["request"]
        project_id = validated_data["project_id"]
        desc_args = {
            "name": validated_data["name"],
            "desc": validated_data.get("desc", ""),
            "project_id": project_id,
            "edit_mode": TemplateEditMode.YAML.value,
        }
        template = utils.create_template_with_perm_check(request, project_id, desc_args)
        files2res.create_resources(template, validated_data["show_version"], validated_data["template_files"])
        return template


class UpdateShowVersionSLZ(ShowVersionNameSLZ):
    old_show_version_id = serializers.IntegerField()


class UpdateTemplateSLZ(YamlTemplateSLZ):
    show_version = UpdateShowVersionSLZ(required=False)
    template_files = serializers.ListField(child=UpdateTemplateFileSLZ(), required=False, allow_empty=False)
    updated_timestamp = serializers.FloatField()

    def update(self, template, validated_data):
        if validated_data["updated_timestamp"] < template.updated.timestamp():
            raise error_codes.ExpiredError(_("模板集内容已被更改，请刷新页面"))

        request = self.context["request"]
        desc_args = {
            "name": validated_data["name"],
            "desc": validated_data.get("desc", ""),
        }
        template = utils.update_template_with_perm_check(request, template, desc_args)
        if validated_data.get("show_version"):
            files2res.update_resources(template, validated_data["show_version"], validated_data.get("template_files"))
        return template


class GetTemplateFilesSLZ(serializers.Serializer):
    id = serializers.SerializerMethodField()
    show_version = serializers.SerializerMethodField()
    template_files = serializers.SerializerMethodField()
    name = serializers.SerializerMethodField()
    desc = serializers.SerializerMethodField()
    locker = serializers.SerializerMethodField()
    is_locked = serializers.SerializerMethodField()
    updated_timestamp = serializers.SerializerMethodField()

    def get_id(self, obj):
        return obj["template"].id

    def get_show_version(self, obj):
        return OrderedDict({"name": obj["show_version"].name, "show_version_id": obj["show_version"].id})

    def get_template_files(self, obj):
        version_id = obj["show_version"].real_version_id
        if self.context["with_file_content"]:
            return res2files.get_template_files(version_id, "id", "name", "content")
        else:
            return res2files.get_template_files(version_id, "id", "name")

    def get_name(self, obj):
        return obj["template"].name

    def get_desc(self, obj):
        return obj["template"].desc

    def get_locker(self, obj):
        return obj["template"].locker

    def get_is_locked(self, obj):
        return obj["template"].is_locked

    def get_updated_timestamp(self, obj):
        return obj["template"].updated.timestamp()


class PreviewResourceFileSLZ(serializers.Serializer):
    id = serializers.CharField()
    name = serializers.CharField()


class PreviewTemplateFileSLZ(TemplateFileSLZ):
    files = serializers.ListField(child=PreviewResourceFileSLZ(), allow_empty=False)


class TemplateReleaseSLZ(serializers.Serializer):
    show_version = GetShowVersionSLZ()
    namespace_id = serializers.IntegerField()
    template_files = serializers.ListField(child=PreviewTemplateFileSLZ(), allow_empty=False)
    is_preview = serializers.BooleanField()
    template_variables = serializers.JSONField(default={})

    def to_internal_value(self, data):
        data = super().to_internal_value(data)

        template_files = []
        for res_file in data["template_files"]:
            res_file_ids = [f["id"] for f in res_file["files"]]
            res_file = res2files.get_resource_file(res_file["resource_name"], res_file_ids, "id", "name", "content")
            template_files.append(res_file)
        data["template_files"] = template_files

        return data
