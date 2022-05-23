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

import attr
from celery import shared_task
from rest_framework.exceptions import ValidationError

from backend.container_service.clusters.base.models import CtxCluster
from backend.helm.toolkit.deployer import ReleaseArgs, helm_install, helm_uninstall, helm_upgrade, make_valuesfile_flag
from backend.helm.toolkit.kubehelm.options import RawFlag
from backend.resources.namespace import Namespace
from backend.resources.namespace.constants import K8S_SYS_NAMESPACE

from .models import InstalledTool, Tool


class ToolArgs(ReleaseArgs):
    @classmethod
    def from_tool(cls, itool: InstalledTool, options: List[RawFlag]):
        return cls(
            project_id=itool.project_id,
            cluster_id=itool.cluster_id,
            name=itool.release_name,
            namespace=itool.namespace,
            operator=itool.updator,
            chart_url=itool.chart_url,
            options=options,
        )


class HelmCmd:
    def __init__(self, project_id: str, cluster_id: str):
        self.project_id = project_id
        self.cluster_id = cluster_id

    def install(self, request_user, itool: InstalledTool) -> InstalledTool:
        # 创建命名空间
        if itool.namespace not in K8S_SYS_NAMESPACE:
            try:
                self._create_namespace(request_user, itool.namespace)
            except Exception as e:
                result_handler(f'create namespace {itool.namespace} in {self.cluster_id} error: {e}', itool.id)

        # 设置额外的 install 参数
        if itool.extra_options:
            options = [RawFlag(itool.extra_options)]
        else:
            options = []

        if itool.values:
            options.append(make_valuesfile_flag(itool.values))

        helm_install.apply_async(
            (request_user.token.access_token, attr.asdict(ToolArgs.from_tool(itool, options))),
            link=result_handler.s(itool.id),
        )
        return itool

    def upgrade(self, request_user, itool: InstalledTool) -> InstalledTool:
        # 更新组件时通过 helm upgrade --install 操作，解决组件 install 时可能未成功生成 release 的问题
        options = [RawFlag('--install')]

        # 设置额外的 install 参数
        if itool.extra_options:
            options.append(RawFlag(itool.extra_options))

        if itool.values:
            options.append(make_valuesfile_flag(itool.values))

        helm_upgrade.apply_async(
            (request_user.token.access_token, attr.asdict(ToolArgs.from_tool(itool, options))),
            link=result_handler.s(itool.id),
        )

        return itool

    def uninstall(self, request_user, itool: InstalledTool):
        helm_uninstall.apply_async(
            (request_user.token.access_token, attr.asdict(ToolArgs.from_tool(itool, options=[]))),
            link=result_handler.s(itool.id, True),
        )

    def _create_namespace(self, request_user, namespace: str):
        ctx_cluster = CtxCluster.create(
            token=request_user.token.access_token, id=self.cluster_id, project_id=self.project_id
        )
        return Namespace(ctx_cluster).get_or_create_cc_namespace(namespace, request_user.username)


class ToolManager:
    """组件管理器: 管理组件的安装, 更新和卸载"""

    def __init__(self, project_id: str, cluster_id: str, tool_id: int):
        """初始化

        :param project_id 项目 ID
        :param cluster_id 集群 ID
        :param tool_id 组件库中的组件 ID
        """
        self.project_id = project_id
        self.cluster_id = cluster_id

        try:
            self.tool = Tool.objects.get(id=tool_id)
        except Tool.DoesNotExist:
            raise ValidationError(f'invalid tool_id({tool_id})')

        self.cmd = HelmCmd(project_id=project_id, cluster_id=cluster_id)

    def install(self, request_user, values: Optional[str] = None) -> InstalledTool:
        """安装组件

        :param request_user: 操作者信息(request.user)
        :param values: 安装组件时的初始配置
        """
        itool = InstalledTool.create_by_tool(
            request_user.username, self.tool, self.project_id, self.cluster_id, values
        )

        # chart_url 为空字符串表示该组件未使用 Helm 部署，设置成部署状态后返回
        if not itool.chart_url:
            itool.success()
            return itool

        return self.cmd.install(request_user, itool)

    def upgrade(self, request_user, chart_url: str, values: Optional[str] = None) -> InstalledTool:
        """更新组件

        :param request_user: 操作者信息(request.user)
        :param chart_url: 更新版本的 chart url
        :param values: 更新组件时的新配置
        """
        itool = InstalledTool.objects.get(tool=self.tool, project_id=self.project_id, cluster_id=self.cluster_id)
        itool.on_upgrade(request_user.username, chart_url, values)
        return self.cmd.upgrade(request_user, itool)

    def uninstall(self, request_user):
        """卸载组件

        :param request_user: 操作者信息(request.user)

        TODO 增加删除审计
        """
        itool = InstalledTool.objects.get(tool=self.tool, project_id=self.project_id, cluster_id=self.cluster_id)
        itool.on_delete(request_user.username)
        self.cmd.uninstall(request_user, itool)


@shared_task
def result_handler(err_msg: Optional[str], itool_id: int, is_uninstall: bool = False):
    """异步流转安装任务的状态"""
    itool = InstalledTool.objects.get(id=itool_id)

    if err_msg:
        itool.fail(err_msg)
        return

    # 如果是卸载组件, 成功后清除安装记录
    if is_uninstall:
        itool.delete()
        return

    # 其他操作记录成功
    itool.success()
