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

from backend.bcs_web.apis.views import BaseAPIViewSet
from backend.container_service.clusters.base.utils import get_clusters
from backend.resources.utils.dynamic.discovery import DiscovererCache


class ClusterViewSet(BaseAPIViewSet):
    def list(self, request, project_id_or_code):
        clusters = get_clusters(request.user.token.access_token, request.project.project_id)
        return Response(clusters)


class ClusterDiscovererCacheViewSet(BaseAPIViewSet):
    def invalidate(self, request, project_id_or_code, cluster_id):
        """主动使集群缓存信息失效"""
        # 缓存集群信息的KEY
        cluster_cache_key = f"osrcp-{cluster_id}.json"
        DiscovererCache(cluster_cache_key).invalidate()
        return Response()
