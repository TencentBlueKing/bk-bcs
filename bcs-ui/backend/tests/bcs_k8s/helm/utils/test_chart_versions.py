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
from datetime import datetime

import mock
import pytest

from backend.helm.helm.models import chart, repo
from backend.helm.helm.utils import chart_versions
from backend.tests.bcs_mocks.fake_bk_repo import FakeBkRepoMod

pytestmark = pytest.mark.django_db

fake_project_id = "project-id"
fake_project_code = "project-code"
fake_versions = ["0.1.0", "0.1.1"]
fake_name = "test"


@pytest.fixture
def create_chart_and_versions():
    repo_obj = repo.Repository.objects.create(name=fake_name)
    chart_obj = chart.Chart.objects.create(name=fake_name, repository=repo_obj)
    for version in fake_versions:
        chart.ChartVersion.objects.create(name=fake_name, version=version, created=datetime.now(), chart=chart_obj)


def test_update_bcs_chart_records(create_chart_and_versions):
    chart_obj = chart.Chart.objects.get(name=fake_name)
    # 删除其中一个版本
    chart_versions.update_and_delete_chart_versions(fake_project_id, fake_project_code, chart_obj, fake_versions[:1])
    assert chart.Chart.objects.filter(name=fake_name).exists()
    assert chart.ChartVersion.objects.filter(name=fake_name)[0].version == fake_versions[1]
    # 删除chart
    chart_versions.update_and_delete_chart_versions(fake_project_id, fake_project_code, chart_obj, fake_versions[-1:])
    assert not chart.Chart.objects.filter(name=fake_name).exists()
    assert not chart.ChartVersion.objects.filter(name=fake_name).exists()


fake_chart_data = chart_versions.ChartData(project_name="projectname", repo_name="reponame", chart_name="demo")
fake_repo_auth = chart_versions.RepoAuth(username="admin", password="adminpwd")


@mock.patch("backend.helm.helm.utils.chart_versions.BkRepoClient", new=FakeBkRepoMod)
def test_get_chart_version_list():
    version_list = chart_versions.get_chart_version_list(fake_chart_data, fake_repo_auth)
    assert isinstance(version_list, list)


@mock.patch("backend.helm.helm.utils.chart_versions.BkRepoClient", new=FakeBkRepoMod)
def test_batch_delete_chart_versions():
    chart_versions.batch_delete_chart_versions(fake_chart_data, fake_repo_auth, ["0.1.0", "0.1.1"])
