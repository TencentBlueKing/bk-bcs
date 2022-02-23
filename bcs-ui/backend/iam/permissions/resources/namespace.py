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

from backend.iam.permissions import decorators
from backend.iam.permissions.exceptions import AttrValidationError
from backend.iam.permissions.perm import PermCtx, Permission, ResCreatorAction
from backend.iam.permissions.request import IAMResource, ResourceRequest
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum
from backend.utils.basic import md5_digest

from .cluster import ClusterPermission, related_cluster_perm
from .constants import ResourceType


def calc_iam_ns_id(cluster_id: str, name: str) -> Optional[str]:
    """
    计算(压缩)出注册到权限中心的命名空间 ID，具备唯一性. 当前的算法并不能完全避免冲突，但概率较低。
    note: 权限中心对资源 ID 有长度限制，不超过32位。长度越长，处理性能越低

    :param cluster_id: 集群 ID
    :param name: 命名空间名，k8s 限定最长63个字符
    :return: iam_ns_id 是命名空间注册到权限中心的资源 ID，它是对结构`集群ID:命名空间name`的一个压缩，
             如 `BCS-K8S-40000:default` 会被处理成 `40000:5f03d33dde`。其中，保留集群数字 ID 的目的是用于
             NamespaceProvider 中的 fetch_instance_info 方法
    """
    cluster_idx = cluster_id.split('-')[-1]
    iam_ns_id = f'{cluster_idx}:{md5_digest(name)[8:16]}{name[:2]}'
    if len(iam_ns_id) > 32:
        raise ValueError(f'iam_ns_id({iam_ns_id}) more than 32 characters')
    return iam_ns_id


class NamespaceAction(str, StructuredEnum):
    CREATE = EnumField('namespace_create', label='namespace_create')
    VIEW = EnumField('namespace_view', label='namespace_view')
    UPDATE = EnumField('namespace_update', label='namespace_update')
    DELETE = EnumField('namespace_delete', label='namespace_delete')


@attr.dataclass
class NamespaceCreatorAction(ResCreatorAction):
    cluster_id: str
    name: str
    resource_type: str = ResourceType.Namespace

    def to_data(self) -> Dict:
        data = super().to_data()
        return {
            'id': calc_iam_ns_id(self.cluster_id, self.name),
            'name': self.name,
            'ancestors': [
                {'system': self.system, 'type': ResourceType.Project, 'id': self.project_id},
                {'system': self.system, 'type': ResourceType.Cluster, 'id': self.cluster_id},
            ],
            **data,
        }


@attr.dataclass
class NamespacePermCtx(PermCtx):
    project_id: str = ''
    cluster_id: str = ''
    name: Optional[str] = None  # 命名空间名
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


class NamespaceRequest(ResourceRequest):
    resource_type: str = ResourceType.Namespace
    attr = {'_bk_iam_path_': f'/project,{{project_id}}/cluster,{{cluster_id}}/'}

    def _make_attribute(self, res_id: str) -> Dict:
        return {
            '_bk_iam_path_': self.attr['_bk_iam_path_'].format(
                project_id=self.attr_kwargs['project_id'], cluster_id=self.attr_kwargs['cluster_id']
            )
        }

    def _validate_attr_kwargs(self):
        if not self.attr_kwargs.get('project_id'):
            raise AttrValidationError('missing project_id or project_id is invalid')

        if not self.attr_kwargs.get('cluster_id'):
            raise AttrValidationError('missing cluster_id or cluster_id is invalid')


class related_namespace_perm(decorators.RelatedPermission):

    module_name: str = ResourceType.Namespace

    def _convert_perm_ctx(self, instance, args, kwargs) -> PermCtx:
        """仅支持第一个参数是 PermCtx 子类实例"""
        if len(args) <= 0:
            raise TypeError('missing NamespacePermCtx instance argument')
        if isinstance(args[0], PermCtx):
            return NamespacePermCtx(
                username=args[0].username,
                project_id=args[0].project_id,
                cluster_id=args[0].cluster_id,
                name=args[0].name,
            )
        else:
            raise TypeError('missing NamespacePermCtx instance argument')


class namespace_perm(decorators.Permission):
    module_name: str = ResourceType.Namespace


class NamespacePermission(Permission):
    """命名空间权限"""

    resource_type: str = ResourceType.Namespace
    resource_request_cls: Type[ResourceRequest] = NamespaceRequest
    parent_res_perm = ClusterPermission()

    @related_cluster_perm(method_name='can_view')
    def can_create(self, perm_ctx: NamespacePermCtx, raise_exception: bool = True) -> bool:
        return self.can_action(perm_ctx, NamespaceAction.CREATE, raise_exception)

    @related_cluster_perm(method_name='can_view')
    def can_view(self, perm_ctx: NamespacePermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_action(perm_ctx, NamespaceAction.VIEW, raise_exception)

    @related_cluster_perm(method_name='can_view')
    def can_update(self, perm_ctx: NamespacePermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [NamespaceAction.UPDATE, NamespaceAction.VIEW], raise_exception)

    @related_cluster_perm(method_name='can_view')
    def can_delete(self, perm_ctx: NamespacePermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [NamespaceAction.DELETE, NamespaceAction.VIEW], raise_exception)

    def make_res_request(self, res_id: str, perm_ctx: NamespacePermCtx) -> ResourceRequest:
        return self.resource_request_cls(res_id, project_id=perm_ctx.project_id, cluster_id=perm_ctx.cluster_id)

    def get_parent_chain(self, perm_ctx: NamespacePermCtx) -> List[IAMResource]:
        return [
            IAMResource(ResourceType.Project, perm_ctx.project_id),
            IAMResource(ResourceType.Cluster, perm_ctx.cluster_id),
        ]

    def get_resource_id(self, perm_ctx: NamespacePermCtx) -> Optional[str]:
        return perm_ctx.iam_ns_id
