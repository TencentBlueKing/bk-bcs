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
from typing import List
from urllib import parse


def get_cors_allowed_origins(raw_urls: List[str]) -> List[str]:
    """获取允许的origin列表(精确过滤)"""

    origins = []
    for url in raw_urls:
        parsed = parse.urlparse(url)
        origin = f'{parsed.scheme}://{parsed.netloc}'
        origins.append(origin)

        if not parsed.port:
            continue

        # 如果 origin 带端口，将不带端口的也加入 origins 中(主要是处理80和443)
        origins.append(origin[: origin.rfind(str(parsed.port)) - 1])

    # 去重返回
    return list(set(origins))
