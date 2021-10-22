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
import base64
import json
import logging

from backend.resources.constants import K8sResourceKind
from backend.utils.basic import getitems

from .api_response import response
from .resource import CoreAPIClassMixins, Resource

logger = logging.getLogger(__name__)


class Pod(Resource, CoreAPIClassMixins):
    resource_kind = K8sResourceKind.Pod.value

    def get_reference_name(self, extra):
        extra_info = json.loads(base64.b64decode(extra))
        return extra_info['data.metadata.ownerReferences.name']

    def filter_pods_by_reference_name(self, pod_name, pod_item, filter_reference_name, pods_map, render_pod_info):
        reference_name = self.get_reference_name(filter_reference_name)
        reference_name_list = [item['name'] for item in getitems(pod_item, ['metadata', 'ownerReferences'], [])]
        if (reference_name in reference_name_list) or (set(reference_name) & set(reference_name_list)):
            pods_map[pod_name] = render_pod_info

        return pods_map

    def filter_pods_by_ips(self, pod_name, pod_item, host_ips, pods_map, render_pod_info):
        pod_host_ip = getitems(pod_item, ['status', 'hostIP'], '')
        if pod_host_ip in host_ips:
            pods_map[pod_name] = render_pod_info

        return pods_map

    def filter_pods(self, namespace, filter_reference_name, filter_pod_name, host_ips):
        resp = self.api_instance.list_namespaced_pod(namespace, _preload_content=False)
        data = json.loads(resp.data)
        pods_map = {}
        for info in data.get('items') or []:
            pod_name = getitems(info, ['metadata', 'name'], '')
            pod_namespace = getitems(info, ['metadata', 'namespace'], '')
            item = self.render_resource(self.resource_kind, info, pod_name, pod_namespace)
            # 如果过滤参数都不存在时，返回所有
            if not (filter_reference_name or filter_pod_name or host_ips):
                pods_map[pod_name] = item

            # TODO: 是否有必要把下面几个拆分到方法中
            # 过滤reference
            if filter_reference_name:
                pods_map = self.filter_pods_by_reference_name(pod_name, info, filter_reference_name, pods_map, item)
            # 过滤pod name
            if filter_pod_name and pod_name == filter_pod_name:
                pods_map[pod_name] = item
            # 过滤host ip
            if host_ips:
                pods_map = self.filter_pods_by_ips(pod_name, info, host_ips, pods_map, item)
        # 转换为list，防止view层出现错误`TypeError: 'dict_values' object does not support indexing`
        return list(pods_map.values())

    def get_all_pod(self, host_ips):
        resp = self.api_instance.list_pod_for_all_namespaces(_preload_content=False)
        data = json.loads(resp.data)
        pod_list = []
        for info in data.get('items') or []:
            pod_name = getitems(info, ['metadata', 'name'], '')
            pod_namespace = getitems(info, ['metadata', 'namespace'], '')
            item = self.render_resource(self.resource_kind, info, pod_name, pod_namespace)
            if host_ips:
                pod_host_ip = getitems(info, ['status', 'hostIP'], '')
                if pod_host_ip in host_ips:
                    pod_list.append(item)
            else:
                pod_list.append(item)

        return pod_list

    @response(format_data=False)
    def get_pod(self, host_ips=None, field=None, extra=None, params=None):
        if params and params.get('namespace'):
            return self.filter_pods(params['namespace'], extra, params.get('name'), host_ips)
        return self.get_all_pod(host_ips)

    @response()
    def delete_pod(self, namespace, name):
        return self.api_instance.delete_namespaced_pod(name, namespace)
