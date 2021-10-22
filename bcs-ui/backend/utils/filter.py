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
from typing import Dict, List


def filter_by_ips(source: List[Dict], ips: List, key: str, fuzzy: bool = False) -> List:
    """
    通过指定 一个或多个 IP 进行过滤

    :param source: 原始数据，单个包含 IP 信息
    :param ips: 指定的 IP 列表（NOTE 若 GET 请求使用 ',' 连接 IP 需在序列化器中处理，此处不做兼容）
    :param key: 指定的 IP 字段名称
    :param fuzzy: 是否进行模糊匹配
    :return: 过滤结果
    """
    if not (ips and key):
        return source

    ret = []
    for item in source:
        if fuzzy:
            for ip in ips:
                if ip in item[key]:
                    ret.append(item)
                    break
        else:
            # 间隔符为 , ; 或空白符号
            for item_ip in re.split(r'[;,\s]+', item[key]):
                if item_ip in ips:
                    ret.append(item)
                    break
    return ret
