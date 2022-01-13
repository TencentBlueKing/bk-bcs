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
from rest_framework.test import APIRequestFactory

from backend.iam.open_apis.constants import MethodType, ResourceType
from backend.iam.open_apis.views import ResourceAPIView
from backend.templatesets.legacy_apps.configuration import models

pytestmark = pytest.mark.django_db

factory = APIRequestFactory()


@pytest.fixture
def template_qsets(project_id):
    template1 = models.Template.objects.create(project_id=project_id, name='nginx-test')
    template2 = models.Template.objects.create(project_id=project_id, name='redis-test')
    return [template1, template2]


class TestTemplatesetAPI:
    def test_list_instance(self, project_id, template_qsets):
        request = factory.post(
            '/apis/iam/v1/templatesets/',
            {
                'method': MethodType.LIST_INSTANCE,
                'type': ResourceType.Templateset,
                'page': {'offset': 0, 'limit': 5},
                "filter": {'parent': {'id': project_id}},
            },
        )
        p_view = ResourceAPIView.as_view()
        response = p_view(request)
        data = response.data
        assert data['count'] == 2
        assert data['results'][0]['display_name'] in ['nginx-test', 'redis-test']

    def test_fetch_instance_info_with_tplset_ids(self, template_qsets):
        request = factory.post(
            '/apis/iam/v1/templatesets/',
            {
                'method': 'fetch_instance_info',
                'type': 'templateset',
                'page': {'offset': 0, 'limit': 5},
                "filter": {'ids': [template_qsets[0].id]},
            },
        )
        p_view = ResourceAPIView.as_view()
        response = p_view(request)
        data = response.data
        assert len(data) == 1

    def test_search_instance(self, project_id, template_qsets):
        # 匹配到关键字
        request = factory.post(
            '/apis/iam/v1/templatesets/',
            {
                'method': MethodType.SEARCH_INSTANCE,
                'type': ResourceType.Templateset,
                'page': {'offset': 0, 'limit': 5},
                "filter": {'parent': {'id': project_id}, 'keyword': 'test'},
            },
        )
        p_view = ResourceAPIView.as_view()
        response = p_view(request)
        data = response.data
        assert data['count'] == 2

        # 匹配不到关键字
        request = factory.post(
            '/apis/iam/v1/templatesets/',
            {
                'method': MethodType.SEARCH_INSTANCE,
                'type': ResourceType.Templateset,
                'page': {'offset': 0, 'limit': 5},
                'filter': {'keyword': 'ttt', 'parent': {'id': project_id}},
            },
        )
        p_view = ResourceAPIView.as_view()
        response = p_view(request)
        data = response.data
        assert data['count'] == 0
