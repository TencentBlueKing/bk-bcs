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

from django.utils.crypto import get_random_string
from django.utils.translation import ugettext_lazy as _

logger = logging.getLogger(__name__)


def create_prometheus_data_flow(username, project_id, cc_app_id, english_name, dataset):
    """prometheus 类型的Metric申请数据平台的dataid，并配置默认的清洗入库规则"""
    return True, _("数据平台功能暂未开启")


def apply_dataid_by_metric(biz_id, dataset, operator):
    # 数据平台功能没有开启，则直接返回
    return True, 0


def get_metric_data_name(metric_name, project_id):
    """
    数据平台(raw_data_name)长度限制为 30 个字符
    metric_name 最大长度为 28 个字符
    """
    metric_name_len = len(metric_name)
    rand = get_random_string(30 - metric_name_len - 1)
    raw_data_name = f"{metric_name}_{rand}"
    return raw_data_name
