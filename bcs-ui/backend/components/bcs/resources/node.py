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

from kubernetes.client.rest import ApiException

from backend.resources.constants import K8sResourceKind
from backend.utils.errcodes import ErrorCode

from .api_response import response
from .resource import CoreAPIClassMixins, Resource

logger = logging.getLogger(__name__)


class Node(Resource, CoreAPIClassMixins):
    resource_kind = K8sResourceKind.Node.value

    def get_nodes_hostname(self):
        nodes_hostname = {}
        try:
            resp = self.api_instance.list_node()
            # format: [{'address': '', 'type': 'InternalIP'}, {'address': '', 'type': 'Hostname'}]
            for node in resp.items:
                host_address = node.status.addresses
                nodes_hostname[host_address[0].address] = host_address[1].address
        except ApiException as err:
            logger.exception('get node error, %s', err)

        return nodes_hostname

    @response()
    def operate_node(self, ip, data):
        nodes_hostname = self.get_nodes_hostname()
        if not (nodes_hostname and nodes_hostname.get(ip)):
            return {'code': ErrorCode.UnknownError, 'message': 'not found node info in cluster'}
        return self.api_instance.patch_node(nodes_hostname[ip], data)

    def disable_agent(self, ip):
        spec = {'spec': {'unschedulable': True}}
        return self.operate_node(ip, spec)

    def enable_agent(self, ip):
        spec = {'spec': {'unschedulable': False}}
        return self.operate_node(ip, spec)

    def create_node_labels(self, ip, labels):
        metadata = {'metadata': {'labels': labels}}
        return self.operate_node(ip, metadata)

    @response()
    def get_node_detail(self, ip):
        nodes_hostname = self.get_nodes_hostname()
        if not (nodes_hostname and nodes_hostname.get(ip)):
            return {'code': ErrorCode.UnknownError, 'message': 'not found node info in cluster'}
        return self.api_instance.read_node(nodes_hostname[ip])

    @response()
    def list_node(self, label_selector=None):
        return self.api_instance.list_node(label_selector=label_selector)
