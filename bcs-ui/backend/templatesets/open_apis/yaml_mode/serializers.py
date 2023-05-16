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

from backend.accounts import bcs_perm
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.resources.namespace.utils import get_namespaces_by_cluster_id
from backend.templatesets.legacy_apps.configuration import constants, models
from backend.templatesets.legacy_apps.configuration.yaml_mode.res2files import get_resource_file, get_template_files


def get_namespace_id(access_token, project_id, cluster_id, namespace):
    namespaces = get_namespaces_by_cluster_id(access_token, project_id, cluster_id)
    for ns in namespaces:
        if ns["name"] == namespace:
            return ns["id"]
    raise serializers.ValidationError(_("项目(id:{})下不存在命名空间({}/{})").format(project_id, cluster_id, namespace))


def add_fields_in_template_files(version_id, req_template_files):
    """
    add id and content fields in template_files
    """
    try:
        ventity = models.VersionedEntity.objects.get(id=version_id)
    except models.VersionedEntity.DoesNotExist:
        raise serializers.ValidationError(f"template version(id:{version_id}) does not exist")

    entity = ventity.get_entity()
    template_files = []
    for res_file in req_template_files:
        res_name = res_file["resource_name"]
        res_file_ids = entity[res_name].split(",")
        resource_file = get_resource_file(res_name, res_file_ids, "id", "name", "content")

        if "files" not in res_file:
            template_files.append(resource_file)
            continue

        if not res_file["files"]:
            raise serializers.ValidationError(f"empty parameter files in template_files({res_name})")

        resource_file_map = {f["name"]: f for f in resource_file["files"]}
        files = [resource_file_map[f["name"]] for f in res_file["files"]]
        template_files.append({"resource_name": res_name, "files": files})

    return template_files


class NamespaceInfoSLZ(serializers.Serializer):
    cluster_id = serializers.CharField()
    name = serializers.CharField()


class TemplateReleaseSLZ(serializers.Serializer):
    project_id = serializers.CharField()
    template_name = serializers.CharField()
    show_version_name = serializers.CharField()
    template_files = serializers.JSONField(required=False)
    namespace_info = NamespaceInfoSLZ()
    template_variables = serializers.JSONField(default={})

    def _validate_template_files(self, data):
        """
        template_files: [{'resource_name': 'Deployment', 'files': [{'name': ''}]}]
        """
        if "template_files" not in data:
            data["template_files"] = get_template_files(data["show_version"].real_version_id, "id", "name", "content")
            return

        template_files = data["template_files"]

        if not template_files:
            raise serializers.ValidationError("empty parameter template_files")

        try:
            data["template_files"] = add_fields_in_template_files(data["show_version"].real_version_id, template_files)
        except Exception as err:
            raise serializers.ValidationError(f"invalid parameter template_files: {err}")

    def _validate_namespace_info(self, data):
        request = self.context["request"]
        namespace_info = data["namespace_info"]
        namespace_info["id"] = get_namespace_id(
            request.user.token.access_token, data["project_id"], namespace_info["cluster_id"], namespace_info["name"]
        )
        perm_ctx = NamespaceScopedPermCtx(
            username=request.user.username,
            project_id=data["project_id"],
            cluster_id=namespace_info["cluster_id"],
            name=namespace_info["name"],
        )
        NamespaceScopedPermission().can_use(perm_ctx)

    def validate(self, data):
        template_name = data["template_name"]
        try:
            template = models.Template.objects.get(
                project_id=data["project_id"], name=template_name, edit_mode=constants.TemplateEditMode.YAML.value
            )
            data["template"] = template
        except models.Template.DoesNotExist:
            raise serializers.ValidationError(_("YAML模板集(name:{})不存在").format(template_name))

        try:
            show_version = models.ShowVersion.objects.get(name=data["show_version_name"], template_id=template.id)
            data["show_version"] = show_version
        except models.ShowVersion.DoesNotExist:
            raise serializers.ValidationError(
                _("YAML模板集(name:{})不存在版本{}").format(template_name, data["show_version_name"])
            )

        self._validate_namespace_info(data)
        self._validate_template_files(data)

        return data
