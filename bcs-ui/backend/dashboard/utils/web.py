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
from typing import Dict, Union

from django.utils.translation import ugettext_lazy as _

from backend.dashboard.exceptions import ResourceTypeUnsupported
from backend.iam.permissions.client import IAMClient
from backend.iam.permissions.resources.cluster import ClusterRequest
from backend.iam.permissions.resources.cluster_scoped import ClusterScopedAction
from backend.iam.permissions.resources.constants import ResourceType
from backend.iam.permissions.resources.namespace import NamespaceRequest
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedAction


def get_res_inst_multi_actions_perms(
    username: str, project_id: str, cluster_id: str, namespace: Union[str, None], resource_type: ResourceType
) -> Dict:
    """获取指定实例的多个操作权限"""
    if resource_type not in [ResourceType.Namespace, ResourceType.Cluster]:
        raise ResourceTypeUnsupported()

    if resource_type == ResourceType.Namespace:
        resources = NamespaceRequest(project_id=project_id, cluster_id=cluster_id).make_resources(namespace)
        action_ids = [
            NamespaceScopedAction.VIEW,
            NamespaceScopedAction.CREATE,
            NamespaceScopedAction.UPDATE,
            NamespaceScopedAction.DELETE,
        ]
    else:
        resources = ClusterRequest(project_id).make_resources(cluster_id)
        action_ids = [
            ClusterScopedAction.VIEW,
            ClusterScopedAction.CREATE,
            ClusterScopedAction.UPDATE,
            ClusterScopedAction.DELETE,
        ]

    return IAMClient().resource_inst_multi_actions_allowed(username, action_ids, resources)


def gen_base_web_annotations(username: str, project_id: str, cluster_id: str, namespace: str) -> Dict:
    """生成资源视图相关的页面控制信息，用于控制按钮展示等"""
    resource_type = ResourceType.Namespace if namespace else ResourceType.Cluster
    perms = get_res_inst_multi_actions_perms(username, project_id, cluster_id, namespace, resource_type)
    tip = _('当前用户没有该操作的权限')

    # 使用 web-console 要求 use 权限，即 view, create, update, delete 总和
    can_use_web_console = all(perms.values())
    # TODO 由于资源视图 webAnnotations 未标准化，因此暂时不提供按钮级别的禁用，改接口报错后弹窗提醒申请权限
    return {
        'perms': {
            'page': {
                'create_btn': {'clickable': True, 'tip': ''},
                'update_btn': {'clickable': True, 'tip': ''},
                'delete_btn': {'clickable': True, 'tip': ''},
                'reschedule_pod_btn': {'clickable': True, 'tip': ''},
                'web_console_btn': {'clickable': can_use_web_console, 'tip': '' if can_use_web_console else tip},
            }
        }
    }
