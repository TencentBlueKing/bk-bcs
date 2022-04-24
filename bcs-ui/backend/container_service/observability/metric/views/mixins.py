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
from typing import Callable, Dict, List

import arrow
from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.bcs_web.audit_log import client as activity_client
from backend.bcs_web.audit_log.constants import ActivityStatus, ActivityType, ResourceType
from backend.components import paas_cc
from backend.container_service.clusters.base.utils import append_shared_clusters
from backend.container_service.observability.metric import constants
from backend.container_service.projects.base.constants import LIMIT_FOR_ALL_DATA
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.utils.basic import getitems
from backend.utils.datetime import get_duration_seconds
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


class ServiceMonitorMixin:
    """ 一些通用的方法 """

    def _handle_endpoints(self, endpoints: List[Dict]) -> List[Dict]:
        """
        Endpoints 配置填充数据

        :param endpoints: 原始数据
        :return: 完成补充的数据
        """
        for endpoint in endpoints:
            endpoint.setdefault('path', constants.DEFAULT_ENDPOINT_PATH)
            endpoint['interval'] = get_duration_seconds(endpoint.get('interval'), constants.DEFAULT_ENDPOINT_INTERVAL)
        return endpoints

    def _handle_items(self, cluster_id: str, cluster_map: Dict, namespace_map: Dict, manifest: Dict) -> List[Dict]:
        """
        ServiceMonitor 配置填充数据

        :param cluster_id: 集群 ID
        :param cluster_map: {cluster_id: cluster_info}
        :param namespace_map: {(cluster_id, name): id}
        :param manifest: ServiceMonitor 配置信息
        :return: ServiceMonitor 列表
        """
        items = manifest.get('items') or []
        new_items = []

        for item in items:
            try:
                labels = item['metadata'].get('labels') or {}
                item['metadata'] = {
                    k: v for k, v in item['metadata'].items() if k not in constants.INNER_USE_SERVICE_METADATA_FIELDS
                }
                item['cluster_id'] = cluster_id
                item['namespace'] = item['metadata']['namespace']
                item['namespace_id'] = namespace_map.get((cluster_id, item['metadata']['namespace']))
                item['name'] = item['metadata']['name']
                item['instance_id'] = f"{item['namespace']}/{item['name']}"
                item['service_name'] = labels.get(constants.SM_SERVICE_NAME_LABEL)
                item['cluster_name'] = cluster_map[cluster_id]['name']
                item['environment'] = cluster_map[cluster_id]['environment']
                item['metadata']['service_name'] = labels.get(constants.SM_SERVICE_NAME_LABEL)
                item['create_time'] = (
                    arrow.get(item['metadata']['creationTimestamp'])
                    .to(settings.TIME_ZONE)
                    .format('YYYY-MM-DD HH:mm:ss')
                )
                if isinstance(item['spec'].get('endpoints'), list):
                    item['spec']['endpoints'] = self._handle_endpoints(item['spec']['endpoints'])
                new_items.append(item)
            except Exception as e:
                logger.error('handle item error, %s, %s', e, item)

        new_items = sorted(new_items, key=lambda x: x['create_time'], reverse=True)
        return new_items

    def _validate_namespace_use_perm(self, project_id: str, cluster_id: str, namespaces: List):
        """ 检查是否有命名空间的使用权限 """
        permission = NamespaceScopedPermission()
        for ns in namespaces:
            if ns in constants.SM_NO_PERM_NAMESPACE:
                raise error_codes.APIError(_('不允许操作命名空间 {}').format(ns))

            # 检查是否有命名空间的使用权限
            # TODO 针对多个，考虑批量去解
            perm_ctx = NamespaceScopedPermCtx(
                username=self.request.user.username, project_id=project_id, cluster_id=cluster_id, name=ns
            )
            permission.can_use(perm_ctx)

    def _activity_log(
        self,
        project_id: str,
        username: str,
        resource_name: str,
        description: str,
        activity_type: ActivityType,
        activity_status: ActivityStatus,
    ) -> None:
        """ 操作记录方法 """
        client = activity_client.ContextActivityLogClient(
            project_id=project_id, user=username, resource_type=ResourceType.Metric, resource=resource_name
        )
        # 根据不同的操作类型，使用不同的记录方法
        log_func = {
            ActivityType.Add: client.log_add,
            ActivityType.Delete: client.log_delete,
            ActivityType.Retrieve: client.log_note,
        }[activity_type]
        log_func(activity_status=activity_status, description=description)

    def _get_cluster_map(self, project_id: str) -> Dict:
        """
        获取集群配置信息

        :param project_id: 项目 ID
        :return: {cluster_id: cluster_info}
        """
        resp = paas_cc.get_all_clusters(self.request.user.token.access_token, project_id)
        # `data.results` 可能为 None，做类型兼容处理
        clusters = getitems(resp, 'data.results', []) or []
        # 添加共享集群
        clusters = append_shared_clusters(clusters)
        return {i['cluster_id']: i for i in clusters}

    def _get_namespace_map(self, project_id: str) -> Dict:
        """
        获取命名空间配置信息

        :param project_id: 项目 ID
        :return: {(cluster_id, name): id}
        """
        resp = paas_cc.get_namespace_list(self.request.user.token.access_token, project_id, limit=LIMIT_FOR_ALL_DATA)
        # `data.results` 可能为 None，做类型兼容处理
        namespaces = getitems(resp, 'data.results', []) or []
        return {(i['cluster_id'], i['name']): i['id'] for i in namespaces}

    def _single_service_monitor_operate_handler(
        self,
        client_func: Callable,
        operate_str: str,
        project_id: str,
        activity_type: ActivityType,
        namespace: str,
        name: str,
        manifest: Dict = None,
        log_success: bool = False,
    ) -> Dict:
        """
        执行单个 ServiceMonitor 操作类的通用处理逻辑

        :param client_func: k8s client 方法
        :param operate_str: 操作描述，可选值为：创建，更新，删除
        :param project_id: 项目 ID
        :param activity_type: 操作类型
        :param namespace: 命名空间
        :param name: ServiceMonitor 名称
        :param manifest: 完整配置信息（创建用），若为 None 则非创建逻辑
        :param log_success: 操作成功时记录日志
        :return:
        """
        username = self.request.user.username
        result = (
            client_func(namespace, manifest) if activity_type == ActivityType.Add else client_func(namespace, name)
        )
        if result.get('status') == 'Failure':
            message = _('{} Metrics [{}/{}] 失败: {}').format(operate_str, namespace, name, result.get('message', ''))
            self._activity_log(project_id, username, name, message, activity_type, ActivityStatus.Failed)
            raise error_codes.APIError(result.get('message', ''))

        # 仅当指定需要记录 成功信息 才记录
        if log_success:
            message = _('{} Metrics [{}/{}] 成功').format(operate_str, namespace, name)
            self._activity_log(project_id, username, name, message, activity_type, ActivityStatus.Succeed)
        return result
