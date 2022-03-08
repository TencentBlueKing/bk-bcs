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
import json
import logging

from rest_framework import viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.components import paas_cc
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.iam.permissions.resources.templateset import (
    TemplatesetAction,
    TemplatesetCreatorAction,
    TemplatesetPermission,
    TemplatesetRequest,
)
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

from ..mixins import TemplatePermission
from ..models import get_template_by_project_and_id
from ..showversion.serializers import GetLatestShowVersionSLZ, GetShowVersionSLZ
from . import init_tpls, serializers
from .deployer import DeployController
from .release import ReleaseData, ReleaseDataProcessor

logger = logging.getLogger(__name__)


class InitialTemplatesViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_initial_templates(self, request, project_id):
        return Response(init_tpls.get_initial_templates())


class YamlTemplateViewSet(viewsets.ViewSet, TemplatePermission):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def _request_data(self, request, **kwargs):
        request_data = request.data.copy() or {}
        logger.info(json.dumps(request_data))
        request_data.update(**kwargs)
        return request_data

    def create_template(self, request, project_id):
        """
        request.data = {
            'name': '',
            'desc': '',
            'show_version': {
                'name': '',
            }
            'template_files': [{
                'resource_name': 'Deployment',
                'files': [{'name': 'nginx.yaml', 'content': 'Kind:Deployment', 'action': 'create'}]
            }]
        }
        """
        data = self._request_data(request, project_id=project_id)
        serializer = serializers.CreateTemplateSLZ(data=data, context={"request": request})
        serializer.is_valid(raise_exception=True)
        template = serializer.save()

        self.iam_perm.grant_resource_creator_actions(
            TemplatesetCreatorAction(
                template_id=template.id,
                name=template.name,
                project_id=template.project_id,
                creator=request.user.username,
            )
        )
        return Response({"template_id": template.id})

    def update_template(self, request, project_id, template_id):
        """
        request.data = {
            'name': '',
            'desc': '',
            'updated_timestamp': 1595907256.0,
            'show_version': {
                'name': '',
                'old_show_version_id': '',
            }
            'template_files': [{
                'resource_name': 'Deployment',
                'files': [{'name': 'nginx.yaml', 'content': 'Kind:Deployment', 'action': 'update', 'id': 3}]
            }]
        }
        """
        template = get_template_by_project_and_id(project_id, template_id)
        data = self._request_data(request, project_id=project_id)
        serializer = serializers.UpdateTemplateSLZ(template, data=data, context={"request": request})
        serializer.is_valid(raise_exception=True)
        template = serializer.save()
        return Response({"template_id": template.id})

    def get_template_by_show_version(self, request, project_id, template_id, show_version_id):
        serializer = GetShowVersionSLZ(data=self.kwargs)
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        template = validated_data["template"]
        self.can_view_template(request, template)

        with_file_content = request.query_params.get("with_file_content")
        with_file_content = False if with_file_content == "false" else True

        serializer = serializers.GetTemplateFilesSLZ(validated_data, context={"with_file_content": with_file_content})
        return Response(serializer.data)

    @response_perms(
        action_ids=[TemplatesetAction.VIEW, TemplatesetAction.UPDATE, TemplatesetAction.INSTANTIATE],
        permission_cls=TemplatesetPermission,
        resource_id_key='id',
    )
    def get_template(self, request, project_id, template_id):
        serializer = GetLatestShowVersionSLZ(data=self.kwargs)
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        template = validated_data["template"]
        self.can_view_template(request, template)

        serializer = serializers.GetTemplateFilesSLZ(validated_data, context={"with_file_content": True})
        return PermsResponse(serializer.data, TemplatesetRequest(project_id=project_id))


class TemplateReleaseViewSet(viewsets.ViewSet, TemplatePermission):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def _request_data(self, request, **kwargs):
        request_data = request.data.copy() or {}
        request_data["show_version"] = kwargs
        return request_data

    # TODO use resources module function
    def _get_namespace_info(self, access_token, project_id, namespace_id):
        resp = paas_cc.get_namespace(access_token, project_id, namespace_id)
        if resp.get("code") != 0:
            raise error_codes.APIError(f"get namespace(id:{namespace_id}) info error: {resp.get('message')}")
        return resp.get("data")

    def _raw_release_data(self, project_id, initial_data):
        show_version = initial_data["show_version"]
        namespace_info = self._get_namespace_info(
            self.request.user.token.access_token, project_id, initial_data["namespace_id"]
        )
        raw_release_data = ReleaseData(
            project_id=project_id,
            namespace_info=namespace_info,
            show_version=show_version["show_version"],
            template_files=initial_data["template_files"],
            template_variables=initial_data["template_variables"],
        )
        return raw_release_data

    def preview_or_apply(self, request, project_id, template_id, show_version_id):
        """
        request.data = {
            'is_preview': True,
            'namespace_id': 'test',
            'template_files': [{
                'resource_name': 'Deployment',
                'files': [{'name': 'nginx.yaml', 'id': 3}]
            }],
            'template_variables': {}
        }
        """
        data = self._request_data(
            request, project_id=project_id, template_id=template_id, show_version_id=show_version_id
        )
        serializer = serializers.TemplateReleaseSLZ(data=data)
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        template = validated_data["show_version"]["template"]
        self.can_use_template(request, template)

        resp = paas_cc.get_namespace(request.user.token.access_token, project_id, validated_data["namespace_id"])
        if resp.get('code') != 0:
            return Response(
                {
                    'code': 400,
                    'message': f"查询命名空间(namespace_id:{project_id}-{validated_data['namespace_id']})出错:{resp.get('message')}",
                }
            )

        namespace_info = resp['data']
        perm_ctx = NamespaceScopedPermCtx(
            username=request.user.username,
            project_id=project_id,
            cluster_id=namespace_info['cluster_id'],
            name=namespace_info['name'],
        )
        NamespaceScopedPermission().can_use(perm_ctx)

        processor = ReleaseDataProcessor(
            user=self.request.user, raw_release_data=self._raw_release_data(project_id, validated_data)
        )
        release_data = processor.release_data()

        if validated_data["is_preview"]:
            return Response(release_data.template_files)

        controller = DeployController(user=self.request.user, release_data=release_data)
        controller.apply()

        return Response()
