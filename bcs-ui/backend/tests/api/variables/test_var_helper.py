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

from backend.templatesets.var_mgmt.open_apis.var_helper import compose_data


class TestVariables:
    @pytest.mark.parametrize(
        "var_id_data_map, var_id_key_name_map, expected_data",
        [
            ({}, {}, []),
            (
                {1: '{"value": 1}'},
                {1: {"key": "test", "name": "test"}},
                [{"key": "test", "name": "test", "id": 1, "value": 1}],
            ),
            (
                {1: '{"value": 1}', 2: '{"value": 1}'},
                {1: {"key": "test", "name": "test"}, 2: {"key": "test1", "name": "test1"}},
                [
                    {"key": "test", "name": "test", "id": 1, "value": 1},
                    {"key": "test1", "name": "test1", "id": 2, "value": 1},
                ],
            ),
            (
                {1: '{"value": 1}', 2: '{"value": 1}'},
                {1: {"key": "test", "name": "test"}, 3: {"key": "test1", "name": "test1"}},
                [{"key": "test", "name": "test", "id": 1, "value": 1}],
            ),
        ],
    )
    def test_compose_data(self, var_id_data_map, var_id_key_name_map, expected_data):
        data = compose_data(var_id_data_map, var_id_key_name_map)
        assert data == expected_data
