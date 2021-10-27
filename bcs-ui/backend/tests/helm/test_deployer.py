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
import contextlib
from unittest.mock import MagicMock

from backend.helm.toolkit.deployer import VALUESFILE_KEY, HelmDeployer, ReleaseArgs
from backend.helm.toolkit.kubehelm.helm import KubeHelmClient
from backend.utils import client as bcs_client


@contextlib.contextmanager
def mock_make_helm_client(access_token=None, project_id=None, cluster_id=None):
    """创建连接k8s集群的client"""
    yield KubeHelmClient(), None


def _get_values_file_path(options):
    for flag in options:
        if isinstance(flag, dict) and flag.get("--values"):
            return flag.get("--values")


def test_helm_install(settings, release_args, mock_run_command_with_retry, request_user, project_id, cluster_id):
    bcs_client.make_helm_client = MagicMock(side_effect=mock_make_helm_client)
    helm_args = ReleaseArgs(
        project_id=project_id,
        cluster_id=cluster_id,
        name=release_args.name,
        namespace=release_args.namespace,
        chart_url=release_args.chart_url,
        operator="admin",
        options=[{VALUESFILE_KEY: {"file_name": "component_values.yaml", "file_content": "replicaCount: 2"}}],
    )
    deployer = HelmDeployer(request_user.token.access_token, helm_args)
    deployer.install()
    mock_run_command_with_retry.assert_called_with(
        max_retries=0,
        cmd_args=[
            settings.HELM3_BIN,
            "install",
            release_args.name,
            "--namespace",
            release_args.namespace,
            release_args.chart_url,
            "--values",
            _get_values_file_path(deployer.helm_args.options),
        ],
    )


def test_helm_upgrade(settings, release_args, mock_run_command_with_retry, request_user, project_id, cluster_id):
    bcs_client.make_helm_client = MagicMock(side_effect=mock_make_helm_client)
    helm_args = ReleaseArgs(
        project_id=project_id,
        cluster_id=cluster_id,
        name=release_args.name,
        namespace=release_args.namespace,
        chart_url=release_args.chart_url,
        operator="admin",
        options=[{VALUESFILE_KEY: {"file_name": "component_values.yaml", "file_content": "replicaCount: 2"}}],
    )
    deployer = HelmDeployer(request_user.token.access_token, helm_args)
    deployer.upgrade()
    mock_run_command_with_retry.assert_called_with(
        max_retries=0,
        cmd_args=[
            settings.HELM3_BIN,
            "upgrade",
            release_args.name,
            "--namespace",
            release_args.namespace,
            release_args.chart_url,
            "--values",
            _get_values_file_path(deployer.helm_args.options),
        ],
    )
