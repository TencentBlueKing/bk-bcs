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

常量定义
"""
import copy
import re
from enum import Enum

from django.conf import settings

from backend.web_console.constant_bk import *  # noqa

# 输入分行标识
INPUT_LINE_BREAKER = "\r"
# 输出分行标识
OUTPUT_LINE_BREAKER = "\r\n"

STDIN_CHANNEL = 0
STDOUT_CHANNEL = 1
STDERR_CHANNEL = 2
ERROR_CHANNEL = 3
RESIZE_CHANNEL = 4

# bash 颜色标识
ANSI_ESCAPE = re.compile(r"\x1B\[[0-?]*[ -/]*[@-~]")

# hello 字符
HELLO_MESSAGE = "Welcome To BCS Console"


# ping/pong时间间隔
WEBSOCKET_PING_INTERVAL = 10
# pod清理时间间隔
CLEAN_USER_POD_INTERVAL = 60
# 锁偏差时间常量
LOCK_SHIFT = -2

# 链接自动断开时间, 30分钟
TICK_TIMEOUT = 60 * 30
# 自动登出时间
LOGIN_TIMEOUT = 60 * 60 * 24
# 清理POD，4个小时
USER_POD_EXPIRE_TIME = 3600 * 4
# Context 过期时间, 12个小时
USER_CTX_EXPIRE_TIME = 3600 * 12

WEB_CONSOLE_HEARTBEAT_KEY = "bcs::web_console::heartbeat"
LABEL_WEB_CONSOLE_CREATE_TIMESTAMP = "io.tencent.web_console.create_timestamp"
NAMESPACE = "web-console"

# 1080p页面测试得来
DEFAULT_COLS = 211
DEFAULT_ROWS = 25

# 默认命令, bash 有宽度兼容性问题, 默认都使用 sh 命令
DEFAULT_COMMAND = [
    '/bin/sh',
    '-c',
    'export TERM=xterm-256color; export PS1="\\u:\W$ "; exec /bin/sh',
]


class WebConsoleMode(Enum):
    # 用户自己集群
    INTERNEL = "internel"
    # 平台集群
    EXTERNAL = "external"
