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

from kubernetes import client

from backend.resources.constants import K8sResourceKind

from .api_response import response
from .resource import CoreAPIClassMixins, Resource

logger = logging.getLogger(__name__)


class Secret(Resource, CoreAPIClassMixins):
    resource_kind = K8sResourceKind.Secret.value

    def get_secret_by_namespace(self, params):
        resp = self.api_instance.list_namespaced_secret(params['namespace'])
        secret_list = []
        for info in resp.items:
            item = self.render_resource_for_preload_content(
                self.resource_kind, info, info.metadata.name, info.metadata.namespace
            )
            params_name = params.get('name')
            if params_name:
                if info.metadata.name == params_name:
                    secret_list.append(item)
                    break
            else:
                secret_list.append(item)
        return secret_list

    def get_all_secret(self):
        secret_list = []
        resp = self.api_instance.list_secret_for_all_namespaces()
        for info in resp.items:
            item = self.render_resource_for_preload_content(
                'Secret', info, info.metadata.name, info.metadata.namespace
            )
            secret_list.append(item)
        return secret_list

    @response(format_data=False)
    def get_secret(self, params):
        if params.get('namespace'):
            return self.get_secret_by_namespace(params)
        return self.get_all_secret()

    @response()
    def create_secret(self, namespace, data):
        data.pop('apiVersion', None)
        return self.api_instance.create_namespaced_secret(namespace, data)

    @response()
    def update_secret(self, namespace, name, data):
        data.pop('apiVersion', None)
        return self.api_instance.replace_namespaced_secret(name, namespace, data)

    @response()
    def delete_secret(self, namespace, name):
        body = client.V1DeleteOptions()
        return self.api_instance.delete_namespaced_secret(name, namespace, body=body)
