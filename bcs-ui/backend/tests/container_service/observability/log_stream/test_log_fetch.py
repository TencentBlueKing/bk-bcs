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

from backend.container_service.observability.log_stream import utils


@pytest.mark.django_db
class TestLogStream:
    @pytest.mark.skip(reason='暂时跳过标准日志部分单元测试')
    def test_fetch(self, api_client, project_id, cluster_id, namespace, pod_name, container_name):
        """测试获取日志"""
        response = api_client.get(
            f'/api/logs/projects/{project_id}/clusters/{cluster_id}/namespaces/{namespace}/pods/{pod_name}/stdlogs/?container_name={container_name}'  # noqa
        )
        assert response.json()['code'] == 0

    @pytest.mark.skip(reason='暂时跳过标准日志部分单元测试')
    def test_create_session(self, api_client, project_id, cluster_id, namespace, pod_name, container_name):
        response = api_client.post(
            f'/api/logs/projects/{project_id}/clusters/{cluster_id}/namespaces/{namespace}/pods/{pod_name}/stdlogs/sessions/',  # noqa
            {"container_name": container_name},
        )

        result = response.json()
        assert result['code'] == 0
        assert len(result['data']['session_id']) > 0
        assert result['data']['ws_url'].startswith("ws://")

    def test_refine_k8s_logs(self, log_content):
        logs = utils.refine_k8s_logs(log_content, None)
        assert len(logs) == 10
        assert logs[0].time == '2021-05-19T12:03:52.516011121Z'

    def test_calc_since_time(self, log_content):
        logs = utils.refine_k8s_logs(log_content, None)
        sine_time = utils.calc_since_time(logs[0].time, logs[-1].time)
        assert sine_time == '2021-05-19T12:03:10.125788125Z'

    def test_calc_previous_page(self, log_content):
        logs = utils.refine_k8s_logs(log_content, None)
        page = utils.calc_previous_page(logs, {'container_name': "", "previous": ""}, "")
        assert page != ""
