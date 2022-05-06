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
import json

import pytest

from backend.resources.workloads.pod.formatter import PodFormatter
from backend.tests.resources.formatter.conftest import WORKLOAD_CONFIG_DIR


@pytest.fixture(scope="module", autouse=True)
def pod_configs():
    with open(f'{WORKLOAD_CONFIG_DIR}/pod.json') as fr:
        configs = json.load(fr)
    return configs


class TestPodFormatter:
    def test_format_dict(self, pod_configs):
        """测试 format_dict 方法"""
        result = PodFormatter().format_dict(pod_configs['status_running'])
        assert set(result.keys()) == {
            'status',
            'readyCnt',
            'totalCnt',
            'restartCnt',
            'images',
            'age',
            'createTime',
            'updateTime',
            'namespace',
            'hostIP',
            'name',
            'podIP',
        }
        assert result['images'] == ['k8s.gcr.io/echoserver:1.4']
