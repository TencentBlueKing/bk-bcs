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

Test codes for backend.resources module
"""
import pytest

from backend.container_service.clusters.base import CtxCluster
from backend.resources.client import BcsAPIEnvironmentQuerier, BcsKubeConfigurationService
from backend.tests.testing_utils.mocks.bcs_api import StubBcsApiClient
from backend.tests.testing_utils.mocks.paas_cc import StubPaaSCCClient
from backend.utils.exceptions import ComponentError

pytestmark = pytest.mark.django_db


@pytest.fixture(autouse=True)
def setup_settings(settings):
    """Setup required settings for unittests"""
    settings.BCS_API_ENV = {'stag': 'my_stag', 'prod': 'my_prod'}
    settings.BCS_SERVER_HOST = {
        'my_stag': 'https://my-stag-bcs-server.example.com',
        'my_prod': 'https://my-prod-bcs-server.example.com',
    }
    settings.BCS_API_PRE_URL = 'https://bcs-api.example.com'


fake_cc_get_cluster_result_ok = {'code': 0, 'result': True, 'data': {'environment': 'stag'}}
fake_cc_get_cluster_result_failed = {'code': 100, 'result': False}


class TestBcsAPIEnvironmentQuerier:
    def test_normal(self, project_id, cluster_id):
        cluster = CtxCluster.create(cluster_id, project_id, token='token')
        querier = BcsAPIEnvironmentQuerier(cluster)
        with StubPaaSCCClient.get_cluster.mock(return_value=fake_cc_get_cluster_result_ok):
            api_env_name = querier.do()

        assert api_env_name == 'my_stag'

    def test_failed(self, project_id, cluster_id):
        cluster = CtxCluster.create(cluster_id, project_id, token='token')
        querier = BcsAPIEnvironmentQuerier(cluster)
        with StubPaaSCCClient.get_cluster.mock(return_value=fake_cc_get_cluster_result_failed):
            with pytest.raises(ComponentError):
                assert querier.do()


class TestBcsKubeConfigurationService:
    def test_make_configuration(self, project_id, cluster_id):
        cluster = CtxCluster.create(cluster_id, project_id, token='token')
        config_service = BcsKubeConfigurationService(cluster)

        faked_credentials = {'server_address_path': '/example-foo-cluster', 'user_token': 'faked-foo-token'}
        with StubBcsApiClient.get_cluster_credentials.mock(return_value=faked_credentials):
            config = config_service.make_configuration()

        assert config.host == 'https://my-stag-bcs-server.example.com/example-foo-cluster'
        assert config.api_key['authorization'] == f'Bearer {faked_credentials["user_token"]}'
