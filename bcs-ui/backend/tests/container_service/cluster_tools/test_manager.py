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
from unittest.mock import patch

from backend.container_service.cluster_tools.constants import OpType, ToolStatus
from backend.container_service.cluster_tools.manager import HelmCmd, ToolManager, result_handler
from backend.helm.toolkit.deployer import make_valuesfile_flag


@patch.object(HelmCmd, '_create_namespace')
def test_tool_manager(mock_create_namespace, tool, request_user, project_id, cluster_id):
    assert tool.name == 'GameStatefulSet'

    with patch('backend.helm.toolkit.deployer.helm_install.apply_async') as mock_helm_install, patch(
        'backend.helm.toolkit.deployer.helm_upgrade.apply_async'
    ) as mock_helm_upgrade, patch('backend.helm.toolkit.deployer.helm_uninstall.apply_async') as mock_helm_uninstall:
        # 测试组件 install
        manager = ToolManager(project_id, cluster_id, tool.id)
        itool = manager.install(request_user)

        assert itool.status == ToolStatus.PENDING
        assert itool.chart_version == '0.6.0-beta3'
        mock_helm_install.assert_called_with(
            (
                request_user.token.access_token,
                {
                    'project_id': project_id,
                    'cluster_id': cluster_id,
                    'name': itool.release_name,
                    'namespace': itool.namespace,
                    'chart_url': itool.chart_url,
                    'operator': request_user.username,
                    'options': [],
                },
            ),
            link=result_handler.s(itool.id, OpType.INSTALL.value),
        )
        itool.success()

        # 测试组件 upgrade
        values = "replicaCount: 2"
        itool = manager.upgrade(request_user, chart_url=itool.chart_url, values=values)
        assert itool.status == ToolStatus.PENDING
        mock_helm_upgrade.assert_called_with(
            (
                request_user.token.access_token,
                {
                    'project_id': project_id,
                    'cluster_id': cluster_id,
                    'name': itool.release_name,
                    'namespace': itool.namespace,
                    'chart_url': itool.chart_url,
                    'operator': request_user.username,
                    'options': ['--install', make_valuesfile_flag(values)],
                },
            ),
            link=result_handler.s(itool.id, OpType.UPGRADE.value),
        )
        itool.success()

        # 测试卸载 uninstall
        manager.uninstall(request_user)
        mock_helm_uninstall.assert_called_with(
            (
                request_user.token.access_token,
                {
                    'project_id': project_id,
                    'cluster_id': cluster_id,
                    'name': itool.release_name,
                    'namespace': itool.namespace,
                    'chart_url': itool.chart_url,
                    'operator': request_user.username,
                    'options': [],
                },
            ),
            link=result_handler.s(itool.id, OpType.UNINSTALL.value),
        )
