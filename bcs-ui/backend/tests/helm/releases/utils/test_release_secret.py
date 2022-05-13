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
from django.db.models.query import QuerySet
from mock import patch

from backend.helm.app.models import App
from backend.helm.helm.models import chart, repo
from backend.helm.releases.utils.release_secret import RecordReleases, get_release_detail
from backend.tests.helm.releases.contents.release_data import release_list

pytestmark = pytest.mark.django_db

FAKE_NAME = FAKE_VERSTION_NAME = "test"
FAKE_NAMESPACE = "default"


def test_get_release_detail(ctx_cluster, default_namespace, release_name, revision, parsed_release_data):
    with patch(
        "backend.helm.releases.utils.release_secret.list_namespaced_releases",
        return_value=[parsed_release_data],
    ):
        release_detail = get_release_detail(ctx_cluster, default_namespace, release_name)
    assert release_detail["name"] == release_name
    assert release_detail["namespace"] == default_namespace
    assert release_detail["version"] == revision


class TestRecordReleases:
    @pytest.fixture(autouse=True)
    def create_db_records(self, project_id, cluster_id):
        repo_obj = repo.Repository.objects.create(name=FAKE_NAME, project_id=project_id)
        chart_obj = chart.Chart.objects.create(name=FAKE_NAME, repository=repo_obj)
        version_obj = chart.ChartVersion.objects.create(
            name=FAKE_NAME, version=FAKE_VERSTION_NAME, created=datetime.now(), chart=chart_obj
        )
        chart_obj.defaultChartVersion = version_obj
        chart_obj.save(update_fields=["defaultChartVersion"])
        snapshot = chart.ChartVersionSnapshot.objects.create(
            name=FAKE_NAME, version=FAKE_VERSTION_NAME, created=datetime.now()
        )
        chart.ChartRelease.objects.create(repository=repo_obj, chart=chart_obj, chartVersionSnapshot=snapshot)

    @pytest.fixture
    def client(self, ctx_cluster):
        return RecordReleases(ctx_cluster=ctx_cluster, namespace=FAKE_NAMESPACE)

    def test_get_app_names(self, client):
        app_names = client._get_app_names()
        assert type(app_names) is QuerySet

    def test_get_chart_map(self, client):
        chart_map = client._get_chart_map()
        assert FAKE_NAME in chart_map

    def test_get_chart_version(self, client):
        chart = client._get_chart_map()[FAKE_NAME]
        chart_version = client._get_chart_version(FAKE_VERSTION_NAME, chart)
        assert chart_version is not None
        assert chart_version.name == FAKE_VERSTION_NAME

    def test_record_chart_release(self, client):
        chart = client._get_chart_map()[FAKE_NAME]
        chart_version = client._get_chart_version(FAKE_VERSTION_NAME, chart)
        chart_release = client._record_chart_release(chart, chart_version, "")
        assert chart_release.chart.name == FAKE_NAME
        assert chart_release.chartVersionSnapshot.version == FAKE_VERSTION_NAME

    def test_get_releases(self, client):
        with patch(
            "backend.helm.releases.utils.release_secret.list_namespaced_releases",
            return_value=release_list,
        ):
            releases = client._get_releases()
        # 名称为test的release已经记录到db中， 所以返回的仅有两条数据
        assert len(releases) == 1
        assert releases[0]["name"] == FAKE_NAME

    def test_get_values_content(self, client):
        value_content = client._get_values_content(release_list[-1])
        assert "replica: 3" in value_content

    def test_record(self, client):
        with patch(
            "backend.helm.releases.utils.release_secret.list_namespaced_releases",
            return_value=release_list,
        ), patch(
            "backend.helm.releases.utils.release_secret.collect_system_variable",
            return_value={
                "SYS_NAMESPACE": FAKE_NAMESPACE,
                "SYS_CC_APP_ID": "1",
                "SYS_PROJECT_KIND": "k8s",
                "SYS_PROJECT_CODE": "test",
            },
        ), patch(
            "backend.helm.helm.bcs_variable.get_data_id_by_project_id",
            return_value={"standard_data_id": 1, "non_standard_data_id": 1},
        ), patch(
            "backend.helm.releases.utils.release_secret.PaaSCCClient.get_cluster_namespace_list",
            return_value={
                "code": 0,
                "results": [
                    {
                        "id": 1,
                        "name": FAKE_NAMESPACE,
                        "project_id": client.ctx_cluster.project_id,
                        "cluster_id": client.ctx_cluster.id,
                    }
                ],
            },
        ), patch(
            "backend.helm.helm.bcs_variable.paas_cc.get_jfrog_domain", return_value="http://example.test.com"
        ):
            client.record()
        # 校验已经写入数据
        assert App.objects.filter(name=FAKE_NAME).exists()
