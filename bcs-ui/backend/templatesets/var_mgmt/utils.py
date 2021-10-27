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
import re

from backend.templatesets.legacy_apps.configuration.utils import get_all_template_config


def get_variable_quote_num(variable_key, project_id) -> int:
    quote_num = 0
    key_pattern = re.compile(r'"([^"]+)":\s*"([^"]*{{%s}}[^"]*)"' % variable_key)
    template_configs = get_all_template_config(project_id)

    for c in template_configs:
        config_str = c.get('config')
        match_list = key_pattern.findall(config_str)
        quote_num += len(match_list)

    return quote_num
