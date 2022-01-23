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
import pytest
from django.conf import settings

from backend.iam.permissions.exceptions import PermissionDeniedError
from backend.iam.permissions.perm import ActionResourcesRequest
from backend.iam.permissions.resources.constants import ResourceType
from backend.iam.permissions.resources.project import ProjectAction, ProjectCreatorAction, ProjectPermCtx
from backend.tests.iam.conftest import generate_apply_url

from . import roles

pytestmark = pytest.mark.django_db


class TestProjectPermission:
    """
    测试项目资源权限
    note: 仅测试 project_create 和 project_view 两种代表性的权限，其他操作权限逻辑重复
    """

    def test_can_create(self, project_permission_obj):
        """测试场景：有项目创建权限"""
        perm_ctx = ProjectPermCtx(username=roles.ADMIN_USER)
        assert project_permission_obj.can_create(perm_ctx)

    def test_can_not_create(self, project_permission_obj):
        """测试场景：无项目创建权限"""

        # 无权限不抛出异常
        username = roles.NO_PROJECT_USER
        perm_ctx = ProjectPermCtx(username=username)
        assert not project_permission_obj.can_create(perm_ctx, raise_exception=False)

        # 无权限抛出异常
        with pytest.raises(PermissionDeniedError) as exec:
            project_permission_obj.can_create(perm_ctx)
        assert exec.value.code == PermissionDeniedError.code
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            action_request_list=[ActionResourcesRequest(ProjectAction.CREATE, resource_type=ResourceType.Project)],
        )

    def test_can_view(self, project_permission_obj, project_id):
        """测试场景：有项目查看权限"""
        perm_ctx = ProjectPermCtx(username=roles.ADMIN_USER, project_id=project_id)
        assert project_permission_obj.can_view(perm_ctx)

    def test_can_not_view(self, project_permission_obj, project_id):
        """测试场景：无项目查看权限"""

        # 无权限不抛出异常
        username = roles.NO_PROJECT_USER
        perm_ctx = ProjectPermCtx(username=username, project_id=project_id)
        assert not project_permission_obj.can_view(perm_ctx, raise_exception=False)

        # 无权限抛出异常
        with pytest.raises(PermissionDeniedError) as exec:
            project_permission_obj.can_view(perm_ctx)
        assert exec.value.code == PermissionDeniedError.code
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    ProjectAction.VIEW,
                    resource_type=ResourceType.Project,
                    resources=[project_id],
                )
            ],
        )

    def test_can_edit_not_view(self, project_permission_obj, project_id):
        """测试场景：有项目编辑权限(同时无项目查看权限)"""
        username = roles.PROJECT_NO_VIEW_USER
        perm_ctx = ProjectPermCtx(username=username, project_id=project_id)
        with pytest.raises(PermissionDeniedError) as exec:
            project_permission_obj.can_edit(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    ProjectAction.VIEW,
                    resource_type=ResourceType.Project,
                    resources=[project_id],
                ),
            ],
        )


class TestProjectCreatorAction:
    def test_to_data(self, bk_user, project_id):
        action = ProjectCreatorAction(bk_user.username, project_id=project_id, name=project_id)
        assert action.to_data() == {
            'id': project_id,
            'name': project_id,
            'creator': bk_user.username,
            'type': ResourceType.Project,
            'system': settings.BK_IAM_SYSTEM_ID,
        }
