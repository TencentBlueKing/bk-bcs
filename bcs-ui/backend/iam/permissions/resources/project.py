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
from typing import Dict, Type

import attr

from backend.iam.permissions import decorators
from backend.iam.permissions.perm import PermCtx, Permission, ResCreatorAction
from backend.iam.permissions.request import ResourceRequest
from backend.packages.blue_krill.data_types.enum import EnumField, StructuredEnum

from .constants import ResourceType


class ProjectAction(str, StructuredEnum):
    CREATE = EnumField('project_create', label='project_create')
    VIEW = EnumField('project_view', label='project_view')
    EDIT = EnumField('project_edit', label='project_edit')


@attr.s
class ProjectRequest(ResourceRequest):
    resource_type = attr.ib(init=False, default=ResourceType.Project)

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'ProjectRequest':
        """从字典构建对象"""
        return cls()


@attr.dataclass
class ProjectCreatorAction(ResCreatorAction):
    name: str
    resource_type: str = ResourceType.Project

    def to_data(self) -> Dict:
        data = super().to_data()
        return {'id': self.project_id, 'name': self.name, **data}


@attr.s
class ProjectPermCtx(PermCtx):
    project_id = attr.ib(validator=attr.validators.instance_of(str), default='')

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'ProjectPermCtx':
        return cls(
            username=init_data['username'],
            force_raise=init_data.get('force_raise', False),
            project_id=init_data.get('project_id', ''),
        )

    @property
    def resource_id(self) -> str:
        return self.project_id


class related_project_perm(decorators.RelatedPermission):
    module_name: str = ResourceType.Project


class ProjectPermission(Permission):
    """项目权限"""

    resource_type: str = ResourceType.Project
    resource_request_cls: Type[ResourceRequest] = ProjectRequest
    perm_ctx_cls = ProjectPermCtx

    def can_create(self, perm_ctx: ProjectPermCtx, raise_exception: bool = True) -> bool:
        return self.can_action(perm_ctx, ProjectAction.CREATE, raise_exception)

    def can_view(self, perm_ctx: ProjectPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_action(perm_ctx, ProjectAction.VIEW, raise_exception, use_cache=True)

    def can_edit(self, perm_ctx: ProjectPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [ProjectAction.EDIT, ProjectAction.VIEW], raise_exception)
