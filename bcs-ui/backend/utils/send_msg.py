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

from backend.components.enterprise.cmsi import send_mail, send_sms, send_weixin


def send_message(receiver, message, title=None, send_way=None):
    """统一的发送消息方式
    发送方式(send_way)参数忽略，所有的通知都邮件、短信、微信全部发送
    """
    if isinstance(receiver, list):
        receiver = ','.join(receiver)
    send_mail(title, message, receiver)
    send_weixin(title, message, receiver)
    send_sms(message, receiver)


try:
    from .send_msg_ext import send_message  # noqa
except ImportError:
    pass
