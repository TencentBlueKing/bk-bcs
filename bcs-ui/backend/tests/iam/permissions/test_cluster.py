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
from backend.iam.permissions.request import ActionResourcesRequest, IAMResource
from backend.iam.permissions.resources.cluster import ClusterAction, ClusterCreatorAction, ClusterPermCtx, cluster_perm
from backend.iam.permissions.resources.constants import ResourceType
from backend.iam.permissions.resources.project import ProjectAction
from backend.tests.iam.conftest import generate_apply_url

from . import roles

pytestmark = pytest.mark.django_db


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
                    resource_type=ResourceType.Project,
                    resources=[project_id],
                ),
                ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id]),
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
                    resource_type=ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id]),
            ],
        )

    def test_can_view_but_no_project(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：有集群查看权限"""
        perm_ctx = ClusterPermCtx(username=roles.CLUSTER_NO_PROJECT_USER, project_id=project_id, cluster_id=cluster_id)
        assert cluster_permission_obj.can_view(perm_ctx)

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
                    resource_type=ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ProjectAction.VIEW,
                    resource_type=ResourceType.Project,
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
                    resource_type=ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    resource_type=ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id]),
            ],
        )

    def test_can_manage_but_no_project(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：有集群管理权限(但是无项目权限)"""
        username = roles.CLUSTER_NO_PROJECT_USER
        perm_ctx = ClusterPermCtx(username=username, project_id=project_id, cluster_id=cluster_id)
        with pytest.raises(PermissionDeniedError) as exec:
            cluster_permission_obj.can_manage(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id])],
        )

    def test_can_manage_but_no_view(self, cluster_permission_obj, project_id, cluster_id):
        """测试场景：有集群管理权限(但是无集群查看权限)"""
        username = roles.CLUSTER_MANAGE_NOT_VIEW_USER
        perm_ctx = ClusterPermCtx(username=username, project_id=project_id, cluster_id=cluster_id)
        with pytest.raises(PermissionDeniedError) as exec:
            cluster_permission_obj.can_manage(perm_ctx)
        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    resource_type=ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id]),
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
                    resource_type=ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    resource_type=ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id]),
            ],
        )


class TestClusterCreatorAction:
    def test_to_data(self, bk_user, project_id, cluster_id):
        action = ClusterCreatorAction(
            creator=bk_user.username, project_id=project_id, cluster_id=cluster_id, name=cluster_id
        )
        assert action.to_data() == {
            'id': cluster_id,
            'name': cluster_id,
            'type': ResourceType.Cluster,
            'system': settings.BK_IAM_SYSTEM_ID,
            'creator': bk_user.username,
            'ancestors': [{'system': settings.BK_IAM_SYSTEM_ID, 'type': ResourceType.Project, 'id': project_id}],
        }
