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

from .manager import ToolManager
from .models import InstalledTool, Tool
from .serializers import ClusterToolSLZ, InstalledToolSLZ, UpgradeToolSLZ


class ToolsViewSet(SystemViewSet):
    """组件库"""

    def list(self, request, project_id, cluster_id):
        """查询集群中可用的组件"""
        serializer = ClusterToolSLZ(
            Tool.objects.all(), many=True, context={'project_id': project_id, 'cluster_id': cluster_id}
        )
        return Response(serializer.data)

    def retrieve(self, request, project_id, cluster_id, tool_id):
        """获取组件安装详情"""
        try:
            itool = InstalledTool.objects.get(tool__id=tool_id, project_id=project_id, cluster_id=cluster_id)
        except InstalledTool.DoesNotExist:
            return Response({})

        serializer = InstalledToolSLZ(itool)
        return Response(serializer.data)

    def install(self, request, project_id, cluster_id, tool_id):
        """安装组件"""
        manager = ToolManager(project_id, cluster_id, tool_id)
        manager.install(request.user, values=request.data.get('values'))
        return Response()

    def upgrade(self, request, project_id, cluster_id, tool_id):
        """更新组件"""
        params = self.params_validate(UpgradeToolSLZ)
        manager = ToolManager(project_id, cluster_id, tool_id)
        manager.upgrade(request.user, params['chart_url'], params['values'])
        return Response()

    def uninstall(self, request, project_id, cluster_id, tool_id):
        """卸载组件"""
        manager = ToolManager(project_id, cluster_id, tool_id)
        manager.uninstall(request.user)
        return Response()
