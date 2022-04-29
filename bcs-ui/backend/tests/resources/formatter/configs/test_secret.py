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

from backend.resources.configs.secret.formatter import SecretsFormatter


@pytest.fixture(scope="module", autouse=True)
def secret_configs():
    configs = """
        {
            "apiVersion": "v1",
            "data": {
                "tls.crt": "YWhzZGtsaGZpc3VkaGZsd3...",
                "tls.key": "WVdoelpHdHNhR1pwYzNWa2..."
            },
            "kind": "Secret",
            "metadata": {
                "annotations": {
                    "kubectl.kubernetes.io/last-applied-configuration": "..."
                },
                "creationTimestamp": "2021-04-29T11:21:10Z",
                "name": "testsecret-tls",
                "namespace": "default",
                "resourceVersion": "420768",
                "uid": "dc08e35b-1569-43e5-8f10-779388d34109"
            },
            "type": "kubernetes.io/tls"
        }
    """
    return json.loads(configs)


class TestConfigMapFormatter:
    def test_format_dict(self, secret_configs):
        """测试 format_dict 方法"""
        result = SecretsFormatter().format_dict(secret_configs)
        assert set(result.keys()) == {'data', 'age', 'createTime', 'updateTime'}
        assert result['data'] == ['tls.crt', 'tls.key']
