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
from .namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from .project import ProjectPermission, related_project_perm


class TemplatesetAction(str, StructuredEnum):
    CREATE = EnumField("templateset_create", label="templateset_create")
    VIEW = EnumField("templateset_view", label="templateset_view")
    UPDATE = EnumField("templateset_update", label="templateset_update")
    DELETE = EnumField("templateset_delete", label="templateset_delete")
    INSTANTIATE = EnumField("templateset_instantiate", label="templateset_instantiate")
    COPY = EnumField("templateset_copy", label="templateset_copy")


@attr.dataclass
class TemplatesetCreatorAction(ResCreatorAction):
    template_id: str
    name: str
    resource_type: str = ResourceType.Templateset

    def to_data(self) -> Dict:
        data = super().to_data()
        return {
            'id': str(self.template_id),
            'name': self.name,
            'ancestors': [{'system': self.system, 'type': ResourceType.Project, 'id': self.project_id}],
            **data,
        }


@attr.s
class TemplatesetPermCtx(PermCtx):
    project_id = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])
    template_id = attr.ib(converter=str, default='')

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'TemplatesetPermCtx':
        return cls(
            username=init_data['username'],
            force_raise=init_data.get('force_raise', False),
            project_id=init_data['project_id'],
            template_id=init_data.get('template_id', ''),
        )

    @property
    def resource_id(self) -> str:
        return self.template_id

    def get_parent_chain(self) -> List[IAMResource]:
        return [IAMResource(ResourceType.Project, self.project_id)]


@attr.s
class TemplatesetRequest(ResourceRequest):
    project_id = attr.ib(validator=[attr.validators.instance_of(str), validate_empty])
    resource_type = attr.ib(init=False, default=ResourceType.Templateset)
    request_attrs = attr.ib(init=False, default={'_bk_iam_path_': f'/project,{{project_id}}/'})

    @classmethod
    def from_dict(cls, init_data: Dict) -> 'TemplatesetRequest':
        """从字典构建对象"""
        return cls(project_id=init_data['project_id'])

    def _make_attribute(self, res_id: str) -> Dict:
        return {'_bk_iam_path_': self.request_attrs['_bk_iam_path_'].format(project_id=self.project_id)}


class related_templateset_perm(decorators.RelatedPermission):
    module_name: str = ResourceType.Templateset


class templateset_perm(decorators.Permission):
    module_name: str = ResourceType.Templateset


class TemplatesetPermission(Permission):
    """模板集权限"""

    resource_type: str = ResourceType.Templateset
    resource_request_cls: Type[ResourceRequest] = TemplatesetRequest
    perm_ctx_cls = TemplatesetPermCtx
    parent_res_perm = ProjectPermission()

    @related_project_perm(method_name="can_view")
    def can_create(self, perm_ctx: TemplatesetPermCtx, raise_exception: bool = True) -> bool:
        return self.can_action(perm_ctx, TemplatesetAction.CREATE, raise_exception)

    @related_project_perm(method_name="can_view")
    def can_view(self, perm_ctx: TemplatesetPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_action(perm_ctx, TemplatesetAction.VIEW, raise_exception)

    @related_project_perm(method_name='can_view')
    def can_copy(self, perm_ctx: TemplatesetPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [TemplatesetAction.COPY, TemplatesetAction.VIEW], raise_exception)

    @related_project_perm(method_name='can_view')
    def can_update(self, perm_ctx: TemplatesetPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [TemplatesetAction.UPDATE, TemplatesetAction.VIEW], raise_exception)

    @related_project_perm(method_name='can_view')
    def can_delete(self, perm_ctx: TemplatesetPermCtx, raise_exception: bool = True) -> bool:
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(perm_ctx, [TemplatesetAction.DELETE, TemplatesetAction.VIEW], raise_exception)

    @related_project_perm(method_name='can_view')
    def can_instantiate(self, perm_ctx: TemplatesetPermCtx, raise_exception: bool = True) -> bool:
        """校验是否有实例化操作的权限"""
        perm_ctx.validate_resource_id()
        return self.can_multi_actions(
            perm_ctx, [TemplatesetAction.INSTANTIATE, TemplatesetAction.VIEW], raise_exception
        )

    def can_instantiate_in_ns(
        self, perm_ctx: TemplatesetPermCtx, cluster_id: str, namespace: str, raise_exception: bool = True
    ) -> bool:
        """校验是否有权限实例化到指定命名空间下"""
        self.can_instantiate(perm_ctx, raise_exception)

        namespace_scoped_perm_ctx = NamespaceScopedPermCtx(
            username=perm_ctx.username,
            project_id=perm_ctx.project_id,
            cluster_id=cluster_id,
            name=namespace,
        )
        return NamespaceScopedPermission().can_use(namespace_scoped_perm_ctx, raise_exception)
