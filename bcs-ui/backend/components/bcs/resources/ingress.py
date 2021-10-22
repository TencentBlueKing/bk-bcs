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

from kubernetes import client

from backend.resources.constants import K8sResourceKind
from backend.utils.basic import getitems

from .api_response import response
from .resource import ExtensionsAPIClassMixins, Resource

logger = logging.getLogger(__name__)


class Ingress(Resource, ExtensionsAPIClassMixins):
    resource_kind = K8sResourceKind.Ingress.value

    @response()
    def create_ingress(self, namespace, data):
        return self.api_instance.create_namespaced_ingress(namespace, data)

    @response()
    def delete_ingress(self, namespace, name):
        body = client.V1DeleteOptions()
        return self.api_instance.delete_namespaced_ingress(name, namespace, body=body)

    def get_service_by_namespace(self, params):
        resp = self.api_instance.list_namespaced_ingress(params['namespace'], _preload_content=False)
        data = json.loads(resp.data)
        ingress_list = []
        for info in data.get('items') or []:
            resource_name = getitems(info, ['metadata', 'name'], '')
            item = self.render_resource('Ingress', info, resource_name, getitems(info, ['metadata', 'namespace'], ''))
            if params.get('name'):
                if resource_name == params['name']:
                    ingress_list.append(item)
                    break
            else:
                ingress_list.append(item)
        return ingress_list

    def get_all_ingress(self):
        ingress_list = []
        resp = self.api_instance.list_ingress_for_all_namespaces(_preload_content=False)
        data = json.loads(resp.data)
        for info in data.get('items') or []:
            item = self.render_resource(
                'Ingress',
                info,
                getitems(info, ['metadata', 'name'], ''),
                getitems(info, ['metadata', 'namespace'], ''),
            )
            ingress_list.append(item)
        return ingress_list

    @response(format_data=False)
    def get_ingress(self, params):
        if params.get('namespace'):
            return self.get_ingress_by_namespace(params)
        return self.get_all_ingress()

    @response()
    def update_ingress(self, namespace, name, data):
        return self.api_instance.replace_namespaced_ingress(name, namespace, data)
