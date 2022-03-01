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

from backend.tests.bcs_mocks.misc import FakePaaSCCMod, FakeProjectPermissionAllowAll

pytestmark = pytest.mark.django_db


@pytest.fixture(autouse=True)
def patch_permissions():
    """Patch permission checks to allow API requests, includes:

    - paas_cc module: return faked project infos
    - ProjectPermission: allow all permission checks
    - get_api_public_key: return None
    """
    with mock.patch('backend.utils.permissions.paas_cc', new=FakePaaSCCMod()), mock.patch(
        'backend.utils.permissions.ProjectPermission', new=FakeProjectPermissionAllowAll
    ), mock.patch('backend.uniapps.application.base_views.paas_cc', new=FakePaaSCCMod()), mock.patch(
        'backend.apps.utils.paas_cc', new=FakePaaSCCMod()
    ), mock.patch(
        'backend.components.apigw.get_api_public_key', return_value=None
    ):
        yield


class TestConfigMaps:
    @pytest.mark.skip(reason='暂时跳过该单元测试')
    def test_get(self, api_client, project_id, use_fake_k8sclient):
        """This is sample API test which use faked k8sclient object"""
        response = api_client.get(f'/api/resource/{project_id}/configmaps/', format='json')
        assert response.json()['code'] == 0
