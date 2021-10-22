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
from backend.container_service.observability.log_stream import utils


def test_refine_k8s_logs(log_content):
    logs = utils.refine_k8s_logs(log_content, None)
    assert len(logs) == 10
    assert logs[0].time == '2021-05-19T12:03:52.516011121Z'


def test_calc_since_time(log_content):
    logs = utils.refine_k8s_logs(log_content, None)
    sine_time = utils.calc_since_time(logs[0].time, logs[-1].time)
    assert sine_time == '2021-05-19T12:03:10.125788125Z'


def test_calc_previous_page(log_content):
    logs = utils.refine_k8s_logs(log_content, None)
    page = utils.calc_previous_page(logs, {'container_name': "", "previous": ""}, "")
    assert page != ""
