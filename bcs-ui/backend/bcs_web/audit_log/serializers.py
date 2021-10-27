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
import arrow
from django.utils.translation import ugettext_lazy as _
from rest_framework import serializers

from .models import UserActivityLog


class ActivityLogSLZ(serializers.ModelSerializer):
    description = serializers.SerializerMethodField()

    def get_description(self, obj):
        description = obj.description
        if description:
            return _(obj.description)
        return description

    class Meta:
        model = UserActivityLog
        fields = (
            "id",
            "project_id",
            "activity_time",  # 时间
            "activity_type",  # 操作类型
            "resource",  # 操作对象名称
            "resource_type",  # 操作对象类型
            "resource_id",  # 操作对象id
            "activity_status",  # 状态
            "user",  # 发起者
            "description",  # 简要描述
        )


class ActivityLogGetSLZ(serializers.Serializer):
    activity_type = serializers.CharField(required=False)
    activity_status = serializers.CharField(required=False)
    resource_type = serializers.CharField(required=False)
    begin_time = serializers.DateTimeField(required=False)
    end_time = serializers.DateTimeField(required=False)


class EventSLZ(serializers.Serializer):
    offset = serializers.IntegerField(min_value=0)
    limit = serializers.IntegerField(min_value=1)
    cluster_id = serializers.CharField(required=False)
    kind = serializers.CharField(required=False)
    level = serializers.CharField(required=False)
    component = serializers.CharField(required=False)
    begin_time = serializers.DateTimeField(required=False)
    end_time = serializers.DateTimeField(required=False)

    def validate_begin_time(self, begin_time):
        """转换为时间戳"""
        return arrow.get(begin_time).timestamp

    def validate_end_time(self, end_time):
        """转换为时间戳"""
        return arrow.get(end_time).timestamp
