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
import json

from django.conf import settings
from django.db import migrations
from iam.contrib.iam_migration.migrator import IAMMigrator

MIGRATION_JSON_FILE_NAME = '0001_bk_bcs_app_20200612-1102_iam.json'


def render_json_conf_file(*args, **kwargs):
    """ 将容器服务 SaaS API Host 渲染到配置文件中 """
    file_path = f'{settings.BASE_DIR}/support-files/iam/{MIGRATION_JSON_FILE_NAME}'
    with open(file_path, 'r') as fr:
        conf = json.loads(fr.read())
    # 替换掉 BCS_SAAS_URL
    conf['operations'][0]['data']['provider_config']['host'] = settings.DEVOPS_BCS_API_URL
    with open(file_path, 'w') as fw:
        fw.write(json.dumps(conf, ensure_ascii=False, indent=2))


def forward_func(*args, **kwargs):
    migrator = IAMMigrator(MIGRATION_JSON_FILE_NAME)
    migrator.migrate()


class Migration(migrations.Migration):
    dependencies = []

    operations = [
        migrations.RunPython(render_json_conf_file),
        migrations.RunPython(forward_func)
    ]
