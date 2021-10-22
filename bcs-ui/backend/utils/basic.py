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
import base64
import hashlib
import json
import logging
import re
from enum import Enum
from functools import reduce
from typing import Any, Dict, List, Union

import arrow
from django.conf import settings
from django.utils import timezone
from rest_framework import fields
from rest_framework.exceptions import ValidationError

logger = logging.getLogger(__name__)
MOSAIC_REG = re.compile(r"(?<=\d{3})(\d{4})(?=\d{4})")


class ChoicesEnum(Enum):
    """Enum with choices"""

    @classmethod
    def get_choices(cls):
        return cls._choices_labels.value

    @classmethod
    def get_choice_label(cls, value):
        if isinstance(value, Enum):
            value = value.value
        return dict(cls.get_choices()).get(value, value)

    @classmethod
    def choice_values(cls):
        return [item[0] for item in cls.get_choices()]

    @classmethod
    def choice_labels(cls):
        return [item[1] for item in cls.get_choices()]


def get_client_ip(request):
    """获取客户端IP"""
    x_forwarded_for = request.META.get("HTTP_X_FORWARDED_FOR")
    if x_forwarded_for:
        ip = x_forwarded_for.split(",")[0]
    else:
        ip = request.META.get("REMOTE_ADDR", "")
    return ip


def mosaic_phone(phone):
    return re.sub(MOSAIC_REG, "****", phone)


def normalize_metric(metric):
    """监控返回标准，保留两位小数"""
    return float("%.2f" % float(metric))


class RequestToken(object):
    def __init__(self, access_token):
        self.access_token = access_token


class RequestUser(object):
    def __init__(self, username, access_token):
        self.username = username
        self.token = RequestToken(access_token)


class RequestProject(object):
    def __init__(self, project_code):
        self.english_name = project_code


class RequestClass(object):
    def __init__(self, username, access_token, project_code):
        self.user = RequestUser(username, access_token)
        self.project = RequestProject(project_code)


def normalize_datetime(time):
    """time format is YYYY-MM-DD HH:mm:ss"""
    arrow_time = arrow.get(time, tzinfo=settings.TIME_ZONE)
    return arrow_time.datetime.strftime(settings.REST_FRAMEWORK["DATETIME_FORMAT"])


def getitems(obj: Dict, items: Union[List, str], default: Any = None) -> Any:
    """
    递归获取数据
    注意：使用字符串作为键路径时，须确保 Key 值均为字符串

    :param obj: Dict 类型数据
    :param items: 键列表：['foo', 'bar']，或者用 "." 连接的键路径： ".foo.bar" 或 "foo.bar"
    :param default: 默认值
    :return: value
    """
    if not isinstance(obj, dict):
        raise TypeError('Dict object support only!')
    if isinstance(items, str):
        items = items.strip(".").split(".")
    try:
        return reduce(lambda x, i: x[i], items, obj)
    except (IndexError, KeyError, TypeError):
        return default


def get_with_placeholder(obj: Dict, items: Union[List, str], placeholder: str = '--') -> Any:
    """
    获取 Dict Key 值，若不存在使用统一占位符

    :param obj: Dict 类型对象
    :param items: 键列表，格式参考 getitems 注释
    :param placeholder: 指定的占位符，默认为 '--'
    :return: value
    """
    return getitems(obj, items, placeholder)


def get_bcs_component_version(cluster_version, bcs_component_version_info, default_version):
    if not cluster_version:
        return default_version

    for component_version, patterns in bcs_component_version_info.items():
        for pattern in patterns:
            if pattern.match(cluster_version):
                return component_version

    return default_version


def normalize_time(time):
    # create_time format: '2019-12-16T09:10:59Z'
    d_time = arrow.get(time).datetime
    return timezone.localtime(d_time).strftime(settings.REST_FRAMEWORK["DATETIME_FORMAT"])


def str2bool(source, default=False):
    """str转换为bool
    True: "true", "True", "1", 1
    False: "false", "False", "0", 0
    """
    try:
        return fields.BooleanField().to_internal_value(source)
    except ValidationError:
        return default


def b64encode_json(data: Any) -> bytes:
    """返回base64.b64encode(bytes(json.dumps(data), 'utf-8'))"""
    return base64.b64encode(bytes(json.dumps(data), "utf-8"))


def md5_digest(content: str) -> str:
    h = hashlib.md5()
    h.update(content.encode("utf-8"))
    return h.hexdigest()
