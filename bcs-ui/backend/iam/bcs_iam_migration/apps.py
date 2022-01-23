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
from __future__ import unicode_literals

import os
from pathlib import Path

import jinja2
from django.apps import AppConfig
from django.conf import settings

iam_context = {
    'BK_IAM_SYSTEM_ID': settings.BK_IAM_SYSTEM_ID,
    'APP_CODE': settings.APP_CODE,
    'BK_IAM_PROVIDER_PATH_PREFIX': settings.BK_IAM_PROVIDER_PATH_PREFIX,
}


def render_migrate_json():
    """根据模板生成最终的 migrate json 文件"""
    iam_tpl_path = Path(settings.BASE_DIR) / 'support-files' / 'iam_tpl'
    iam_tpl = Path(settings.BASE_DIR) / 'support-files' / 'iam'
    iam_tpl.mkdir(exist_ok=True)

    j2_env = jinja2.Environment(loader=jinja2.FileSystemLoader(iam_tpl_path), trim_blocks=True)
    for dir in iam_tpl_path.iterdir():
        j2_env.get_template(dir.name).stream(**iam_context).dump(str(iam_tpl / dir.name[:-3]))


class BcsIamMigrationConfig(AppConfig):
    name = "backend.iam.bcs_iam_migration"
    label = "bcs_iam_migration"

    def ready(self):
        render_migrate_json()
