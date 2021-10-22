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
import logging

from rest_framework import viewsets
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.components import paas_cc
from backend.utils.errcodes import ErrorCode
from backend.utils.renderers import BKAPIRenderer

logger = logging.getLogger(__name__)


class NamespaceViewSet(viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def list_namespaces(self, request, project_id, cluster_id):
        resp = paas_cc.get_cluster_namespace_list(request.user.token.access_token, project_id, cluster_id)
        if resp.get('code') != ErrorCode.NoError:
            logger.error('get namespaces error, %s', resp.get('message'))
            return Response({'namespaces': []})

        namespaces = resp.get('data', {}).get('results', [])
        return Response({'namespaces': namespaces})
