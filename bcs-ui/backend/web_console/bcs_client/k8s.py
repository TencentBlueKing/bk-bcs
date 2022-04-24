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
import json
import logging
from urllib.parse import urlparse

from tornado.httpclient import HTTPRequest

from backend.web_console import constants
from backend.web_console.bcs_client.base import BCSClientBase

logger = logging.getLogger(__name__)


class K8SClientBase(BCSClientBase):
    def __init__(self, url, rows, cols, msg_handler, user_token):
        super().__init__(url, rows, cols, msg_handler)

        headers = {"Authorization": f"Bearer {user_token}"}
        self.url = HTTPRequest(url, headers=headers, validate_cert=False)
        self.connect()

    def write_message(self, message):
        """写入消息,需要写入std channel
        0 for stdin
        1 for stdout
        2 for stderr
        """
        message = chr(constants.STDIN_CHANNEL) + message
        self.ws.write_message(message)

    def set_pty_size(self, rows: int, cols: int):
        """自动宽度适应"""
        if not self.ws or self.ws.stream.closed():
            logger.info("session %s, close, set_pty_size just ignore", self)
            return

        # bash长度需要比页面长度，否则换行有问题
        message = json.dumps({"Width": cols, "Height": rows})
        message = chr(constants.RESIZE_CHANNEL) + message
        self.ws.write_message(message)

    def handle_message(self, message):
        """消息处理"""
        # 登入业务容器会cat或者curl下二进制的文件，utf-8解码失败后，使用latin1解码
        channel = message[0]
        if channel not in [constants.STDOUT_CHANNEL, constants.STDERR_CHANNEL]:
            return

        raw_message = message[1:]
        return raw_message


class KubectlInternalClient(K8SClientBase):
    """kubectl容器启动在用户自己集群"""

    MODE = "k8s_kubectl_internal"

    def post_connected(self):
        """k8s client添加心跳"""
        self.msg_handler.heartbeat()

        super().post_connected()

    @classmethod
    def create_client(cls, msg_handler, context, rows, cols):
        host = urlparse(context["admin_server_address"])
        if host.scheme == "https":
            scheme = "wss"
        else:
            scheme = "ws"

        bcs_address = host._replace(scheme=scheme).geturl()

        ws_url = f"{bcs_address}/api/v1/namespaces/{context['namespace']}/pods/{context['user_pod_name']}/exec"
        ws_url += "?command=/bin/bash&stderr=true&stdout=true&stdin=true&tty=true"
        client = cls(ws_url, rows, cols, msg_handler, context["admin_user_token"])
        return client


class KubectlExternalClient(KubectlInternalClient):
    """kubectl容器启动在共享集群"""

    MODE = "k8s_kubectl_external"


class ContainerDirectClient(K8SClientBase):
    MODE = "k8s_container_direct"

    @classmethod
    def create_client(cls, msg_handler, context, rows, cols):
        """k8s单个容器web_socket"""
        host = urlparse(context["server_address"])
        if host.scheme == "https":
            scheme = "wss"
        else:
            scheme = "ws"

        # 使用自己的集群ID
        bcs_address = host._replace(scheme=scheme).geturl()
        command = cls.get_command_params(context)

        ws_url = f"{bcs_address}/api/v1/namespaces/{context['namespace']}/pods/{context['user_pod_name']}/exec"
        ws_url += f"?{command}&container={context['container_name']}&stderr=true&stdout=true&stdin=true&tty=true"
        client = cls(ws_url, rows, cols, msg_handler, context["user_token"])
        return client
