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
from backend.components.base import ComponentAuth
from backend.components.paas_cc import PaaSCCClient
from backend.container_service.clusters.base.utils import get_cluster_type
from backend.dashboard.utils.resp import ListApiRespBuilder
from backend.iam.permissions.resources.cluster import ClusterPermCtx, ClusterPermission
from backend.resources.namespace.client import Namespace
from backend.utils.error_codes import error_codes


class NamespaceViewSet(SystemViewSet):
    """Namespace 相关接口"""

    def list(self, request, project_id, cluster_id):
        # TODO 优化实现(序列化中校验)
        cc_client = PaaSCCClient(auth=ComponentAuth(request.user.token.access_token))
        resp = cc_client.get_cluster(project_id, cluster_id)
        if resp['result'] is False:
            raise error_codes.APIError((f"获取集群信息失败，错误信息：{resp['message']}"))

        client = Namespace(request.ctx_cluster)
        cluster_perm_ctx = ClusterPermCtx(
            username=request.user.username,
            project_id=project_id,
            cluster_id=cluster_id,
        )
        ClusterPermission().can_view(cluster_perm_ctx)

        response_data = ListApiRespBuilder(
            client, cluster_type=get_cluster_type(cluster_id), project_code=request.project.english_name
        ).build()
        return Response(response_data)
