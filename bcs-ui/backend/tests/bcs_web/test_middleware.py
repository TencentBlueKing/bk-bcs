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
from typing import Optional

import pytest

from backend.bcs_web.middleware import get_cookie_domain_by_host


@pytest.mark.parametrize(
    'cookie_domain, request_domain, expect_domain',
    [
        ('.example.com', 'bcs.example.com', '.example.com'),
        (None, 'bcs.example.com', None),
        ('', 'bcs.example.com', ''),
        ('.example.com;.qq.example.com', 'bcs.qq.example.com', '.qq.example.com'),
    ],
)
def test_get_cookie_domain_by_host(cookie_domain: Optional[str], request_domain: str, expect_domain: Optional[str]):
    assert get_cookie_domain_by_host(cookie_domain, request_domain) == expect_domain
