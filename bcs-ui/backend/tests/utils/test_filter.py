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

from backend.utils.filter import filter_by_ips


@pytest.mark.parametrize(
    "source_ips, ips, ret_ips, fuzzy",
    [
        (['127.0.0.1', '127.0.0.2'], ['127.0.1'], [], True),
        (['127.0.0.1', '127.0.0.2'], ['127.0.0.3'], [], True),
        (['127.0.0.1', '127.0.0.2'], ['127.0.0.1'], ['127.0.0.1'], True),
        (['127.0.0.1', '127.0.0.2'], ['127.0.0'], ['127.0.0.1', '127.0.0.2'], True),
        (['127.0.0.31', '127.0.0.2; 127.0.0.3'], ['127.0.0.3'], ['127.0.0.31', '127.0.0.2; 127.0.0.3'], True),
        (['127.0.0.1', '127.0.0.2'], ['127.0.0'], [], False),
        (['127.0.0.1', '127.0.0.2'], ['127.0.0.3'], [], False),
        (['127.0.0.31', '127.0.0.2,127.0.0.3'], ['127.0.0.3'], ['127.0.0.2,127.0.0.3'], False),
        (
            ['127.0.0.1', '127.0.0.2,127.0.0.3'],
            ['127.0.0.1', '127.0.0.2'],
            ['127.0.0.1', '127.0.0.2,127.0.0.3'],
            False,
        ),
    ],
)
def test_str2bool(source_ips, ips, ret_ips, fuzzy):
    source = [{'ip': ip} for ip in source_ips]
    ret = filter_by_ips(source, ips, key='ip', fuzzy=fuzzy)
    assert ret_ips == [item['ip'] for item in ret]
