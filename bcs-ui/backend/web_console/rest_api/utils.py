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
import copy
import hashlib
import string
from urllib.parse import urlparse

from django.conf import settings
from django.utils.encoding import smart_bytes

from backend.bcs_web.audit_log import client as activity_client
from backend.web_console.bcs_client import k8s
from backend.web_console.constants import WebConsoleMode

DNS_ALLOW_CHARS = string.ascii_lowercase + string.digits + "-"


def get_mesos_context(client, container_id: str) -> dict:
    """通过taskgroup和container_id过滤host_ip"""
    resp = client.get_mesos_app_taskgroup(
        field="data.containerStatuses.containerID,data.metadata.name,data.containerStatuses.name,data.hostIP",
    )
    taskgroups = resp.get("data") or []
    context = {}

    for taskgroup in taskgroups:
        for container in taskgroup["data"]["containerStatuses"]:
            if container_id == container["containerID"]:
                context = {
                    "taskgroup_name": taskgroup["data"]["metadata"]["name"],
                    "host_ip": taskgroup["data"]["hostIP"],
                    "container_name": container["name"],
                }
                return context
    return context


def get_k8s_context(client, container_id: str) -> dict:
    """通过containder_id获取pod, namespace信息"""
    resp = client.get_pod(field="resourceName,data.status.containerStatuses,namespace")
    pods = resp.get("data") or []
    context = {}
    for pod in pods:
        namespace = pod["namespace"]
        pod_name = pod["resourceName"]
        try:
            for container in pod["data"]["status"]["containerStatuses"]:
                # 必须是ready状态
                if not container["ready"]:
                    continue

                if container["containerID"] == f"docker://{container_id}":
                    context["namespace"] = namespace
                    context["pod_name"] = pod_name
                    context["container_name"] = container["name"]
                    return context
        except Exception:
            pass
    return context


def get_k8s_cluster_context(client, project_id, cluster_id):
    """获取集群信息"""
    context = copy.deepcopy(client.context)

    # 原始集群ID
    context["source_cluster_id"] = cluster_id

    # 内部版bcs返回server_address
    if "server_address" in context:
        server_address_path = urlparse(context["server_address"]).path
    else:
        # 社区版直接bcs_api返回server_address_path
        server_address_path = context["server_address_path"]

    # API调用地址, 可以为http/https
    server_address = f"{client._bcs_server_host}{server_address_path}"
    context["server_address"] = server_address.rstrip("/")

    # Kubectl Config地址, 必须为https
    https_server_address = f"{client._bcs_https_server_host}{server_address_path}"
    context["https_server_address"] = https_server_address.rstrip("/")

    return context


def get_k8s_admin_context(client, context, mode):
    if mode == WebConsoleMode.EXTERNAL.value:
        context["mode"] = k8s.KubectlExternalClient.MODE
        # 外部模式使用固定的admin_token和集群ID
        context["admin_user_token"] = settings.WEB_CONSOLE_EXTERNAL_CLUSTER["API_TOKEN"]
        context["admin_cluster_identifier"] = settings.WEB_CONSOLE_EXTERNAL_CLUSTER["ID"]
        context["admin_server_address"] = "{}/tunnels/clusters/{}".format(
            settings.WEB_CONSOLE_EXTERNAL_CLUSTER["API_HOST"], settings.WEB_CONSOLE_EXTERNAL_CLUSTER["ID"]
        )

    else:
        context["mode"] = k8s.KubectlInternalClient.MODE
        # 内部模式user_token, identifier为当前的集群id
        context["admin_user_token"] = context["user_token"]
        context["admin_cluster_identifier"] = context["identifier"]
        context["admin_server_address"] = context["server_address"]

    return context


def get_k8s_pod_spec(client):
    pod_spec = settings.WEB_CONSOLE_POD_SPEC.copy()
    if client._bcs_server_host_ip:
        pod_spec["hostAliases"] = [
            {"ip": client._bcs_server_host_ip, "hostnames": [urlparse(client._bcs_https_server_host).hostname]}
        ]

    return pod_spec


def activity_log(project_id, cluster_id, cluster_name, username, status, message=None):
    """操作记录"""
    with activity_client.ContextActivityLogClient(
        project_id=project_id,
        resource_id=cluster_id,
        user=username,
        resource_type="web_console",
        resource=cluster_name,
    ).log_start() as ual:

        if status is True:
            ual.update_log(activity_status="succeed", description="启动WebConsole成功")
        else:
            ual.update_log(activity_status="failed", description="启动WebConsole失败：%s" % message)


def get_username_slug(username: str) -> str:
    """k8s configmap名字限制，需要转换下
    https://tools.ietf.org/html/rfc1123#section-2
    """
    hash_id = hashlib.sha1(smart_bytes(username)).hexdigest()[:12]
    _username = "".join(char for char in username.lower() if char in DNS_ALLOW_CHARS)
    return f"{_username}-{hash_id}"
