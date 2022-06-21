# -*- coding: utf-8 -*-
"""
Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community
Edition) available.
Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://opensource.org/licenses/MIT

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
"""
import json
import time
from pathlib import Path

from django.conf import settings
from django.core.management.base import BaseCommand, CommandError
from rest_framework import exceptions, serializers

from backend.container_service.cluster_tools.models import Tool


class ToolSLZ(serializers.ModelSerializer):
    # 去除 chart_name 的 unique=True 约束
    chart_name = serializers.CharField()

    class Meta:
        model = Tool
        exclude = ['version']


class Command(BaseCommand):
    help = 'Register the cluster tools'

    def handle(self, *args, **options):
        self._register()

    def _register(self):
        """创建或者更新组件信息"""
        latest_file = self._get_latest_file()
        tools_data = json.loads(latest_file.read_text())
        try:
            serializer = ToolSLZ(data=tools_data['tools'], many=True)
            serializer.is_valid(raise_exception=True)
        except exceptions.ValidationError as e:
            raise CommandError(f'Register {latest_file.name} failed: {e}')

        tool_version = tools_data['version']
        for tool in serializer.validated_data:
            chart_name = tool.pop('chart_name')
            tool['version'] = tool_version
            Tool.objects.update_or_create(chart_name=chart_name, defaults=tool)

        self.stdout.write(self.style.SUCCESS(f'Register {latest_file.name} successfully'))

    def _get_latest_file(self) -> Path:
        """获取最新的版本文件"""
        file_path = Path(settings.BASE_DIR) / 'support-files' / 'cluster_tools'
        # 设定一个较早时间的文件
        latest_file_name = 'tools-190001010000.json'
        latest_time = self._str_to_time(latest_file_name)
        for f in file_path.iterdir():
            f_time = self._str_to_time(f.name)
            if f_time > latest_time:
                latest_file_name = f.name
                latest_time = f_time

        return file_path / latest_file_name

    def _str_to_time(self, file_name: str) -> time.struct_time:
        """转换文件名中的时间字符串"""
        try:
            time_str = file_name.split('.json')[0].split('tools-')[1]
            return time.strptime(time_str, '%Y%m%d%H%M')
        except Exception as e:
            raise CommandError(f'Register failed: {e}')
