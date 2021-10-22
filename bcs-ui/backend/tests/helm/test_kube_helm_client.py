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
import pytest

from backend.helm.toolkit.kubehelm.helm import KubeHelmClient


def test_do_install(settings, mock_run_command_with_retry):
    release_name = "test-example"
    namespace = "test"
    chart_url = "http://repo.example.com/charts/example-0.1.0.tgz"
    options = [
        {"--values": "bcs.yaml"},
        {"--values": "bcs-saas.yaml"},
        {"--username": "admin"},
        {"--password": "admin"},
        "--atomic",
    ]

    client = KubeHelmClient()
    client.do_install_or_upgrade("install", release_name, namespace, chart_url, options=options)
    mock_run_command_with_retry.assert_called_with(
        max_retries=0,
        cmd_args=[
            settings.HELM3_BIN,
            "install",
            release_name,
            "--namespace",
            namespace,
            chart_url,
            "--values",
            "bcs.yaml",
            "--values",
            "bcs-saas.yaml",
            "--username",
            "admin",
            "--password",
            "admin",
            "--atomic",
        ],
    )
