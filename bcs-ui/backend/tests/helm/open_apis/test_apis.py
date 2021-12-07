# -*- coding: utf-8 -*-
#
# Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
# Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://opensource.org/licenses/MIT
#
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.
#
from unittest import mock

import pytest

from backend.helm.helm.constants import PUBLIC_REPO_NAME
from backend.helm.helm.models import repo
from backend.tests.bcs_mocks.misc import FakePaaSCCMod, FakeProjectPermissionAllowAll
from backend.tests.testing_utils.base import generate_random_string
from backend.utils.cache import region

pytestmark = pytest.mark.django_db

fake_project_id = generate_random_string(8)
fake_name = "unittest-cluster"
fake_url = "http://example.repo.com"
fake_username = "admin"
fake_password = "admintest"


@pytest.fixture(autouse=True)
def create_db_records():
    def _create(name, project_id, repo_url):
        repo_obj = repo.Repository.objects.create(name=name, project_id=project_id, url=repo_url)
        cred = {"username": fake_username, "password": fake_password}
        repo.RepositoryAuth.objects.create(type="basic", credentials=cred, repo=repo_obj)

    # 添加公共仓库
    _create(PUBLIC_REPO_NAME, fake_project_id, f"{fake_url}/public/public")
    # 添加项目仓库
    _create(fake_name, fake_project_id, f"{fake_url}/{fake_name}/{fake_name}")


def test_chart_repo(api_client):
    url = f"/apis/helm/projects/{fake_project_id}/repo/"
    resp = api_client.get(url)
    resp_json = resp.json()
    assert resp_json["code"] == 0
    assert resp_json["data"]["url"] == f"{fake_url}/{fake_name}/{fake_name}"
    assert resp_json["data"]["username"] == fake_username


def test_shared_chart_repo(api_client):
    url = f"/apis/helm/public_repo/"
    resp = api_client.get(url)
    resp_json = resp.json()
    assert resp_json["code"] == 0
    assert resp_json["data"]["url"] == f"{fake_url}/public/public"
    assert resp_json["data"]["username"] == fake_username
