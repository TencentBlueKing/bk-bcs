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

from backend.container_service.cluster_tools import models


@pytest.fixture
def tool(db):
    return models.Tool.objects.create(
        name="GameStatefulSet", chart_name='bcs-gamestatefulset-operator', default_version='0.6.0-beta3'
    )


@pytest.fixture(autouse=True)
def patch_get_random_string():
    with mock.patch('backend.helm.toolkit.deployer.get_random_string', new=lambda *args, **kwargs: '12345678'):
        yield
