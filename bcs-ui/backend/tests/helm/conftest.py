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

import pytest

from backend.helm.toolkit.deployer import HelmDeployer, ReleaseArgs
from backend.helm.toolkit.kubehelm.helm import KubeHelmClient


@pytest.fixture(scope="module", autouse=True)
def mock_run_command_with_retry():
    with patch.object(KubeHelmClient, "_run_command_with_retry", return_value=(None, None)) as mock_method:
        yield mock_method


@pytest.fixture(scope="module", autouse=True)
def mock_inject_auth_options():
    with patch.object(HelmDeployer, "_inject_auth_options"):
        yield


@pytest.fixture(autouse=True)
def use_bin_helm3(settings):
    settings.HELM3_BIN = "/bin/helm3"


@pytest.fixture
def release_args(project_id, cluster_id):
    return ReleaseArgs(
        project_id=project_id,
        cluster_id=cluster_id,
        name="gamestatefulset",
        namespace="bcs-system",
        operator="admin",
        chart_url="http://repo.example.com/charts/bcs-gamestatefulset-operator-0.6.0.tgz",
    )
