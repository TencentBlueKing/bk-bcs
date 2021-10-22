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
from backend.components.bcs import BCSClientBase
from backend.components.utils import http_delete, http_get, http_post

CLUSTERKEEP_ENDPOINT = "{host_prefix}/v4/clusterkeeper"
STORAGE_PREFIX = "{host_prefix}/v4/storage"
# 针对特定接口的超时时间
DEFAULT_TIMEOUT = 20
DEFAULT_K8S_VERSION = "1.8.3"


class BCSClient(BCSClientBase):
    """Mesos和K8S共有的API"""

    @property
    def storage_host(self):
        return STORAGE_PREFIX.format(host_prefix=self.api_host)

    @property
    def cluster_keeper_host(self):
        return CLUSTERKEEP_ENDPOINT.format(host_prefix=self.api_host)

    def create_cluster(self, cluster_type, operator, ip_list, data=()):
        url = '{host}/clusters/{cluster_id}'.format(
            host=self.cluster_keeper_host,
            cluster_id=self.cluster_id,
        )
        data = dict(data)
        data.update(
            {
                'project_id': self.project_id,
                'access_token': self.access_token,
                'operator': operator,
                'clusterType': cluster_type,
                'ipList': ip_list,
                'version': DEFAULT_K8S_VERSION,
            }
        )
        return http_post(url, json=data, headers=self.headers, timeout=DEFAULT_TIMEOUT)

    def add_cluster_node(self, cluster_type, operator, ip_list, cc_app_id, need_nat=True, data=()):
        url = '{host}/clusters/{cluster_id}/nodes'.format(
            host=self.cluster_keeper_host,
            cluster_id=self.cluster_id,
        )
        data = dict(data)
        data.update(
            {
                'project_id': self.project_id,
                'access_token': self.access_token,
                'operator': operator,
                'clusterType': cluster_type,
                'ipList': ip_list,
                'ccAppID': str(cc_app_id),
                'needNat': need_nat,
            }
        )
        return http_post(url, json=data, headers=self.headers, timeout=DEFAULT_TIMEOUT)

    def delete_cluster_node(self, cluster_type, operator, ip_list, data=()):
        """删除集群节点"""
        url = '{host}/clusters/{cluster_id}/nodes'.format(
            host=self.cluster_keeper_host,
            cluster_id=self.cluster_id,
        )
        data = dict(data)
        data.update(
            {
                'project_id': self.project_id,
                'access_token': self.access_token,
                'operator': operator,
                'clusterType': cluster_type,
                'ipList': ip_list,
            }
        )
        return http_delete(url, json=data, headers=self.headers, timeout=DEFAULT_TIMEOUT)

    def get_task_result(self, task_id):
        url = '{host}/tasks/{task_id}'.format(
            host=self.cluster_keeper_host,
            task_id=task_id,
        )
        data = {
            'project_id': self.project_id,
        }
        return http_get(url, params=data, headers=self.headers)

    def get_events(self, params):
        """获取事件
        注意需要针对不同的环境进行查询
        """
        url = "{host}/events".format(host=self.storage_host)
        resp = http_get(url, params=params, headers=self.headers)
        return resp
