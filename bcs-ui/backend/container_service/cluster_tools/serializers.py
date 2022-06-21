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
from typing import Dict, List, Union

from rest_framework import serializers

from .models import InstalledTool, Tool


class ClusterToolSLZ(serializers.ModelSerializer):
    installed_info = serializers.SerializerMethodField()
    supported_actions = serializers.SerializerMethodField()

    class Meta:
        model = Tool
        exclude = ('extra_options', 'namespace')

    def get_installed_info(self, obj: Tool) -> Dict[str, str]:
        try:
            t = InstalledTool.objects.get(
                tool=obj, project_id=self.context['project_id'], cluster_id=self.context['cluster_id']
            )
            return {
                'cluster_id': t.cluster_id,
                'chart_version': t.chart_version,
                'values': t.values,
                'status': t.status,
                'message': t.message,
            }
        except InstalledTool.DoesNotExist:
            return {}

    def get_supported_actions(self, obj: Tool) -> List[str]:
        return [action.strip() for action in obj.supported_actions.split(',')]


class UpgradeToolSLZ(serializers.Serializer):
    chart_url = serializers.CharField()
    values = serializers.CharField(default="", allow_blank=True)


class InstalledToolSLZ(serializers.ModelSerializer):
    chart_version = serializers.ReadOnlyField()
    tool_info = serializers.SerializerMethodField()

    class Meta:
        model = InstalledTool
        exclude = ('id', 'tool', 'extra_options', 'deleted_time', 'is_deleted')

    def get_tool_info(self, obj: InstalledTool) -> Dict[str, Union[int, str]]:
        tool = obj.tool
        return {'id': tool.id, 'name': tool.name, 'description': tool.description}
