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
from datetime import datetime

from django.conf import settings
from django.utils.translation import ugettext_lazy as _
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.accounts.bcs_perm import Cluster, Namespace
from backend.bcs_web.audit_log import client
from backend.components import ops, paas_cc
from backend.container_service.clusters import constants, serializers
from backend.container_service.clusters.base import utils as cluster_utils
from backend.container_service.clusters.constants import ClusterState
from backend.container_service.clusters.models import ClusterInstallLog, ClusterOperType, CommonStatus
from backend.container_service.clusters.utils import can_use_hosts
from backend.container_service.projects.base.constants import ProjectKindName
from backend.helm.app.models import App
from backend.templatesets.legacy_apps.instance.models import (
    InstanceConfig,
    InstanceEvent,
    MetricConfig,
    VersionInstance,
)
from backend.uniapps.network.models import K8SLoadBlance
from backend.utils.cache import rd_client
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.ratelimit import RateLimiter
from backend.utils.renderers import BKAPIRenderer

from .configs import k8s

logger = logging.getLogger(__name__)
ACTIVITY_RESOURCE_TYPE = 'cluster'
DEFAULT_K8S_VERSION = getattr(settings, 'K8S_VERSION', 'v1.12.3') or 'v1.12.3'
NO_RES = "**"
CLUSTER_ENVIRONMENT = 'prod'


class BaseCluster:
    render_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def update_cluster_status(self, status=CommonStatus.Initializing):
        cluster_info = paas_cc.update_cluster(self.access_token, self.project_id, self.cluster_id, {'status': status})
        if cluster_info.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(cluster_info.get('message'))
        return cluster_info.get('data', {})

    def delete_cluster(self, cluster_id):
        """未请求到ops api，异常出现时，删除集群"""
        resp = paas_cc.delete_cluster(self.access_token, self.project_id, cluster_id)
        if resp.get('code') != ErrorCode.NoError:
            logger.error('Request paas cc api error, resp: %s' % json.dumps(resp))

    def get_request_config(self, cluster_id, version, environment):
        kind_type_map = {'k8s': k8s.ClusterConfig}
        self.get_cluster_base_config(cluster_id, version=version, environment=environment)

        client = kind_type_map[self.kind_name](self.config, self.area_info, cluster_name=self.cluster_name)
        kwargs = {
            "access_token": self.access_token,
            "project_id": self.project_id,
            "cluster_state": self.data["cluster_state"],
        }
        return client.get_request_config(cluster_id, self.data['master_ips'], need_nat=self.data['need_nat'], **kwargs)

    def save_snapshot(self, cluster_id, config):
        data = {
            'cluster_id': cluster_id,
            'configure': json.dumps(config),
            'creator': self.username,
            'project_id': self.project_id,
        }
        resp = paas_cc.save_cluster_snapshot(self.access_token, data)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(resp.get('message'))

    def save_task_url(self, log, data):
        log_params = log.log_params
        log_params['task_url'] = data.get('task_url') or ''
        log.set_params(log_params)

    def get_ops_backend(self, cluster_state):
        """根据集群state获取ops对应的backend
        - 通过bcs创建集群: gcloud_v3
        - 导入已存在集群: gcloud_v3_contain
        """
        if cluster_state == constants.ClusterState.BCSNew.value:
            return "gcloud_v3"
        return "gcloud_v3_contain"

    def create_cluster_via_bcs(
        self, cluster_id, cc_module, config=None, version='1.12.3', environment='prod', websvr=None
    ):  # noqa
        """调用bcs接口创建集群"""
        self.cluster_id = cluster_id
        kind_version_map = {'k8s': DEFAULT_K8S_VERSION}
        if not config:
            config = self.get_request_config(cluster_id, kind_version_map[self.kind_name], environment)
        # 下发配置时，去除version
        config.pop('control_ip', None)
        websvr = config.pop('websvr', []) or websvr
        config.pop('version', None)
        config.pop("platform", None)
        # 存储参数，便于任务失败重试
        params = {
            'project_id': self.project_id,
            'cluster_id': cluster_id,
            'cluster_name': self.cluster_name,
            'module_id_list': cc_module,
            'username': self.username,
            'master_ips': self.data['master_ips'],
            'kind': self.project_info['kind'],
            'kind_name': self.kind_name,
            'need_nat': self.data['need_nat'],
            'control_ip': self.control_ip,
            'config': config,
            'version': version,
            'environment': self.data['environment'],
            'cc_app_id': self.cc_app_id,
            'area_name': self.area_name,
            'project_name': self.project_info['project_name'],
            'websvr': websvr,
        }

        log = ClusterInstallLog.objects.create(
            project_id=self.project_id,
            cluster_id=cluster_id,
            token=self.access_token,
            status=CommonStatus.Initializing,
            params=json.dumps(params),
            operator=self.request.user.username,
            oper_type=ClusterOperType.ClusterInstall,
            is_polling=True,
        )

        platform = self.get_ops_backend(self.data["cluster_state"])
        try:
            task_info = ops.create_cluster(
                self.access_token,
                self.project_id,
                self.kind_name,
                cluster_id,
                self.data['master_ips'],
                config,
                cc_module,
                self.control_ip,
                self.cc_app_id,
                self.username,
                websvr,
                platform=platform,
            )
        except Exception as err:
            logger.exception('Create cluster error: %s', err)
            task_info = {'code': ErrorCode.UnknownError}

        # 存储快照
        config['version'] = kind_version_map[self.kind_name]
        config['control_ip'] = self.control_ip
        config['websvr'] = websvr
        config["platform"] = platform
        self.save_snapshot(cluster_id, config)
        if task_info.get('code') != ErrorCode.NoError:
            log.set_finish_polling_status(True, False, CommonStatus.InitialFailed)
            # 兼容log
            log.log = json.dumps(constants.BCS_OPS_ERROR_INFO)
            log.save()
            # 更新集群状态
            self.update_cluster_status(status=CommonStatus.InitialFailed)
            logger.error(task_info.get('message'))
            return log
            # raise error_codes.APIError.f("初始化集群失败，请联系管理员处理!")
        data = task_info.get('data') or {}
        task_id = data.get('task_id')
        if not task_id:
            raise error_codes.APIError(_("获取初始化任务ID异常，返回任务ID为{}，请联系管理员处理").format(task_id))
        log.set_task_id(task_id)
        self.save_task_url(log, data)
        return log

    def register_cluster(self, cluster_info):
        cluster_perm = Cluster(self.request, self.project_id, cluster_info['cluster_id'])
        cluster_perm.register(cluster_info["cluster_id"], cluster_info["name"], cluster_info["environment"])


class CreateCluster(BaseCluster):
    def __init__(self, request, project_id):
        self.request = request
        self.project_id = project_id
        self.access_token = request.user.token.access_token
        self.bk_token = request.COOKIES.get('bk_token')
        self.project_info = request.project
        self.username = request.user.username
        self.kind_name = ProjectKindName

    def check_data(self):
        slz = serializers.CreateClusterSLZ(
            data=self.request.data,
            context={"project_id": self.project_id, "access_token": self.access_token, "default_coes": self.kind_name},
        )
        slz.is_valid(raise_exception=True)
        self.data = slz.validated_data
        self.data['creator'] = self.username
        # 强制只有prod环境
        self.data['environment'] = CLUSTER_ENVIRONMENT
        # 获取绑定的CMDB业务ID
        self.cc_app_id = self.project_info.get('cc_app_id')
        if not self.cc_app_id:
            raise error_codes.CheckFailed(_("当前项目没有绑定业务，不允许操作容器服务"))
        # NOTE: 选择第一个master节点作为中控机IP
        self.control_ip = self.data['master_ips'][:1]
        # 检查是否有权限操作IP
        can_use_hosts(self.cc_app_id, self.username, self.data["master_ips"])
        self.area_info = self.get_area_info()
        self.data['area_id'] = self.area_info['id']

    def get_area_info(self):
        """获取指定区域配置"""
        area_info = paas_cc.get_area_list(self.access_token)
        if area_info.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(area_info.get('message'))
        area_info_data = area_info.get('data') or {}
        area_list = area_info_data.get('results') or []
        if not area_list:
            raise error_codes.APIError(_("获取区域配置信息为空，请确认后重试"))
        data = area_list[0]
        if not data:
            raise error_codes.CheckFailed(_("获取区域配置为空，请确认后重试"))
        return data

    def get_cluster_base_config(self, cluster_id, version, environment='prod'):
        params = {'ver_id': version, 'environment': environment, 'kind': self.kind_name}
        base_cluster_config = paas_cc.get_base_cluster_config(self.access_token, self.project_id, params)
        if base_cluster_config.get('code') != ErrorCode.NoError:
            # delete created cluster record
            self.delete_cluster(cluster_id)
            raise error_codes.APIError(base_cluster_config.get('message'))
        config = json.loads(base_cluster_config.get('data', {}).get('configure', '{}'))
        if not config:
            raise error_codes.CheckFailed(_("获取集群基本配置失败"))
        config['version'] = version
        self.config = config

    def check_cluster_perm(self):
        res_type = "cluster_prod" if self.data["environment"] == "prod" else "cluster_test"
        perm_cluster = Cluster(self.request, self.project_id, NO_RES, resource_type=res_type)
        perm_cluster.can_create(raise_exception=True)

    def create(self):
        """集群初始化流程
        1. 申请集群ID
        2. 创建set及module
        3. 触发OPS api
        """
        self.check_data()
        # 权限校验
        self.check_cluster_perm()
        cluster_data = copy.deepcopy(self.data)
        cluster_data['master_ips'] = [{'inner_ip': ip} for ip in self.data['master_ips']]
        cluster_data["state"] = cluster_data["cluster_state"]
        # 添加类型参数
        cluster_data["type"] = cluster_data["coes"]
        self.cluster_name = self.data['name']
        # 创建set
        with client.ContextActivityLogClient(
            project_id=self.project_id,
            user=self.username,
            resource_type=ACTIVITY_RESOURCE_TYPE,
            resource=cluster_data['name'],
        ).log_add() as ual:
            resp = paas_cc.create_cluster(self.access_token, self.project_id, cluster_data)
            if resp.get('code') != ErrorCode.NoError or not resp.get('data'):
                raise error_codes.APIError(resp.get('message', _("创建集群失败")))
            cluster_info = resp.get('data')
            # 现阶段平台侧不主动创建CMDB set&module，赋值为空列表
            module_id_list = []

        # 兼容不创建set后，提取出区域名称
        self.area_name = self.area_info['name']
        ual.update_log(resource_id=cluster_info.get("cluster_id"))
        log = self.create_cluster_via_bcs(cluster_info['cluster_id'], module_id_list)
        if not log.is_finished and log.is_polling:
            log.polling_task()
        # 注册集群信息到AUTH， 便于申请权限
        self.register_cluster(cluster_info)

        return Response(cluster_info)


class ReinstallCluster(BaseCluster):
    def __init__(self, request, project_id, cluster_id):
        self.request = request
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.access_token = request.user.token.access_token
        self.username = request.user.username
        self.project_info = request.project
        self.cc_app_id = request.project.get('cc_app_id')

    def ratelimit(self):
        rate_limiter = RateLimiter(rd_client, self.cluster_id)
        rate_limiter.add_rule(1, {"second": 15})
        try:
            resp = rate_limiter.acquire()
        except Exception as error:
            logger.error('%s, %s' % (error_codes.ConfigError.code_num, "获取token出现异常,详情:%s" % error))
        if not resp.get('allowed'):
            raise error_codes.CheckFailed(_("已经触发操作，请勿重复操作"))

    def get_cluster_info(self):
        cluster_info = paas_cc.get_cluster(self.access_token, self.project_id, self.cluster_id)
        if cluster_info.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(cluster_info.get('message'))
        if not cluster_info.get('data'):
            raise error_codes.APIError(_("查询集群: {} 对应的信息为空").format(self.cluster_id))
        return cluster_info

    def get_cluster_last_log(self):
        log = ClusterInstallLog.objects.filter(
            project_id=self.project_id,
            cluster_id=self.cluster_id,
        ).last()
        if not log:
            raise error_codes.CheckFailed(_("没有查询对应的任务日志"))
        return log

    def check_cluster_perm(self):
        cluster_perm = Cluster(self.request, self.project_id, self.cluster_id)
        cluster_perm.can_create(raise_exception=True)

    def reinstall(self):
        # 由于前端轮训，因此，为防止重复触发，使用访问频率控制
        self.ratelimit()
        self.check_cluster_perm()
        cluster_info = self.get_cluster_info()
        data = cluster_info.get('data', {})
        if data.get('status') in [CommonStatus.Initializing, CommonStatus.Removing]:
            raise error_codes.CheckFailed(_("集群正在操作中，请勿重复操作"))
        # 获取任务日志
        log = self.get_cluster_last_log()
        # 更新集群状态，使其处于执行中状态
        data = self.update_cluster_status()
        # 获取存储下发的配置信息
        params = json.loads(log.params)
        with client.ContextActivityLogClient(
            project_id=self.project_id,
            user=self.username,
            resource_type=ACTIVITY_RESOURCE_TYPE,
            resource=self.cluster_id,
            resource_id=self.cluster_id,
        ).log_modify():
            self.data = {
                'master_ips': params['master_ips'],
                'need_nat': params['need_nat'],
                'environment': params['environment'],
                'cluster_id': self.cluster_id,
                'name': params['cluster_name'],
                "cluster_state": data.get("state") or constants.ClusterState.BCSNew.value,
            }
            self.cluster_name = params['cluster_name']
            self.kind_name = params['kind_name']
            self.area_name = params['area_name']
            self.control_ip = params['control_ip']
            websvr = params['websvr']
            log = self.create_cluster_via_bcs(
                self.cluster_id, params['module_id_list'], config=params.get('config'), websvr=websvr
            )
            if not log.is_finished and log.is_polling:
                log.polling_task()
        self.register_cluster(self.data)

        return Response(data)


class DeleteCluster(BaseCluster):
    def __init__(self, request, project_id, cluster_id):
        self.request = request
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.access_token = request.user.token.access_token
        self.username = request.user.username
        self.project_info = request.project
        self.kind_name = ProjectKindName
        self.cc_app_id = request.project.get('cc_app_id')

    def get_cluster_snapshot(self):
        snapshot_info = paas_cc.get_cluster_snapshot(self.access_token, self.project_id, self.cluster_id)
        if snapshot_info.get('code') != ErrorCode.NoError:
            self.update_cluster_status(status=CommonStatus.RemoveFailed)
            raise error_codes.APIError(snapshot_info.get('message'))
        return snapshot_info.get('data', {})

    def get_cluster_ns_list(self):
        cluster_ns_info = paas_cc.get_cluster_namespace_list(
            self.access_token, self.project_id, self.cluster_id, desire_all_data=True
        )
        if cluster_ns_info.get('code') != ErrorCode.NoError:
            self.update_cluster_status(status=CommonStatus.RemoveFailed)
            raise error_codes.APIError(cluster_ns_info.get('message'))
        data = cluster_ns_info.get('data', {}).get('results') or []
        # 删除命名空间权限记录
        for info in data:
            perm_client = Namespace(self.request, self.project_id, info["id"])
            perm_client.delete()
        return [int(info['id']) for info in data]

    def delete_namespaces(self):
        """删除集群下的命名空间"""
        resp = paas_cc.delete_cluster_namespace(self.access_token, self.project_id, self.cluster_id)
        if resp.get('code') != ErrorCode.NoError:
            self.update_cluster_status(status=CommonStatus.RemoveFailed)
            raise error_codes.APIError(resp.get('message'))

    def clean_instance(self, ns_id_list):
        """调整命名空间对应的实例等信息的删除状态"""
        instance_info = InstanceConfig.objects.filter(namespace__in=ns_id_list)
        version_instance_id_list = [info.instance_id for info in instance_info]
        # 更新instance为已删除
        datetime_now = datetime.now()
        instance_info.update(deleted_time=datetime_now, is_deleted=True)
        # 更新version instance为已删除
        VersionInstance.objects.filter(id__in=version_instance_id_list).update(
            deleted_time=datetime_now, is_deleted=True
        )
        # 更新instance event为已删除
        InstanceEvent.objects.filter(instance_id__in=version_instance_id_list).update(
            deleted_time=datetime_now, is_deleted=True
        )
        # 更新metric config为已删除
        MetricConfig.objects.filter(instance_id__in=version_instance_id_list, namespace__in=ns_id_list).update(
            deleted_time=datetime_now, is_deleted=True
        )

    def clean_lb(self):
        K8SLoadBlance.objects.filter(project_id=self.project_id, cluster_id=self.cluster_id).update(
            is_deleted=True, deleted_time=datetime.now()
        )

    def delete_helm_release(self):
        App.objects.filter(project_id=self.project_id, cluster_id=self.cluster_id).delete()

    def get_cluster_master(self):
        cluster_masters = paas_cc.get_master_node_list(self.access_token, self.project_id, self.cluster_id)
        if cluster_masters.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(cluster_masters.get('message'))
        results = cluster_masters.get('data', {}).get('results') or []
        return [info['inner_ip'] for info in results if info.get('inner_ip')]

    def delete_cluster_via_bcs(self):
        # 获取集群下对应的master ip
        master_ip_list = self.get_cluster_master()
        params = {
            'project_id': self.project_id,
            'cluster_id': self.cluster_id,
            'cc_app_id': self.cc_app_id,
            'host_ips': master_ip_list,
            'project_code': self.project_info['english_name'],
            'username': self.username,
        }
        # 创建记录
        log = ClusterInstallLog.objects.create(
            project_id=self.project_id,
            cluster_id=self.cluster_id,
            token=self.access_token,
            status=CommonStatus.Removing,
            params=json.dumps(params),
            operator=self.username,
            oper_type=ClusterOperType.ClusterRemove,
            is_polling=True,
        )
        # NOTE: 因为后端的任务调度引擎针对删除集群的流程，跳过异常；因此，当出现调用接口异常时，也跳过异常处理
        try:
            task_info = ops.delete_cluster(
                self.access_token,
                self.project_id,
                self.kind_name,
                self.cluster_id,
                master_ip_list,
                self.control_ip,
                self.cc_app_id,
                self.username,
                self.websvr,
                self.config,
                platform=self.config.pop("platform", "gcloud_v3"),
            )
            if task_info.get("code") != ErrorCode.NoError:
                logger.warning(
                    "调用删除集群流程失败，项目ID: %s, 集群ID: %s, error: %s",
                    self.project_id,
                    self.cluster_id,
                    task_info.get("message"),
                )
                self.delete_cluster(self.cluster_id)
                return
        except Exception as e:
            logger.error("调用删除集群流程异常，error: %s", e)
            self.delete_cluster(self.cluster_id)
            return

        data = task_info.get('data') or {}
        task_id = data.get('task_id')
        if not task_id:
            raise error_codes.APIError(_("获取标准运维任务ID失败，返回任务ID为{}，请联系管理员处理").format(task_id))
        log.set_task_id(task_id)
        self.save_task_url(log, data)
        return log

    def check_cluster_perm(self):
        cluster_perm = Cluster(self.request, self.project_id, self.cluster_id)
        cluster_perm.can_delete(raise_exception=True)

    def delete_cluster_nodes(self, access_token, project_id, cluster_id):
        """删除集群下的节点"""
        # 查询集群下节点
        nodes = cluster_utils.get_cluster_nodes(access_token, project_id, cluster_id)
        if not nodes:
            return
        node_status = [{"inner_ip": node["inner_ip"], "status": CommonStatus.Removed} for node in nodes]
        cluster_utils.update_cc_nodes_status(access_token, project_id, cluster_id, node_status)

    def delete(self):
        # 添加权限限制
        self.check_cluster_perm()
        # 针对导入/纳管的集群，需要删除节点记录
        cluster = cluster_utils.get_cluster(self.access_token, self.project_id, self.cluster_id)
        if cluster["state"] == ClusterState.Existing.value:
            self.delete_cluster_nodes(self.access_token, self.project_id, self.cluster_id)
        # 更新集群状态为删除中
        data = self.update_cluster_status(status=CommonStatus.Removing)
        # 删除平台数据库记录的已经实例化，但是没有删除的实例信息
        ns_id_list = self.get_cluster_ns_list()
        self.clean_instance(ns_id_list)
        self.delete_namespaces()
        self.delete_helm_release()
        # 下发删除任务
        with client.ContextActivityLogClient(
            project_id=self.project_id,
            user=self.username,
            resource_type=ACTIVITY_RESOURCE_TYPE,
            resource=data['name'],
            resource_id=self.cluster_id,
        ).log_delete():
            snapshot = self.get_cluster_snapshot()
            self.config = json.loads(snapshot.get('configure', '{}'))
            self.control_ip = self.config.pop('control_ip', [])
            self.websvr = self.config.pop('websvr', [])
            task = self.delete_cluster_via_bcs()
            if task and not task.is_finished and task.is_polling:
                task.polling_task()

        return Response({})
