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

from django.conf import settings
from rest_framework import response, viewsets
from rest_framework.renderers import BrowsableAPIRenderer

from backend.accounts.bcs_perm import Cluster
from backend.components import paas_cc
from backend.container_service.clusters.base import utils as cluster_utils
from backend.container_service.clusters.utils import get_cmdb_hosts
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer

logger = logging.getLogger(__name__)


class ClusterPermBase:
    def can_view_cluster(self, request, project_id, cluster_id):
        perm = Cluster(request, project_id, cluster_id)
        perm.can_view(raise_exception=True)


class ClusterMasterInfo(ClusterPermBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_master_ips(self, request, project_id, cluster_id):
        """get master inner ip info"""
        master_resp = paas_cc.get_master_node_list(request.user.token.access_token, project_id, cluster_id)
        if master_resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(master_resp.get("message"))
        data = master_resp.get("data") or {}
        master_ip_info = data.get("results") or []
        return [info["inner_ip"] for info in master_ip_info if info.get("inner_ip")]

    def cluster_masters(self, request, project_id, cluster_id):
        self.can_view_cluster(request, project_id, cluster_id)
        # 获取master
        masters = cluster_utils.get_cluster_masters(request.user.token.access_token, project_id, cluster_id)
        # 返回master对应的主机信息
        # 因为先前
        host_property_filter = {
            "condition": "OR",
            "rules": [
                {"field": "bk_host_innerip", "operator": "equal", "value": info["inner_ip"]} for info in masters
            ],
        }
        username = settings.ADMIN_USERNAME
        cluster_masters = get_cmdb_hosts(username, request.project.cc_app_id, host_property_filter)

        return response.Response(cluster_masters)
