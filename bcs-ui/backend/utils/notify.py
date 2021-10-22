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
from celery import shared_task
from django.conf import settings

from backend.utils.func_controller import get_func_controller
from backend.utils.send_msg import send_message

NOTIFY_MANAGER_FUNC_CODE = "notify_manager"


@shared_task
def notify_manager(message):
    """管理员通知"""
    wx_message = '[%s-%s] %s' % (settings.PLAT_SHOW_NAME, settings.PAAS_ENV, message)
    enabled, wlist = get_func_controller(NOTIFY_MANAGER_FUNC_CODE)

    send_message(wlist, wx_message, title=None, send_way='wx')
    send_message(wlist, message, title=None, send_way='rtx')
