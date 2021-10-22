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
from contextlib import contextmanager
from types import MethodType
from typing import Callable
from unittest import mock


class mockable_function:
    """增加一个快捷 mock 函数入口"""

    def __init__(self, func: Callable):
        self._func = func

    def __call__(self, *args, **kwargs):
        return self._func(*args, **kwargs)

    def __get__(self, obj, obj_type=None):
        """实现描述符协议，兼容装饰类方法时的情况"""
        if obj is None:
            return self
        return MethodType(self, obj)

    def mock(self, *args, **kwargs):
        """返回 Mock 当前函数的上下文管理器"""

        @contextmanager
        def _mocking_func():
            mocked_obj = mock.MagicMock(*args, **kwargs)
            orig_func = self._func
            self._func = mocked_obj
            yield
            self._func = orig_func

        return _mocking_func()
