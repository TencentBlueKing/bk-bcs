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
from typing import Optional

import attr

from ..exceptions import PermissionDeniedError
from .cluster_scoped import ClusterScopedPermCtx, ClusterScopedPermission
from .namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission


@attr.dataclass
class ProjectScopedPermCtx:
    username: str
    project_id: str
    cluster_id: str
    namespace: Optional[str] = None


def can_apply_in_cluster(perm_ctx: ProjectScopedPermCtx, raise_exception: bool = True) -> bool:
    """用于模板集或 Helm，校验用户是否有权限将资源(包含集群域和命名空间域)部署到集群中"""
    messages = []
    action_request_list = []
    force_raise = False

    if perm_ctx.namespace:
        try:
            namespace_scoped_perm_ctx = NamespaceScopedPermCtx(
                username=perm_ctx.username,
                project_id=perm_ctx.project_id,
                cluster_id=perm_ctx.cluster_id,
                name=perm_ctx.namespace,
            )
            NamespaceScopedPermission().can_use_ignore_related_perms(namespace_scoped_perm_ctx, True)
        except PermissionDeniedError as e:
            messages.append(e.message)
            action_request_list.extend(e.action_request_list)
            force_raise = True

    try:
        cluster_scoped_perm_ctx = ClusterScopedPermCtx(
            username=perm_ctx.username,
            project_id=perm_ctx.project_id,
            cluster_id=perm_ctx.cluster_id,
            force_raise=force_raise,
        )
        ClusterScopedPermission().can_use_ignore_related_perms(cluster_scoped_perm_ctx, True)
    except PermissionDeniedError as e:
        messages.append(e.message)
        action_request_list.extend(e.action_request_list)

    # 有权限，直接返回 True
    if not messages:
        return True

    if raise_exception:
        raise PermissionDeniedError(
            message=';'.join(messages), username=perm_ctx.username, action_request_list=action_request_list
        )

    return False
