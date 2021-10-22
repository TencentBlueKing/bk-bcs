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
from django.utils.translation import ugettext_lazy as _
from rest_framework import response, viewsets
from rest_framework.renderers import BrowsableAPIRenderer

from backend.bcs_web.audit_log import client
from backend.container_service.clusters import serializers as node_serializers
from backend.container_service.clusters.models import CommonStatus
from backend.container_service.clusters.module_apis import get_cluster_node_mod
from backend.utils.renderers import BKAPIRenderer

from .base import ClusterPerm, Nodes

node = get_cluster_node_mod()


class DeleteNodeRecordViewSet(Nodes, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def delete(self, request, project_id, cluster_id, node_id):
        access_token = request.user.token.access_token
        node_info = self.get_node_by_id(access_token, project_id, cluster_id, node_id)
        # set current node status is removed
        self.update_nodes_in_cluster(
            access_token, project_id, cluster_id, [node_info['inner_ip']], CommonStatus.Removed
        )

        return response.Response()


class BatchReinstallNodes(ClusterPerm, Nodes, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_cluster_nodes(self, request, project_id, cluster_id):
        node_list = self.get_node_list(request, project_id, cluster_id)
        return {info['id']: info for info in node_list}

    def reinstall_nodes(self, request, project_id, cluster_id):
        """当初始化失败时，允许用户批量重装
        1. 检测节点必须为当前项目下的同一集群
        2. 检测节点状态必须为初始化失败状态
        3. 下发配置，并更改节点状态
        """
        # 校验集群的编辑权限
        self.can_edit_cluster(request, project_id, cluster_id)
        # 获取集群下的节点
        cluster_nodes = self.get_cluster_nodes(request, project_id, cluster_id)
        # 获取请求参数
        slz = node_serializers.BatchReinstallNodesSLZ(data=request.data, context={'cluster_nodes': cluster_nodes})
        slz.is_valid(raise_exception=True)
        req_data = slz.validated_data
        # 获取集群信息
        cluster_info = self.get_cluster_info(request, project_id, cluster_id)
        # 获取节点IP
        node_id_ip_map = {id: cluster_nodes[id]['inner_ip'] for id in req_data['node_id_list']}
        # 下发继续初始化流程
        node_ip_list = node_id_ip_map.values()
        node_ips = ','.join(node_ip_list)
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type='node',
            resource=node_ips[:512],
            resource_id=','.join([str(id) for id in node_id_ip_map.keys()])[:256],
            description=_("集群: {}, batch reinstall nodes: {}").format(cluster_id, node_ips),
        ).log_modify():
            # 更改节点状态为初始化中
            self.update_nodes_in_cluster(
                request.user.token.access_token, project_id, cluster_id, node_ip_list, CommonStatus.Initializing
            )
            # 下发流程，触发重试任务
            node_client = node.BatchReinstallNodes(request, project_id, cluster_info, node_id_ip_map)
            node_client.reinstall()

        return response.Response()
