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
from typing import Dict, List, Optional, Type

import attr

from backend.iam.permissions.exceptions import AttrValidationError
from backend.iam.permissions.perm import PermCtx, Permission
from backend.iam.permissions.request import IAMResource, ResourceRequest
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum

from .cluster import ClusterPermission, related_cluster_perm
from .constants import ResourceType
from .namespace import NamespaceAction, NamespaceRequest, calc_iam_ns_id


class NamespaceScopedAction(str, StructuredEnum):
    CREATE = EnumField('namespace_scoped_create', label='namespace_scoped_create')
    VIEW = EnumField('namespace_scoped_view', label='namespace_scoped_view')
    UPDATE = EnumField('namespace_scoped_update', label='namespace_scoped_update')
    DELETE = EnumField('namespace_scoped_delete', label='namespace_scoped_delete')


@attr.dataclass
class NamespaceScopedPermCtx(PermCtx):
    project_id: str = ''
    cluster_id: str = ''
    name: str = ''  # 命名空间名
    iam_ns_id: Optional[str] = None  # 注册到权限中心的命名空间ID

    def __attrs_post_init__(self):
        """权限中心的 resource_id 长度限制为32位"""
        if self.name:
            self.iam_ns_id = calc_iam_ns_id(self.cluster_id, self.name)

    @property
    def resource_id(self) -> str:
        return self.iam_ns_id

    def validate(self):
        super().validate()
        if not self.project_id:
            raise AttrValidationError('project_id must not be empty')
        if not self.cluster_id:
            raise AttrValidationError('cluster_id must not be empty')
        if not self.name:
            raise AttrValidationError('name must not be empty')

    def to_request_attrs(self) -> Dict[str, str]:
        return {'project_id': self.project_id, 'cluster_id': self.cluster_id}


class NamespaceScopedPermission(Permission):
    """命名空间域资源权限控制"""

    resource_type: str = ResourceType.Namespace
    resource_request_cls: Type[ResourceRequest] = NamespaceRequest
    parent_res_perm = ClusterPermission()

    @related_cluster_perm(method_name='can_view')
    def can_create(self, perm_ctx: NamespaceScopedPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [NamespaceScopedAction.CREATE, NamespaceAction.VIEW], raise_exception)

    @related_cluster_perm(method_name='can_view')
    def can_view(self, perm_ctx: NamespaceScopedPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [NamespaceScopedAction.VIEW, NamespaceAction.VIEW], raise_exception)

    @related_cluster_perm(method_name='can_view')
    def can_update(self, perm_ctx: NamespaceScopedPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(
            perm_ctx, [NamespaceScopedAction.UPDATE, NamespaceScopedAction.VIEW, NamespaceAction.VIEW], raise_exception
        )

    @related_cluster_perm(method_name='can_view')
    def can_delete(self, perm_ctx: NamespaceScopedPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(
            perm_ctx, [NamespaceScopedAction.DELETE, NamespaceScopedAction.VIEW, NamespaceAction.VIEW], raise_exception
        )

    @related_cluster_perm(method_name='can_view')
    def can_use(self, perm_ctx: NamespaceScopedPermCtx, raise_exception: bool = True) -> bool:
        """与 can_use_ignore_related_perms 方法的区别是校验上级资源"""
        return self.can_use_ignore_related_perms(perm_ctx, raise_exception)

    def can_use_ignore_related_perms(self, perm_ctx: NamespaceScopedPermCtx, raise_exception: bool = True) -> bool:
        """use 表示 create、update、view、delete 操作的集合，未校验上级资源"""
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(
            perm_ctx,
            [
                NamespaceScopedAction.CREATE,
                NamespaceScopedAction.VIEW,
                NamespaceScopedAction.UPDATE,
                NamespaceScopedAction.DELETE,
                NamespaceAction.VIEW,
            ],
            raise_exception,
        )

    def get_parent_chain(self, perm_ctx: NamespaceScopedPermCtx) -> List[IAMResource]:
        return [
            IAMResource(ResourceType.Project, perm_ctx.project_id),
            IAMResource(ResourceType.Cluster, perm_ctx.cluster_id),
        ]

    def get_resource_id(self, perm_ctx: NamespaceScopedPermCtx) -> Optional[str]:
        return perm_ctx.iam_ns_id
