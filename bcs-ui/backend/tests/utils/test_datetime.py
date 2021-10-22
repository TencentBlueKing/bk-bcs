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

from backend.utils.datetime import get_duration_seconds


@pytest.mark.parametrize(
    'duration, default, expired',
    [
        ('xxxxxx', 100, 100),
        ('3h5m7s', None, 11107),
        ('5m7s', None, 307),
        ('1h7s', None, 3607),
        ('1h0m5s', None, 3605),
        ('1h0m0s', None, 3600),
        ('1h', None, 3600),
        ('10m', None, 600),
        ('8s', None, 8),
        ('0s', None, 0),
    ],
)
def test_get_duration_seconds(duration, default, expired):
    assert get_duration_seconds(duration, default) == expired
