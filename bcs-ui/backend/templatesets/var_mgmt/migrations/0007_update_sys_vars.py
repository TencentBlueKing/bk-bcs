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
# 初始化系统内置变量
from __future__ import unicode_literals

import json

from django.db import migrations

from ..models import Variable

SYS_VARS = [
    {
        "key": "SYS_STANDARD_DATA_ID",
        "name": u"标准日志采集DataId",
        "scope": "global",
    },
    {
        "key": "SYS_NON_STANDARD_DATA_ID",
        "name": u"非标准日志采集DataId",
        "scope": "global",
    },
]


def init_sys_vars(apps, schema_editor):
    for _var in SYS_VARS:
        _d = {
            "name": _var['name'],
            "scope": _var['scope'],
            "project_id": 0,
            "category": "sys",
            "default": json.dumps({"value": ""}),
        }
        Variable.objects.update_or_create(key=_var['key'], defaults=_d)


class Migration(migrations.Migration):

    dependencies = [
        ('variable', '0006_auto_20180409_1118'),
    ]

    operations = [migrations.RunPython(init_sys_vars)]
