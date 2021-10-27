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
from dataclasses import dataclass, field
from typing import Dict, List

from backend.templatesets.legacy_apps.configuration.models import ShowVersion, Template


@dataclass
class ResContext:
    access_token: str
    username: str
    project_id: str
    namespace: str
    cluster_id: str
    template: Template
    show_version: ShowVersion
    instance_entity: Dict[str, List[int]]  # like {"Deployment": [1, 2], "Job": [11, 12]}, 其中整数为db记录项的id
    is_preview: bool
    template_variables: Dict[str, str] = field(default_factory=dict)
    namespace_id: int = 0  # TODO namespace_id待废弃
