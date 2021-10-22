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
from typing import Tuple

from django.db import models
from django.utils import timezone
from jsonfield import JSONField

from backend.utils.models import BaseTSModel

from .managers import RepositoryAuthManager, RepositoryManager


class Repository(BaseTSModel):
    url = models.URLField('URL')
    name = models.CharField('Name', max_length=32)
    description = models.CharField(max_length=512)
    project_id = models.CharField('ProjectID', max_length=32)
    provider = models.CharField('Provider', max_length=32)
    is_provisioned = models.BooleanField('Provisioned?', default=False)
    refreshed_at = models.DateTimeField(null=True)

    # git repo fields
    commit = models.CharField(max_length=64, null=True)
    # TODO: check, if git, use which branch?
    branch = models.CharField(max_length=30, null=True)

    # chart museum fields
    storage_info = JSONField(default={})

    objects = RepositoryManager()

    class Meta:
        unique_together = ("project_id", "name")
        db_table = 'helm_repository'

    def __str__(self):
        return "[{id}]{project}/{name}".format(id=self.id, project=self.project_id, name=self.name)

    def refreshed(self, commit):
        self.refreshed_at = timezone.now()
        self.commit = commit
        self.save()

    @property
    def plain_auths(self):
        auths = list(self.auths.values("credentials", "type", "role"))
        return [
            {
                "type": auth["type"],
                "role": auth["role"],
                "credentials": auth["credentials"],
            }
            for auth in auths
        ]

    @property
    def username_password(self) -> Tuple[str, str]:
        try:
            credentials = list(self.auths.values("credentials"))
            credential = credentials[0]["credentials"]
            return (credential["username"], credential["password"])
        except Exception:
            return ("", "")


class RepositoryAuth(models.Model):
    AUTH_CHOICE = (("BASIC", "BasicAuth"),)

    type = models.CharField('Type', choices=AUTH_CHOICE, max_length=16)
    # ex: {"password":"EJWmMqqGeA5E6JNb","username":"admin-T49e"}
    credentials = JSONField('Credentials', default={})
    repo = models.ForeignKey(Repository, on_delete=models.CASCADE, related_name='auths')
    # TODO: use rbac module instead
    role = models.CharField('Role', max_length=16)

    objects = RepositoryAuthManager()

    @property
    def credentials_decoded(self):
        return self.credentials

    class Meta:
        db_table = 'helm_repo_auth'
