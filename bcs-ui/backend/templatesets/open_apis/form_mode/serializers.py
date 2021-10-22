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

from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from backend.templatesets.legacy_apps.configuration import models
from backend.templatesets.legacy_apps.configuration.constants import K8sResourceName
from backend.templatesets.legacy_apps.configuration.showversion.serializers import GetShowVersionSLZ
from backend.templatesets.legacy_apps.configuration.utils import to_bcs_res_name
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.templatesets.legacy_apps.instance.utils import validate_ns_by_tempalte_id


# TODO refactor validate_instance_entity
def generate_instance_entity(req_instance_resources, instance_resources_id_map):
    """验证前端传过了的待实例化资源是否是该版本的资源"""
    if not req_instance_resources:
        return instance_resources_id_map

    instance_entity = {}
    for res_name, res_info in req_instance_resources.items():
        res_id_list = [res["id"] for res in res_info]
        valid_id_list = instance_resources_id_map.get(res_name, [])
        invalid_id_list = list(set(res_id_list) - set(valid_id_list))
        if invalid_id_list:
            raise ValidationError(f"id {invalid_id_list} is invalid for the entity")
        instance_entity[res_name] = res_id_list
    return instance_entity


def get_instantiated_ns(access_token, template_id, namespace_id, project_id, instance_entity):
    _, instantiated_ns_name, _ = validate_ns_by_tempalte_id(
        template_id, [namespace_id], access_token, project_id, instance_entity
    )
    if instantiated_ns_name:
        return instantiated_ns_name[0]
    return ""


# TODO refactor with GetTemplateFilesSLZ
class GetFormTemplateSLZ(serializers.Serializer):
    show_version = serializers.SerializerMethodField()
    instance_entity = serializers.SerializerMethodField()
    name = serializers.SerializerMethodField()
    desc = serializers.SerializerMethodField()
    locker = serializers.SerializerMethodField()
    is_locked = serializers.SerializerMethodField()

    def get_show_version(self, obj):
        return OrderedDict({"name": obj["show_version"].name, "show_version_id": obj["show_version"].id})

    def get_instance_entity(self, obj):
        version_id = obj["show_version"].real_version_id
        ventity = models.VersionedEntity.objects.get(id=version_id)
        return ventity.get_instance_resources()

    def get_name(self, obj):
        return obj["template"].name

    def get_desc(self, obj):
        return obj["template"].desc

    def get_locker(self, obj):
        return obj["template"].locker

    def get_is_locked(self, obj):
        return obj["template"].is_locked


class TemplateReleaseSLZ(serializers.ModelSerializer):
    release_id = serializers.IntegerField(source="id")

    class Meta:
        model = VersionInstance
        fields = ("release_id", "show_version_id", "show_version_name", "is_bcs_success", "ns_id")


# TODO replace InstanceNamespaceSLZ
class InstanceEntitySLZ(serializers.Serializer):
    instance_entity = serializers.JSONField(required=False)

    def to_internal_value(self, data):
        data = super().to_internal_value(data)
        if "instance_entity" not in data:
            return data

        instance_entity = {}
        project_kind = self.context["project_kind"]
        for res_name, res_info in data["instance_entity"].items():
            bcs_res_name = to_bcs_res_name(project_kind, res_name)
            instance_entity[bcs_res_name] = res_info
        data["instance_entity"] = instance_entity
        return data


class CreateTemplateReleaseSLZ(InstanceEntitySLZ):
    show_version = GetShowVersionSLZ()
    namespace_id = serializers.CharField()
    namespace_variables = serializers.JSONField(required=False)
    is_start = serializers.BooleanField(default=True)

    def validate(self, data):
        show_version_slz_data = data["show_version"]
        show_version_obj = show_version_slz_data["show_version"]

        ventity = models.VersionedEntity.objects.get(id=show_version_obj.real_version_id)
        data["version_id"] = ventity.id
        instance_resources_id_map = ventity.instance_resources_id_map
        req_instance_resources = data.get("instance_entity")
        data["instance_entity"] = generate_instance_entity(req_instance_resources, instance_resources_id_map)
        template = show_version_slz_data["template"]
        data.update(
            {
                "template_id": template.id,
                "template_name": template.name,
                "template": template,
                "project_id": show_version_slz_data["project_id"],
            }
        )

        instantiated_ns_name = get_instantiated_ns(
            self.context["access_token"],
            data["template_id"],
            data["namespace_id"],
            data["project_id"],
            req_instance_resources,
        )
        if instantiated_ns_name:
            raise ValidationError(f"namespace {instantiated_ns_name} has been instantiated")

        data.update(
            {
                "namespaces": data["namespace_id"],
                "ns_list": [data["namespace_id"]],
                "show_version_id": show_version_slz_data["show_version_id"],
                "show_version_name": show_version_obj.name,
                "variable_info": {data["namespace_id"]: data.get("namespace_variables", {})},
            }
        )

        return data


class UpdateTemplateReleaseSLZ(serializers.Serializer):
    resource_name = serializers.ChoiceField(choices=[K8sResourceName.K8sDeployment.value])
    name = serializers.CharField()
    namespace_id = serializers.CharField()
    namespace_variables = serializers.JSONField(required=False)
    project_id = serializers.CharField()
    template_id = serializers.CharField()
    release_id = serializers.CharField()

    def validate(self, data):
        template_id = data["template_id"]
        data["template"] = models.get_template_by_project_and_id(data["project_id"], template_id)

        namespace_id = data["namespace_id"]
        release_id = data["release_id"]
        try:
            VersionInstance.objects.get(id=release_id, template_id=template_id, ns_id=namespace_id)
        except VersionInstance.DoesNotExist:
            raise ValidationError(
                f"release does not exist: release_id({release_id}), template_id({template_id}), "
                f"namespace_id({namespace_id})"
            )

        name = data["name"]
        resource_name = data["resource_name"]
        try:
            InstanceConfig.objects.get(
                name=name, instance_id=release_id, namespace=namespace_id, category=resource_name
            )
        except InstanceConfig.DoesNotExist:
            raise ValidationError(
                f"release does not exist: release_id({release_id}), name({name}), "
                f"namespace_id({namespace_id}), resource_name({resource_name})"
            )
        data["variable_info"] = {namespace_id: data.get("namespace_variables", {})}
        return data
