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
import mock
import pytest
from django.conf import settings

from backend.iam.permissions.exceptions import PermissionDeniedError
from backend.iam.permissions.request import ActionResourcesRequest, IAMResource
from backend.iam.permissions.resources.cluster import ClusterAction
from backend.iam.permissions.resources.constants import ResourceType
from backend.iam.permissions.resources.project import ProjectAction
from backend.iam.permissions.resources.templateset import (
    TemplatesetAction,
    TemplatesetCreatorAction,
    TemplatesetPermCtx,
    templateset_perm,
)
from backend.tests.iam.conftest import generate_apply_url

from . import roles

pytestmark = pytest.mark.django_db


@pytest.fixture(autouse=True)
def patch_can_apply_in_cluster(project_id, cluster_id):
    with mock.patch(
        'backend.iam.permissions.resources.project_scoped.can_apply_in_cluster',
        side_effect=PermissionDeniedError(
            username=roles.PROJECT_TEMPLATESET_USER,
            action_request_list=[
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
            ],
        ),
    ):
        yield


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
            [ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id])],
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
                    resource_type=ResourceType.Templateset,
                    resources=[template_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    TemplatesetAction.VIEW,
                    resource_type=ResourceType.Templateset,
                    resources=[template_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id]),
            ],
        )

    def test_can_not_instantiate_in_cluster(
        self, templateset_permission_obj, project_id, template_id, cluster_id, namespace
    ):
        """测试场景：有模板集实例化权限(但是无实例化到集群中权限)"""
        username = roles.PROJECT_TEMPLATESET_USER
        perm_ctx = TemplatesetPermCtx(username=username, project_id=project_id, template_id=template_id)
        with pytest.raises(PermissionDeniedError) as exec:
            templateset_permission_obj.can_instantiate_in_cluster(perm_ctx, cluster_id, namespace)

        assert exec.value.data['apply_url'] == generate_apply_url(
            username,
            [
                ActionResourcesRequest(
                    ClusterAction.VIEW,
                    ResourceType.Cluster,
                    resources=[cluster_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
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
                    resource_type=ResourceType.Templateset,
                    resources=[template_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(
                    TemplatesetAction.VIEW,
                    resource_type=ResourceType.Templateset,
                    resources=[template_id],
                    parent_chain=[IAMResource(ResourceType.Project, project_id)],
                ),
                ActionResourcesRequest(ProjectAction.VIEW, resource_type=ResourceType.Project, resources=[project_id]),
            ],
        )


class TestTemplatesetCreatorAction:
    def test_to_data(self, bk_user, project_id, form_template):
        action = TemplatesetCreatorAction(
            creator=bk_user.username, project_id=project_id, template_id=form_template.id, name=form_template.name
        )
        assert action.to_data() == {
            'id': str(form_template.id),
            'name': form_template.name,
            'type': ResourceType.Templateset,
            'system': settings.BK_IAM_SYSTEM_ID,
            'creator': bk_user.username,
            'ancestors': [
                {'system': settings.BK_IAM_SYSTEM_ID, 'type': ResourceType.Project, 'id': project_id},
            ],
        }
