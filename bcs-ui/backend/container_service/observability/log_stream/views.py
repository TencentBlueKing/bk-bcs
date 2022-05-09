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

from channels.generic.websocket import AsyncWebsocketConsumer
from django.conf import settings
from django.http import HttpResponse
from django.utils import timezone
from rest_framework.response import Response

from backend.bcs_web.viewsets import SystemViewSet
from backend.resources.workloads.pod.constants import LogFilter
from backend.resources.workloads.pod.log import LogClient
from backend.web_console.session import session_mgr

from . import constants, serializers, utils

logger = logging.getLogger(__name__)


class LogStreamViewSet(SystemViewSet):
    """k8s 原生日志流"""

    def fetch(self, request, project_id: str, cluster_id: str, namespace: str, pod: str):
        """获取日志"""

        data = self.params_validate(serializers.FetchLogsSLZ)

        filter = LogFilter(container_name=data["container_name"], previous=data["previous"])
        if data["started_at"] and data['finished_at']:
            filter.since_time = utils.calc_since_time(data["started_at"], data['finished_at'])
        else:
            filter.tail_lines = data["tail_lines"]

        client = LogClient(request.ctx_cluster, namespace, pod)
        content = client.fetch_log(filter)
        logs = utils.refine_k8s_logs(content, data['started_at'])

        url_path_prefix = (
            f"/api/logs/projects/{project_id}/clusters/{cluster_id}/namespaces/{namespace}/pods/{pod}/stdlogs/"  # noqa
        )
        previous = utils.calc_previous_page(logs, data, url_path_prefix)

        result = {"logs": logs, "previous": previous}
        return Response(result)

    def create_session(self, request, project_id: str, cluster_id: str, namespace: str, pod: str):
        """获取实时日志session"""
        data = self.params_validate(serializers.GetLogSessionSLZ)

        filter = LogFilter(
            container_name=data["container_name"],
            since_time=data["since_time"],
            tail_lines=data["tail_lines"],
        )

        ctx = {
            'username': request.user.username,
            'access_token': request.user.token.access_token,
            'filter': filter,
            'project_id': project_id,
            'cluster_id': cluster_id,
        }

        session = session_mgr.create(project_id, cluster_id)
        session_id = session.set(ctx)

        stream_url = (
            f'/ws/logs/projects/{project_id}/clusters/{cluster_id}/namespaces/{namespace}/pods/{pod}/stdlogs/stream/'
        )

        ws_url = utils.make_ws_url(stream_url, session_id)
        result = {"session_id": session_id, "ws_url": ws_url}
        return Response(result)

    def download(self, request, project_id: str, cluster_id: str, namespace: str, pod: str):
        """下载日志"""
        data = self.params_validate(serializers.DownloadLogsSLZ)

        filter = LogFilter(
            container_name=data["container_name"], previous=data["previous"], tail_lines=constants.MAX_TAIL_LINES
        )

        client = LogClient(request.ctx_cluster, namespace, pod)
        content = client.fetch_log(filter)

        ts = timezone.now().strftime("%Y%m%d%H%M%S")
        filename = f"{pod}-{data['container_name']}-{ts}.log"
        response = HttpResponse(content=content, content_type='application/octet-stream')
        response['Content-Disposition'] = f'attachment; filename="{filename}"'
        return response


class LogStreamHandler(AsyncWebsocketConsumer):
    """日志 Channel / WebSocket 处理"""

    async def connect(self):
        self.namespace = self.scope["url_route"]["kwargs"]["namespace"]
        self.pod = self.scope["url_route"]["kwargs"]["pod"]

        self.ctx_cluster = self.scope['ctx_cluster']
        self.filter = LogFilter(**self.scope['ctx_session']['filter'])

        logger.info("%s connect from client: %s", self.ctx_cluster, self.scope['ctx_session']['username'])
        self.closed = False

        await self.accept()

        await self.reader()

    async def disconnect(self, close_code):
        logger.info("%s disconnect from client: %s", self.ctx_cluster, self.scope['ctx_session']['username'])
        self.closed = True

    async def receive(self, text_data):
        """获取消息, 目前只有推送, 只打印日志"""
        logger.info("receive message: %s", text_data)

    async def reader(self):
        client = LogClient(self.ctx_cluster, self.namespace, self.pod)

        async for line in client.stream(self.filter):
            if self.closed is True:
                return

            # k8s返回使用空格分隔
            t, _, log = line.partition(' ')
            data = {"streams": [{'time': t, 'log': log}]}
            try:
                await self.send(text_data=json.dumps(data))
            except Exception as error:
                logger.error("reader error: %s", error)
                return
