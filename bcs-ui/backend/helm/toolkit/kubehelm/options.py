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
from typing import Any, Dict, List, NewType, Optional, Union

RawFlag = NewType('RawFlag', Union[Dict[str, Union[bool, str]], str])


class Options:
    """
    支持 Helm Options 命令行组装
    """

    def __init__(self, init_options: Optional[List[RawFlag]] = None):
        self._options = []

        if not init_options:
            return

        for raw_flag in init_options:
            self._options.append(_Flag(raw_flag))

    def add(self, raw_flag: RawFlag):
        self._options.append(_Flag(raw_flag))

    def options(self) -> List[str]:
        """
        example init_options: [{"--set": "a=1,b=2"}, {"--values": "data.yaml"}, "--debug", {"--force": True}]
        return options: ["--set", "a=1,b=2", "--values", "data.yaml", "--debug", "--force"]
        """
        options = []

        for flag in self._options:
            options.extend(flag.to_cmd_options())

        return options


class _Flag:
    """提供将 RawFlag 转换成 Helm Options 的功能

    转换 RawFlag 的示例:
    dict: {"--set": "a=1"}  to_cmd_options=> ["--set", "a=1"]
    dict: {"--reuse-db": True} to_cmd_options=> ["--reuse-db"]
    dict: {"--reuse-db": False} to_cmd_options=> []
    str: "--reuse-db" to_cmd_options=> ["--reuse-db"]
    """

    def __init__(self, raw_flag: RawFlag):
        self.raw_flag = raw_flag

    def to_cmd_options(self) -> List[str]:
        raw_flag = self.raw_flag

        if isinstance(raw_flag, str):
            return [raw_flag]
        elif isinstance(raw_flag, dict):
            k = list(raw_flag.keys())[0]
            v = raw_flag[k]
            if v is True:
                return [k]
            elif v:
                return [k, str(v)]
            return []

        raise NotImplementedError(f'unsupported type {type(raw_flag)}')
