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
from ..exceptions import PermissionDeniedError
from .cluster_scoped import ClusterScopedPermCtx, ClusterScopedPermission
from .namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission


def can_instantiate_in_cluster(username: str, project_id: str, cluster_id: str, namespace: str):
    message = ''
    action_request_list = []
    force_raise = False

    try:
        namespace_scoped_perm_ctx = NamespaceScopedPermCtx(
            username=username, project_id=project_id, cluster_id=cluster_id, name=namespace
        )
        NamespaceScopedPermission().can_use(namespace_scoped_perm_ctx, True)
    except PermissionDeniedError as e:
        message = f'{message}; {e.message}'
        action_request_list.extend(e.action_request_list)
        force_raise = True

    try:
        cluster_scoped_perm_ctx = ClusterScopedPermCtx(
            username=username, project_id=project_id, cluster_id=cluster_id, force_raise=force_raise
        )
        ClusterScopedPermission().can_use(cluster_scoped_perm_ctx, True)
    except PermissionDeniedError as e:
        message = f'{message}; {e.message}'
        action_request_list.extend(e.action_request_list)

    if message:
        raise PermissionDeniedError(
            message=message.lstrip('; '), username=username, action_request_list=action_request_list
        )
