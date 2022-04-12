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

Module for generating kubeconfig file
"""
import contextlib
import logging
import tempfile
from dataclasses import dataclass
from typing import List

import yaml

logger = logging.getLogger(__name__)


@dataclass
class Cluster:
    """Cluster Info

    :param cert: path of certification
    :param cert_data: certification in base64 format, will be ignored if cert is provided
    """

    name: str
    server: str
    cert: str = ""
    cert_data: str = ""
    api_version: str = "v1"


@dataclass
class User:
    """Kubernetes user info"""

    name: str
    token: str


@dataclass
class Context:
    """kube config context"""

    name: str
    user: User
    cluster: Cluster


class KubeConfig:
    """The kubeconfig class"""

    def __init__(self, contexts: List[Context]):
        assert contexts, "Must provide at least one context"
        self.contexts = contexts

    @staticmethod
    def format_cluster(cluster):
        """Format a cluster as kubeconfig format"""
        cert_info = {"certificate-authority-data": cluster.cert_data}
        if cluster.cert:
            cert_info = {"certificate-authority": cluster.cert}

        cert_info = {"insecure-skip-tls-verify": True}

        return {
            "name": cluster.name,
            "cluster": {"server": cluster.server, "api-version": cluster.api_version, **cert_info},
        }

    @staticmethod
    def format_user(user):
        """Format an user as kubeconfig format"""
        return {"name": user.name, "user": {"token": user.token}}

    @staticmethod
    def format_context(context):
        """Format a context as kubeconfig format"""
        return {
            "name": context.name,
            "context": {
                "user": context.user.name,
                "cluster": context.cluster.name,
            },
        }

    def dumps(self, current_context: str = ""):
        """Represent current config as a kubeconfig file

        :returns: kubeconfig file content
        """
        clusters = {}
        users = {}
        for context in self.contexts:
            clusters[context.cluster.name] = self.format_cluster(context.cluster)
            users[context.user.name] = self.format_user(context.user)

        current_context = current_context or self.contexts[0].name
        payload = {
            "apiVersion": "v1",
            "kind": "Config",
            "clusters": list(clusters.values()),
            "users": list(users.values()),
            "contexts": [self.format_context(context) for context in self.contexts],
            "current-context": current_context,
        }
        return yaml.dump(payload, default_flow_style=False)

    @contextlib.contextmanager
    def as_tempfile(self):
        """A context manager which dump current config to a temp kubeconfig file"""

        with tempfile.NamedTemporaryFile() as fp:
            fp.write(self.dumps().encode())
            fp.flush()
            yield fp.name
