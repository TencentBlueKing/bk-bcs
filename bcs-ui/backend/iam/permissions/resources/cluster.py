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

from backend.iam.permissions import decorators
from backend.iam.permissions.perm import PermCtx, Permission, ResCreatorAction, validate_empty
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


@attr.s
class ClusterPermCtx(PermCtx):
    project_id = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])
    cluster_id = attr.ib(validator=attr.validators.instance_of(str), default='')

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'ClusterPermCtx':
        return cls(
            username=init_data['username'],
            force_raise=init_data.get('force_raise', False),
            project_id=init_data['project_id'],
            cluster_id=init_data.get('cluster_id', ''),
        )

    @property
    def resource_id(self) -> str:
        return self.cluster_id

    def get_parent_chain(self) -> List[IAMResource]:
        return [IAMResource(ResourceType.Project, self.project_id)]


@attr.s
class ClusterRequest(ResourceRequest):
    project_id = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])
    resource_type = attr.ib(init=False, default=ResourceType.Cluster)
    request_attrs = attr.ib(init=False, default={'_bk_iam_path_': f'/project,{{project_id}}/'})

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'ClusterRequest':
        """从字典构建对象"""
        return cls(project_id=init_data['project_id'])

    def _make_attribute(self, res_id: str) -> Dict:
        return {'_bk_iam_path_': self.request_attrs['_bk_iam_path_'].format(project_id=self.project_id)}


class related_cluster_perm(decorators.RelatedPermission):
    module_name: str = ResourceType.Cluster


class cluster_perm(decorators.Permission):
    module_name: str = ResourceType.Cluster


class ClusterPermission(Permission):
    """集群权限"""

    resource_type: str = ResourceType.Cluster
    resource_request_cls: Type[ResourceRequest] = ClusterRequest
    perm_ctx_cls = ClusterPermCtx
    parent_res_perm = ProjectPermission()

    @related_project_perm(method_name='can_view')
    def can_create(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True) -> bool:
        return self.can_action(perm_ctx, ClusterAction.CREATE, raise_exception)

    @related_project_perm(method_name='can_view')
    def can_view(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_action(perm_ctx, ClusterAction.VIEW, raise_exception, use_cache=True)

    @related_project_perm(method_name='can_view')
    def can_manage(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [ClusterAction.MANAGE, ClusterAction.VIEW], raise_exception)

    @related_project_perm(method_name='can_view')
    def can_delete(self, perm_ctx: ClusterPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [ClusterAction.DELETE, ClusterAction.VIEW], raise_exception)
