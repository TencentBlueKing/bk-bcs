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
import json
import logging
from datetime import datetime

from django.utils.translation import ugettext_lazy as _
from rest_framework.exceptions import ValidationError
from rest_framework.response import Response

from backend.accounts.bcs_perm import Cluster
from backend.bcs_web.audit_log import client
from backend.components import ops, paas_cc
from backend.components.bcs import k8s as bcs_k8s
from backend.container_service.clusters import constants, serializers
from backend.container_service.clusters.base import get_cluster
from backend.container_service.clusters.constants import ClusterState
from backend.container_service.clusters.models import CommonStatus, NodeLabel, NodeOperType, NodeStatus, NodeUpdateLog
from backend.container_service.clusters.utils import can_use_hosts
from backend.container_service.projects.base.constants import ProjectKindName
from backend.utils.cache import rd_client
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.ratelimit import RateLimiter
from backend.utils.renderers import BKAPIRenderer

from .configs import k8s

logger = logging.getLogger(__name__)

DEFAULT_NODE_LIMIT = 10000
ACTIVITY_RESOURCE_TYPE = 'node'


class BaseNode(object):
    render_classes = (BKAPIRenderer,)

    def get_cluster_snapshot(self):
        snapshot_info = paas_cc.get_cluster_snapshot(self.access_token, self.project_id, self.cluster_id)
        if snapshot_info.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(snapshot_info.get('message'))
        return snapshot_info.get('data', {})

    def update_nodes(self, node_ips, status=CommonStatus.Initializing):
        if not node_ips:
            return
        data = [{'cluster_id': self.cluster_id, 'inner_ip': ip, 'status': status} for ip in node_ips]
        resp = paas_cc.update_node_with_cluster(self.access_token, self.project_id, data={'updates': data})
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(resp.get('message'))

    def update_cluster_nodes(self, node_ips, status=CommonStatus.Initializing):
        """更新阶段状态，并返回更新后的信息"""
        data = [{'inner_ip': ip, 'status': status} for ip in node_ips]
        resp = paas_cc.update_node_list(self.access_token, self.project_id, self.cluster_id, data=data)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(resp.get('message'))
        return resp.get('data') or []

    def get_node_ip(self):
        node_info = paas_cc.get_node(self.access_token, self.project_id, self.node_id, cluster_id=self.cluster_id)
        if node_info.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(node_info.get('message'))
        return node_info.get('data') or {}

    def get_request_config(self, op_type=constants.OpType.ADD_NODE.value):
        kind_type_map = {'k8s': k8s.NodeConfig}
        snapshot_info = self.get_cluster_snapshot()
        snapshot_config = json.loads(snapshot_info.get('configure', '{}'))
        if snapshot_config.get('common'):
            self.master_ip_list = list(snapshot_config['common']['cluster_masters'].values())
        else:
            self.master_ip_list = snapshot_config.get('master_iplist', '').split(',')
        client = kind_type_map[self.kind_name](snapshot_config, op_type=op_type)
        try:
            resp_config = client.get_request_config(
                self.access_token, self.project_id, self.cluster_id, self.master_ip_list, self.ip_list
            )
            return resp_config
        except Exception as err:
            logger.error('Get node config error, detail: %s', err)
            # 更新下节点状态
            self.update_nodes(self.ip_list, status=CommonStatus.Removed)
            raise error_codes.CheckFailed(_("获取节点初始化配置异常"))

    def save_task_url(self, log, data):
        log_params = log.log_params
        log_params['task_url'] = data.get('task_url') or ''
        log.set_params(log_params)

    def create_node_by_bcs(self, node_info_list, control_ip=None, config=None, websvr=None):  # noqa
        if not config:
            config = self.get_request_config()

        control_ip = config.pop('control_ip', []) or control_ip
        websvr = config.pop('websvr', []) or websvr
        node_info = {i['inner_ip']: '[%s]' % i['id'] for i in node_info_list}
        params = {
            'project_id': self.project_id,
            'cluster_id': self.cluster_id,
            'username': self.username,
            'kind': self.project_info['kind'],
            'kind_name': self.kind_name,
            'need_nat': self.need_nat,
            'control_ip': control_ip,
            'config': config,
            'cc_app_id': self.cc_app_id,
            'node_info': node_info,
            'master_ip_list': self.master_ip_list,
            'module_id_list': self.module_id_list,
            'websvr': websvr,
        }
        log = NodeUpdateLog.objects.create(
            project_id=self.project_id,
            token=self.access_token,
            cluster_id=self.cluster_id,
            operator=self.username,
            params=json.dumps(params),
            oper_type=NodeOperType.NodeInstall,
            node_id=",".join(node_info.values()),
            status=CommonStatus.Initializing,
            is_polling=True,
        )
        task_info = ops.add_cluster_node(
            self.access_token,
            self.project_id,
            self.kind_name,
            self.cluster_id,
            self.master_ip_list,
            self.ip_list,
            config,
            control_ip,
            self.cc_app_id,
            self.username,
            self.module_id_list,
            websvr,
        )
        if task_info.get('code') != ErrorCode.NoError:
            log.set_finish_polling_status(True, False, CommonStatus.InitialFailed)
            # 兼容log
            log.log = json.dumps(constants.BCS_OPS_ERROR_INFO)
            log.save()
            # 更新节点状态
            self.update_nodes(self.ip_list, status=CommonStatus.InitialFailed)
            return log
            # raise error_codes.APIError.f("初始化节点失败，请联系管理员处理!")
        data = task_info.get('data') or {}
        task_id = data.get('task_id')
        if not task_id:
            raise error_codes.APIError(_("获取标准运维任务ID失败，返回任务为{}，请联系管理员处理").format(task_id))
        log.set_task_id(task_id)
        # record the task url by params
        self.save_task_url(log, data)
        return log

    def check_perm(self):
        perm_client = Cluster(self.request, self.project_id, self.cluster_id)
        perm_client.can_edit(raise_exception=True)


class CreateNode(BaseNode):
    def __init__(self, request, project_id, cluster_id):
        self.request = request
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.access_token = request.user.token.access_token
        self.username = request.user.username
        self.bk_token = request.COOKIES.get('bk_token')
        self.project_info = request.project
        self.kind_name = ProjectKindName
        self.cc_app_id = request.project.get('cc_app_id')

    def check_data(self):
        slz = serializers.CreateNodeSLZ(data=self.request.data)
        slz.is_valid(raise_exception=True)
        self.data = slz.validated_data

    def check_node_ip(self):
        project_node_list = [
            info['inner_ip'] for info in self.project_nodes if info['status'] not in [CommonStatus.Removed]
        ]
        intersection = set(project_node_list) & set(self.ip_list)
        if intersection:
            raise error_codes.CheckFailed(_("部分主机已经使用，IP为{}").format(','.join(intersection)))

    def get_node_list(self):
        cluster_node_info = paas_cc.get_node_list(
            self.access_token, self.project_id, self.cluster_id, params={'limit': DEFAULT_NODE_LIMIT}
        )
        if cluster_node_info.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(cluster_node_info.get('message'))
        return cluster_node_info.get('data', {}).get('results', [])

    def get_removed_remained_ips(self):
        removed_ips, remained_ips = [], []
        project_node_list = [
            info['inner_ip'] for info in self.project_nodes if info['status'] in [CommonStatus.Removed]
        ]
        for ip in self.ip_list:
            if ip in project_node_list:
                removed_ips.append(ip)
            else:
                remained_ips.append(ip)
        return removed_ips, remained_ips

    def add_nodes(self, remained_ips):
        if not remained_ips:
            return
        data = [
            {
                'creator': self.username,
                'name': ip,
                'inner_ip': ip,
                'description': ip,
                'device_class': '',
                'status': CommonStatus.Initializing,
            }
            for ip in remained_ips
        ]
        resp = paas_cc.create_node(self.access_token, self.project_id, self.cluster_id, {'objects': data})
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(resp.get('message'))

    def get_cluster_info(self):
        resp = paas_cc.get_cluster(self.access_token, self.project_id, self.cluster_id)
        if resp.get('code') != ErrorCode.NoError or not resp.get('data'):
            raise error_codes.APIError(resp.get('message', _("获取集群信息为空")))
        return resp['data']

    def create(self):
        """添加节点
        1. 检查节点是否可用
        2. 触发OPS api
        """
        # 校验集群edit权限
        self.check_perm()
        # 校验数据
        self.check_data()
        self.ip_list = [ip.split(',')[0] for ip in self.data['ip']]
        # 检测IP是否被占用
        can_use_hosts(self.project_info["cc_app_id"], self.username, self.ip_list)
        self.project_nodes = paas_cc.get_all_cluster_hosts(self.access_token)
        self.check_node_ip()
        # 获取已经存在的IP，直接更新使用
        removed_ips, remained_ips = self.get_removed_remained_ips()
        # 更新IP
        self.update_nodes(removed_ips)
        # 添加IP
        self.add_nodes(remained_ips)
        # 获取节点是否需要NAT
        cluster_info = self.get_cluster_info()
        self.need_nat = cluster_info.get('need_nat', True)
        # 现阶段平台侧不主动创建CMDB set&module，赋值为空列表
        self.module_id_list = []
        # 请求ops api
        with client.ContextActivityLogClient(
            project_id=self.project_id,
            user=self.username,
            resource_type=ACTIVITY_RESOURCE_TYPE,
            resource=','.join(self.ip_list)[:32],
        ).log_add():
            # 更新所有节点为初始化中
            node_info_list = self.update_cluster_nodes(self.ip_list)
            log = self.create_node_by_bcs(node_info_list)
            if not log.is_finished and log.is_polling:
                log.polling_task()
        return Response({})


class ReinstallNode(BaseNode):
    def __init__(self, request, project_id, cluster_id, node_id):
        self.request = request
        self.access_token = request.user.token.access_token
        self.username = request.user.username
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.node_id = node_id
        self.bk_token = request.COOKIES.get('bk_token')
        self.project_info = request.project
        self.kind_name = ProjectKindName
        self.cc_app_id = request.project.get('cc_app_id')

    def get_node_ip(self):
        node_info = paas_cc.get_node(self.access_token, self.project_id, self.node_id)
        if node_info.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(node_info.get('message'))
        return node_info.get('data') or {}

    def get_node_last_log(self):
        log = NodeUpdateLog.objects.filter(
            project_id=self.project_id, cluster_id=self.cluster_id, node_id__contains='[%s]' % self.node_id
        ).last()
        if not log:
            raise error_codes.APIError(_("没有查询到节点添加记录，请联系管理员处理!"))
        return log

    def ratelimit(self):
        rate_limiter = RateLimiter(rd_client, '%s_%s' % (self.cluster_id, self.node_id))
        rate_limiter.add_rule(1, {"second": 15})
        try:
            resp = rate_limiter.acquire()
        except Exception as error:
            logger.error('%s, %s' % (error_codes.ConfigError.code_num, "获取token出现异常,详情:%s" % error))
        if not resp.get('allowed'):
            raise error_codes.CheckFailed(_("已经触发操作，请勿重复操作"))

    def reinstall(self):
        self.ratelimit()
        # 校验集群编辑权限
        self.check_perm()
        # 通过node id获取Ip信息
        node_info = self.get_node_ip()
        log = self.get_node_last_log()
        node_ip = node_info.get('inner_ip')
        params = json.loads(log.params)
        # 校验权限
        if node_info.get('status') not in [NodeStatus.Removed, NodeStatus.InitialFailed]:
            raise error_codes.CheckFailed(_("IP: {}正在操作中，请勿重复操作").format(node_ip))
        with client.ContextActivityLogClient(
            project_id=self.project_id,
            user=self.username,
            resource_type=ACTIVITY_RESOURCE_TYPE,
            resource=node_ip,
            resource_id=self.node_id,
        ).log_modify():
            self.update_nodes([node_ip])
            # 调用OPS api
            self.need_nat = params['need_nat']
            self.master_ip_list = params['master_ip_list']
            self.module_id_list = params['module_id_list']
            self.ip_list = [node_ip]
            log = self.create_node_by_bcs(
                [node_info], control_ip=params['control_ip'], config=params['config'], websvr=params['websvr']
            )
            if not log.is_finished and log.is_polling:
                log.polling_task()

        return Response({})


class DeleteNodeBase(BaseNode):
    def delete_via_bcs(self, request, project_id, cluster_id, kind_name, node_info):
        self.ip_list = list(node_info.keys())
        params = {
            'project_id': project_id,
            'cluster_id': cluster_id,
            'username': request.user.username,
            'node_info': node_info,
            'ip_list': self.ip_list,
            'kind': request.project['kind'],
            'kind_name': kind_name,
            'cc_app_id': request.project['cc_app_id'],
        }
        log = NodeUpdateLog.objects.create(
            project_id=project_id,
            cluster_id=cluster_id,
            token=request.user.token.access_token,
            node_id=','.join(node_info.values()),
            params=json.dumps(params),
            operator=request.user.username,
            oper_type=NodeOperType.NodeRemove,
            is_polling=True,
            is_finished=False,
        )
        config = self.get_request_config(op_type=constants.OpType.DELETE_NODE.value)
        control_ip = config.pop('control_ip', [])
        websvr = config.pop('websvr', [])
        try:
            task_info = ops.delete_cluster_node(
                request.user.token.access_token,
                project_id,
                kind_name,
                cluster_id,
                self.master_ip_list,
                self.ip_list,
                config,
                control_ip,
                request.project['cc_app_id'],
                request.user.username,
                websvr,
            )
        except Exception as err:
            logger.exception('request bcs ops error, detail: %s', err)
            task_info = {'code': ErrorCode.UnknownError}
        if task_info.get('code') != ErrorCode.NoError:
            log.set_finish_polling_status(True, False, CommonStatus.RemoveFailed)
            # 更改节点状态
            self.update_nodes(self.ip_list, status=CommonStatus.RemoveFailed)
        data = task_info.get('data') or {}
        task_id = data.get('task_id')
        if not task_id:
            raise error_codes.APIError(_("获取标准运维任务ID失败，返回任务为{}，请联系管理员处理").format(task_id))
        log.set_task_id(task_id)
        self.save_task_url(log, data)
        return log

    def can_delete_node(self, access_token, project_id, cluster_id):
        cluster = get_cluster(access_token, project_id, cluster_id)
        # 针对导入/纳管的集群，不允许通过平台删除节点
        if cluster["state"] == ClusterState.Existing.value:
            raise ValidationError(_("导入的集群不允许通过平台删除节点"))


class DeleteNode(DeleteNodeBase):
    def __init__(self, request, project_id, cluster_id, node_id):
        self.request = request
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.node_id = node_id
        self.access_token = request.user.token.access_token
        self.username = request.user.username
        self.bk_token = request.COOKIES.get('bk_token')
        self.project_info = request.project
        self.kind_name = ProjectKindName
        self.cc_app_id = request.project.get('cc_app_id')

    def k8s_container_num(self):
        client = bcs_k8s.K8SClient(self.access_token, self.project_id, self.cluster_id, None)
        host_pod_info = client.get_pod(
            host_ips=[self.node_ip], field=','.join(['data.status.containerStatuses', 'data.metadata.namespace'])
        )
        if host_pod_info.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(host_pod_info.get('message'))
        count = 0
        for i in host_pod_info.get('data', []):
            namespace = i.get('data', {}).get('metadata', {}).get('namespace')
            if namespace in constants.K8S_SKIP_NS_LIST:
                continue
            count += len(i.get('data', {}).get('status', {}).get('containerStatuses', []))
        return count

    def check_host_exist_container(self):
        """获取node下是否有容器运行"""
        container_count = getattr(self, '%s_container_num' % self.kind_name)()
        if container_count:
            raise error_codes.CheckFailed(_("当前节点下存在运行容器, 请先清理容器!"))

    def check_host_removing(self, node_info):
        if node_info.get('status') in [NodeStatus.Removing, CommonStatus.Scheduling]:
            raise error_codes.CheckFailed(_("当前节点正在删除，请勿重复操作!"))

    def check_host_stop_scheduler(self, node_info):
        status = [
            NodeStatus.ToRemoved,
            NodeStatus.Removable,
            NodeStatus.RemoveFailed,
            CommonStatus.ScheduleFailed,
            NodeStatus.InitialFailed,
        ]
        if node_info.get("status") not in status:
            raise error_codes.CheckFailed(_("节点必须要先停用，才可以删除，请确认!"))

    def delete_node_via_bcs(self):
        node_info = {self.node_ip: '[%s]' % self.node_id}
        log = self.delete_via_bcs(self.request, self.project_id, self.cluster_id, self.kind_name, node_info)
        return log

    def delete_node_labels(self):
        NodeLabel.objects.filter(node_id=self.node_id, is_deleted=False).update(
            is_deleted=True, deleted_time=datetime.now(), updator=self.username, labels=json.dumps({})
        )

    def delete(self):
        """删除节点
        1. 检查节点状态
        2. 查询容器数量
        3. 调用ops删除节点
        """
        self.check_perm()
        # 查询节点
        node_info = self.get_node_ip()
        # 检查操作
        self.check_host_removing(node_info)
        self.check_host_stop_scheduler(node_info)
        self.node_ip = node_info.get('inner_ip')
        # 检测容器是否存在
        self.check_host_exist_container()
        # 判断节点是否允许删除
        self.can_delete_node(self.request.user.token.access_token, self.project_id, self.cluster_id)
        # 调用BCS
        with client.ContextActivityLogClient(
            project_id=self.project_id,
            user=self.username,
            resource_type=ACTIVITY_RESOURCE_TYPE,
            resource=self.node_ip,
            resource_id=self.node_id,
        ).log_delete():
            # 更新状态
            self.update_nodes([self.node_ip], status=CommonStatus.Removing)
            # 删除节点上的label
            self.delete_node_labels()
            log = self.delete_node_via_bcs()
            if not log.is_finished and log.is_polling:
                log.polling_task()
        return Response({})

    def force_delete(self):
        """强制删除
        1. 检查状态
        2. 调用ops删除节点
        """
        self.check_perm()
        # 判断节点是否允许删除
        self.can_delete_node(self.request.user.token.access_token, self.project_id, self.cluster_id)
        # 查询节点
        node_info = self.get_node_ip()
        self.check_host_stop_scheduler(node_info)
        self.check_host_removing(node_info)

        # 调用BCS
        self.node_ip = node_info.get('inner_ip')
        with client.ContextActivityLogClient(
            project_id=self.project_id,
            user=self.username,
            resource_type=ACTIVITY_RESOURCE_TYPE,
            resource=self.node_ip,
            resource_id=self.node_id,
        ).log_delete():
            # 更新状态
            self.update_nodes([self.node_ip], status=CommonStatus.Removing)
            # 删除节点上的label
            self.delete_node_labels()
            log = self.delete_node_via_bcs()
            if not log.is_finished and log.is_polling:
                log.polling_task()
        return Response({})


class BatchDeleteNode(DeleteNodeBase):
    def __init__(self, request, project_id, cluster_id, node_list):
        self.request = request
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.node_list = node_list
        self.kind_name = ProjectKindName
        self.access_token = request.user.token.access_token

    def delete_nodes(self):
        # 判断节点是否允许删除
        self.can_delete_node(self.request.user.token.access_token, self.project_id, self.cluster_id)

        node_info = {node['inner_ip']: '[%s]' % node['id'] for node in self.node_list}
        # 更新节点状态为删除中
        self.update_cluster_nodes([node["inner_ip"] for node in self.node_list], NodeStatus.Removing)
        log = self.delete_via_bcs(self.request, self.project_id, self.cluster_id, self.kind_name, node_info)
        if not log.is_finished and log.is_polling:
            log.polling_task()


class BatchReinstallNodes(BaseNode):
    def __init__(self, request, project_id, cluster_info, node_id_ip_map):
        self.request = request
        self.project_id = project_id
        self.cluster_id = cluster_info['cluster_id']
        self.cluster_info = cluster_info
        self.node_id_ip_map = node_id_ip_map
        self.access_token = request.user.token.access_token
        self.username = request.user.username
        self.project_info = request.project
        self.kind_name = ProjectKindName
        self.cc_app_id = self.project_info.cc_app_id

    def reinstall(self):
        self.need_nat = self.cluster_info.get('need_nat', True)
        # 现阶段平台侧不主动创建CMDB set&module，赋值为空列表
        self.module_id_list = []
        self.ip_list = list(self.node_id_ip_map.values())
        log = self.create_node_by_bcs([{'inner_ip': ip, 'id': id} for id, ip in self.node_id_ip_map.items()])
        if not log.is_finished and log.is_polling:
            log.polling_task()
