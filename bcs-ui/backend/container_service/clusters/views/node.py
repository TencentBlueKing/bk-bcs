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
import copy
import json
import logging
import re
from typing import Dict, List

from django.utils.translation import ugettext_lazy as _
from rest_framework import viewsets
from rest_framework.exceptions import ValidationError
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.accounts import bcs_perm
from backend.bcs_web.audit_log import client
from backend.bcs_web.viewsets import SystemViewSet
from backend.components import data as data_api
from backend.components import paas_cc
from backend.components.bcs import k8s
from backend.container_service.clusters import constants
from backend.container_service.clusters import serializers as slzs
from backend.container_service.clusters import utils as cluster_utils
from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.base.utils import get_cluster_nodes
from backend.container_service.clusters.driver.k8s import K8SDriver
from backend.container_service.clusters.models import CommonStatus, NodeLabel, NodeStatus, NodeUpdateLog
from backend.container_service.clusters.module_apis import get_cluster_node_mod, get_gse_mod
from backend.container_service.clusters.tools.node import query_cluster_nodes
from backend.container_service.clusters.utils import cluster_env_transfer, status_transfer
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.paginator import custom_paginator
from backend.utils.renderers import BKAPIRenderer

# 导入相应模块
node = get_cluster_node_mod()
gse = get_gse_mod()

logger = logging.getLogger(__name__)


class NodeBase:
    def can_view_cluster(self, request, project_id, cluster_id):
        """has view cluster perm"""
        cluster_perm = bcs_perm.Cluster(request, project_id, cluster_id)
        cluster_perm.can_view(raise_exception=True)

    def can_edit_cluster(self, request, project_id, cluster_id):
        cluster_perm = bcs_perm.Cluster(request, project_id, cluster_id)
        return cluster_perm.can_edit(raise_exception=True)

    def get_node_list(self, request, project_id, cluster_id):
        """get cluster node list"""
        node_resp = paas_cc.get_node_list(
            request.user.token.access_token,
            project_id,
            cluster_id,
            params={'limit': constants.DEFAULT_NODE_LIMIT},
        )
        if node_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(node_resp.get('message'))
        return node_resp.get('data') or {}

    def get_all_cluster(self, request, project_id):
        resp = paas_cc.get_all_clusters(request.user.token.access_token, project_id)
        if (resp.get('code') != ErrorCode.NoError) or (not resp.get('data')):
            raise error_codes.APIError('search cluster error')
        return resp.get('data') or {}

    def get_cluster_env(self, request, project_id):
        """get cluster env map"""
        data = self.get_all_cluster(request, project_id)
        results = data.get('results') or []
        return {
            info['cluster_id']: cluster_env_transfer(info['environment']) for info in results if info.get('cluster_id')
        }

    def get_node_by_id(self, request, project_id, cluster_id, node_id):
        """get node info by node id"""
        resp = paas_cc.get_node(request.user.token.access_token, project_id, node_id, cluster_id=cluster_id)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(resp.get('message'))
        return resp.get('data') or {}

    def get_project_cluster(self, request, project_id):
        """get cluster info"""
        data = self.get_all_cluster(request, project_id)
        results = data.get('results') or []
        return {info['cluster_id']: info['name'] for info in results}

    def update_nodes_in_cluster(self, request, project_id, cluster_id, node_ips, status):
        """update nodes with same cluster"""
        data = [{'inner_ip': ip, 'status': status} for ip in node_ips]
        resp = paas_cc.update_node_list(request.user.token.access_token, project_id, cluster_id, data=data)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(resp.get('message'))
        return resp.get('data') or []

    def get_cluster(self, request, project_id, cluster_id):
        cluster_resp = paas_cc.get_cluster(request.user.token.access_token, project_id, cluster_id)
        if cluster_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError.f(cluster_resp.get('message'))
        return cluster_resp.get('data') or {}


class NodeHandler:
    def filter_node(self, data, filter_ip, filter_key='InnerIP'):
        """filter rule:
        - single ip: fuzzy search
        - other: precise search
        """
        if not filter_ip:
            return data
        if isinstance(filter_ip, str):
            filter_ip = filter_ip.split(',')
        filter_data = []
        # fuzzy search
        if len(filter_ip) == 1:
            filter_data = [info for info in data if filter_ip[0].strip() in info[filter_key]]
        else:
            for info in data:
                if info[filter_key] in filter_ip:
                    filter_data.append(info)
                if len(filter_data) == len(filter_ip):
                    break
        return filter_data

    def clean_node(self, data):
        """remove the specific status node item
        Note: remove the node of the 'removed' status
        """
        return [info for info in data if info.get('status') not in [NodeStatus.Removed]]

    def get_order_by(self, request, project_id, data, ordering):
        if not (ordering and data['results']):
            return data
        # reverse order
        node_ip_list = [i['inner_ip'] for i in data['results']]
        cc_app_id = request.project['cc_app_id']
        # split the asc or desc
        if ordering.startswith('-'):
            metric, reverse = ordering[1:], False
        else:
            metric, reverse = ordering, True

        result = data_api.get_node_metrics_order(metric, cc_app_id, node_ip_list).get('list') or []
        # metric sort
        if metric == 'mem':
            result = sorted(result, key=lambda x: x['used'] / x['total'], reverse=reverse)
        elif metric == 'cpu_summary':
            result = sorted(result, key=lambda x: x['usage'], reverse=reverse)
        elif metric == 'disk':
            result = sorted(result, key=lambda x: x['in_use'], reverse=reverse)
        order_by = [i['ip'] for i in result]

        order_by = order_by[::-1] if reverse else order_by

        def index(ip):
            try:
                return order_by.index(ip)
            except ValueError:
                return 0

        data['results'] = sorted(data['results'], key=lambda x: (index(x['inner_ip']), x['inner_ip']), reverse=True)
        return data


class NodeLabelBase:
    def get_labels_by_node(self, request, project_id, node_id_list):
        node_label_info = NodeLabel.objects.filter(node_id__in=node_id_list, project_id=project_id, is_deleted=False)
        return node_label_info.values("id", "project_id", "node_id", "cluster_id", "labels")

    def delete_node_label(self, request, node_id):
        """set node label deleted"""
        try:
            cluster_utils.delete_node_labels_record(NodeLabel, [node_id], request.user.username)
        except Exception as err:
            logger.error('delete node label error, %s', err)


class NodeCreateListViewSet(NodeBase, NodeHandler, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_data(self, request):
        slz = slzs.ListNodeSLZ(data=request.GET)
        slz.is_valid(raise_exception=True)
        return dict(slz.validated_data)

    def get_post_data(self, request):
        slz = slzs.ListNodeSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        return dict(slz.validated_data)

    def add_container_count(self, request, project_id, cluster_id, node_list):
        host_ip_list = [info['inner_ip'] for info in node_list]
        try:
            driver = K8SDriver(request, project_id, cluster_id)
            host_container_map = driver.get_host_container_count(host_ip_list)
        except Exception as e:
            logger.exception(f"通过BCS API查询主机container数量异常, 详情: {e}")
            host_container_map = {}
        for info in node_list:
            info["containers"] = host_container_map.get(info["inner_ip"], 0)
        return node_list

    def compose_data_with_containers(self, request, project_id, cluster_id, with_containers, data):
        if not (with_containers and data):
            return data
        # add container count
        return self.add_container_count(request, project_id, cluster_id, data)

    def add_env_perm(self, request, project_id, cluster_id, data, cluster_env_info):
        nodes_results = bcs_perm.Cluster.hook_perms(request, project_id, [{'cluster_id': cluster_id}])
        for info in data.get('results') or []:
            info['permissions'] = nodes_results[0]['permissions']
            info['cluster_env'] = cluster_env_info.get(cluster_id, '')

    def get_create_node_perm(self, request, project_id, cluster_id):
        perm_client = bcs_perm.Cluster(request, project_id, cluster_id)
        return perm_client.can_edit(raise_exception=False)

    def filter_node_with_labels(self, cluster_id, data, filter_label_list):
        """filter node list by node labels
        filter_label_list format: [{'a': '1'}, {'a': '2'}, {'b': '1'}]
        """
        if not filter_label_list:
            return data
        node_id_info_map = {info['id']: info for info in data}
        node_labels = NodeLabel.objects.filter(cluster_id=cluster_id, is_deleted=False)
        filter_data = []
        for info in node_labels:
            labels = info.node_labels
            for filter_label in filter_label_list:
                key = list(filter_label)[-1]
                if key in labels and labels[key] == filter_label[key] and info.node_id in node_id_info_map:
                    filter_data.append(node_id_info_map[info.node_id])
                    break
        return filter_data

    def filter_nodes_by_status(self, node_list, status_list):
        if not status_list:
            return node_list
        return [node for node in node_list if node["status"] in status_list]

    def data_handler_for_nodes(self, request, project_id, cluster_id, data):
        self.can_view_cluster(request, project_id, cluster_id)
        node_list = self.get_node_list(request, project_id, cluster_id)
        # filter by request ip
        node_list = self.filter_node(node_list.get('results') or [], data.get('ip'), filter_key="inner_ip")
        node_list = self.filter_node_with_labels(cluster_id, node_list, data.get('labels'))
        # 通过节点状态过滤节点
        node_list = self.filter_nodes_by_status(node_list, data["status_list"])
        node_list = self.clean_node(node_list)
        # pagination for node list
        ip_offset = data.pop('offset', 0)
        ip_limit = data.pop('limit', constants.DEFAULT_PAGE_LIMIT)
        pagination_data = custom_paginator(node_list, limit=ip_limit, offset=ip_offset)
        # add
        pagination_data['results'] = self.compose_data_with_containers(
            request, project_id, cluster_id, data.get('with_containers'), pagination_data['results']
        )
        # order the node list
        ordering = data.get('ordering')
        if ordering:
            pagination_data = self.get_order_by(request, project_id, pagination_data, ordering)

        cluster_env_info = self.get_cluster_env(request, project_id)
        # add perm
        self.add_env_perm(request, project_id, cluster_id, pagination_data, cluster_env_info)

        has_create_perm = self.get_create_node_perm(request, project_id, cluster_id)
        return {'code': ErrorCode.NoError, 'data': pagination_data, 'permissions': {'create': has_create_perm}}

    def post_node_list(self, request, project_id, cluster_id):
        """post request for node list"""
        data = self.get_post_data(request)
        node_list_with_perm = self.data_handler_for_nodes(request, project_id, cluster_id, data)
        return Response(node_list_with_perm)

    def list(self, request, project_id, cluster_id):
        """get node list
        note: pagination by backend
        """
        # get request data
        data = self.get_data(request)
        node_list_with_perm = self.data_handler_for_nodes(request, project_id, cluster_id, data)
        return Response(node_list_with_perm)

    def create(self, request, project_id, cluster_id):
        node_client = node.CreateNode(request, project_id, cluster_id)
        return node_client.create()

    def list_nodes_ip(self, request, project_id, cluster_id):
        """获取集群下节点的IP"""
        nodes = get_cluster_nodes(request.user.token.access_token, project_id, cluster_id)
        return Response([info["inner_ip"] for info in nodes])


class NodeGetUpdateDeleteViewSet(NodeBase, NodeLabelBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_request_params(self, request):
        slz = slzs.UpdateNodeSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        return slz.validated_data

    def node_handler(self, request, project_id, cluster_id, node_info):
        driver = K8SDriver(request, project_id, cluster_id)
        if node_info['status'] == constants.ClusterManagerNodeStatus.REMOVABLE:
            driver.disable_node(node_info['inner_ip'])
        elif node_info['status'] == constants.ClusterManagerNodeStatus.RUNNING:
            driver.enable_node(node_info['inner_ip'])
        else:
            raise error_codes.CheckFailed(f'node of the {node_info["status"]} does not allow operation')

    def update(self, request, project_id, cluster_id, inner_ip):
        self.can_edit_cluster(request, project_id, cluster_id)
        # get params
        params = self.get_request_params(request)
        # 记录node的操作，这里包含disable: 停止调度，enable: 允许调度
        # 根据状态进行判断，当前端传递的是normal时，是要允许调度，否则是停止调度
        node_info = {"inner_ip": inner_ip, "status": params["status"]}
        operate = "enable" if node_info["status"] == constants.ClusterManagerNodeStatus.RUNNING else "disable"
        log_desc = (
            f'project: {request.project.project_name}, cluster: {cluster_id}, {operate} node: {node_info["inner_ip"]}'
        )
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type='node',
            resource=node_info['inner_ip'],
            description=log_desc,
        ).log_modify():
            self.node_handler(request, project_id, cluster_id, node_info)
        return Response()


class FailedNodeDeleteViewSet(NodeBase, NodeLabelBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def delete(self, request, project_id, cluster_id, node_id):
        """Delete failed node"""
        self.delete_node_label(request, node_id)
        node_client = node.DeleteNode(request, project_id, cluster_id, node_id)
        return node_client.force_delete()


class NodeUpdateLogView(NodeBase, viewsets.ModelViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)
    serializer_class = slzs.NodeInstallLogSLZ
    queryset = NodeUpdateLog.objects.all()

    def get_queryset(self, project_id, cluster_id, node_id):
        return (
            super()
            .get_queryset()
            .filter(project_id=project_id, cluster_id=cluster_id, node_id__icontains='[%s]' % node_id)
            .order_by('-create_at')
        )

    def get_display_status(self, curr_status):
        return status_transfer(curr_status, constants.NODE_RUNNING_STATUS, constants.NODE_FAILED_STATUS)

    def get_node_ip(self, access_token, project_id, cluster_id, node_id):
        resp = paas_cc.get_node(access_token, project_id, node_id, cluster_id=cluster_id)
        if resp.get("code") != ErrorCode.NoError:
            logger.error("request paas cc node api error, %s", resp.get("message"))
            return None
        return resp.get("data", {}).get("inner_ip")

    def get_log_data(self, request, logs, project_id, cluster_id, node_id):
        if not logs:
            return {'status': 'none'}
        latest_log = logs[0]
        status = self.get_display_status(latest_log.status)
        data = {
            'project_id': project_id,
            'cluster_id': cluster_id,
            'status': status,
            'log': [],
            "task_url": latest_log.log_params.get("task_url") or "",
            "error_msg_list": [],
        }
        for info in logs:
            info.status = self.get_display_status(info.status)
            slz = slzs.NodeInstallLogSLZ(instance=info)
            data['log'].append(slz.data)

        return data

    def get(self, request, project_id, cluster_id, node_id):
        self.can_view_cluster(request, project_id, cluster_id)
        # get log
        logs = self.get_queryset(project_id, cluster_id, node_id)
        data = self.get_log_data(request, logs, project_id, cluster_id, node_id)
        return Response(data)


class NodeContainers(NodeBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_params(self, request, project_id, cluster_id):
        slz = slzs.NodeSLZ(
            data=request.GET, context={'request': request, 'project_id': project_id, 'cluster_id': cluster_id}
        )
        slz.is_valid(raise_exception=True)
        return slz.data

    def list(self, request, project_id, cluster_id):
        """获取节点下的容器列表"""
        self.can_view_cluster(request, project_id, cluster_id)
        # get params
        params = self.get_params(request, project_id, cluster_id)
        # get containers
        driver = K8SDriver(request, project_id, cluster_id)
        containers = driver.flatten_container_info(params['res_id'])

        return Response(containers)


class NodeLabelQueryCreateViewSet(NodeBase, NodeLabelBase, viewsets.ViewSet):
    def label_key_handler(self, pre_labels, curr_labels):
        """处理label的key"""
        ret_data = {}
        pre_label_keys = pre_labels.keys()
        curr_keys = curr_labels.keys()
        same_keys = list(set(pre_label_keys) & set(curr_keys))
        diff_keys = list(set(pre_label_keys) ^ set(curr_keys))
        # 相同的key，如果value不一样，也设置为mix value
        for key in same_keys:
            if pre_labels[key] != curr_labels[key]:
                ret_data[key] = constants.DEFAULT_MIX_VALUE
            else:
                ret_data[key] = pre_labels[key]
        # 不同的key，都设置为mix value
        for key in diff_keys:
            ret_data[key] = constants.DEFAULT_MIX_VALUE
        return ret_data

    def label_syntax(self, node_labels, exist_node_without_label=False):
        """处理节点标签
        如果为mix value，则设置为*****-----$$$$$
        """
        ret_data = {}
        for info in node_labels:
            labels = json.loads(info.get("labels") or "{}")
            if not labels:
                continue
            if exist_node_without_label:
                ret_data.update(labels)
            else:
                if not ret_data:
                    ret_data.update(labels)
                else:
                    ret_data = self.label_key_handler(ret_data, labels)

        if exist_node_without_label:
            ret_data = {key: constants.DEFAULT_MIX_VALUE for key in ret_data}

        return ret_data

    def get_node_labels(self, request, project_id):
        """获取节点标签"""
        # 获取节点ID
        node_ids = request.GET.get("node_ids")
        cluster_id = request.GET.get("cluster_id")
        if not node_ids:
            raise error_codes.CheckFailed(_("节点信息不存在，请确认后重试!"))
        # 以半角逗号分隔
        node_id_list = [int(node_id) for node_id in node_ids.split(",") if str(node_id).isdigit()]
        # 判断节点属于项目
        all_nodes = self.get_node_list(request, project_id, cluster_id).get('results') or []
        if not all_nodes:
            raise error_codes.APIError(_("当前项目下没有节点!"))
        all_node_id_list = [info["id"] for info in all_nodes]
        diff_node_id_list = set(node_id_list) - set(all_node_id_list)
        if diff_node_id_list:
            return Response(
                {
                    "code": ErrorCode.UserError,
                    "message": _("节点ID [{}] 不属于当前项目，请确认").format(",".join(diff_node_id_list)),
                }
            )

        node_label_list = self.get_labels_by_node(request, project_id, node_id_list)
        # 校验权限
        cluster_id_list = [info["cluster_id"] for info in all_nodes if info["id"] in node_id_list]
        for cluster_id in set(cluster_id_list):
            perm_client = bcs_perm.Cluster(request, project_id, cluster_id)
            perm_client.can_view(raise_exception=True)
        if not node_label_list:
            return Response({"code": ErrorCode.NoError, "data": {}})

        # 如果有多个节点，并且有的节点不存在标签，则全部value为mix value
        exist_node_without_label = False
        if len(node_label_list) != len(node_id_list):
            exist_node_without_label = True
        for info in node_label_list:
            if not info.get("labels"):
                exist_node_without_label = True
        ret_data = self.label_syntax(node_label_list, exist_node_without_label=exist_node_without_label)
        return Response({"code": ErrorCode.NoError, "data": ret_data})

    def get_create_label_params(self, request):
        slz = slzs.NodeLabelParamsSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        node_id_labels = slz.data
        return node_id_labels.get("node_id_list"), node_id_labels.get("node_label_info")

    def label_regex(self, node_label_info):
        """校验label满足正则"""
        prefix_part_regex = re.compile(
            r"^(?=^.{3,253}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$"
        )
        name_part_regex = re.compile(r"^[a-z0-9A-Z][\w.-]{0,61}[a-z0-9A-Z]$|^[a-z0-9A-Z]$")
        val_regex = re.compile(r"^[a-z0-9A-Z][\w.-]{0,61}[a-z0-9A-Z]$|^[a-z0-9A-Z]$")
        if not node_label_info:
            return
        for key, val in node_label_info.items():
            if key in constants.DEFAULT_SYSTEM_LABEL_KEYS:
                raise error_codes.APIError(_("[{}]为系统默认key，禁止使用，请确认").format(key))
            # 针对key的限制
            if key.count("/") == 1:
                split_list = key.split("/")
                if not prefix_part_regex.match(split_list[0]):
                    raise error_codes.APIError(_("键[{}]不符合规范，请参考帮助文档!").format(key))
                if not name_part_regex.match(split_list[-1]):
                    raise error_codes.APIError(_("键[{}]不符合规范，请参考帮助文档!").format(key))
            else:
                if not name_part_regex.match(key):
                    raise error_codes.APIError(_("键[{}]不符合规范，请参考帮助文档!").format(key))
            # 针对val的校验
            if val != constants.DEFAULT_MIX_VALUE and not val_regex.match(val):
                raise error_codes.APIError(_("键[{}]对应的值[{}]不符合规范，请参考帮助文档!").format(key, val))

    def get_label_operation(self, exist_node_labels, post_data, node_id_list, all_node_id_ip_map):
        """获取节点标签，并且和数据库中作对比，识别到添加、删除、更新操作对应的key:value
        format: {
            id: {
                ip: "",
                add: {
                    key: val
                },
                update: {
                    key: val
                },
                delete: {
                    key: val
                },
                existed: {
                },
            }
        }
        """
        label_operation_map = {}
        existed_node_id_list = []
        # 已经存在的记录调整
        for info in exist_node_labels:
            node_id = info["node_id"]
            existed_node_id_list.append(node_id)
            label_operation_map[node_id] = {
                "new": False,
                "cluster_id": all_node_id_ip_map[node_id]["cluster_id"],
                "ip": all_node_id_ip_map[node_id]["inner_ip"],
                "add": {},
                "update": {},
                "delete": {},
                "existed": {},
            }
            labels = json.loads(info["labels"] or "{}")
            if not labels:
                label_operation_map[node_id]["add"] = {
                    key: val for key, val in post_data.items() if val != constants.DEFAULT_MIX_VALUE
                }
            else:
                post_data_copy = copy.deepcopy(post_data)
                for key, val in labels.items():
                    if key not in post_data:
                        label_operation_map[node_id]["delete"][key] = val
                        continue
                    if post_data[key] != constants.DEFAULT_MIX_VALUE:
                        label_operation_map[node_id]["update"][key] = post_data[key]
                    else:
                        label_operation_map[node_id]["existed"][key] = val
                    post_data_copy.pop(key, None)
                label_operation_map[node_id]["add"].update(
                    {key: val for key, val in post_data_copy.items() if val != constants.DEFAULT_MIX_VALUE}
                )
        # 新添加的node调整
        for node_id in set(node_id_list) - set(existed_node_id_list):
            item = {key: val for key, val in post_data.items() if val != constants.DEFAULT_MIX_VALUE}
            if not item:
                continue
            label_operation_map[node_id] = {
                "new": True,
                "cluster_id": all_node_id_ip_map[node_id]["cluster_id"],
                "ip": all_node_id_ip_map[node_id]["inner_ip"],
                "add": item,
                "update": {},
                "delete": {},
                "existed": {},
            }
        return label_operation_map

    def check_perm(self, request, project_id, all_node_id_ip_map, node_id_list):
        # 校验权限
        cluster_id_list = [
            info["cluster_id"] for node_id, info in all_node_id_ip_map.items() if node_id in node_id_list
        ]
        for cluster_id in set(cluster_id_list):
            perm_client = bcs_perm.Cluster(request, project_id, cluster_id)
            perm_client.can_view(raise_exception=True)

    def create_node_label_via_k8s(self, request, project_id, label_operation_map):
        """K8S打Label"""
        for node_id, info in label_operation_map.items():
            client = k8s.K8SClient(request.user.token.access_token, project_id, info["cluster_id"], None)
            online_node_info = client.get_node_detail(info["ip"])
            if online_node_info.get("code") != ErrorCode.NoError:
                raise error_codes.APIError(online_node_info.get("message"))
            online_metadata = (online_node_info.get("data") or {}).get("metadata") or {}
            online_labels = online_metadata.get("labels") or {}
            online_labels.update(info["add"])
            online_labels.update(info["update"])
            for label_key in info["delete"]:
                online_labels.pop(label_key, None)
            online_labels["$patch"] = "replace"
            # 写入操作
            k8s_resp = client.create_node_labels(info["ip"], online_labels)
            if k8s_resp.get("code") != ErrorCode.NoError:
                raise error_codes.APIError(k8s_resp.get("message"))

    def create_or_update(self, request, project_id, label_operation_map):
        for node_id, info in label_operation_map.items():
            if info["new"]:
                # 创建之前先检查是否有删除的，然后替换
                node_label_obj = NodeLabel.objects.filter(node_id=node_id)
                if node_label_obj.exists():
                    node_label_obj.update(
                        creator=request.user.username,
                        project_id=project_id,
                        cluster_id=info["cluster_id"],
                        labels=json.dumps(info["add"]),
                        is_deleted=False,
                    )
                else:
                    NodeLabel.objects.create(
                        creator=request.user.username,
                        project_id=project_id,
                        cluster_id=info["cluster_id"],
                        node_id=node_id,
                        labels=json.dumps(info["add"]),
                    )
            else:
                node_label_info = NodeLabel.objects.get(node_id=node_id, is_deleted=False)
                existed_labels = json.loads(node_label_info.labels or "{}")
                existed_labels.update(info["add"])
                existed_labels.update(info["update"])
                for key in info["delete"]:
                    existed_labels.pop(key, None)
                node_label_info.updator = request.user.username
                node_label_info.labels = json.dumps(existed_labels)
                node_label_info.save()

    def create_node_labels(self, request, project_id):
        """添加节点标签"""
        # 解析参数
        node_id_list, node_label_info = self.get_create_label_params(request)
        # 校验label中key和value
        self.label_regex(node_label_info)
        # 获取数据库中节点的label
        # NOTE: 节点为正常状态时，才允许设置标签
        project_node_info = self.get_node_list(request, project_id, None).get('results') or []
        if not project_node_info:
            raise error_codes.APIError(_("当前项目下节点为空，请确认"))
        all_node_id_list = []
        all_node_id_ip_map = {}
        for info in project_node_info:
            all_node_id_list.append(info["id"])
            all_node_id_ip_map[info["id"]] = {"inner_ip": info["inner_ip"], "cluster_id": info["cluster_id"]}
            if info['id'] in node_id_list and info['status'] != CommonStatus.Normal:
                raise error_codes.CheckFailed(_("节点不是正常状态时，不允许设置标签"))
        diff_node_id_list = set(node_id_list) - set(all_node_id_list)
        if diff_node_id_list:
            raise error_codes.CheckFailed(_("节点ID [{}] 不属于当前项目，请确认").format(",".join(diff_node_id_list)))
        # 校验权限
        self.check_perm(request, project_id, all_node_id_ip_map, node_id_list)
        # 匹配数据
        pre_node_labels = self.get_labels_by_node(request, project_id, node_id_list)
        label_operation_map = self.get_label_operation(
            pre_node_labels, node_label_info, node_id_list, all_node_id_ip_map
        )
        # k8s 是以节点为维度
        self.create_node_label_via_k8s(request, project_id, label_operation_map)
        # 写入数据库
        self.create_or_update(request, project_id, label_operation_map)

        client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="node",
            resource=str(node_id_list),
            resource_id=str(node_id_list),
            extra=json.dumps(node_label_info),
            description=_("节点打标签"),
        ).log_add(activity_status="succeed")
        return Response({"code": 0, "message": _("创建成功!")})


class NodeLabelListViewSet(NodeBase, NodeLabelBase, SystemViewSet):
    def compose_nodes(
        self, node_id_info: Dict, label_info: List, project_code: str, cluster_name_env: Dict, nodes: Dict
    ) -> List:
        # map for node id and node label
        label_info_dict = {info['node_id']: info for info in label_info}
        node_info_with_label = []
        # compose the node info
        for node_id, info in node_id_info.items():
            info['labels'] = []
            info['project_code'] = project_code
            info.update(cluster_name_env.get(info['cluster_id']) or {})
            label_info = label_info_dict.get(node_id)
            if label_info:
                label_slz = json.loads(label_info.get('labels') or '{}')
                label_list = [{key: val} for key, val in label_slz.items()]
                info['labels'] = label_list

            # 添加集群 host name和污点信息
            node_info = nodes.get(info["inner_ip"], {})
            info["host_name"] = node_info.get("host_name", "")
            info["taints"] = node_info.get("taints", {})

            node_info_with_label.append(info)
        return node_info_with_label

    def exclude_removed_status_node(self, data):
        node_id_info_map = {info['id']: info for info in data if info['status'] not in [NodeStatus.Removed]}
        return node_id_info_map

    def get_cluster_id(self, request):
        cluster_id = request.query_params.get('cluster_id')
        return None if cluster_id in ['all', None] else cluster_id

    def get_cluster_id_info_map(self, request, project_id):
        """get cluster info map
        format: {'cluster_id': {'cluster_name': xxx, 'cluster_env': xxx}}
        """
        data = self.get_all_cluster(request, project_id)
        results = data.get('results') or []
        return {
            info['cluster_id']: {
                'cluster_env': cluster_env_transfer(info['environment']),
                'cluster_name': info['name'],
            }
            for info in results
            if info.get('cluster_id')
        }

    def list(self, request, project_id):
        # get cluster id by request
        cluster_id = self.get_cluster_id(request)
        # get node info
        node_list = self.get_node_list(request, project_id, cluster_id)
        node_list = node_list.get('results') or []
        if not node_list:
            return Response({'code': 0, 'result': []})
        node_id_info_map = self.exclude_removed_status_node(node_list)
        # get node labels
        node_label_list = self.get_labels_by_node(request, project_id, node_id_info_map.keys())
        # render cluster id, cluster name and cluster environment
        cluster_name_env = self.get_cluster_id_info_map(request, project_id)
        # 获取节点的taint
        ctx_cluster = CtxCluster.create(token=request.user.token.access_token, id=cluster_id, project_id=project_id)
        nodes = query_cluster_nodes(ctx_cluster)
        node_list = self.compose_nodes(
            node_id_info_map, node_label_list, request.project['english_name'], cluster_name_env, nodes
        )
        # add perm for node
        nodes_results = bcs_perm.Cluster.hook_perms(request, project_id, node_list)

        return Response({'count': len(node_list), 'results': nodes_results})


class RescheduleNodePods(NodeBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def reschedule_pods_taskgroups(self, request, project_id, cluster_id, inner_ip):
        driver = K8SDriver(request, project_id, cluster_id)
        driver.reschedule_host_pods(inner_ip, raise_exception=False)

    def put(self, request, project_id, cluster_id, inner_ip):
        """重新调度节点上的POD or Taskgroup
        主要目的是由于主机裁撤或者机器故障，需要替换机器
        步骤:
        1. 停止节点调度(前置条件)
        2. 查询节点上的所有pod
        3. 重新调度
        """
        self.can_edit_cluster(request, project_id, cluster_id)
        log_desc = f"project: {request.project.project_name}, cluster: {cluster_id}, node: {inner_ip}, reschedule pods"
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type='node',
            resource=inner_ip,
            description=log_desc,
        ).log_modify():
            # reschedule the pod or taskgroup
            self.reschedule_pods_taskgroups(request, project_id, cluster_id, inner_ip)
        return Response({"code": 0, "message": "task started, please pay attention to the change of container count"})


class NodeForceDeleteViewSet(NodeBase, NodeLabelBase, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def delete(self, request, project_id, cluster_id, node_id):
        self.delete_node_label(request, node_id)
        node_client = node.DeleteNode(request, project_id, cluster_id, node_id)
        return node_client.force_delete()

    def delete_oper(self, request, project_id, cluster_id, node_id):
        """强制删除节点
        1. 判断是否已经停用，如果没有停用进行停止调度操作
        2. 如果有pod/taskgroup，删除上面的pod/taskgroup
        3. 调用移除节点
        """
        self.delete_node_label(request, node_id)
        node_client = node.DeleteNode(request, project_id, cluster_id, node_id)
        return node_client.force_delete()


class BatchUpdateDeleteNodeViewSet(NodeGetUpdateDeleteViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_request_params(self, request):
        slz = slzs.BatchUpdateNodesSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        return slz.validated_data

    def node_list_handler(self, request, project_id, cluster_id, node_list):
        for info in node_list:
            self.node_handler(request, project_id, cluster_id, info)

    def update_nodes_status(self, request, project_id, cluster_id, node_list, ip_list):
        driver = K8SDriver(request, project_id, cluster_id)
        node_container_data = driver.get_host_container_count(ip_list)
        update_data = []
        for info in node_list:
            curr_node_container_count = node_container_data.get(info['inner_ip']) or 0
            if curr_node_container_count == 0 and info['status'] == NodeStatus.ToRemoved:
                info['status'] = NodeStatus.Removable
            update_data.append({'inner_ip': info['inner_ip'], 'status': info['status']})
        resp = paas_cc.update_node_list(request.user.token.access_token, project_id, cluster_id, data=update_data)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(resp.get('message'))
        return resp.get('data') or []

    def batch_update_nodes(self, request, project_id, cluster_id):
        self.can_edit_cluster(request, project_id, cluster_id)
        params = self.get_request_params(request)
        inner_ip_list = params["inner_ip_list"]
        # 组装参数
        node_list = [{"inner_ip": inner_ip, "status": params["status"]} for inner_ip in inner_ip_list]
        # 记录node的操作，这里包含disable: 停止调度，enable: 允许调度
        # 根据状态进行判断，当前端传递的是normal时，是要允许调度，否则是停止调度
        operate = "enable" if params["status"] == constants.ClusterManagerNodeStatus.RUNNING else "disable"
        log_desc = f'project: {request.project.project_name}, cluster: {cluster_id}, {operate} node: {inner_ip_list}'
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type='node',
            resource=inner_ip_list[: constants.IP_LIST_RESERVED_LENGTH],
            description=log_desc,
        ).log_modify():
            self.node_list_handler(request, project_id, cluster_id, node_list)

        return Response()
