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
from dataclasses import fields
from typing import Any


def make_dataclass_from_dict(data_cls, init_data: dict) -> Any:
    """与dataclasses.make_dataclass不同，make_dataclass_from_dict支持排除掉init_data中的非属性数据"""
    return data_cls(**{k: v for k, v in init_data.items() if k in [f.name for f in fields(data_cls)]})
