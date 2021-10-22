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
import os
import uuid
from unittest import mock

import pytest
from kubernetes import client

from backend.tests.conftest import TESTING_API_SERVER_URL
from backend.tests.testing_utils.mocks.collection import StubComponentCollection


class FakeBcsKubeConfigurationService:
    """Fake configuration service which return local apiserver as config"""

    def __init__(self, *args, **kwargs):
        pass

    def make_configuration(self):
        configuration = client.Configuration()
        configuration.api_key = {"authorization": f'Bearer {os.environ.get("TESTING_SERVER_API_KEY")}'}
        configuration.verify_ssl = False
        configuration.host = TESTING_API_SERVER_URL
        return configuration


@pytest.fixture(autouse=True)
def setup_fake_cluster_dependencies():
    # 替换所有 Comp 系统为测试专用的 Stub 系统；替换集群地址为测试用 API Server
    with mock.patch(
        'backend.container_service.core.ctx_models.ComponentCollection', new=StubComponentCollection
    ), mock.patch(
        'backend.resources.utils.kube_client.BcsKubeConfigurationService', new=FakeBcsKubeConfigurationService
    ):
        yield
