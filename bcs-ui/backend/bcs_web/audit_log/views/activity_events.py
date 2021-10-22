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

from rest_framework import views
from rest_framework.renderers import BrowsableAPIRenderer

from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import BKAPIResponse

from .. import serializers
from ..utils import get_event, get_project_clusters

logger = logging.getLogger(__name__)


class ActivityEventView(views.APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get(self, request, project_id):
        serializer = serializers.EventSLZ(data=request.GET)
        serializer.is_valid(raise_exception=True)

        validated_data = serializer.validated_data
        access_token = request.user.token.access_token

        empty_data = {'count': 0, 'results': []}

        clusters = get_project_clusters(access_token, project_id)
        if not clusters.data:
            return BKAPIResponse(empty_data)

        cluster_id = validated_data.get('cluster_id')
        if cluster_id and cluster_id not in clusters.data:
            return BKAPIResponse(empty_data)

        if not cluster_id:
            cluster_id = ",".join(clusters.data.keys())
            validated_data['cluster_id'] = cluster_id

        results, count = get_event(access_token, project_id, validated_data, clusters.data)

        data = {
            "count": count,
            "next": None,
            "previous": None,
            "results": results,
        }

        return BKAPIResponse(data)
