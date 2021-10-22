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

from .api_response import response
from .resource import BaseMixins, FilterResourceData, Resource

logger = logging.getLogger(__name__)

preferred_version = "apps/v1beta1"
updated_preferred_version = "apps/v1beta2"


class DaemonSet(Resource, FilterResourceData, BaseMixins):
    resource_kind = K8sResourceKind.DaemonSet.value

    def get_api_class(self, api_client):
        resp = client.AppsApi(api_client).get_api_group()
        group_version = resp.preferred_version.group_version
        # NOTE: 针对1.8对应的preferred_version: apps/v1beta1调整为apps/v1beta2
        # 因为针对1.8集群，返回的preferred_version有问题，找不到对应的resource
        if group_version == preferred_version:
            group_version = updated_preferred_version
        return self.compose_api_class(group_version)

    @response()
    def create_daemonset(self, namespace, data):
        return self.api_instance.create_namespaced_daemon_set(namespace, data)

    @response()
    def delete_daemonset(self, namespace, name):
        body = client.V1DeleteOptions()
        return self.api_instance.delete_namespaced_daemon_set(name, namespace, body=body)

    @response(format_data=False)
    def get_daemonset(self, params):
        # 因为view层向下传递时是多个namespace+name, 需要增加过滤
        resp = self.api_instance.list_daemon_set_for_all_namespaces(_preload_content=False)
        data = json.loads(resp.data)
        return self.filter_data(self.resource_kind, data, params)

    @response()
    def update_daemonset(self, namespace, name, data):
        return self.api_instance.replace_namespaced_daemon_set(name, namespace, data)

    @response()
    def patch_daemonset(self, namespace, name, data):
        return self.api_instance.patch_namespaced_daemon_set(name, namespace, data)
