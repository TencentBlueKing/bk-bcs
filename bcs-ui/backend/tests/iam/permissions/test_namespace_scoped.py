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
from backend.iam.permissions.resources import NamespaceScopedPermCtx
from backend.iam.permissions.resources.cluster import ClusterAction
from backend.iam.permissions.resources.constants import ResourceType
from backend.iam.permissions.resources.namespace import calc_iam_ns_id
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedAction
from backend.iam.permissions.resources.project import ProjectAction
from backend.tests.iam.conftest import generate_apply_url

from . import roles


class TestNamespaceScopedPermission:
    """
    命名空间域资源权限
    """

    def test_can_update(self, namespace_scoped_permission_obj, project_id, cluster_id, namespace_name):
        """测试场景：有命名空间域资源创建权限(同时有集群/项目查看权限)"""
        perm_ctx = NamespaceScopedPermCtx(
            username=roles.ADMIN_USER, project_id=project_id, cluster_id=cluster_id, name=namespace_name
        )
        assert namespace_scoped_permission_obj.can_update(perm_ctx)

    def test_can_update_but_not_view(self, namespace_scoped_permission_obj, project_id, cluster_id, namespace_name):
        """测试场景：有命名空间域更新但是无命名空间域查看"""
        perm_ctx = NamespaceScopedPermCtx(
            username=roles.NAMESPACE_SCOPED_NO_VIEW_USER,
            project_id=project_id,
            cluster_id=cluster_id,
            name=namespace_name,
        )
        with pytest.raises(PermissionDeniedError) as exec:
            namespace_scoped_permission_obj.can_update(perm_ctx)

        iam_ns_id = calc_iam_ns_id(cluster_id, namespace_name)
        assert exec.value.data['apply_url'] == generate_apply_url(
            roles.NAMESPACE_SCOPED_NO_VIEW_USER,
            [
                ActionResourcesRequest(
                    NamespaceScopedAction.VIEW,
                    resource_type=ResourceType.Namespace,
                    resources=[iam_ns_id],
                    parent_chain=[
                        IAMResource(ResourceType.Project, project_id),
                        IAMResource(ResourceType.Cluster, cluster_id),
                    ],
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

    def test_can_use(self, namespace_scoped_permission_obj, project_id, cluster_id, namespace_name):
        """测试场景：有命名空间域资源使用权限(同时有集群/项目查看权限)"""
        perm_ctx = NamespaceScopedPermCtx(
            username=roles.ADMIN_USER, project_id=project_id, cluster_id=cluster_id, name=namespace_name
        )
        assert namespace_scoped_permission_obj.can_use(perm_ctx)
