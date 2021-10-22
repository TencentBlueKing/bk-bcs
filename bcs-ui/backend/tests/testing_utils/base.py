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
import contextlib
import random
from typing import Dict

RANDOM_CHARACTER_SET = 'abcdefghijklmnopqrstuvwxyz0123456789'


def generate_random_string(length=30, chars=RANDOM_CHARACTER_SET):
    """Generates a non-guessable OAuth token"""
    rand = random.SystemRandom()
    return ''.join(rand.choice(chars) for x in range(length))


def dict_is_subequal(data: Dict, full_data: Dict) -> bool:
    """检查两个字典是否相等，忽略在 `full_data` 中有，但 `data` 里没有提供的 key"""
    for key, value in data.items():
        if key not in full_data:
            return False
        if value != full_data[key]:
            return False
    return True


@contextlib.contextmanager
def nullcontext():
    """A context manager which does nothing"""
    yield
