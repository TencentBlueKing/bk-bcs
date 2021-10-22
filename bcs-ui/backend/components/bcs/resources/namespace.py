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


class Namespace(Resource, CoreAPIClassMixins):
    resource_kind = K8sResourceKind.Namespace.value

    def get_body(self, metadata=None, spec=None):
        """render request body"""
        # include api_version, kind, metadata, spec
        # body = client.V1Namespace(api_version=, kind=, metadata=, spec=)
        pass

    @response()
    def create_namespace(self, body):
        return self.api_instance.create_namespace(body)

    @response()
    def delete_namespace(self, name):
        body = client.V1DeleteOptions()
        return self.api_instance.delete_namespace(name, body=body)

    @response(format_data=False)
    def get_namespace(self, params):
        resp = self.api_instance.list_namespace()
        namespace_list = []
        for info in resp.items:
            item = self.render_resource_for_preload_content(
                self.resource_kind, info, info.metadata.name, info.metadata.namespace
            )
            # 适配先前逻辑
            item['clusterId'] = params['cluster_id']
            namespace_list.append(item)
        return namespace_list
