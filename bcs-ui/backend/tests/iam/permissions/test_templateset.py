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
from unittest import mock

import pytest

from backend.iam.permissions.exceptions import PermissionDeniedError
from backend.iam.permissions.request import ActionResourcesRequest, IAMResource
from backend.iam.permissions.resources.constants import ResourceType
from backend.iam.permissions.resources.project import ProjectAction, ProjectPermission
from backend.iam.permissions.resources.templateset import (
    TemplatesetAction,
    TemplatesetPermCtx,
    TemplatesetPermission,
    templateset_perm,
)
from backend.tests.iam.conftest import generate_apply_url

from ..fake_iam import FakeProjectPermission, FakeTemplatesetPermission
from . import roles


@pytest.fixture
def templateset_permission_obj():
    templateset_patcher = mock.patch.object(TemplatesetPermission, '__bases__', (FakeTemplatesetPermission,))
    project_patcher = mock.patch.object(ProjectPermission, '__bases__', (FakeProjectPermission,))
    with templateset_patcher, project_patcher:
        templateset_patcher.is_local = True  # 标注为本地属性，__exit__ 的时候恢复成 patcher.temp_original
        project_patcher.is_local = True
        yield TemplatesetPermission()


class TestTemplatesetPermission:
    """
    模板集资源权限
    note: 仅测试 templateset_instantiate，其他操作权限逻辑重复(和TestClusterPermission类似)
    """

    def test_can_instantiate(self, templateset_permission_obj, project_id, template_id):
        """测试场景：有模板集实例化权限"""
        username = roles.ADMIN_USER
        perm_ctx = TemplatesetPermCtx(username=username, project_id=project_id, template_id=template_id)
        assert templateset_permission_obj.can_instantiate(perm_ctx)

    def test_can_instantiate_but_no_project(self, templateset_permission_obj, project_id, template_id):
        """测试场景：有模板集实例化权限(同时无项目查看权限)"""
        username = roles.TEMPLATESET_NO_PROJECT_USER
        perm_ctx = TemplatesetPermCtx(username=username, project_id=project_id, template_id=template_id)
        with pytest.raises(PermissionDeniedError) as exec:
            templateset_permission_obj.can_instantiate(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    ProjectAction.VIEW, resource_type=ProjectPermission.resource_type, resources=[project_id]
                )
            ],
        )

    def test_can_not_instantiate(self, templateset_permission_obj, project_id, template_id):
        """测试场景：无模板集实例化权限(同时无项目查看权限)"""
        username = roles.ANONYMOUS_USER
        perm_ctx = TemplatesetPermCtx(username=username, project_id=project_id, template_id=template_id)
        with pytest.raises(PermissionDeniedError) as exec:
            templateset_permission_obj.can_instantiate(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    TemplatesetAction.INSTANTIATE,
                    resource_type=TemplatesetPermission.resource_type,
                    resources=[template_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    TemplatesetAction.VIEW,
                    resource_type=TemplatesetPermission.resource_type,
                    resources=[template_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ProjectAction.VIEW, resource_type=ProjectPermission.resource_type, resources=[project_id]
                ),
            ],
        )


@templateset_perm(method_name='can_instantiate')
def instantiate_templateset(perm_ctx: TemplatesetPermCtx):
    """"""


class TestTemplatesetPermDecorator:
    def test_can_instantiate(self, templateset_permission_obj, project_id, template_id):
        """测试场景：有模板集实例化权限"""
        perm_ctx = TemplatesetPermCtx(username=roles.ADMIN_USER, project_id=project_id, template_id=template_id)
        instantiate_templateset(perm_ctx)

    def test_can_not_instantiate(self, templateset_permission_obj, project_id, template_id):
        """测试场景：无模板集实例化权限(同时无项目查看权限)"""
        username = roles.ANONYMOUS_USER
        perm_ctx = TemplatesetPermCtx(username=username, project_id=project_id, template_id=template_id)
        with pytest.raises(PermissionDeniedError) as exec:
            instantiate_templateset(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    TemplatesetAction.INSTANTIATE,
                    resource_type=TemplatesetPermission.resource_type,
                    resources=[template_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    TemplatesetAction.VIEW,
                    resource_type=TemplatesetPermission.resource_type,
                    resources=[template_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ProjectAction.VIEW, resource_type=ProjectPermission.resource_type, resources=[project_id]
                ),
            ],
        )
