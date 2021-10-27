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
import os.path
from unittest import mock

import pytest
import yaml
from kubernetes.dynamic.exceptions import ResourceNotFoundError

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.hpa import client as hpa_client
from backend.tests.conftest import TEST_NAMESPACE
from backend.utils.basic import getitems

from ..conftest import FakeBcsKubeConfigurationService

BASE_DIR = os.path.dirname(os.path.abspath(__file__))

# 预先加载默认配置
with open(os.path.join(BASE_DIR, "sample_cpu_hpa.yaml")) as fh:
    hpa_manifest = yaml.load(fh.read())

# 检查逻辑，确保使用的是测试用的命名空间
if getitems(hpa_manifest, 'metadata.namespace') != TEST_NAMESPACE:
    hpa_manifest['metadata']['namespace'] = TEST_NAMESPACE

hpa_name = getitems(hpa_manifest, 'metadata.name')


class TestHPA:
    @pytest.fixture(autouse=True)
    def use_faked_configuration(self):
        """Replace ConfigurationService with fake object"""
        with mock.patch(
            'backend.resources.utils.kube_client.BcsKubeConfigurationService',
            new=FakeBcsKubeConfigurationService,
        ):
            yield

    @pytest.fixture(autouse=True)
    def use_fake_db(self):
        with mock.patch("backend.templatesets.legacy_apps.instance.models.InstanceConfig.objects"):
            yield

    @pytest.fixture
    def hpa_client(self, project_id, cluster_id):
        try:
            return hpa_client.HPA(CtxCluster.create(token='token', project_id=project_id, id=cluster_id))
        except ResourceNotFoundError:
            pytest.skip('Can not initialize HPA client, skip')

    @pytest.fixture
    def sample_hpa(self, hpa_client):
        hpa_client.update_or_create(namespace=TEST_NAMESPACE, name=hpa_name, body=hpa_manifest, is_format=False)
        yield
        hpa_client.delete_wait_finished(namespace=TEST_NAMESPACE, name=hpa_name, namespace_id="", username="")

    def test_list(self, hpa_client, sample_hpa):
        hpa_list = hpa_client.list(namespace=TEST_NAMESPACE)
        assert len(hpa_list) > 0

    def test_update_or_create(self, hpa_client, sample_hpa):
        res, created = hpa_client.update_or_create(body=hpa_manifest, is_format=False)
        assert created is False

    def test_delete(self, hpa_client):
        hpa_client.update_or_create(namespace=TEST_NAMESPACE, name=hpa_name, body=hpa_manifest, is_format=False)

        result = hpa_client.delete_ignore_nonexistent(namespace=TEST_NAMESPACE, name=hpa_name)
        assert result.status == 'Success'
