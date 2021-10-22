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
from typing import List, Optional

from django.db import models


class RepositoryManager(models.Manager):
    def get_repos(self, repo_ids: List[int]) -> models.query.QuerySet:
        """查询repo信息"""
        return self.filter(id__in=repo_ids).values("id", "name", "url")


class ChartManager(models.Manager):
    def get_queryset(self):
        return super().get_queryset().filter(deleted=False)

    def get_charts(self, project_id: str, repo_id: Optional[int] = None) -> models.query.QuerySet:
        """获取chart列表
        返回chart所属的repo信息及chart的默认(最新)版本
        """
        qs = self.filter(repository__project_id=project_id).order_by("-changed_at")
        if repo_id is not None:
            qs = qs.filter(repository__id=repo_id)

        return qs


class ChartVersionManager(models.Manager):
    def get_versions(self, version_ids: int) -> models.query.QuerySet:
        """通过版本ID获取版本信息"""
        return self.filter(id__in=version_ids).values("id", "name", "version")


class RepositoryAuthManager(models.Manager):
    pass


class ChartVersionSnapshotManager(models.Manager):
    def make_snapshot(self, chart_version):
        snapshot, created = self.get_or_create(
            digest=chart_version.digest,
            defaults={
                "name": chart_version.name,
                "home": chart_version,
                "description": chart_version.description,
                "engine": chart_version.engine,
                "maintainers": chart_version.maintainers,
                "sources": chart_version.sources,
                "urls": chart_version.urls,
                "files": chart_version.files,
                "questions": chart_version.questions,
                "version": chart_version.version,
                "digest": chart_version.digest,
                "created": chart_version.created,
                "version_id": chart_version.id,
            },
        )
        return snapshot
