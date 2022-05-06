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

from backend.resources.workloads.statefulset.formatter import StatefulSetFormatter
from backend.tests.resources.formatter.conftest import WORKLOAD_CONFIG_DIR


@pytest.fixture(scope="module", autouse=True)
def statefulset_configs():
    with open(f'{WORKLOAD_CONFIG_DIR}/statefulset.json') as fr:
        configs = json.load(fr)
    return configs


class TestStatefulsetFormatter:
    def test_format_dict(self, statefulset_configs):
        """测试 format_dict 方法"""
        result = StatefulSetFormatter().format_dict(statefulset_configs['normal'])
        assert set(result.keys()) == {'images', 'age', 'createTime', 'updateTime'}
        assert result['images'] == ['k8s.gcr.io/nginx-slim:0.8']
