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

from backend.iam.permissions.filter import ProjectFilter
from backend.tests.testing_utils.base import generate_random_string

project_id1 = generate_random_string(32)
project_id2 = generate_random_string(32)
project_id3 = generate_random_string(32)


@pytest.mark.parametrize(
    'policies, expect_project_id_list, expect_op',
    [
        ({'field': 'project.id', 'op': OP.ANY, 'value': []}, [], OP.ANY),
        ({'field': 'project.id', 'op': OP.EQ, 'value': project_id1}, [project_id1], OP.IN),
        ({'field': 'project.id', 'op': OP.IN, 'value': [project_id1, project_id2]}, [project_id1, project_id2], OP.IN),
        (
            {
                'op': OP.OR,
                'content': [
                    {'op': OP.IN, 'field': 'project.id', 'value': [project_id1, project_id2]},
                    {'op': OP.IN, 'field': 'project.id', 'value': [project_id1, project_id3]},
                ],
            },
            [project_id1, project_id2, project_id3],
            OP.IN,
        ),
        (
            {
                'op': OP.AND,
                'content': [
                    {'op': OP.ANY, 'field': 'project.id', 'value': []},
                    {'op': OP.IN, 'field': 'project.id', 'value': [project_id1, project_id3]},
                ],
            },
            [],
            OP.ANY,
        ),
    ],
)
def test_make_dict_filter(policies, expect_project_id_list, expect_op):
    dict_filter = ProjectFilter._make_dict_filter(policies)
    assert not (set(dict_filter['value']) - set(expect_project_id_list))
    assert dict_filter['op'] == expect_op
