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

from backend.container_service.clusters.base.models import CtxCluster

from .contents import release_data


@pytest.fixture
def ctx_cluster(cluster_id, project_id):
    return CtxCluster.create(id=cluster_id, token="token", project_id=project_id)


@pytest.fixture
def release_name():
    return "bk-redis"


@pytest.fixture
def revision():
    return 1


@pytest.fixture
def default_namespace():
    return "default"


@pytest.fixture
def parsed_release_data():
    return release_data.release_data


@pytest.fixture
def release_secret_data():
    return release_data.secret_data
