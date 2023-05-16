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
from dataclasses import dataclass

from django.conf import settings

from backend.templatesets.legacy_apps.instance import constants as instance_constants
from backend.utils.basic import get_bcs_component_version

try:
    from backend.container_service.observability.datalog.utils import get_data_id_by_project_id
except ImportError:
    from backend.container_service.observability.datalog_ce.utils import get_data_id_by_project_id


def get_kubectl_version(cluster_version, kubectl_version_info, default_kubectl_version):
    # CE 版本目前仅提供默认版本的 kubectld
    if settings.EDITION == settings.COMMUNITY_EDITION:
        return default_kubectl_version

    return get_bcs_component_version(cluster_version, kubectl_version_info, default_kubectl_version)


@dataclass
class BCSInjectData:
    source_type: str
    creator: str
    updator: str
    version: str
    project_id: str
    app_id: str
    cluster_id: str
    namespace: str
    stdlog_data_id: str


def get_stdlog_data_id(project_id):
    data_info = get_data_id_by_project_id(project_id)
    return str(data_info.get('standard_data_id'))
