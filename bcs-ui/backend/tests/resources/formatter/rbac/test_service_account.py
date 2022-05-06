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

from backend.resources.rbac.service_account.formatter import ServiceAccountFormatter


@pytest.fixture(scope="module", autouse=True)
def service_account_configs():
    configs = """
        {
            "apiVersion": "v1",
            "kind": "ServiceAccount",
            "metadata": {
                "creationTimestamp": "2021-04-13T09:02:15Z",
                "name": "default",
                "namespace": "default",
                "resourceVersion": "404",
                "uid": "9e8ca0ca-8e4a-4e68-a875-1476f9b1aa68"
            },
            "secrets": [
                {
                    "name": "default-token-kvb6t"
                }
            ]
        }
    """
    return json.loads(configs)


class TestServiceAccountFormatter:
    def test_format_dict(self, service_account_configs):
        """测试 format_dict 方法"""
        result = ServiceAccountFormatter().format_dict(service_account_configs)
        assert set(result.keys()) == {'secrets', 'age', 'createTime', 'updateTime'}
        assert result['secrets'] == 1
