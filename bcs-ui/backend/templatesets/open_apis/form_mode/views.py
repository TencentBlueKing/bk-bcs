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
import copy

from rest_framework.response import Response

from backend.bcs_web.apis import views
from backend.templatesets.legacy_apps.configuration.mixins import TemplatePermission
from backend.templatesets.legacy_apps.configuration.models import get_template_by_project_and_id
from backend.templatesets.legacy_apps.configuration.showversion.serializers import GetShowVersionSLZ
from backend.templatesets.legacy_apps.instance.models import VersionInstance

from . import serializers
from .deployer import DeployController
from .release import ReleaseDataProcessor


class TemplateViewSet(views.BaseAPIViewSet, TemplatePermission):

    # TODO replace TemplateResourceView.get, but need support ResourceRequestSLZ
    def get_template_by_show_version(self, request, *args, **kwargs):
        req_data = copy.deepcopy(self.kwargs)
        req_data["project_id"] = request.project.project_id

        serializer = GetShowVersionSLZ(data=req_data)
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        self.can_view_template(request, validated_data["template"])
        serializer = serializers.GetFormTemplateSLZ(validated_data)

        return Response(serializer.data)


class TemplateReleaseViewSet(views.NoAccessTokenBaseAPIViewSet, TemplatePermission):
    def _request_data(self, request, project_id, template_id, show_version_id):
        request_data = request.data.copy() or {}
        show_version = {"show_version_id": show_version_id, "template_id": template_id, "project_id": project_id}
        request_data["show_version"] = show_version
        return request_data

    def _merge_path_params(self, request, **kwargs):
        request_data = request.data.copy() or {}
        request_data.update(**kwargs)
        return request_data

    def _filter_release(self, request, project_id, template_id, context):
        template = get_template_by_project_and_id(project_id, template_id)
        self.can_view_template(request, template)

        qsets = VersionInstance.objects.filter(template_id=template_id)
        namespace_id = request.query_params.get("namespace_id")
        if namespace_id:
            qsets = qsets.filter(ns_id=namespace_id)

        if context.get("latest"):
            serializer = serializers.TemplateReleaseSLZ(qsets.latest("created"))
        else:
            serializer = serializers.TemplateReleaseSLZ(qsets, many=True)
        return serializer.data

    def list_releases(self, request, project_id_or_code, template_id):
        """
        query_params = {'namespace_id': ''}
        """
        data = self._filter_release(request, request.project.project_id, template_id, context={})
        return Response(data)

    def get_latest_release(self, request, project_id_or_code, template_id):
        """
        query_params = {'namespace_id': ''}
        """
        data = self._filter_release(request, request.project.project_id, template_id, context={"latest": True})
        return Response(data)

    def create_release(self, request, project_id_or_code, template_id, show_version_id):
        """
        request.data = {
            'namespace_id': 19873,
            'namespace_variables': {'image_tag': '1.0'}
            'instance_entity': {
                'Deployment': [{'name': 'nginx-deployment', 'id': 3 # 必须传入id}]
            }
        }
        """
        data = self._request_data(request, request.project.project_id, template_id, show_version_id)
        serializer = serializers.CreateTemplateReleaseSLZ(
            data=data,
            context={"project_kind": request.project.project_id, "access_token": request.user.token.access_token},
        )
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        template = validated_data["template"]
        self.can_use_template(request, template)

        processor = ReleaseDataProcessor(validated_data)
        release_data = processor.release_data("create")

        controller = DeployController(user=self.request.user, project_kind=request.project.kind)
        release_id = controller.create_release(release_data)
        return Response({"release_id": release_id})

    def update_resource(self, request, project_id_or_code, template_id, release_id):
        """
        request.data = {
            'resource_name': 'Deployment',
            'name': 'nginx',
            'namespace_id': '',
            'namespace_variables': {'image_tag': '1.0'}
        }
        """
        path_params = {"project_id": request.project.project_id, "template_id": template_id, "release_id": release_id}
        serializer = serializers.UpdateTemplateReleaseSLZ(data=self._merge_path_params(request, **path_params))
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        template = validated_data["template"]
        self.can_use_template(request, template)

        processor = ReleaseDataProcessor(validated_data)
        release_data = processor.release_data("update")

        controller = DeployController(user=self.request.user, project_kind=request.project.kind)
        release_id = controller.update_release(release_data)
        return Response({"release_id": release_id})
