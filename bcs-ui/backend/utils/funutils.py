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
from urllib.parse import urlparse


def convert_mappings(mappings, data, reversed=False, default=NotImplemented):
    result = {}
    for k1, k2 in mappings.items():
        if reversed:
            key, target = k2, k1
        else:
            key, target = k1, k2
        if target not in data and default is NotImplemented:
            continue
        result[key] = data[target]
    return result


def num_transform(num, format='to_zore'):
    """数字转换
    to_zore: 标识负值转换为0
    """
    return {'to_zore': lambda x: x if x > 0 else 0}.get(format)(num)


def remove_url_domain(url: str) -> str:
    """去掉域名, 调用域名在前端指定"""
    parsed_url = urlparse(url)
    # 去掉http, https, 域名
    domain_less_url = parsed_url._replace(scheme="", netloc="").geturl()
    return domain_less_url
