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
from rest_framework import viewsets
from rest_framework.response import Response

from backend.bcs_web.apis.authentication import JWTAuthentication
from backend.bcs_web.apis.permissions import AccessTokenPermission
from backend.container_service.projects.base import list_projects
from backend.utils.renderers import BKAPIRenderer


class ProjectsViewSet(viewsets.ViewSet):
    authentication_classes = (JWTAuthentication,)
    renderer_classes = (BKAPIRenderer,)
    permission_classes = (AccessTokenPermission,)

    def list_projects(self, request):
        projects = list_projects(request.user.token.access_token)
        return Response(projects)
