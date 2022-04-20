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


class ImageQuerySLZ(serializers.Serializer):
    limit = serializers.IntegerField(required=True)
    offset = serializers.IntegerField(required=True)
    filters = serializers.CharField(required=False)
    search = serializers.CharField(required=False)


class SingleImageCollectSLZ(serializers.Serializer):
    image_repo = serializers.CharField(required=True)
    has_collected = serializers.BooleanField(required=True)
    image_project = serializers.CharField(required=False)


class SingleImageUpdateSLZ(serializers.Serializer):
    image_repo = serializers.CharField(required=True)
    desc = serializers.CharField(required=True)


class AvailableTagSLZ(serializers.Serializer):
    repo = serializers.CharField(required=True)
    is_pub = serializers.BooleanField(required=True)


class ImageDetailSLZ(serializers.Serializer):
    limit = serializers.IntegerField(default=10)
    offset = serializers.IntegerField(default=10)
    image_repo = serializers.CharField(required=True)
