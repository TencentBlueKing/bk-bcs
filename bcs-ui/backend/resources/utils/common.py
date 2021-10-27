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
import arrow


def calculate_age(create_at: str) -> str:
    """
    DashBoard 用，计算当前对象存活时间

    :param create_at: 对象创建时间
    :return: 对象存活时间
    """
    return calculate_duration(create_at)


def calculate_duration(start: str, end: str = None) -> str:
    """
    计算 起始 至 终止时间 间时间间隔（带单位），例：
    1. start: '2021-04-01 12:35:30' end: '2021-04-03 14:00:00' => '2d1h'
    2. start: '2021-04-01 12:35:30' end: '2021-04-01 12:59:59' => '24m29s'

    :param start: 起始时间
    :param end: 终止时间
    :return: 持续时间（带单位）
    """
    if not start:
        return '--'

    start = arrow.get(start)
    end = arrow.get(end) if end else arrow.utcnow()

    if end <= start:
        return '--'

    duration = end - start
    days = duration.days
    seconds = duration.seconds
    minutes = (seconds % 3600) // 60
    hours = seconds // 3600 % 24
    seconds %= 60

    units = ['d', 'h', 'm', 's']
    values = [days, hours, minutes, seconds]
    for idx, v in enumerate(values):
        next_idx = idx + 1
        if v <= 0:
            continue
        if next_idx < len(values) and values[next_idx] > 0:
            return f'{v}{units[idx]}{values[next_idx]}{units[next_idx]}'
        return f'{v}{units[idx]}'
