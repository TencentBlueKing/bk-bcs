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

from backend.components.bcs import k8s as bcs_k8s

from . import k8s


class BCSDriver:

    # 1：k8s 2: mesos
    KIND_BCS_CLIENT_AND_DRIVER = {
        1: {'bcs_client': bcs_k8s.K8SClient, 'driver': k8s.Driver},
    }

    def __init__(self, request, project_id, cluster_id):
        self.request = request
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.k8s_kind_info = self.KIND_BCS_CLIENT_AND_DRIVER[request.project.kind]
        self.bcs_client = self.k8s_kind_info['bcs_client'](
            request.user.token.access_token, project_id, cluster_id, None
        )
        self.driver = self.k8s_kind_info['driver']

    def get_unit_info_by_name(self, params=None):
        """获取pod或taskgroup信息"""
        return self.driver.get_unit_info_by_name(
            self.bcs_client,
            ns_name=params.get('namespace'),
            pod_name=params.get('unit_name'),
            field=params.get('field'),
        )

    def get_events(self, params):
        return self.driver.get_events(self.bcs_client, params)
