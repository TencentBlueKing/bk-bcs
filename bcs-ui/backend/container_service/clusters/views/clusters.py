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

from backend.bcs_web.viewsets import SystemViewSet
from backend.components.paas_cc import get_all_clusters
from backend.container_service.clusters.utils import cluster_env_transfer
from backend.utils.errcodes import ErrorCode


class ClusterViewSet(SystemViewSet):
    def list(self, request, project_id):
        """获取集群列表"""
        clusters = self.get_cluster_list(request, project_id)
        for info in clusters.get("results") or []:
            info["environment"] = cluster_env_transfer(info["environment"])
        return Response(clusters)

    def get_cluster_list(self, request, project_id):
        resp = get_all_clusters(request.user.token.access_token, project_id, desire_all_data=1)
        if resp.get("code") != ErrorCode.NoError:
            return {}
        return resp.get("data") or {}
