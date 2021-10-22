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
from iam import OP

from backend.iam.legacy_perms import ProjectPermission

test_dict_filter_data = [
    ({'op': OP.IN, 'value': [2, 1], 'field': 'project.id'}, {'project_id_list': [1, 2], 'op': OP.IN}),
    ({'op': OP.EQ, 'value': 1, 'field': 'project.id'}, {'project_id_list': [1], 'op': OP.IN}),
    ({'op': OP.ANY, 'value': [], 'field': 'project.id'}, {'project_id_list': [], 'op': OP.ANY}),
    (
        {
            'op': OP.OR,
            'content': [
                {'op': OP.IN, 'field': 'project.id', 'value': [2, 1, 5]},
                {'op': OP.ANY, 'field': 'project.id', 'value': []},
                {'op': OP.EQ, 'field': 'project.id', 'value': 3},
                {'op': OP.IN, 'field': 'project.id', 'value': [4]},
            ],
        },
        {'project_id_list': [], 'op': OP.ANY},
    ),
    (
        {
            'op': OP.OR,
            'content': [
                {'op': OP.IN, 'field': 'project.id', 'value': [2, 1, 5]},
                {'op': OP.EQ, 'field': 'project.id', 'value': 3},
                {'op': OP.IN, 'field': 'fake_project.id', 'value': [4, 6]},
            ],
        },
        {'project_id_list': [1, 2, 3, 5], 'op': OP.IN},
    ),
]


class TestProjectPermission:
    @pytest.mark.parametrize('policies, expected_dict_filter', test_dict_filter_data)
    def test_make_dict_filter(self, policies, expected_dict_filter):
        project_perm = ProjectPermission()
        dict_filter = project_perm._make_dict_filter(policies)
        assert dict_filter['project_id_list'].sort() == expected_dict_filter['project_id_list'].sort()
        assert dict_filter['op'] == expected_dict_filter['op']
