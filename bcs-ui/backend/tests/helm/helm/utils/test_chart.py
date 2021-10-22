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

import pytest

from backend.helm.helm.models import chart, repo
from backend.helm.helm.serializers import ChartSLZ
from backend.helm.helm.utils.chart import ChartList
from backend.tests.testing_utils.base import generate_random_string

pytestmark = pytest.mark.django_db

name_or_version = generate_random_string(4)
project_id = generate_random_string(32)


@pytest.fixture(autouse=True)
def create_db_records():
    repo_obj = repo.Repository.objects.create(name=name_or_version, project_id=project_id)
    chart_obj = chart.Chart.objects.create(name=name_or_version, repository=repo_obj)
    version_obj = chart.ChartVersion.objects.create(
        name=name_or_version, version=name_or_version, created=datetime.now(), chart=chart_obj
    )
    chart_obj.defaultChartVersion = version_obj
    chart_obj.save(update_fields=["defaultChartVersion"])


class TestChartList:
    def get_chart_data(self):
        data = ChartList(project_id).get_chart_data()
        assert len(data) > 0
        assert "defaultChartVersion" in data[0]
        assert "repository" in data[0]
        assert "version" in data[0]["defaultChartVersion"]
        assert "url" in data[0]["repository"]

    def test_get_version_and_repo_ids(self):
        charts = chart.Chart.objects.get_charts(project_id)
        slz = ChartSLZ(charts, many=True)
        charts_data = slz.data
        ver_ids, repo_ids = ChartList(project_id)._get_version_and_repo_ids(charts_data)
        assert isinstance(ver_ids, set)
        assert len(ver_ids) == 1
        assert isinstance(repo_ids, set)
        assert len(repo_ids) == 1
