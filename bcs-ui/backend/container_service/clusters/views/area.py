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

from rest_framework import viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.components import paas_cc
from backend.container_service.clusters.views.utils import get_areas
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer


class AreaListViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def list(self, request, project_id):
        """get the area list"""
        return Response(get_areas(request))


class AreaInfoViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def info(self, request, area_id):
        """get the area info"""
        resp = paas_cc.get_area_info(request.user.token.access_token, area_id)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(f'request bcs cc area info api error, {resp.get("message")}')

        data = resp.get('data') or {}
        data['configuration'] = json.loads(data.pop('configuration', '{}'))

        return Response(data)
