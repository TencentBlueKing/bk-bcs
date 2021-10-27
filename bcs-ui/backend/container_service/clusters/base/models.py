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

用于在系统内部使用的 Cluster 集群建模
"""
from backend.container_service.core.ctx_models import BaseContextedModel


class CtxCluster(BaseContextedModel):
    """集群对象

    :param id: 集群 ID
    :param project_id: 集群所属项目 ID
    """

    def __init__(self, id: str, project_id: str):
        self.id = id
        self.project_id = project_id

    def __str__(self):
        return f'<Cluster: {self.project_id}-{self.id}>'
