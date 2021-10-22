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


class ConfigMap(Resource, CoreAPIClassMixins):
    resource_kind = K8sResourceKind.ConfigMap.value

    @response()
    def create_configmap(self, namespace, data):
        return self.api_instance.create_namespaced_config_map(namespace, data)

    @response()
    def delete_configmap(self, namespace, name):
        body = client.V1DeleteOptions()
        return self.api_instance.delete_namespaced_config_map(name, namespace, body=body)

    @response()
    def update_configmap(self, namespace, name, data):
        return self.api_instance.replace_namespaced_config_map(name, namespace, data)

    def get_configmap_by_namespace(self, params):
        resp = self.api_instance.list_namespaced_config_map(params['namespace'])
        configmap_list = []
        for info in resp.items:
            item = self.render_resource_for_preload_content(
                self.resource_kind, info, info.metadata.name, info.metadata.namespace
            )
            params_name = params.get('name')
            if params_name:
                if info.metadata.name == params_name:
                    configmap_list.append(item)
                    break
            else:
                configmap_list.append(item)
        return configmap_list

    def get_all_configmap(self):
        configmap_list = []
        resp = self.api_instance.list_config_map_for_all_namespaces()
        for info in resp.items:
            item = self.render_resource_for_preload_content(
                self.resource_kind, info, info.metadata.name, info.metadata.namespace
            )
            configmap_list.append(item)
        return configmap_list

    @response(format_data=False)
    def get_configmap(self, params):
        if isinstance(params, dict) and params.get('namespace'):
            return self.get_configmap_by_namespace(params)
        return self.get_all_configmap()
