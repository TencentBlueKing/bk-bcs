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

from backend.resources.configs.configmap.formatter import ConfigMapFormatter


@pytest.fixture(scope="module", autouse=True)
def config_map_configs():
    configs = """
        {
            "apiVersion": "v1",
            "kind": "ConfigMap",
            "data": {
                "game.properties": "enemies=aliens,lives=3",
                "ui.properties": "color.good=purple,color.bad=yellow"
            },
            "metadata": {
                "creationTimestamp": "2021-04-14T06:55:56Z",
                "name": "rgwt919c",
                "namespace": "default",
                "resourceVersion": "20035",
                "uid": "a31699bc-642f-4b49-99d5-cd8908930062"
            }
        }
    """
    return json.loads(configs)


class TestConfigMapFormatter:
    def test_format_dict(self, config_map_configs):
        """测试 format_dict 方法"""
        result = ConfigMapFormatter().format_dict(config_map_configs)
        assert set(result.keys()) == {'data', 'age', 'createTime', 'updateTime'}
        assert result['data'] == ['game.properties', 'ui.properties']
