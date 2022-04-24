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
import os
from pathlib import Path

import pytest

from backend.change_log.change_log import ChangeLog


class TestChangeLog:
    def test_success(self):
        # 获取md文档目录
        base_dir = Path(__file__).parent
        log_dir = base_dir.joinpath("CHANGELOG")
        log_list = ChangeLog(path=log_dir.as_posix()).list()
        assert len(log_list) == 2
        # version和date是通过文件名解析获取
        assert log_list[0]["version"] == "v1.0.1"

    @pytest.mark.parametrize(
        "path,expected_length",
        [
            # 路径不正确
            (Path(__file__).as_posix(), 0),
            # 文件后缀不正确
            (Path(__file__).joinpath("CHANGELOGTEST").as_posix(), 0),
        ],
    )
    def test_fail(self, path, expected_length):
        log_list = ChangeLog(path=path).list()
        assert len(log_list) == expected_length
