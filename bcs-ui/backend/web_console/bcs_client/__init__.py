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
from backend.web_console.bcs_client import k8s


class BCSClientFactory:
    def __init__(self):
        self._bcs_clients = {}

    def register(self, bcs_client_cls):
        mode = bcs_client_cls.MODE
        self._bcs_clients[mode] = bcs_client_cls

    def create(self, mode: str, msg_handler, context, rows, cols):
        bcs_client_cls = self._bcs_clients.get(mode)
        if not bcs_client_cls:
            raise ValueError(f'{mode} not in {self._bcs_clients}')
        return bcs_client_cls.create_client(msg_handler, context, rows, cols)


factory = BCSClientFactory()
factory.register(k8s.ContainerDirectClient)
factory.register(k8s.KubectlInternalClient)
factory.register(k8s.KubectlExternalClient)
