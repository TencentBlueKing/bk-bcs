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
import re
from datetime import timedelta

# 时间表达式 匹配器
TIME_DURATION_PATTERN = re.compile(r'^((?P<hours>\d+)h)?((?P<minutes>\d+)m)?((?P<seconds>\d+)s)?$')


def get_duration_seconds(duration: str, default: int = None) -> int:
    """
    解析时间表达式，获取持续的时间（单位：s）
    支持 时(h)，分(m)，秒(s) 为单位，暂不支持 天，月，年 等

    :param duration: 时间表达式，格式如 3h5m7s
    :param default: 默认值
    :return: duration 对应秒数
    """
    if not duration:
        return default

    match = TIME_DURATION_PATTERN.match(duration)
    if not match:
        return default

    delta = timedelta(**{k: int(v) for k, v in match.groupdict().items() if v})
    return int(delta.total_seconds())
