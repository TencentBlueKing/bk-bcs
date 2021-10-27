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

from backend.helm.toolkit.kubehelm.options import Options

init_options1 = [{"--set": "a=1,b=2"}, {"--values": "data.yaml"}]
flag1 = "--debug"
expect_options1 = ["--set", "a=1,b=2", "--values", "data.yaml", "--debug"]

init_options2 = [{"--set": "a=1,b=2"}, {"--values": "data.yaml"}, {"--debug": False}]
flag2 = {"--reuse-values": True}
expect_options2 = ["--set", "a=1,b=2", "--values", "data.yaml", "--reuse-values"]

init_options3 = ["--set", "a=1,b=2", "--values", "data.yaml", "--debug"]
flag3 = "--reuse-values"
expect_options3 = ["--set", "a=1,b=2", "--values", "data.yaml", "--debug", "--reuse-values"]


@pytest.mark.parametrize(
    "init_options, flag, expect_options",
    [
        (init_options1, flag1, expect_options1),
        (init_options2, flag2, expect_options2),
        (init_options3, flag3, expect_options3),
    ],
)
def test_options_lines(init_options, flag, expect_options):
    opts = Options(init_options)
    opts.add(flag)
    assert opts.options() == expect_options
