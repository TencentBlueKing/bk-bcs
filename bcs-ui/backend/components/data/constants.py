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
import logging

from django.conf import settings

logger = logging.getLogger(__name__)

# 是否开启数据平台功能
IS_DATA_OPEN = False

# 接入ESB后，走ESB访问，确认访问路径
DATA_API_V3_PREFIX = f"{settings.COMPONENT_HOST}/api/c/compapi/data/v3"
# 测试阶段，绕过用户登录态验证
DATA_TOKEN = ""

APP_CODE = settings.APP_ID
APP_SECRET = settings.APP_TOKEN

EXPIRE_TIME = "7d"

# eslog不再支持新增, 按照数据平台要求切换至sz4集群
STORAGE_CLUSTER = "eslog-sz4"

DockerMetricFields = {
    "cpu_summary": ["cpuusage", "id", "container_name"],  # 使用率 cpuusage
    "mem": ["rss", "total", "rss_pct", "id", "container_name"],  # 使用率 rss/total
    "disk": ["used_pct", "device_name", "container_name"],
    "net": ["rxbytes", "txbytes", "rxpackets", "txpackets", "container_name"],
}

NodeMetricFields = {
    "cpu_summary": ["usage"],
    "mem": ["total", "used"],
    "disk": ["in_use", "device_name"],
    "net": ["speedSent", "speedRecv", "device_name"],
    "io": ["rkb_s", "wkb_s", "util", "device_name"],
}

DEFAULT_SEARCH_SIZE = 100
API_URL = f"{DATA_API_V3_PREFIX}/dataquery/query/"

# 替换DockerMetricFields
try:
    from .constants_ext import DockerMetricFields  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
