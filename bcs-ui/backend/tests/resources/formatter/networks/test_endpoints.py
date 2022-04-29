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

from backend.resources.networks.endpoints.formatter import EndpointsFormatter
from backend.tests.resources.formatter.conftest import NETWORK_CONFIG_DIR


@pytest.fixture(scope="module", autouse=True)
def endpoints_configs():
    with open(f'{NETWORK_CONFIG_DIR}/endpoints.json') as fr:
        configs = json.load(fr)
    return configs


class TestEndpointsFormatter:
    def test_format_dict(self, endpoints_configs):
        """测试 format_dict 方法"""
        result = EndpointsFormatter().format_dict(endpoints_configs['normal'])
        assert set(result.keys()) == {'endpoints', 'age', 'createTime', 'updateTime'}
        assert result['endpoints'] == [
            '127.xxx.xxx.xx1:80',
            '127.xxx.xxx.xx2:80',
            '127.xxx.xxx.xx3:80',
            '127.xxx.xxx.xx4:80',
            '127.xxx.xxx.xx5:80',
            '127.xxx.xxx.xx6:80',
        ]
