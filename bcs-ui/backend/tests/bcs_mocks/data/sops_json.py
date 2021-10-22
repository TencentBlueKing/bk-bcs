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
fake_task_id = 123
fake_task_url = "http://test.com"

create_task_ok = {"task_id": fake_task_id, "task_url": fake_task_url}
start_task_ok = {"task_id": fake_task_id}

get_task_status_ok = {
    'children': {
        'n93767c5d8d83d94a22bda6423358fda': {
            'children': {},
            'elapsed_time': 0,
            'error_ignorable': False,
            'finish_time': '2021-03-18 16:03:42 +0800',
            'id': 'n93767c5d8d83d94a22bda6423358fda',
            'loop': 1,
            'name': "<class 'pipeline.core.flow.event.EmptyStartEvent'>",
            'retry': 0,
            'skip': False,
            'start_time': '2021-03-18 16:03:42 +0800',
            'state': 'FINISHED',
            'state_refresh_at': '2021-03-18T08:03:42.394Z',
            'version': 'd5d697dc92c73aea97461d93c24b886a',
        },
        'n9a9632fba9e39efa587cfc3a0666e1a': {
            'children': {},
            'elapsed_time': 734,
            'error_ignorable': False,
            'finish_time': '',
            'id': 'n9a9632fba9e39efa587cfc3a0666e1a',
            'loop': 1,
            'name': '申领服务器（轮询）',
            'retry': 0,
            'skip': False,
            'start_time': '2021-03-18 16:03:44 +0800',
            'state': 'RUNNING',
            'state_refresh_at': '2021-03-18T08:03:44.738Z',
            'version': 'faed5baeb6003518bff75235aea6b6bb',
        },
        'nc7e829074ca3024a393e1187cfde9f5': {
            'children': {},
            'elapsed_time': 2,
            'error_ignorable': False,
            'finish_time': '2021-03-18 16:03:44 +0800',
            'id': 'nc7e829074ca3024a393e1187cfde9f5',
            'loop': 1,
            'name': '申请CVM服务器',
            'retry': 0,
            'skip': False,
            'start_time': '2021-03-18 16:03:42 +0800',
            'state': 'FINISHED',
            'state_refresh_at': '2021-03-18T08:03:44.662Z',
            'version': '982549095a6b35b2bceebadee64c516d',
        },
        'n93767c5d8d83d94a22bda6423358fd1': {
            'children': {},
            'elapsed_time': 0,
            'error_ignorable': False,
            'finish_time': '2021-03-18 16:03:42 +0800',
            'id': 'n93767c5d8d83d94a22bda6423358fda',
            'loop': 1,
            'name': "<class 'pipeline.core.flow.event.EmptyEndEvent'>",
            'retry': 0,
            'skip': False,
            'start_time': '2021-03-18 16:03:42 +0800',
            'state': 'FINISHED',
            'state_refresh_at': '2021-03-18T08:03:42.394Z',
            'version': 'd5d697dc92c73aea97461d93c24b886a',
        },
    },
    'elapsed_time': 737,
    'error_ignorable': False,
    'finish_time': '',
    'id': 'n9355daa30c330188f02d5a75bc17e2a',
    'loop': 1,
    'name': "<class 'pipeline.core.pipeline.Pipeline'>",
    'retry': 0,
    'skip': False,
    'start_time': '2021-03-18 16:03:42 +0800',
    'state': 'RUNNING',
    'state_refresh_at': '2021-03-18T08:03:42.374Z',
    'version': '',
}
