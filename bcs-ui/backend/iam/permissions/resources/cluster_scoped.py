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
from typing import Dict, List, Type

import attr

from backend.iam.permissions.perm import PermCtx, Permission, validate_empty
from backend.iam.permissions.request import IAMResource, ResourceRequest
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum

from .cluster import ClusterAction, ClusterRequest
from .constants import ResourceType
from .project import ProjectPermission, related_project_perm


class ClusterScopedAction(str, StructuredEnum):
    CREATE = EnumField('cluster_scoped_create', label='cluster_scoped_create')
    VIEW = EnumField('cluster_scoped_view', label='cluster_scoped_view')
    UPDATE = EnumField('cluster_scoped_update', label='cluster_scoped_update')
    DELETE = EnumField('cluster_scoped_delete', label='cluster_scoped_delete')


@attr.s
class ClusterScopedPermCtx(PermCtx):
    project_id = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])
    cluster_id = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'ClusterScopedPermCtx':
        return cls(
            username=init_data['username'],
            force_raise=init_data['force_raise'],
            project_id=init_data['project_id'],
            cluster_id=init_data['cluster_id'],
        )

    @property
    def resource_id(self) -> str:
        return self.cluster_id

    def get_parent_chain(self) -> List[IAMResource]:
        return [IAMResource(ResourceType.Project, self.project_id)]


class ClusterScopedPermission(Permission):
    """集群域资源权限控制"""

    resource_type: str = ResourceType.Cluster
    resource_request_cls: Type[ResourceRequest] = ClusterRequest
    perm_ctx_cls = ClusterScopedPermCtx
    parent_res_perm = ProjectPermission()

    @related_project_perm(method_name='can_view')
    def can_create(self, perm_ctx: ClusterScopedPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [ClusterScopedAction.CREATE, ClusterAction.VIEW], raise_exception)

    @related_project_perm(method_name='can_view')
    def can_view(self, perm_ctx: ClusterScopedPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [ClusterScopedAction.VIEW, ClusterAction.VIEW], raise_exception)

    @related_project_perm(method_name='can_view')
    def can_update(self, perm_ctx: ClusterScopedPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(
            perm_ctx, [ClusterScopedAction.UPDATE, ClusterScopedAction.VIEW, ClusterAction.VIEW], raise_exception
        )

    @related_project_perm(method_name='can_view')
    def can_delete(self, perm_ctx: ClusterScopedPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(
            perm_ctx, [ClusterScopedAction.DELETE, ClusterScopedAction.VIEW, ClusterAction.VIEW], raise_exception
        )

    @related_project_perm(method_name='can_view')
    def can_use(self, perm_ctx: ClusterScopedPermCtx, raise_exception: bool = True) -> bool:
        """与 can_use_ignore_related_perms 方法的区别是校验上级资源"""
        return self.can_use_ignore_related_perms(perm_ctx, raise_exception)

    def can_use_ignore_related_perms(self, perm_ctx: ClusterScopedPermCtx, raise_exception: bool = True) -> bool:
        """use 表示 create、update、view、delete 操作的集合，未校验上级资源"""
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(
            perm_ctx,
            [
                ClusterScopedAction.CREATE,
                ClusterScopedAction.VIEW,
                ClusterScopedAction.UPDATE,
                ClusterScopedAction.DELETE,
                ClusterAction.VIEW,
            ],
            raise_exception,
        )
