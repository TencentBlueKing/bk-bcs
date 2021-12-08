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

单元测试可用环境变量说明
值                          说明                          默认值
TESTING_API_SERVER_URL  	测试环境/本地集群URL           'http://localhost:28180'
TESTING_SERVER_API_KEY	    测试环境/本地集群 api_key      None
TEST_PROJECT_ID		        指定的测试项目 ID              32位随机字符串
TEST_CLUSTER_ID	            指定的测试集群 ID              8位随机字符串
TEST_NAMESPACE	            用于单元测试的命名空间          'default'
TEST_POD_NAME	            用于单元测试的 Pod 名称        8位随机字符串
TEST_CONTAINER_NAME	        用于单元测试的容器名称          8位随机字符串
"""
import os
from unittest import mock

import pytest
from django.contrib.auth import get_user_model
from kubernetes import client
from rest_framework.test import APIClient

from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.constants import ClusterType
from backend.templatesets.legacy_apps.configuration import models
from backend.templatesets.legacy_apps.configuration.constants import TemplateEditMode
from backend.tests.testing_utils.base import generate_random_string
from backend.tests.testing_utils.mocks.k8s_client import get_dynamic_client
from backend.tests.testing_utils.mocks.viewsets import FakeSystemViewSet, FakeUserViewSet
from backend.utils import FancyDict

# 单元测试用集群 ApiServer
TESTING_API_SERVER_URL = os.environ.get("TESTING_API_SERVER_URL", 'http://localhost:28180')

# 全局 patch SystemViewSet & UserViewSet & get_dynamic_client
from backend.bcs_web import viewsets  # noqa

viewsets.SystemViewSet = FakeSystemViewSet
viewsets.UserViewSet = FakeUserViewSet

from backend.resources import resource  # noqa

resource.get_dynamic_client = get_dynamic_client

# 单元测试用常量，用于不便使用 pytest.fixture 的地方
TEST_PROJECT_ID = os.environ.get("TEST_PROJECT_ID", generate_random_string(32))
TEST_CLUSTER_ID = os.environ.get("TEST_CLUSTER_ID", f"BCS-K8S-{generate_random_string(5, chars='12345')}")
TEST_NAMESPACE = os.environ.get("TEST_NAMESPACE", 'default')
# 测试用共享集群 ID
TEST_SHARED_CLUSTER_ID = os.environ.get('TEST_SHARED_CLUSTER_ID', 'BCS-K8S-95001')


def fake_get_cluster_type(cluster_id: str) -> ClusterType:
    if cluster_id == TEST_SHARED_CLUSTER_ID:
        return ClusterType.SHARED
    return ClusterType.SINGLE


from backend.container_service.clusters.base import utils as cluster_base_utils  # noqa

cluster_base_utils.get_cluster_type = fake_get_cluster_type


@pytest.fixture
def cluster_id():
    """使用环境变量或者生成一个随机集群 ID"""
    # 集群 ID 后五位为纯数字
    return os.environ.get("TEST_CLUSTER_ID", f"BCS-K8S-{generate_random_string(5, chars='12345')}")


@pytest.fixture
def project_id():
    """使用环境变量或者生成一个随机项目 ID"""
    return os.environ.get("TEST_PROJECT_ID", generate_random_string(32))


@pytest.fixture
def request_user():
    return FancyDict({"username": "admin", "token": FancyDict({"access_token": "test_access_token"})})


@pytest.fixture
def random_name():
    """生成一个随机 name"""
    return generate_random_string(8)


@pytest.fixture
def namespace():
    """使用环境变量或者生成一个命名空间"""
    return os.environ.get("TEST_NAMESPACE", generate_random_string(8))


@pytest.fixture
def pod_name():
    """使用环境变量或者生成一个随机 Pod 名称"""
    return os.environ.get("TEST_POD_NAME", generate_random_string(8))


@pytest.fixture
def container_name():
    """使用环境变量或者生成一个随机容器名称"""
    return os.environ.get("TEST_CONTAINER_NAME", generate_random_string(8))


@pytest.fixture
def bk_user():
    User = get_user_model()
    user = User.objects.create(username=generate_random_string(6))

    # Set token attribute
    user.token = mock.MagicMock()
    user.token.access_token = generate_random_string(12)
    user.token.expires_soon = lambda: False

    return user


@pytest.fixture
def api_client(request, bk_user):
    """Return an authenticated client"""
    client = APIClient()
    client.force_authenticate(user=bk_user)
    return client


@pytest.fixture
def testing_kubernetes_apiclient():
    """返回连接单元测试 apiserver 的 ApiClient 实例"""
    configuration = client.Configuration()
    configuration.api_key = {"authorization": f'Bearer {os.environ.get("TESTING_SERVER_API_KEY")}'}
    configuration.verify_ssl = False
    configuration.host = TESTING_API_SERVER_URL
    return client.ApiClient(configuration)


@pytest.fixture
def use_fake_k8sclient(cluster_id):
    """替换代码中所有的 k8s.K8SClient() 调用，使其连接用于测试的 apiserver"""
    fake_cluster_context = {
        'id': cluster_id,
        'provider': 2,
        'creator_id': 100,
        'identifier': f'{cluster_id}-x',
        'created_at': '2020-01-01T00:00:00',
        'server_address_path': '',
        'user_token': 'fake_user_token',
    }
    with mock.patch(
        'backend.components.bcs.k8s_client.make_cluster_context', return_value=fake_cluster_context
    ), mock.patch(
        'backend.components.bcs.BCSClientBase._bcs_server_host',
        new_callable=mock.PropertyMock,
        return_value=TESTING_API_SERVER_URL,
    ):
        yield


@pytest.fixture
def ctx_cluster(cluster_id, project_id):
    return CtxCluster.create(id=cluster_id, token=generate_random_string(12), project_id=project_id)


@pytest.fixture
def form_template(project_id):
    template = models.Template.objects.create(
        project_id=project_id, name='nginx', edit_mode=TemplateEditMode.PageForm.value
    )
    return template
