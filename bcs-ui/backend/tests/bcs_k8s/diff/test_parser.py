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

from backend.helm.toolkit.diff import parser

manifest1 = b"""apiVersion: v1
kind: Deployment
metadata:
  name: nginx
  namespace: default
"""
expect_manifest_list1 = [b"apiVersion: v1\nkind: Deployment\nmetadata:\n  name: nginx\n  namespace: default\n"]

manifest2 = b"""---
apiVersion: v1
kind: Deployment
metadata:
  name: nginx
  namespace: default"""
expect_manifest_list2 = [b"---\napiVersion: v1\nkind: Deployment\nmetadata:\n  name: nginx\n  namespace: default"]


manifest3 = b"""apiVersion: v1
kind: Deployment
metadata:
  name: nginx
  namespace: default
  
---

apiVersion: v1
kind: ConfigMap
metadata:
  name: perf-config
data: 
  perf-conf: |
    ---
    shmkey: 6069
    ---
    processes:
    - id: id
    ---"""  # noqa

expect_manifest_list3 = [
    b"apiVersion: v1\nkind: Deployment\nmetadata:\n  name: nginx\n  namespace: default\n  ",
    b'\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: perf-config\ndata: \n  perf-conf: |\n    ---\n    '
    b'shmkey: 6069\n    ---\n    processes:\n    - id: id\n    ---',
]


@pytest.mark.parametrize(
    'manifest, expect_manifest_list',
    [(manifest1, expect_manifest_list1), (manifest2, expect_manifest_list2), (manifest3, expect_manifest_list3)],
)
def test_split_manifest(manifest, expect_manifest_list):
    contents = parser.split_manifest(manifest)
    assert contents == expect_manifest_list
