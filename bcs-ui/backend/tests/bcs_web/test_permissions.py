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
from typing import Dict

import mock
import pytest
from rest_framework import viewsets
from rest_framework.response import Response
from rest_framework.test import APIRequestFactory, force_authenticate

from backend.bcs_web.permissions import AccessProjectPermission, ProjectEnableBCS
from backend.tests.bcs_mocks.misc import FakeProjectPermissionAllowAll
from backend.tests.testing_utils.base import generate_random_string
from backend.tests.testing_utils.mocks.paas_cc import StubPaaSCCClient
from backend.tests.testing_utils.mocks.utils import mockable_function
from backend.utils.cache import region

pytestmark = pytest.mark.django_db

factory = APIRequestFactory()

HAS_PERM_PROJECT_ID = generate_random_string(32)


class FakeProjectPermission(FakeProjectPermissionAllowAll):
    """对于 project_id = HAS_PERM_PROJECT_ID, 权限放行"""

    def can_view(self, perm_ctx, raise_exception=False):
        if perm_ctx.project_id == HAS_PERM_PROJECT_ID:
            return True
        return False


class FakePaaSCCClient(StubPaaSCCClient):
    """project_id != HAS_PERM_PROJECT_ID 时, 设置项目未绑定cc业务，表示未开启容器服务"""

    @mockable_function
    def get_project(self, project_id: str) -> Dict:
        p = self.make_project_data(project_id)
        if project_id != HAS_PERM_PROJECT_ID:
            p["cc_app_id"] = 0
        return p


@pytest.fixture(autouse=True)
def patch_permissions():
    """Patch permission checks to allow API requests, includes:

    - paas_cc module: return faked project infos
    - get_api_public_key: return None
    - ProjectPermission.can_view: return True if is mocked project id else False
    """
    with mock.patch('backend.bcs_web.permissions.PaaSCCClient', new=FakePaaSCCClient), mock.patch(
        'backend.components.apigw.get_api_public_key', return_value=None
    ), mock.patch('backend.bcs_web.permissions.ProjectPermission', new=FakeProjectPermission):
        yield


class AccessProjectView(viewsets.ViewSet):
    permission_classes = (AccessProjectPermission,)

    def get(self, request, project_id):
        return Response({"project_id": project_id})


class ProjectEnableBCSView(viewsets.ViewSet):
    permission_classes = (ProjectEnableBCS,)

    def get(self, request, project_id):
        return Response({"project_id": request.project.project_id})


class TestCustomPermissions:
    """
    测试自定义 Permission， 参考 https://github.com/encode/django-rest-framework/blob/master/tests/test_permissions.py
    """

    def test_access_project_permission(self, bk_user, project_id):
        request = factory.get('/1', format='json')
        force_authenticate(request, user=bk_user)

        p_view = AccessProjectView.as_view({'get': 'get'})

        # 无权限的项目
        response = p_view(request, project_id=project_id)
        assert response.data.get('message') == "no project permissions"
        # 有权限的项目
        response = p_view(request, project_id=HAS_PERM_PROJECT_ID)
        assert response.data.get('project_id') == HAS_PERM_PROJECT_ID
        assert region.get(f'BK_DEVOPS_BCS:PROJECT_ID:{HAS_PERM_PROJECT_ID}') == HAS_PERM_PROJECT_ID

    def test_project_has_bcs(self, bk_user):
        request = factory.get('/1', format='json')
        force_authenticate(request, user=bk_user)

        p_view = ProjectEnableBCSView.as_view({'get': 'get'})

        # 未启用BCS的项目
        response = p_view(request, project_id=generate_random_string(32))
        assert response.data.get('message') == "project does not enable bcs"
        # 启用BCS的项目
        response = p_view(request, project_id=HAS_PERM_PROJECT_ID)
        assert response.data.get('project_id') == HAS_PERM_PROJECT_ID
        assert region.get(f'BK_DEVOPS_BCS:ENABLED_BCS_PROJECT:{HAS_PERM_PROJECT_ID}').project_id == HAS_PERM_PROJECT_ID
