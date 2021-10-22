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
from kubernetes.client.rest import ApiException

from backend.utils.basic import getitems

from .api_response import response
from .resource import FilterResourceData, Resource, StorageAPIClassMixins

logger = logging.getLogger(__name__)


class StorageClass(Resource, FilterResourceData, StorageAPIClassMixins):
    @response(format_data=False)
    def list_storage_class(self):
        resp = self.api_instance.list_storage_class(_preload_content=False)
        data = json.loads(resp.data)

        return data
