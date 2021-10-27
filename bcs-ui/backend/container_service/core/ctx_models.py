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

保存上下文模型相关内容
"""
from backend.components.base import ComponentAuth
from backend.components.collection import ComponentCollection


class ModelContext:
    """保存当前模型上下文相关内容的对象，可能包括：

    - 当前模型所使用的 ComponentAuth 对象（通常来自用户）
    """

    def __init__(self, auth: ComponentAuth):
        self.auth = auth

    @classmethod
    def from_token(cls, access_token: str) -> 'ModelContext':
        """根据当前 token 生成 context 对象"""
        auth = ComponentAuth(access_token)
        return cls(auth)


class BaseContextedModelMeta(type):
    """模型元类，提供数据校验等功能"""

    def __new__(cls, name, bases, attrs):
        """替换类的 __init__ 方法，接收额外的 context 参数"""
        result_cls = super().__new__(cls, name, bases, attrs)
        _orig_init = result_cls.__init__

        def __init__(self, *args, **kwargs):
            context = kwargs.pop('context', None)
            if not context:
                raise TypeError(f'"context" parameter is required to initialize {result_cls.__name__}')

            self.context = context
            _orig_init(self, *args, **kwargs)

        result_cls.__init__ = __init__
        return result_cls


class BaseContextedModel(metaclass=BaseContextedModelMeta):
    """包含上下文对象的模型基类，为模型类增加额外的初始化参数：context

    :param context: 代表当前对象上下文对象，里面包含鉴权信息等
    """

    @classmethod
    def create(cls, *args, **kwargs):
        """创建一个新对象

        :param token: 当前代表用户身份的 access_token 对象
        :param args/kwargs: 用于创建对象的原始字段
        """
        token = kwargs.pop('token', None)
        if not token:
            raise TypeError('"token" parameter is required')
        context = ModelContext.from_token(token)
        return cls(context=context, *args, **kwargs)

    @property
    def context(self):
        """获取当前上下文对象"""
        return self._context

    @context.setter
    def context(self, context: ModelContext):
        self._context = context

    @property
    def comps(self):
        """获取由当前上下文对象初始化的所有 component 系统组合"""
        return ComponentCollection(self.context.auth)
