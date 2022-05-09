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
from typing import List

from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.components.bcs.k8s import K8SClient
from backend.container_service.clusters.base.utils import get_cluster_type, get_shared_cluster_proj_namespaces
from backend.container_service.clusters.constants import ClusterType
from backend.container_service.observability.metric.constants import (
    INNER_USE_LABEL_PREFIX,
    INNER_USE_SERVICE_METADATA_FIELDS,
)

logger = logging.getLogger(__name__)


class ServiceViewSet(SystemViewSet):
    """Metric Service 相关接口"""

    def list(self, request, project_id, cluster_id):
        """获取可选 Service 列表"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, env=None)
        resp = client.get_service({'env': 'k8s'})
        services = self._slim_down_service(resp.get('data') or [])

        # 共享集群需要再过滤下属于当前项目的命名空间
        if get_cluster_type(cluster_id) == ClusterType.SHARED:
            project_namespaces = get_shared_cluster_proj_namespaces(request.ctx_cluster, request.project.english_name)
            services = [svc for svc in services if svc['namespace'] in project_namespaces]

        return Response(services)

    def _slim_down_service(self, service_list: List) -> List:
        """
        去除冗余的 Service 信息

        :param service_list: Service 列表
        :return: 去除冗余信息后的 Service 列表
        """
        for service in service_list:
            service['data']['metadata'] = {
                k: v for k, v in service['data']['metadata'].items() if k not in INNER_USE_SERVICE_METADATA_FIELDS
            }
            labels = service['data']['metadata'].get('labels')
            if not labels:
                continue
            service['data']['metadata']['labels'] = dict(
                sorted([(k, v) for k, v in labels.items() if not self._is_inner_use_label(k)])
            )
        return service_list

    def _is_inner_use_label(self, label_key: str) -> bool:
        """
        判断 Label 是否为内部使用的（不展示给前端）

        :param label_key: Label 键名
        :return: True / False
        """
        if label_key in ['io.tencent.bcs.controller.name']:
            return False
        # 若前缀符合，则认为是内部使用的
        for prefix in INNER_USE_LABEL_PREFIX:
            if label_key.startswith(prefix):
                return True
        return False
