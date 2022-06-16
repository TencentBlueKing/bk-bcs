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
import pytest
from django.conf import settings

from backend.helm.helm.models.repo import Repository, RepositoryAuth

pytestmark = pytest.mark.django_db

FAKE_NAME = "unittest-proj"
FAKE_REPO_DOMAIN = "http://example.repo.com"
FAKE_USERNAME = "admin"
FAKE_PASSWORD = "admintest"
# 路径含义: 项目/仓库
FAKE_SHARED_REPO_PATH = "/public/public"
FAKE_DEDICATED_REPO_PATH = "/{FAKE_NAME}/{FAKE_NAME}"


@pytest.fixture(autouse=True)
def create_db_records(project_id):
    def _create(name, repo_url):
        repo_obj = Repository.objects.create(name=name, project_id=project_id, url=repo_url)
        cred = {"username": FAKE_USERNAME, "password": FAKE_PASSWORD}
        RepositoryAuth.objects.create(type="basic", credentials=cred, repo=repo_obj)

    # 添加公共仓库
    _create(settings.BCS_SHARED_CHART_REPO_NAME, f"{FAKE_REPO_DOMAIN}{FAKE_SHARED_REPO_PATH}")
    # 添加项目仓库
    _create(FAKE_NAME, f"{FAKE_REPO_DOMAIN}{FAKE_DEDICATED_REPO_PATH}")


def test_chart_repo(api_client, project_id):
    url = f"/apis/helm/projects/{project_id}/repo/"
    resp = api_client.get(url)
    resp_json = resp.json()
    assert resp_json["code"] == 0
    assert resp_json["data"]["url"] == f"{FAKE_REPO_DOMAIN}{FAKE_DEDICATED_REPO_PATH}"
    assert resp_json["data"]["username"] == FAKE_USERNAME


def test_shared_chart_repo(api_client):
    url = f"/apis/helm/public_repo/"
    resp = api_client.get(url)
    resp_json = resp.json()
    assert resp_json["code"] == 0
    assert resp_json["data"]["url"] == f"{FAKE_REPO_DOMAIN}{FAKE_SHARED_REPO_PATH}"
    assert resp_json["data"]["username"] == FAKE_USERNAME
