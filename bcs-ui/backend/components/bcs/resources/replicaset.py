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
import re

from django.utils.translation import ugettext_lazy as _
from kubernetes import client
from kubernetes.client.rest import ApiException

from backend.utils.basic import getitems
from backend.utils.error_codes import error_codes

from .api_response import response
from .resource import Resource

logger = logging.getLogger(__name__)


class ReplicaSet(Resource):
    @property
    def api_versions(self):
        resp = client.AppsApi(self.api_client).get_api_group()
        return [f"{''.join([i.capitalize() for i in ver.group_version.split('/')])}Api" for ver in resp.versions]

    def _parse_owner_ref_name(self, params):
        owner_ref_name = None
        extra = params.get('extra')
        if extra:
            extra_info = json.loads(base64.b64decode(extra))
            owner_ref_name = extra_info['data.metadata.ownerReferences.name']
        # NOTE: type of owner_ref_name maybe list, str
        if isinstance(owner_ref_name, str):
            owner_ref_name = re.split(r',|;', owner_ref_name)
        return owner_ref_name

    def _request_api(self, func_name, *args, **kwargs):
        for api_version in self.api_versions:
            try:
                api_instance = getattr(client, api_version)(self.api_client)
                return getattr(api_instance, func_name)(*args, **kwargs)
            except Exception as e:
                logger.error("call %s error, apiversion: %s, error: %s", func_name, api_version, e)
        raise error_codes.APIError(_("查询RS失败：K8S SDK中未查询到合适的版本"))

    @response(format_data=False)
    def get_replicaset(self, params):
        owner_ref_name = self._parse_owner_ref_name(params)

        if params.get('namespace'):
            resp = self._request_api("list_namespaced_replica_set", params["namespace"])
        else:
            resp = self._request_api("list_replica_set_for_all_namespaces")

        if not owner_ref_name:
            return [
                {
                    "resourceName": info.metadata.name,
                    "data": {"status": {"replicas": getattr(info.status, "replicas", 0)}},
                }
                for info in resp.items
            ]

        # filter replica set by owner reference name
        replicaset_list = []
        for info in resp.items:
            # 过滤掉owner references为空的情况
            if not getattr(info.metadata, "owner_references", None):
                continue
            owner_ref_name_list = [ref.name for ref in info.metadata.owner_references]
            if set(owner_ref_name) & set(owner_ref_name_list):
                replicaset_list.append(
                    {
                        "resourceName": info.metadata.name,
                        "data": {"status": {"replicas": getattr(info.status, "replicas", 0)}},
                    }
                )
        return replicaset_list
