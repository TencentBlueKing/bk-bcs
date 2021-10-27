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

模板实例化过程中用到的通用方法
"""
import collections
import copy

from django.template import Context, Template


def update_nested_dict(orginal_dict, update_dict):
    """"""
    new_dict = copy.deepcopy(orginal_dict)
    for k, v in update_dict.items():
        if isinstance(v, collections.Mapping):
            new_dict[k] = update_nested_dict(new_dict.get(k, {}), v)
        else:
            new_dict[k] = v
    return new_dict


def render_mako_context(content, context):
    """通过mako模板做变量替换
    note: 所有的变量必须添加到 context 中
    """
    return Template(f"{{% autoescape off %}}{content}{{% endautoescape %}}").render(Context(context))
