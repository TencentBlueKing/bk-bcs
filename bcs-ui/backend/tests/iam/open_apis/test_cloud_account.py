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
import random

import mock
import pytest
from django.utils.crypto import get_random_string
from rest_framework.test import APIRequestFactory

from backend.iam.open_apis.constants import MethodType, ResourceType
from backend.iam.open_apis.views import ResourceAPIView
from backend.tests.testing_utils.mocks.cm import StubClusterManagerClient

factory = APIRequestFactory()


@pytest.fixture(autouse=True)
def patch_cm_client():
    with mock.patch(
        'backend.iam.open_apis.providers.cloud_account.ClusterManagerClient', new=StubClusterManagerClient
    ):
        yield


class TestCloudAccountAPI:
    def test_list_instance(self, project_id):
        request = factory.post(
            '/apis/iam/v1/cloud_accounts/',
            {
                'method': MethodType.LIST_INSTANCE,
                'type': ResourceType.Cloudaccount,
                'page': {'offset': 0, 'limit': 5},
                'filter': {'parent': {'id': project_id}},
            },
        )
        p_view = ResourceAPIView.as_view()
        response = p_view(request)
        data = response.data
        assert data['count'] == 2

    def test_fetch_instance_info_with_ids(self):
        filter_ids = [get_random_string(8) for _ in range(random.randint(0, 10))]
        request = factory.post(
            '/apis/iam/v1/cloud_accounts/',
            {
                'method': MethodType.FETCH_INSTANCE_INFO,
                'type': ResourceType.Cloudaccount,
                'filter': {'ids': filter_ids},
            },
        )
        p_view = ResourceAPIView.as_view()
        response = p_view(request)
        data = response.data
        assert len(data) == len(filter_ids)
