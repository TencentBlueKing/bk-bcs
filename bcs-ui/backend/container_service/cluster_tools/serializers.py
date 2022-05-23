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
from typing import Dict

from rest_framework import serializers

from .models import InstalledTool, Tool


class ClusterToolSZL(serializers.ModelSerializer):
    installed_info = serializers.SerializerMethodField()

    class Meta:
        models = Tool
        fields = '__all__'

    def get_installed_info(self, obj: Tool) -> Dict[str, str]:
        try:
            t = InstalledTool.objects.get(
                tool=obj, project_id=self.context['project_id'], cluster_id=self.context['cluster_id']
            )
            return {'cluster_id': t.cluster_id, 'values': t.values, 'status': t.status, 'message': t.message}
        except InstalledTool.DoesNotExist:
            return {}


class UpgradeToolSLZ(serializers.Serializer):
    chart_url = serializers.CharField()
    values = serializers.CharField(default="", allow_blank=True)


class InstalledToolSLZ(serializers.ModelSerializer):
    chart_version = serializers.ReadOnlyField()

    class Meta:
        model = InstalledTool
        fields = '__all__'
