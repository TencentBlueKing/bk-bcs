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
from rest_framework.response import Response

from backend.bcs_web.audit_log.audit.decorators import log_audit_on_view
from backend.bcs_web.audit_log.constants import ActivityType
from backend.bcs_web.viewsets import SystemViewSet
from backend.container_service.clusters.base.utils import get_cluster
from backend.container_service.clusters.constants import K8S_SKIP_NS_LIST
from backend.container_service.clusters.constants import ClusterManagerNodeStatus as node_status
from backend.container_service.clusters.tools import node, resp
from backend.container_service.clusters.utils import get_nodes_repr
from backend.iam.permissions.resources.cluster import ClusterPermCtx, ClusterPermission
from backend.resources.node.client import Node
from backend.resources.workloads.pod.scheduler import PodsRescheduler

from . import serializers as slz
from .auditors import NodeAuditor


class NodeViewSets(SystemViewSet):
    def list_nodes(self, request, project_id, cluster_id):
        """查询集群下nodes
        NOTE: 限制查询一个集群下的节点
        """
        # 以集群中节点为初始数据，如果bcs cc中节点不在集群中，处于初始化中或者初始化失败，也需要展示
        cluster_nodes = node.query_cluster_nodes(request.ctx_cluster)
        bcs_cc_nodes = node.query_nodes_from_cm(request.ctx_cluster)
        # 组装数据
        cluster = get_cluster(request.user.token.access_token, request.project.project_id, cluster_id)
        client = node.NodesData(bcs_cc_nodes, cluster_nodes, cluster_id, cluster.get("name", ""))
        return Response(client.nodes())

    @log_audit_on_view(NodeAuditor, activity_type=ActivityType.Modify)
    def set_labels(self, request, project_id, cluster_id):
        """设置节点标签"""
        params = self.params_validate(slz.NodeLabelListSLZ)
        node_client = Node(request.ctx_cluster)
        node_client.set_labels_for_multi_nodes(params["node_label_list"])

        # 获取节点名称，用于审计
        node_names = get_nodes_repr([n["node_name"] for n in params["node_label_list"]])
        request.audit_ctx.update_fields(resource=node_names, extra=params, description=_("节点设置标签"))

        return Response()

    @log_audit_on_view(NodeAuditor, activity_type=ActivityType.Modify)
    def set_taints(self, request, project_id, cluster_id):
        """设置污点"""
        params = self.params_validate(slz.NodeTaintListSLZ)
        node_client = Node(request.ctx_cluster)
        node_client.set_taints_for_multi_nodes(params["node_taint_list"])

        # 获取节点名称，用于审计
        node_names = get_nodes_repr([n["node_name"] for n in params["node_taint_list"]])
        request.audit_ctx.update_fields(resource=node_names, extra=params, description=_("节点设置污点"))

        return Response()

    def query_labels(self, request, project_id, cluster_id):
        """查询node的标签

        TODO: 关于labels和taints是否有必要合成一个，通过前端传递参数判断查询类型
        """
        params = self.params_validate(slz.QueryNodeListSLZ)
        builder = resp.NodeRespBuilder(request.ctx_cluster)
        return Response(builder.query_labels(params["node_name_list"]))

    def query_taints(self, request, project_id, cluster_id):
        """查询node的污点"""
        params = self.params_validate(slz.QueryNodeListSLZ)
        node_client = Node(request.ctx_cluster)
        return Response(node_client.filter_nodes_field_data("taints", params["node_name_list"]))

    @log_audit_on_view(NodeAuditor, activity_type=ActivityType.Modify)
    def set_schedule_status(self, request, project_id, cluster_id):
        """设置节点调度状态
        通过传递状态, 设置节点的调度状态
        """
        params = self.params_validate(slz.NodeStatusSLZ)
        # NOTE: 如果状态为REMOVABLE，则期望的是停止调度状态, 以便于进一步操作
        unschedulable = True if params["status"] == node_status.REMOVABLE else False
        client = Node(request.ctx_cluster)
        client.set_nodes_schedule_status(unschedulable, params["node_name_list"])

        request.audit_ctx.update_fields(
            resource=get_nodes_repr(params["node_name_list"]),
            extra=params,
            description=_("节点停止调度") if unschedulable else _("节点允许调度"),
        )

        return Response()

    def detail(self, request, project_id, cluster_id, name):
        """节点详情"""
        # 需要集群的查看权限
        perm_ctx = ClusterPermCtx(username=request.user.username, project_id=project_id, cluster_id=cluster_id)
        ClusterPermission().can_view(perm_ctx)
        return Response(node.NodeDetailQuerier(name, request.ctx_cluster).detail())


class MasterViewSet(SystemViewSet):
    def list(self, request, project_id, cluster_id):
        # 需要集群的查看权限
        # TODO: 后面支持权限中心V3后，使用新的权限校验
        perm_ctx = ClusterPermCtx(username=request.user.username, project_id=project_id, cluster_id=cluster_id)
        ClusterPermission().can_view(perm_ctx)
        # 获取master详情
        masters = node.BcsClusterMaster(
            ctx_cluster=request.ctx_cluster, biz_id=request.project.cc_app_id
        ).list_masters()
        return Response(masters)


class BatchReschedulePodsViewSet(SystemViewSet):
    @log_audit_on_view(NodeAuditor, activity_type=ActivityType.Modify)
    def reschedule(self, request, project_id, cluster_id):
        """批量重新调度节点上的pods"""
        data = self.params_validate(slz.ClusterNodesSLZ)
        PodsRescheduler(request.ctx_cluster).reschedule_by_nodes(data["host_ips"], K8S_SKIP_NS_LIST)

        # NOTE: 这里记录为节点的审计
        request.audit_ctx.update_fields(
            resource=get_nodes_repr(data["host_ips"]), extra=data, description=_("批量 POD 重新调度")
        )

        return Response()
