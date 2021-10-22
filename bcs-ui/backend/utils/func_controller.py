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

平台功能开关
"""
import logging

from django.conf import settings

from backend.container_service.projects.models import FunctionController

logger = logging.getLogger(__name__)


def get_func_controller(func_code):
    # 直接开启的功能开关，不需要在db中配置
    if func_code in settings.DIRECT_ON_FUNC_CODE:
        return True, []

    try:
        ref = FunctionController.objects.filter(func_code=func_code).first()
        if not ref:
            return (False, [])

        if ref.wlist:
            wlist = [i.strip() for i in ref.wlist.split(';')]
        else:
            wlist = []
        return (ref.enabled, wlist)
    except Exception:
        logger.exception("get_func_controller error")
        return (False, [])
