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

from .log_link import get_log_links
from .manager import CollectConfManager
from .serializers import QueryLogLinksSLZ, UpdateOrCreateCollectConfSLZ


class LogCollectViewSet(SystemViewSet):
    def retrieve(self, request, project_id, cluster_id, pk):
        manager = CollectConfManager(request.user.token.access_token, project_id, cluster_id)
        return Response(manager.get_config(config_id=pk))

    def create(self, request, project_id, cluster_id):
        """创建日志采集规则"""
        validate_data = self.params_validate(
            UpdateOrCreateCollectConfSLZ, project_id=project_id, cluster_id=cluster_id
        )
        manager = CollectConfManager(request.user.token.access_token, project_id, cluster_id)
        manager.create_config(request.user.username, config=validate_data)
        return Response()

    def list(self, request, project_id, cluster_id):
        """查询日志采集规则"""
        manager = CollectConfManager(request.user.token.access_token, project_id, cluster_id)
        return Response(manager.list_configs())

    def update(self, request, project_id, cluster_id, pk):
        """更新日志采集规则"""
        validate_data = self.params_validate(
            UpdateOrCreateCollectConfSLZ, project_id=project_id, cluster_id=cluster_id
        )
        manager = CollectConfManager(request.user.token.access_token, project_id, cluster_id)
        manager.update_config(request.user.username, config_id=pk, config=validate_data)
        return Response()

    def destroy(self, request, project_id, cluster_id, pk):
        """删除日志采集规则"""
        manager = CollectConfManager(request.user.token.access_token, project_id, cluster_id)
        manager.delete_config(request.user.username, config_id=pk)
        return Response()


class LogLinksViewSet(SystemViewSet):
    def get_log_links(self, request, project_id):
        params = self.params_validate(QueryLogLinksSLZ)
        log_links = get_log_links(project_id, params['bk_biz_id'], container_ids=params.get('container_ids'))
        return Response(log_links)
