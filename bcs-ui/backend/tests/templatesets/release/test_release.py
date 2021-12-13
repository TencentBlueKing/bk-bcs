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

from backend.templatesets import models
from backend.templatesets.release.manager import AppReleaseManager
from backend.utils.basic import getitems

from .fake_manager import FakeReleaseResourceManager, fake_async_run

pytestmark = pytest.mark.django_db


@pytest.fixture(autouse=True)
def use_dummy_settings_config(settings):
    settings.DEVOPS_ARTIFACTORY_HOST = "harbor-api.service.consul"


class TestReleaseDataGenerator:
    def test_form_generator(self, form_show_version, form_release_data):
        for res in form_release_data.resource_list:
            assert res.name == getitems(res.manifest, 'metadata.name')
            assert res.kind == getitems(res.manifest, 'kind')
            assert res.namespace == getitems(res.manifest, 'metadata.namespace')
            assert res.version == form_show_version.name
            assert getitems(res.manifest, 'webCache') is None
            assert 'io.tencent.bcs.cluster' in getitems(res.manifest, 'metadata.annotations')

            if res.kind == 'Service':
                assert getitems(res.manifest, 'spec.type') == 'ClusterIP'

    def test_yaml_generator(self, yaml_show_version, yaml_release_data):
        assert len(yaml_release_data.resource_list) == 4

        for res in yaml_release_data.resource_list:
            assert res.name == getitems(res.manifest, 'metadata.name')
            assert res.kind == getitems(res.manifest, 'kind')
            assert res.namespace == getitems(res.manifest, 'metadata.namespace')
            assert res.version == yaml_show_version.name

            if res.kind == 'Endpoints':
                assert getitems(res.manifest, 'subsets')[0]['addresses'][0]['ip'] == '0.0.0.1'
                continue

            if res.kind == 'Pod':
                assert getitems(res.manifest, 'spec.containers')[0]['image'] == 'redis:5.0.4'


class TestReleaseManager:
    def test_update_or_create(self, bk_user, project_id, yaml_release_data):
        with mock.patch(
            'backend.templatesets.release.manager.ReleaseResourceManager', new=FakeReleaseResourceManager
        ), mock.patch('backend.templatesets.release.manager.async_run', new=fake_async_run):
            release_manager = AppReleaseManager(dynamic_client=None)
            app_release, _ = release_manager.update_or_create(bk_user.username, release_data=yaml_release_data)

            assert models.ResourceInstance.objects.filter(app_release=app_release).count() == 4
