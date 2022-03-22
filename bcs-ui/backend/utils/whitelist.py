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
# TODO: apps/whitelist是否可以统一
from backend.utils.func_controller import get_func_controller


def can_access_webconsole(app_code: str, project_id_or_code: str) -> bool:
    """蓝鲸应用是否可以访问webconsole接口
    NOTE：存储内容包含app_code和project信息(包含project_code和project_id)，格式app_code:project_id_or_code
    """
    func_code = "APP_ACCESS_WEBCONSOLE"
    enabled, wlist = get_func_controller(func_code)
    return enabled or f"{app_code}:{project_id_or_code}" in wlist


def is_app_open_api_trusted(app_code: str) -> bool:
    """
    校验访问 open api 的蓝鲸应用是可信任的，用以通过传递的username获取用户信息

    :param app_code: 蓝鲸应用编码
    :return: 返回是否可信任
    """
    func_code = "TRUSTED_APPS_FOR_OPEN_API"
    enabled, wlist = get_func_controller(func_code)
    wlist.extend(["bk_bcs_monitor", "bk_harbor", "bk_bcs", "workbench", "helm-plugin"])
    return enabled or app_code in wlist
