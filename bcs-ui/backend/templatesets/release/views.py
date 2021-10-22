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
from dataclasses import asdict

from rest_framework.response import Response

from backend.bcs_web import viewsets
from backend.resources.utils.kube_client import get_dynamic_client
from backend.utils.data_types import make_dataclass_from_dict

from .. import models
from . import serializers
from .generator.generator import ReleaseDataGenerator
from .generator.res_context import ResContext
from .manager import AppReleaseManager


class ReleaseViewSet(viewsets.SystemViewSet):
    def preview_manifests(self, request, project_id):
        release_data = self._release_data(request, project_id, is_preview=True)
        return Response([asdict(res) for res in release_data.resource_list])

    def create(self, request, project_id):
        return self._update_or_create(request, project_id)

    def update(self, request, project_id, release_id):
        return self._update_or_create(request, project_id)

    def list(self, request, project_id):
        serializer = serializers.ListReleaseSLZ(data=request.data)
        serializer.is_valid(raise_exception=True)

        query_params = {'project_id': project_id}
        query_params.update(serializer.validated_data)

        serializer = serializers.ReleaseSLZ(models.AppRelease.objects.filter(**query_params), many=True)
        return Response(serializer.data)

    def get(self, request, project_id, release_id):
        serializer = serializers.ReleaseSLZ(models.AppRelease.objects.get(id=release_id, project_id=project_id))
        return Response(serializer.data)

    def delete(self, request, project_id, release_id):
        app_release = models.AppRelease.objects.get(id=release_id, project_id=project_id)
        release_manager = AppReleaseManager(
            dynamic_client=get_dynamic_client(request.use.token.access_token, project_id, app_release.cluster_id)
        )
        release_manager.delete(request.user.username, release_id)
        return Response()

    def _release_data(self, request, project_id, is_preview=False):
        req_data = self.get_request_data(request, project_id=project_id, is_preview=is_preview)
        serializer = serializers.GetReleaseResourcesSLZ(data=req_data)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data
        validated_data.update({'access_token': request.user.token.access_token, 'username': request.user.username})

        res_ctx = make_dataclass_from_dict(ResContext, validated_data)
        release_data = ReleaseDataGenerator(name=validated_data['name'], res_ctx=res_ctx).generate()

        return release_data

    def _update_or_create(self, request, project_id):
        release_data = self._release_data(request, project_id)
        release_manager = AppReleaseManager(
            dynamic_client=get_dynamic_client(request.use.token.access_token, project_id, release_data.cluster_id)
        )
        release_manager.update_or_create(request.user.username, release_data=release_data)
        return Response()
