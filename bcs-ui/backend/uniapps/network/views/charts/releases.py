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
import logging

from django.db.models import Q

from backend.helm.app.models import App
from backend.helm.helm import bcs_variable

logger = logging.getLogger(__name__)


class HelmReleaseMixin:
    def get_helm_release(self, cluster_id, name, namespace_id=None, namespace=None):
        releases = App.objects.filter(
            Q(namespace_id=namespace_id) | Q(namespace=namespace), cluster_id=cluster_id, name=name
        )
        release = releases.first()
        if not release:
            logger.error(f"没有查询到集群:{cluster_id}, 名称:{name}对应的release信息")
        return release

    def collect_system_variable(self, access_token, project_id, namespace_id):
        return bcs_variable.collect_system_variable(
            access_token=access_token, project_id=project_id, namespace_id=namespace_id
        )
