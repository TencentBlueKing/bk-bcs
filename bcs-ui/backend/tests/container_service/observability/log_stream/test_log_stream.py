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
from channels.testing import WebsocketCommunicator

from backend.accounts.middlewares import BCSChannelAuthMiddlewareStack
from backend.container_service.observability.log_stream.views import LogStreamHandler


@pytest.fixture
def session_id(api_client, project_id, cluster_id, namespace, pod_name, container_name):
    response = api_client.post(
        f'/api/logs/projects/{project_id}/clusters/{cluster_id}/namespaces/{namespace}/pods/{pod_name}/stdlogs/sessions/',  # noqa
        {"container_name": container_name},
    )

    result = response.json()
    return result['data']['session_id']


@pytest.mark.skip(reason='暂时跳过标准日志部分单元测试')
@pytest.mark.django_db
@pytest.mark.asyncio
async def test_log_stream(project_id, cluster_id, namespace, pod_name, session_id):

    app = BCSChannelAuthMiddlewareStack(LogStreamHandler.as_asgi())

    # Test a normal connection
    communicator = WebsocketCommunicator(
        app,
        f'/ws/logs/projects/{project_id}/clusters/{cluster_id}/namespaces/{namespace}/pods/{pod_name}/stdlogs/stream/?session_id={session_id}',  # noqa
    )

    communicator.scope['url_route'] = {
        'kwargs': {
            'project_id': project_id,
            'cluster_id': cluster_id,
            'namespace': namespace,
            'pod': pod_name,
        }
    }

    connected, _ = await communicator.connect()

    assert connected
    # Test sending text
    await communicator.send_to(text_data="hello")

    # Close out
    await communicator.disconnect()
