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
from itertools import groupby as _groupby


def groupby(iterable, key, reverse=False):
    """排序分组"""
    return _groupby(sorted(iterable, key=key, reverse=reverse), key=key)


def recursive_groupby(iterable, keys, reverse=False):
    """迭代分组"""
    if not keys:
        yield ([], iterable)
    else:
        key = keys[0]
        for g, iters in groupby(iterable, key=key, reverse=reverse):
            for ret in recursive_groupby(iters, keys[1:], reverse=reverse):
                ret[0].insert(0, g)
                yield ret
