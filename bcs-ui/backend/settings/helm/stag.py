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
# 预发布环境配置文件
from .base import *  # noqa

# ******************************** 日志 配置 ********************************
BK_LOG_DIR = os.environ.get('BKAPP_LOG_DIR', '/app/logs/')
LOG_CLASS = 'logging.handlers.RotatingFileHandler'

# 预发布环境直接使用 INFO
LOGGING_DIR = os.path.join(BK_LOG_DIR, APP_ID)
LOG_LEVEL = 'INFO'

# 兼容企业版
LOGGING_DIR = os.environ.get('LOGGING_DIR', LOGGING_DIR)

# 自动建立日志目录
if not os.path.exists(LOGGING_DIR):
    os.makedirs(LOGGING_DIR)

LOG_FILE = os.path.join(LOGGING_DIR, f'bcs_ui.log')
LOGGING = get_logging_config(LOG_LEVEL, None, LOG_FILE)

# ******************************** 容器服务相关配置 ********************************

# PaaS域名，发送邮件链接需要
PAAS_HOST = BK_PAAS_HOST
PAAS_ENV = 'stag'
