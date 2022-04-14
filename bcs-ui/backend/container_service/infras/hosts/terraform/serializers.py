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
from rest_framework import serializers

from .models import HostApplyTaskLog


class ApplyHostDataSLZ(serializers.Serializer):
    region = serializers.CharField()
    vpc_name = serializers.CharField()
    cvm_type = serializers.CharField()
    disk_size = serializers.IntegerField()
    replicas = serializers.IntegerField(min_value=1)
    zone_id = serializers.CharField()
    disk_type = serializers.CharField()


class TaskLogSLZ(serializers.ModelSerializer):
    logs = serializers.JSONField()

    class Meta:
        model = HostApplyTaskLog
        fields = ("created", "task_url", "operator", "status", "is_finished", "logs")
