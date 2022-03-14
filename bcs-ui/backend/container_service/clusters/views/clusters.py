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

from backend.accounts.bcs_perm import Cluster
from backend.bcs_web.viewsets import SystemViewSet
from backend.components.paas_cc import get_all_clusters
from backend.container_service.clusters.base import utils as cluster_utils
from backend.container_service.clusters.utils import cluster_env_transfer
from backend.utils.errcodes import ErrorCode


class ClusterViewSet(SystemViewSet):
    def list(self, request, project_id):
        """获取集群列表"""
        cluster_info = self.get_cluster_list(request, project_id)
        cluster_data = cluster_info.get("results") or []
        # add allow delete perm
        for info in cluster_data:
            info["environment"] = cluster_env_transfer(info["environment"])
        perm_can_use = True if request.GET.get("perm_can_use") == "1" else False

        cluster_results = Cluster.hook_perms(request, project_id, cluster_data, filter_use=perm_can_use)
        # add can create cluster perm for prod/test
        can_create_test, can_create_prod = self.get_cluster_create_perm(request, project_id)

        cluster_results = cluster_utils.append_shared_clusters(cluster_results)
        return Response(
            {
                "code": ErrorCode.NoError,
                "data": {"count": len(cluster_results), "results": cluster_results},
                "permissions": {
                    "test": can_create_test,
                    "prod": can_create_prod,
                    "create": can_create_test or can_create_prod,
                },
            }
        )

    def get_cluster_create_perm(self, request, project_id):
        test_cluster_perm = Cluster(request, project_id, "**", resource_type="cluster_test")
        can_create_test = test_cluster_perm.can_create(raise_exception=False)
        prod_cluster_perm = Cluster(request, project_id, "**", resource_type="cluster_prod")
        can_create_prod = prod_cluster_perm.can_create(raise_exception=False)
        return can_create_test, can_create_prod

    def get_cluster_list(self, request, project_id):
        resp = get_all_clusters(request.user.token.access_token, project_id, desire_all_data=1)
        if resp.get("code") != ErrorCode.NoError:
            return {}
        return resp.get("data") or {}
