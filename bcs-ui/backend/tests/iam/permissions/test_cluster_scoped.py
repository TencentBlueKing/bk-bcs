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

from backend.iam.permissions.exceptions import PermissionDeniedError
from backend.iam.permissions.request import ActionResourcesRequest, IAMResource
from backend.iam.permissions.resources.cluster import ClusterAction
from backend.iam.permissions.resources.cluster_scoped import ClusterScopedPermCtx
from backend.iam.permissions.resources.constants import ResourceType
from backend.iam.permissions.resources.project import ProjectAction
from backend.tests.iam.conftest import generate_apply_url

from . import roles


class TestClusterScopedPermission:
    """
    集群域资源权限
    """

    def test_can_create(self, cluster_scoped_permission_obj, project_id, cluster_id):
        """测试场景：有集群域资源创建权限(同时有集群/项目查看权限)"""
        perm_ctx = ClusterScopedPermCtx(username=roles.ADMIN_USER, project_id=project_id, cluster_id=cluster_id)
        assert cluster_scoped_permission_obj.can_create(perm_ctx)

    def test_can_create_but_no_cluster(self, cluster_scoped_permission_obj, project_id, cluster_id):
        """测试场景：有集群域资源创建权限(但是无集群权限)"""
        perm_ctx = ClusterScopedPermCtx(
            username=roles.CLUSTER_SCOPED_NO_CLUSTER_USER, project_id=project_id, cluster_id=cluster_id
        )
        with pytest.raises(PermissionDeniedError) as exec:
            cluster_scoped_permission_obj.can_create(perm_ctx)
        assert exec.value.data['perms']['apply_url'] == generate_apply_url(
            roles.CLUSTER_SCOPED_NO_CLUSTER_USER,
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

    def test_can_use(self, cluster_scoped_permission_obj, project_id, cluster_id):
        """测试场景：有集群域资源使用权限(同时有集群/项目查看权限)"""
        perm_ctx = ClusterScopedPermCtx(username=roles.ADMIN_USER, project_id=project_id, cluster_id=cluster_id)
        assert cluster_scoped_permission_obj.can_use(perm_ctx)
