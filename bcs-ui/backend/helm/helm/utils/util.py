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
import datetime
import logging
import urllib

import pytz
from dateutil import parser
from furl import furl

logger = logging.getLogger(__name__)


def parse_chart_time(time_string):
    # 外部输入的 created 出现各种形式
    # 1. yaml load 解析出来是 datetime 类型
    # 2. 字符串,形如: `2018-02-27T17:49:56.232875637Z`
    # 3. 字符串,形如: `2018-07-30T14:22:31Z`
    try:
        time_value = parser.parse(time_string)
    except Exception as e:
        logger.exception("dateutil.parser.parse:%s failed, %s", time_string, e)
        # if failed, run history implemente
        if not time_string:
            raise ValueError(time_string)

        if isinstance(time_string, datetime.datetime):
            time_value = time_string.astimezone(pytz.utc)
        elif isinstance(time_string, str):
            try:
                time_value = datetime.datetime.strptime(time_string, '%Y-%m-%dT%H:%M:%S.%fZ').astimezone(pytz.utc)
            except Exception:
                time_value = datetime.datetime.strptime(time_string, '%Y-%m-%dT%H:%M:%SZ').astimezone(pytz.utc)
        else:
            raise ValueError(time_string)

    return time_value


class EmptyVaue(Exception):
    pass


def fix_rancher_value_by_type(value, item_type):
    result = value
    if item_type == "boolean":
        if isinstance(value, bool):
            result = value
        else:
            result = True if value == "true" else False

    elif item_type == "int":
        # empty value should be skipped
        if isinstance(value, str) and not value.strip():
            raise EmptyVaue("value is: %s" % value)

        result = int(value)

    elif item_type == "float":
        # empty value should be skipped
        if isinstance(value, str) and not value.strip():
            raise EmptyVaue("value is: %s" % value)

        result = float(value)

    elif item_type in ["string", "password"]:
        result = str(value)

    else:
        result = str(value)

    return result


def merge_rancher_answers(answers, customs):
    parameters = dict()
    if not isinstance(customs, list):
        raise ValueError(customs)

    for item in customs:
        # TODO 前端需要增加类型支持
        parameters[item["name"]] = item["value"]

    if not isinstance(answers, list):
        return parameters
        # raise ValueError(answers)

    for item in answers:
        item_type = item.get("type")
        value = item["value"]

        try:
            value = fix_rancher_value_by_type(value, item_type)
        except EmptyVaue:
            continue
        else:
            parameters[item["name"]] = value

    return parameters


def fix_chart_url(url, repo_url):
    # 修正urls（chartmuseum默认没有带repo.url）
    # parameters:
    #  url: charts/chartmuseum-curator-0.9.0.tgz
    #  repo_url: http://charts.bking.com
    # ==> http://charts.bking.com/charts/chartmuseum-curator-0.9.0.tgz
    # s-1: path ==> integrated url
    hostname = urllib.parse.urlparse(repo_url).netloc
    if hostname not in url:
        integrated_url = urllib.parse.urljoin(repo_url, url)
    else:
        integrated_url = url

    # s-2: replace '//' with '/'
    f = furl(integrated_url)
    f.path.segments = [x for x in f.path.segments if x]
    return f.url
