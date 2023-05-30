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
import codecs
import json
import os

from django.conf import settings
from django.db import migrations
from iam.contrib.iam_migration.migrator import IAMMigrator


def forward_func(apps, schema_editor):
    if settings.EDITION == settings.COMMUNITY_EDITION and os.environ.get("ENABLE_TEMPLATESET_PERMISSION", False):
        migrator = IAMMigrator(Migration.migration_json)
        migrator.migrate()


class Migration(migrations.Migration):
    migration_json = "0027_upsert_action_groups_sg.json"

    dependencies = [('bcs_iam_migration', '0026_bk_bcs_app_202305251457')]

    operations = [migrations.RunPython(forward_func)]
