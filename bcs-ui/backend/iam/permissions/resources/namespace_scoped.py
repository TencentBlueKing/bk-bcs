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
from collections import defaultdict
from typing import Dict, List, Type, Union

import attr

from backend.iam.permissions.perm import PermCtx, Permission, validate_empty
from backend.iam.permissions.request import IAMResource, ResourceRequest
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum

from .cluster import ClusterPermission, related_cluster_perm
from .constants import ResourceType
from .namespace import NamespaceAction, NamespaceRequest, calc_iam_ns_id


class NamespaceScopedAction(str, StructuredEnum):
    """
    note: USE 是复合 action, 权限中心未直接注册
    """

    CREATE = EnumField('namespace_scoped_create', label='namespace_scoped_create')
    VIEW = EnumField('namespace_scoped_view', label='namespace_scoped_view')
    UPDATE = EnumField('namespace_scoped_update', label='namespace_scoped_update')
    DELETE = EnumField('namespace_scoped_delete', label='namespace_scoped_delete')
    USE = EnumField('namespace_scoped_use', label='namespace_scoped_use')


@attr.s
class NamespaceScopedPermCtx(PermCtx):
    project_id = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])
    cluster_id = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])
    name = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])  # 命名空间名
    iam_ns_id = attr.ib(init=False)  # 注册到权限中心的命名空间 ID

    def __attrs_post_init__(self):
        """权限中心的 resource_id 长度限制为32位"""
        self.iam_ns_id = calc_iam_ns_id(self.cluster_id, self.name)

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'NamespaceScopedPermCtx':
        return cls(
            username=init_data['username'],
            force_raise=init_data['force_raise'],
            project_id=init_data['project_id'],
            cluster_id=init_data['cluster_id'],
            name=init_data['name'],
        )

    @property
    def resource_id(self) -> str:
        return self.iam_ns_id

    def get_parent_chain(self) -> List[IAMResource]:
        return [
            IAMResource(ResourceType.Project, self.project_id),
            IAMResource(ResourceType.Cluster, self.cluster_id),
        ]


class NamespaceScopedPermission(Permission):
    """命名空间域资源权限控制"""

    resource_type: str = ResourceType.Namespace
    resource_request_cls: Type[ResourceRequest] = NamespaceRequest
    perm_ctx_cls = NamespaceScopedPermCtx
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

    def resources_actions_allowed(
        self, username: str, action_ids: List[str], res_ids: Union[List[str], str], res_request: ResourceRequest
    ):
        """
        note: 在 Permission.resources_actions_allowed 的基础上, 增加对复合操作 NamespaceScopedAction.USE 的支持
        TODO 如果有其他复合操作需要支持, 再抽象
        """
        multi_actions = [
            NamespaceScopedAction.CREATE,
            NamespaceScopedAction.VIEW,
            NamespaceScopedAction.UPDATE,
            NamespaceScopedAction.DELETE,
            NamespaceAction.VIEW,
        ]

        action_list = list(action_ids)
        if NamespaceScopedAction.USE in action_ids:
            action_list.extend(multi_actions)
            action_list = list(set(action_list))
            action_list.remove(NamespaceScopedAction.USE)

        raw_actions_allowed = super().resources_actions_allowed(username, action_list, res_ids, res_request)

        if NamespaceScopedAction.USE not in action_ids:
            return raw_actions_allowed

        # 只返回 action_ids 对应的权限结果
        ns_actions_allowed = defaultdict(dict)
        for iam_ns_id, actions_allowed in raw_actions_allowed.items():
            for action_id in action_ids:
                if action_id == NamespaceScopedAction.USE:
                    # 当 multi_actions 中的 action_id 都有权限时, NamespaceScopedAction.USE 才有权限
                    ns_actions_allowed[iam_ns_id][NamespaceScopedAction.USE] = all(
                        [actions_allowed[action] for action in multi_actions]
                    )
                else:
                    ns_actions_allowed[iam_ns_id][action_id] = actions_allowed[action_id]

        return ns_actions_allowed
