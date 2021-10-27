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
import os

import django

django.setup()  # noqa

from channels.routing import ProtocolTypeRouter, URLRouter
from django.core.asgi import get_asgi_application

from backend.accounts.middlewares import BCSChannelAuthMiddlewareStack
from backend.container_service.observability.log_stream import routing as log_routing

os.environ.setdefault("DJANGO_SETTINGS_MODULE", "backend.settings.ce.saas_prod")


application = ProtocolTypeRouter(
    {
        "http": get_asgi_application(),
        "websocket": BCSChannelAuthMiddlewareStack(URLRouter(log_routing.websocket_urlpatterns)),
    }
)
