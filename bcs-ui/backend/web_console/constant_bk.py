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

各版本差异常量定义
"""
import logging
import re
from collections import OrderedDict

from django.utils.translation import ugettext_lazy as _

logger = logging.getLogger(__name__)

GUIDE_MESSAGE = [
    "Guide: https://bk.tencent.com/docs/markdown/产品白皮书/Function/web_console/Description.md",
    _("支持常用Bash快捷键; Windows下Ctrl-W为关闭窗口快捷键, 请使用Alt-W代替"),
]

MGR_GUIDE_MESSAGE = [
    "Guide: https://bk.tencent.com/docs/markdown/产品白皮书/Function/web_console/Description.md",
    _("支持常用Bash快捷键; Windows下Ctrl-W为关闭窗口快捷键, 请使用Alt-W代替; 使用Alt-Num切换Tab"),
]

# pod版本
KUBECTLD_VERSION = OrderedDict(
    {
        "1.12.10": [re.compile(r"^[vV]?1\.12\.\w+$")],
        "1.14.10": [re.compile(r"^[vV]?1\.14\.\w+$")],
        "1.16.15": [re.compile(r"^[vV]?1\.16\.\w+$")],
        "1.18.20": [re.compile(r"^[vV]?1\.18\.\w+$")],
        "1.20.12": [re.compile(r"^[vV]?1\.20\.\w+$")],
    }
)

DEFAULT_KUBECTLD_VERSION = "1.20.12"


# 尝试加载额外的常量配置，可能会覆盖当前配置
try:
    from .constant_ext import *  # noqa # type: ignore
except ImportError:
    logger.debug('constant extension "constant_ext" module not found')
