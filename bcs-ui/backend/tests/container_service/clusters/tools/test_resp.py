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

from backend.container_service.clusters.tools.resp import NodeRespBuilder, filter_label_keys, is_reserved_label_key
from backend.resources.constants import NodeConditionStatus

from .conftest import fake_inner_ip


class TestNodeRespBuilder:
    def test_list_nodes(self, client, create_and_delete_node, ctx_cluster):
        resp_builder = NodeRespBuilder(ctx_cluster)
        nodes = resp_builder.list_nodes()
        assert "manifest_ext" in nodes
        manifest_ext = nodes["manifest_ext"]
        for _, ext in manifest_ext.items():
            assert "status" in ext
            assert ext["status"] in NodeConditionStatus.get_values()

    def test_query_labels(self, node_name, client, create_and_delete_node, ctx_cluster):
        resp_builder = NodeRespBuilder(ctx_cluster)
        data = resp_builder.query_labels([node_name])
        assert data[fake_inner_ip]


@pytest.mark.parametrize(
    "label_key,is_reserved", [("test", False), ("kubernetes.io/arch", True), ("xxx.kubernetes.io/instance_type", True)]
)
def test_is_reserved_label_key(label_key, is_reserved):
    assert is_reserved_label_key(label_key) == is_reserved


@pytest.mark.parametrize(
    "keys,expected_keys", [(["kubernetes.io/arch", "test"], ["kubernetes.io/arch"]), (["test", []])]
)
def test_filter_label_keys(keys, expected_keys):
    assert filter_label_keys(keys) == expected_keys
