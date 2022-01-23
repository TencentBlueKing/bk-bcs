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
from backend.components import paas_cc
from backend.helm.app.models import App
from backend.templatesets.legacy_apps.instance.models import InstanceConfig
from backend.templatesets.var_mgmt.models import NameSpaceVariable
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

from . import k8s


class Namespace:
    def __init__(self, access_token, project_id, project_kind):
        self.access_token = access_token
        self.project_id = project_id
        self.client = k8s

    def delete(self, namespace_id):
        # namespace exist
        ns_resp = paas_cc.get_namespace(self.access_token, self.project_id, namespace_id)
        ns_data = ns_resp['data']
        if ns_resp.get('code') != ErrorCode.NoError or not ns_data:
            raise error_codes.APIError(f'query namespace exist error, {ns_resp.get("message")}')
        # get namespace info
        cluster_id, ns_name = ns_data.get('cluster_id'), ns_data.get('name')

        self.client.delete(self.access_token, self.project_id, cluster_id, ns_name)
        # delete db resource record
        InstanceConfig.objects.filter(namespace=namespace_id).delete()
        NameSpaceVariable.objects.filter(ns_id=namespace_id).delete()
        App.objects.filter(namespace_id=namespace_id).delete()

        # delete bcs cc record
        ns_resp = paas_cc.delete_namespace(self.access_token, self.project_id, cluster_id, namespace_id)
        if ns_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(f'delete bcs cc namespace error, {ns_resp.get("message")}')

        return ns_resp

    def list(self, cluster_id):
        return self.client.get_namespace(self.access_token, self.project_id, cluster_id)
