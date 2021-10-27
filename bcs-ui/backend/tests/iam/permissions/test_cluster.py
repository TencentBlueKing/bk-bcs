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
from backend.iam.permissions.resources.cluster import ClusterAction, ClusterPermCtx, ClusterPermission, cluster_perm
from backend.iam.permissions.resources.constants import ResourceType
from backend.iam.permissions.resources.project import ProjectAction, ProjectPermission
from backend.tests.iam.conftest import generate_apply_url

from ..fake_iam import FakeClusterPermission, FakeProjectPermission
from . import roles


@pytest.fixture
def cluster_permission_obj():
    cluster_patcher = mock.patch.object(ClusterPermission, '__bases__', (FakeClusterPermission,))
    project_patcher = mock.patch.object(ProjectPermission, '__bases__', (FakeProjectPermission,))
    with cluster_patcher, project_patcher:
        cluster_patcher.is_local = True  # 标注为本地属性，__exit__ 的时候恢复成 patcher.temp_original
        project_patcher.is_local = True
        yield ClusterPermission()


class TestClusterPermission:
    """
    集群资源权限
    note: 仅测试 cluster_create 和 cluster_view 两种代表性的权限，其他操作权限逻辑重复
    """

    def test_can_create(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：有集群创建权限(同时有项目查看权限)"""
        perm_ctx = ClusterPermCtx(username=roles.ADMIN_USER, project_id=project_id)
        assert cluster_permission_obj.can_create(perm_ctx)

    def test_can_not_create(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：无集群创建权限(同时无项目查看权限)"""
        perm_ctx = ClusterPermCtx(username=roles.ANONYMOUS_USER, project_id=project_id)
        with pytest.raises(PermissionDeniedError) as exec:
            cluster_permission_obj.can_create(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            roles.ANONYMOUS_USER,
            [
                ActionResourcesRequest(
                    ClusterAction.CREATE,
                    resource_type=ProjectPermission.resource_type,
                    resources=[project_id],
                ),
                ActionResourcesRequest(
                    ProjectAction.VIEW, resource_type=ProjectPermission.resource_type, resources=[project_id]
                ),
            ],
        )

    def test_can_view(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：有集群查看权限(同时有项目查看权限)"""
        perm_ctx = ClusterPermCtx(username=roles.ADMIN_USER, project_id=project_id, cluster_id=cluster_id)
        assert cluster_permission_obj.can_view(perm_ctx)

    def test_can_not_view_but_project(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：无集群查看权限(同时有项目查看权限)"""
        self._test_can_not_view(
            roles.PROJECT_NO_CLUSTER_USER,
            cluster_permission_obj,
            project_id,
            cluster_id,
            expected_action_list=[
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    resource_type=cluster_permission_obj.resource_type,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ProjectAction.VIEW, resource_type=ProjectPermission.resource_type, resources=[project_id]
                ),
            ],
        )

    def test_can_view_but_no_project(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：有集群查看权限(同时无项目查看权限)"""
        self._test_can_not_view(
            roles.CLUSTER_NO_PROJECT_USER,
            cluster_permission_obj,
            project_id,
            cluster_id,
            expected_action_list=[
                ActionResourcesRequest(
                    ProjectAction.VIEW,
                    resource_type=ProjectPermission.resource_type,
                    resources=[project_id],
                )
            ],
        )

    def test_can_not_view(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：无集群查看权限(同时无项目查看权限)"""
        self._test_can_not_view(
            roles.ANONYMOUS_USER,
            cluster_permission_obj,
            project_id,
            cluster_id,
            expected_action_list=[
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    resource_type=cluster_permission_obj.resource_type,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ProjectAction.VIEW,
                    resource_type=ProjectPermission.resource_type,
                    resources=[project_id],
                ),
            ],
        )

    def _test_can_not_view(self, username, cluster_permission_obj, project_id, cluster_id, expected_action_list):
        perm_ctx = ClusterPermCtx(username=username, project_id=project_id, cluster_id=cluster_id)
        with pytest.raises(PermissionDeniedError) as exec:
            cluster_permission_obj.can_view(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(username, expected_action_list)

    def test_can_not_manage(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：无集群管理权限(同时无项目查看权限)"""
        username = roles.ANONYMOUS_USER
        perm_ctx = ClusterPermCtx(username=username, project_id=project_id, cluster_id=cluster_id)
        with pytest.raises(PermissionDeniedError) as exec:
            cluster_permission_obj.can_manage(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    ClusterAction.MANAGE,
                    resource_type=ClusterPermission.resource_type,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    resource_type=ClusterPermission.resource_type,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ProjectAction.VIEW, resource_type=ProjectPermission.resource_type, resources=[project_id]
                ),
            ],
        )

    def test_can_manage_but_no_project(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：有集群管理权限(同时无项目权限)"""
        username = roles.CLUSTER_NO_PROJECT_USER
        perm_ctx = ClusterPermCtx(username=username, project_id=project_id, cluster_id=cluster_id)
        with pytest.raises(PermissionDeniedError) as exec:
            cluster_permission_obj.can_manage(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    ProjectAction.VIEW, resource_type=ProjectPermission.resource_type, resources=[project_id]
                )
            ],
        )


@cluster_perm(method_name='can_manage')
def manage_cluster(perm_ctx: ClusterPermCtx):
    """"""


class TestClusterPermDecorator:
    def test_can_manage(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：有集群管理权限(同时有项目查看权限)"""
        perm_ctx = ClusterPermCtx(username=roles.ADMIN_USER, project_id=project_id, cluster_id=cluster_id)
        manage_cluster(perm_ctx)

    def test_can_not_manage(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：无集群管理权限(同时无项目查看权限)"""
        username = roles.ANONYMOUS_USER
        perm_ctx = ClusterPermCtx(username=username, project_id=project_id, cluster_id=cluster_id)
        with pytest.raises(PermissionDeniedError) as exec:
            manage_cluster(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    ClusterAction.MANAGE,
                    resource_type=ClusterPermission.resource_type,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    resource_type=ClusterPermission.resource_type,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ProjectAction.VIEW, resource_type=ProjectPermission.resource_type, resources=[project_id]
                ),
            ],
        )
