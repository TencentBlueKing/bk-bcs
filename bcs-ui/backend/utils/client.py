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
import logging
import tempfile

from django.conf import settings
from kubernetes.client.rest import ApiException
from rest_framework.exceptions import APIException

from backend.components import bcs
from backend.container_service.misc.bke_client import BCSClusterClient
from backend.kube_core.toolkit.dashboard_cli import DashboardClient
from backend.kube_core.toolkit.kubectl import KubectlClusterClient
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


def get_kubectl_config_context(access_token=None, project_id=None, cluster_id=None):
    with make_kubectl_client(access_token=access_token, project_id=project_id, cluster_id=cluster_id) as (
        kubectl,
        error,
    ):
        if error is not None:
            logger.error("get_kubectl_config_context failed, %s", error)
            raise APIException("get_kubectl_config_context failed, %s" % error)

        with open(kubectl.kubeconfig, "r") as f:
            return f.read()


def get_bcs_host(access_token, project_id, cluster_id):
    if not (access_token and project_id and cluster_id):
        return None
    bcs_client = bcs.BCSClientBase(access_token, project_id, cluster_id, None)
    return bcs_client._bcs_https_server_host


@contextlib.contextmanager
def make_kubectl_client(access_token=None, project_id=None, cluster_id=None):
    """make a kubectl client for connection to k8s, it return a tuple of kubectl client and exception"""
    options = dict()
    host = get_bcs_host(access_token, project_id, cluster_id)
    if host:
        bcs_client = get_bcs_client(project_id=project_id, cluster_id=cluster_id, access_token=access_token)
        try:
            with bcs_client.make_kubectl_client() as kubectl_client:
                yield kubectl_client, None
        except Exception as e:
            logger.exception("make kubectl client failed, %s", e)
            yield None, e
    else:
        # default
        kubectl_client = KubectlClusterClient(kubectl_bin=settings.KUBECTL_BIN, kubeconfig=settings.KUBECFG, **options)
        yield kubectl_client, None


def get_bcs_client(project_id, cluster_id, access_token):
    host = get_bcs_host(access_token, project_id, cluster_id)
    if not host:
        raise ValueError(host)

    bcs_client = BCSClusterClient(
        host=host,
        access_token=access_token,
        project_id=project_id,
        cluster_id=cluster_id,
    )
    return bcs_client


@contextlib.contextmanager
def make_kubectl_client_from_kubeconfig(kubeconfig_content, **options):
    with tempfile.NamedTemporaryFile() as fp:
        fp.write(kubeconfig_content.encode())
        fp.flush()
        kubectl_client = KubectlClusterClient(kubectl_bin=settings.KUBECTL_BIN, kubeconfig=fp.name, **options)
        yield kubectl_client


def make_dashboard_ctl_client(kubeconfig, bin_path=settings.DASHBOARD_CTL_BIN):
    return DashboardClient(dashboard_ctl_bin=bin_path, kubeconfig=kubeconfig)


class KubectlClient:
    def __init__(self, access_token, project_id, cluster_id):
        self.access_token = access_token
        self.project_id = project_id
        self.cluster_id = cluster_id

    def _run_with_kubectl(self, operation, namespace, manifests):
        err_msg = ""
        with make_kubectl_client(
            project_id=self.project_id, cluster_id=self.cluster_id, access_token=self.access_token
        ) as (client, err):
            if err is not None:
                if isinstance(err, ApiException):
                    err = f"Code: {err.status}, Reason: {err.reason}"
                err_msg = f"make client failed: {err}"

            if not err_msg:
                try:
                    if operation == "apply":
                        client.ensure_namespace(namespace)
                        client.apply(manifests, namespace)
                    elif operation == "delete":
                        client.ensure_namespace(namespace)
                        client.delete(manifests, namespace)
                except Exception as e:
                    err_msg = f"client {operation} failed: {e}"

        if err_msg:
            raise error_codes.ComponentError(err_msg)

    def apply(self, namespace, manifests):
        self._run_with_kubectl("apply", namespace, manifests)

    def delete(self, namespace, manifests):
        self._run_with_kubectl("delete", namespace, manifests)


@contextlib.contextmanager
def make_helm_client(access_token=None, project_id=None, cluster_id=None):
    """创建连接k8s集群的client"""
    host = get_bcs_host(access_token, project_id, cluster_id)
    if host:
        bcs_client = get_bcs_client(project_id=project_id, cluster_id=cluster_id, access_token=access_token)
        try:
            with bcs_client.make_helm_client() as helm_client:
                yield helm_client, None
        except Exception as e:
            logger.exception("make helm client failed, %s", e)
            yield None, e
    else:
        yield None, APIException("bcs client host not found")
