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
from .resource import CoreAPIClassMixins, Resource

logger = logging.getLogger(__name__)


class Service(Resource, CoreAPIClassMixins):
    resource_kind = K8sResourceKind.Service.value

    def get_service_by_namespace(self, params):
        resp = self.api_instance.list_namespaced_service(params['namespace'], _preload_content=False)
        data = json.loads(resp.data)
        service_list = []
        for info in data.get('items') or []:
            resource_name = getitems(info, ['metadata', 'name'], '')
            resource_namespace = getitems(info, ['metadata', 'namespace'], '')
            item = self.render_resource('Service', info, resource_name, resource_namespace)
            params_name = params.get('name')
            if params_name:
                if resource_name == params_name:
                    service_list.append(item)
                    break
            else:
                service_list.append(item)
        return service_list

    def get_all_service(self):
        service_list = []
        resp = self.api_instance.list_service_for_all_namespaces(_preload_content=False)
        data = json.loads(resp.data)
        for info in data.get('items') or []:
            resource_name = getitems(info, ['metadata', 'name'], '')
            resource_namespace = getitems(info, ['metadata', 'namespace'], '')
            item = self.render_resource(self.resource_kind, info, resource_name, resource_namespace)
            service_list.append(item)
        return service_list

    @response(format_data=False)
    def get_service(self, params):
        if params.get('namespace'):
            return self.get_service_by_namespace(params)
        return self.get_all_service()

    @response()
    def create_service(self, namespace, data):
        return self.api_instance.create_namespaced_service(namespace, data)

    @response()
    def update_service(self, namespace, name, data):
        return self.api_instance.replace_namespaced_service(name, namespace, data)

    @response()
    def delete_serivce(self, namespace, name):
        body = client.V1DeleteOptions()
        return self.api_instance.delete_namespaced_service(name, namespace, body=body)
