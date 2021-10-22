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
from rest_framework.response import Response

from backend.bcs_web.apis.views import NoAccessTokenBaseAPIViewSet
from backend.templatesets.legacy_apps.configuration.mixins import TemplatePermission
from backend.templatesets.legacy_apps.configuration.yaml_mode.deployer import DeployController
from backend.templatesets.legacy_apps.configuration.yaml_mode.release import ReleaseData, ReleaseDataProcessor

from .serializers import TemplateReleaseSLZ


class TemplateReleaseViewSet(NoAccessTokenBaseAPIViewSet, TemplatePermission):
    def _request_data(self, request, **kwargs):
        request_data = request.data.copy() or {}
        request_data.update(**kwargs)
        return request_data

    def apply(self, request, project_id_or_code):
        project_id = request.project.project_id
        data = self._request_data(request, project_id=project_id)
        serializer = TemplateReleaseSLZ(data=data, context={"request": request})
        serializer.is_valid(raise_exception=True)
        validated_data = serializer.validated_data

        self.can_use_template(request, validated_data["template"])

        validated_data = serializer.validated_data
        processor = ReleaseDataProcessor(
            user=self.request.user,
            raw_release_data=ReleaseData(
                project_id=project_id,
                namespace_info=validated_data["namespace_info"],
                show_version=validated_data["show_version"],
                template_files=validated_data["template_files"],
                template_variables=validated_data["template_variables"],
            ),
        )

        controller = DeployController(user=self.request.user, release_data=processor.release_data())

        controller.apply()
        return Response()
