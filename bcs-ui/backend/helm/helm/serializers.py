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
from urllib.parse import urlparse

from django.conf import settings
from rest_framework import serializers
from rest_framework.exceptions import ValidationError

from .models.chart import Chart, ChartRelease, ChartVersion, ChartVersionSnapshot
from .models.repo import Repository, RepositoryAuth


class MinimalRepoSLZ(serializers.ModelSerializer):
    class Meta:
        model = Repository
        fields = ('name', 'url')


class ChartVersionSLZ(serializers.ModelSerializer):
    data = serializers.SerializerMethodField()

    class Meta:
        model = ChartVersion
        fields = ('id', 'name', 'version', 'created', 'data')

    def get_data(self, obj):
        return obj.to_json()


class ChartVersionTinySLZ(serializers.ModelSerializer):
    class Meta:
        model = ChartVersion
        fields = ('id', 'name', 'version', 'created')


class ChartVersionListSLZ(serializers.ModelSerializer):
    versions = ChartVersionTinySLZ(many=True)

    class Meta:
        model = Chart
        fields = ('id', 'name', 'icon', 'versions')


class ChartDetailSLZ(serializers.ModelSerializer):
    data = serializers.SerializerMethodField()

    class Meta:
        model = Chart
        fields = ('id', 'name', 'repository', 'data', 'defaultChartVersion')

    def get_data(self, obj):
        return obj.to_json()


class ChartSLZ(serializers.ModelSerializer):
    updated_at = serializers.DateTimeField()
    changed_at = serializers.DateTimeField()
    created_at = serializers.DateTimeField()
    deleted_at = serializers.DateTimeField()
    annotations = serializers.JSONField()

    class Meta:
        model = Chart
        fields = "__all__"


class ChartWithVersionRepoSLZ(ChartSLZ):
    defaultChartVersion = ChartVersionTinySLZ(read_only=True)
    repository = MinimalRepoSLZ(read_only=True)


class CreateRepoSLZ(serializers.Serializer):
    # url will be auto generated for ChartMuseumProvider
    url = serializers.URLField(required=True)
    name = serializers.CharField(required=True)
    # project_id = serializers.CharField(required=True)
    provider = serializers.CharField(required=True)

    def validate_url(self, value):
        if not urlparse(value).scheme:
            raise ValidationError("url needs scheme")
        return value


class RepositoryAuthSLZ(serializers.ModelSerializer):
    credentials_decoded = serializers.DictField(read_only=True)

    class Meta:
        model = RepositoryAuth
        fields = ('type', 'credentials', 'repo', 'role', 'credentials_decoded')


class RepoSLZ(serializers.ModelSerializer):
    auths = RepositoryAuthSLZ(many=True, read_only=True)
    project_id = serializers.CharField(read_only=True)

    class Meta:
        model = Repository
        fields = (
            'id',
            'name',
            'url',
            'description',
            'project_id',
            'provider',
            'is_provisioned',
            'auths',
            'branch',
            'refreshed_at',
            'commit',
        )
        read_only_fields = ("project_id", "refreshed_at", "commit")


class ChartVersionSnapshotSLZ(serializers.ModelSerializer):
    questions = serializers.JSONField()
    files = serializers.JSONField()
    maintainers = serializers.JSONField()
    sources = serializers.JSONField()
    urls = serializers.JSONField()

    class Meta:
        model = ChartVersionSnapshot
        fields = "__all__"


class ChartReleaseSLZ(serializers.ModelSerializer):
    chartVersionSnapshot = ChartVersionSnapshotSLZ(read_only=True)
    answers = serializers.JSONField()
    customs = serializers.JSONField()
    valuefile = serializers.CharField(
        initial=[],
        label="Values File",
        help_text="Yaml format data",
        style={"base_template": "textarea.html", "rows": 10},
    )

    class Meta:
        model = ChartRelease
        fields = "__all__"


class RepositorySyncSLZ(serializers.Serializer):
    pass


class ChartVersionParamsSLZ(serializers.Serializer):
    version_list = serializers.ListField(child=serializers.CharField(), default=[])


class ChartParamsSLZ(serializers.Serializer):
    is_public_repo = serializers.BooleanField(default=False)


class ChartVersionDetailSLZ(serializers.ModelSerializer):
    data = serializers.SerializerMethodField()

    class Meta:
        model = ChartVersion
        fields = ("id", "name", "version", "data")

    def get_data(self, obj):
        return obj.to_json()
