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

from backend.helm.toolkit.kubehelm.helm import KubeHelmClient

fake_init_cmd_args = ["helm", "install", "name", "--namespace", "namespace"]
fake_chart_path = "/chart_path"
fake_values_path = "/values_path"
fake_post_renderer_config_path = "/config_path"
fake_cmd = fake_init_cmd_args + [fake_chart_path, "--post-renderer", f"{fake_post_renderer_config_path}/ytt_renderer"]
fake_cmd_with_values = fake_cmd + ["--values", fake_values_path]


@pytest.mark.parametrize(
    "chart_path, values_path, post_renderer_config_path, cmd_flags, expect_args",
    [
        (
            fake_chart_path,
            fake_values_path,
            fake_post_renderer_config_path,
            ["--reuse-values"],
            fake_cmd + ["--reuse-values"],
        ),
        (
            fake_chart_path,
            fake_values_path,
            fake_post_renderer_config_path,
            [{"--reuse-values": True}],
            fake_cmd + ["--reuse-values"],
        ),
        (
            fake_chart_path,
            fake_values_path,
            fake_post_renderer_config_path,
            [{"--reuse-values": True}],
            fake_init_cmd_args
            + [fake_chart_path]
            + ["--post-renderer", f"{fake_post_renderer_config_path}/ytt_renderer"]
            + ["--reuse-values"],
        ),
        (
            None,
            fake_values_path,
            fake_post_renderer_config_path,
            [{"--set": "a=v1"}],
            fake_init_cmd_args
            + ["--post-renderer", f"{fake_post_renderer_config_path}/ytt_renderer", "--values", fake_values_path]
            + ["--set", "a=v1"],
        ),
        (
            fake_chart_path,
            fake_values_path,
            None,
            [{"--set": "a=v1"}],
            fake_init_cmd_args + [fake_chart_path, "--values", fake_values_path] + ["--set", "a=v1"],
        ),
        (
            None,
            fake_values_path,
            None,
            [{"--set": "a=v1"}],
            fake_init_cmd_args + ["--values", fake_values_path] + ["--set", "a=v1"],
        ),
        (
            None,
            None,
            None,
            [{"--set": "a=v1"}],
            fake_init_cmd_args + ["--set", "a=v1"],
        ),
    ],
)
def test_compose_cmd_args(chart_path, values_path, post_renderer_config_path, cmd_flags, expect_args):
    client = KubeHelmClient()
    init_cmd_args = fake_init_cmd_args.copy()
    composed_cmd_args = client._compose_cmd_args(
        init_cmd_args, chart_path, values_path, post_renderer_config_path, cmd_flags
    )
    assert composed_cmd_args == expect_args
