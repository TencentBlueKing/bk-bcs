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
from backend.container_service.clusters.mgr.cluster.master import BcsClusterMaster


class ClusterPerm:
    def can_view(self, request, project_id, cluster_id):
        perm = Cluster(request, project_id, cluster_id)
        perm.can_view(raise_exception=True)


class ClusterMastersViewSet(SystemViewSet, ClusterPerm):
    def get(self, request, project_id, cluster_id):
        """获取集群的master节点数据"""
        # 需要集群的查看权限
        self.can_view(request, project_id, cluster_id)
        # 获取master详情
        masters = BcsClusterMaster(ctx_cluster=request.ctx_cluster, biz_id=request.project.cc_app_id).get_masters()
        return Response(masters)
