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
from typing import List, Optional, Type

import attr

from backend.iam.permissions.exceptions import AttrValidationError
from backend.iam.permissions.perm import PermCtx, Permission
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


@attr.dataclass
class ClusterScopedPermCtx(PermCtx):
    project_id: str = ''
    cluster_id: str = ''

    @property
    def resource_id(self) -> str:
        return self.cluster_id

    def validate(self):
        super().validate()
        if not self.project_id:
            raise AttrValidationError('project_id must not be empty')
        if not self.cluster_id:
            raise AttrValidationError('cluster_id must not be empty')


class ClusterScopedPermission(Permission):
    """集群域资源权限控制"""

    resource_type: str = ResourceType.Cluster
    resource_request_cls: Type[ResourceRequest] = ClusterRequest
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

    def can_use(self, perm_ctx: ClusterScopedPermCtx, raise_exception: bool = True) -> bool:
        """use 表示 create、update、view操作的集合，不包括 related_actions 的校验"""
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(
            perm_ctx,
            [
                ClusterScopedAction.CREATE,
                ClusterScopedAction.VIEW,
                ClusterScopedAction.UPDATE,
                ClusterAction.VIEW,
            ],
            raise_exception,
        )

    def make_res_request(self, res_id: str, perm_ctx: ClusterScopedPermCtx) -> ResourceRequest:
        return self.resource_request_cls(res_id, project_id=perm_ctx.project_id)

    def get_parent_chain(self, perm_ctx: ClusterScopedPermCtx) -> List[IAMResource]:
        return [IAMResource(ResourceType.Project, perm_ctx.project_id)]

    def get_resource_id(self, perm_ctx: ClusterScopedPermCtx) -> Optional[str]:
        return perm_ctx.cluster_id
