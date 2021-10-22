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
import logging
from urllib.parse import urlencode

import aiohttp
from kubernetes import watch
from kubernetes.client import CoreV1Api

from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.utils.kube_client import get_dynamic_client, wrap_kube_client_exc

from . import constants

logger = logging.getLogger(__name__)


class LogClient:
    """Pod 子资源日志查询"""

    def __init__(self, ctx_cluster: CtxCluster, namespace: str, pod_name: str):
        self.ctx_cluster = ctx_cluster
        self.dynamic_client = get_dynamic_client(
            ctx_cluster.context.auth.access_token, ctx_cluster.project_id, ctx_cluster.id, use_cache=False
        )

        pod_resource = self.dynamic_client.get_preferred_resource("Pod")
        self.resource = pod_resource.subresources['log']
        self.namespace = namespace
        self.pod_name = pod_name

    def fetch_log(self, filter: constants.LogFilter):
        """获取日志"""
        params = {
            'timestamps': constants.LOG_SHOW_TIMESTAMPS,
            'limitBytes': constants.LOG_MAX_LIMIT_BYTES,
            'container': filter.container_name,
            'previous': filter.previous,
        }

        if filter.since_time:
            params['sinceTime'] = filter.since_time
        else:
            params['tailLines'] = filter.tail_lines

        with wrap_kube_client_exc():
            result = self.dynamic_client.get(self.resource, self.pod_name, self.namespace, query_params=params)

        return result

    def watch(self, filter: constants.LogFilter):
        """获取实时日志"""
        core_v1 = CoreV1Api(self.dynamic_client.client)

        w = watch.Watch()

        s = w.stream(
            core_v1.read_namespaced_pod_log,
            self.pod_name,
            self.namespace,
            tail_lines=filter.tail_lines,
            container=filter.container_name,
            timestamps=constants.LOG_SHOW_TIMESTAMPS,
            follow=True,
        )
        return s

    async def stream(self, filter: constants.LogFilter):
        """异步获取实时日志"""
        host = self.dynamic_client.client.configuration.host
        path = self.resource.path(self.pod_name, self.namespace)

        query_params = {
            'container': filter.container_name,
            'tailLines': filter.tail_lines,
            'sinceTime': filter.since_time,
            'follow': True,
            'timestamps': constants.LOG_SHOW_TIMESTAMPS,
        }

        url = f'{host}{path}?{urlencode(query_params)}'

        headers = {'Accept': 'application/json, */*'}
        headers.update(self.dynamic_client.client.configuration.api_key)

        async with aiohttp.ClientSession() as session:
            async with session.get(url, headers=headers, timeout=constants.STREAM_TIMEOUT, ssl=False) as response:
                response.raise_for_status()

                async for line in response.content:
                    yield line.decode('utf8')
