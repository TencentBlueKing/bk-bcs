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
from django.core.management.base import BaseCommand

from backend.accounts.bcs_perm import tasks


class Command(BaseCommand):
    def add_arguments(self, parser):
        parser.add_argument('--all-project', action='store_true', help='同步所有项目')
        parser.add_argument('project_id', nargs='*', type=str)

    def handle(self, *args, **options):
        tasks.sync_bcs_perm_script(options['all_project'], options['project_id'])
