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

from backend.components import cc


def list_biz_maintainers(biz_id: int) -> List[str]:
    """查询业务的运维角色"""
    return cc.get_app_maintainers("admin", biz_id)


def is_biz_maintainer(biz_id: int, username: str) -> bool:
    """判断用户是否为业务的运维角色"""
    maintainers = list_biz_maintainers(biz_id)
    return username in maintainers
