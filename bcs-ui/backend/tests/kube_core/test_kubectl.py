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
import textwrap
from pathlib import Path

import pytest

from backend.kube_core.toolkit.kubectl import Cluster, Context, KubeConfig, KubectlClusterClient, User
from backend.kube_core.toolkit.kubectl.exceptions import KubectlExecutionError
from backend.tests.conftest import TESTING_API_SERVER_URL

KUBECTL_BIN = "/usr/local/bin/kubectl"

if not Path(KUBECTL_BIN).exists():
    pytestmark = pytest.mark.skip('kubectl bin not found')


class TestKubeConfig:
    def test_dumps_normal(self):
        context = Context(
            name="local-test-default",
            user=User(name="user-default", token="invalid-token"),
            cluster=Cluster(
                name="local-test-server",
                server="http://127.0.0.1:8080/",
                cert_data="invalid-ca",
            ),
        )
        config = KubeConfig(contexts=[context])
        assert "clusters" in config.dumps()


class TestIntegratedClient:
    @pytest.fixture
    def context(self):
        return Context(
            name="local-test-server",
            user=User(name="user-someone", token="fake-token"),
            cluster=Cluster(
                name="local-test-server",
                server=TESTING_API_SERVER_URL,
                cert_data="not a valid cert",
            ),
        )

    def test_invalid_format(self, context):
        config = KubeConfig(contexts=[context])
        with config.as_tempfile() as fname:
            client = KubectlClusterClient(KUBECTL_BIN, kubeconfig=fname)
            with pytest.raises(KubectlExecutionError):
                client.apply("invalid_template")

    def test_normal(self, context):
        config = KubeConfig(contexts=[context])
        with config.as_tempfile() as fname:
            client = KubectlClusterClient(KUBECTL_BIN, kubeconfig=fname)
            tmpl = textwrap.dedent(
                '''
            apiVersion: v1
            data:
                hello: world
            kind: ConfigMap
            metadata:
                creationTimestamp: null
                name: bcs-unitest-c1
            '''
            )
            client.apply(tmpl)
            client.delete(tmpl)

    def test_valid_yaml_invalid_resource(self, context):
        config = KubeConfig(contexts=[context])
        with config.as_tempfile() as fname:
            client = KubectlClusterClient(KUBECTL_BIN, kubeconfig=fname)
            tmpl = textwrap.dedent(
                '''
            - invalid
            - valid
            - resource
            '''
            )
            with pytest.raises(KubectlExecutionError):
                client.apply(tmpl)
            with pytest.raises(KubectlExecutionError):
                client.delete(tmpl)
