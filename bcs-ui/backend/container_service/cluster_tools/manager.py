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
from .models import Tool


class HelmCmd:
    def __init__(self, project_id: str, cluster_id: str):
        self.project_id = project_id
        self.cluster_id = cluster_id

    def install(self):
        """"""

    def upgrade(self):
        """"""

    def uninstall(self):
        """"""


class ToolManager:
    """组件管理器: 管理组件的安装, 更新和卸载"""

    def __init__(self, project_id: str, cluster_id: str, tool_id: int):
        self.project_id = project_id
        self.cluster_id = cluster_id

        self.tool = Tool.objects.get(id=tool_id)
        self.cmd = HelmCmd(project_id=project_id, cluster_id=cluster_id)

    def install(self):
        """"""
        self.cmd.install()

    def upgrade(self):
        """"""
        self.cmd.upgrade()

    def uninstall(self):
        """"""
        self.cmd.uninstall()
