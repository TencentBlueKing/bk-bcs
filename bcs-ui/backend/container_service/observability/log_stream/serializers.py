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

from .constants import DEFAULT_TAIL_LINES


class FetchLogsSLZ(serializers.Serializer):
    """拉取日志"""

    container_name = serializers.CharField()
    tail_lines = serializers.IntegerField(default=DEFAULT_TAIL_LINES)
    started_at = serializers.CharField(default="")
    finished_at = serializers.CharField(default="")
    previous = serializers.BooleanField(default=False)


class GetLogSessionSLZ(serializers.Serializer):
    """获取日志会话"""

    container_name = serializers.CharField()
    tail_lines = serializers.IntegerField(default=0)
    since_time = serializers.CharField(default="")


class DownloadLogsSLZ(serializers.Serializer):
    """下载日志"""

    container_name = serializers.CharField()
    previous = serializers.BooleanField(default=False)
