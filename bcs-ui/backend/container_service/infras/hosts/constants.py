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

logger = logging.getLogger(__name__)


# 默认使用的标准运维业务ID
SOPS_BIZ_ID = ""

# 申请主机流程模板ID
APPLY_HOST_TEMPLATE_ID = ""

try:
    from .constants_ext import *  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
