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

from .constants import ResourceType
from .project import ProjectPermission, related_project_perm


class ClusterAction(str, StructuredEnum):
    CREATE = EnumField('cluster_create', label='cluster_create')
    VIEW = EnumField('cluster_view', label='cluster_view')
    MANAGE = EnumField('cluster_manage', label='cluster_manage')
    DELETE = EnumField('cluster_delete', label='cluster_delete')


@attr.dataclass
class ClusterCreatorAction(ResCreatorAction):
    cluster_id: str
    name: str
    resource_type: str = ResourceType.Cluster

    def to_data(self) -> Dict:
        data = super().to_data()
        return {
            'id': self.cluster_id,
            'name': self.name,
            'ancestors': [{'system': self.system, 'type': ResourceType.Project, 'id': self.project_id}],
            **data,
        }


@attr.dataclass
class ClusterPermCtx(PermCtx):
    project_id: str = ''
    cluster_id: Optional[str] = None

    @property
    def resource_id(self) -> str:
        return self.cluster_id

    def validate(self):
        super().validate()
        if not self.project_id:
            raise AttrValidationError('project_id must not be empty')


class ClusterRequest(ResourceRequest):
    resource_type: str = ResourceType.Cluster
    attr = {'_bk_iam_path_': f'/project,{{project_id}}/'}

    def _make_attribute(self, res_id: str) -> Dict:
        return {'_bk_iam_path_': self.attr['_bk_iam_path_'].format(project_id=self.attr_kwargs['project_id'])}

    def _validate_attr_kwargs(self):
        if not self.attr_kwargs.get('project_id'):
            raise AttrValidationError('missing project_id or project_id is invalid')


class related_cluster_perm(decorators.RelatedPermission):

    module_name: str = ResourceType.Cluster

    def _convert_perm_ctx(self, instance, args, kwargs) -> PermCtx:
        """仅支持第一个参数是 PermCtx 子类实例"""
        if len(args) <= 0:
            raise TypeError('missing ClusterPermCtx instance argument')
        if isinstance(args[0], PermCtx):
            return ClusterPermCtx(
                username=args[0].username, project_id=args[0].project_id, cluster_id=args[0].cluster_id
            )
        else:
            raise TypeError('missing ClusterPermCtx instance argument')


class cluster_perm(decorators.Permission):
    module_name: str = ResourceType.Cluster


class ClusterPermission(Permission):
    """集群权限"""

    resource_type: str = ResourceType.Cluster
    resource_request_cls: Type[ResourceRequest] = ClusterRequest
    parent_res_perm = ProjectPermission()

    @related_project_perm(method_name='can_view')
    def can_create(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True) -> bool:
        return self.can_action(perm_ctx, ClusterAction.CREATE, raise_exception)

    @related_project_perm(method_name='can_view')
    def can_view(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_action(perm_ctx, ClusterAction.VIEW, raise_exception)

    @related_project_perm(method_name='can_view')
    def can_manage(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [ClusterAction.MANAGE, ClusterAction.VIEW], raise_exception)

    @related_project_perm(method_name='can_view')
    def can_delete(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [ClusterAction.DELETE, ClusterAction.VIEW], raise_exception)

    def make_res_request(self, res_id: str, perm_ctx: ClusterPermCtx) -> ResourceRequest:
        return self.resource_request_cls(res_id, project_id=perm_ctx.project_id)

    def get_parent_chain(self, perm_ctx: ClusterPermCtx) -> List[IAMResource]:
        return [IAMResource(ResourceType.Project, perm_ctx.project_id)]

    def get_resource_id(self, perm_ctx: ClusterPermCtx) -> Optional[str]:
        return perm_ctx.cluster_id
