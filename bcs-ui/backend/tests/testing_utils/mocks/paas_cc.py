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
import uuid
from typing import Dict, List

from .utils import mockable_function


class StubPaaSCCClient:
    """使用假数据的 PaaSCCClient 对象"""

    def __init__(self, *args, **kwargs):
        pass

    @mockable_function
    def get_cluster(self, project_id: str, cluster_id: str) -> Dict:
        return self.wrap_resp(self.make_cluster_data(project_id, cluster_id))

    @mockable_function
    def get_cluster_by_id(self, cluster_id: str) -> Dict:
        return self.make_cluster_data_by_id(cluster_id)

    @mockable_function
    def list_clusters(self, cluster_ids: List[str]) -> List:
        return [self.make_cluster_data(uuid.uuid4().hex, cluster_id) for cluster_id in cluster_ids]

    @mockable_function
    def get_project(self, project_id: str) -> Dict:
        return self.make_project_data(project_id)

    @mockable_function
    def get_cluster_namespace_list(self, project_id: str, cluster_id: str) -> Dict:
        return self.make_cluster_namespace_data(project_id, cluster_id)

    @mockable_function
    def get_project_with_code(self, project_id: str) -> Dict:
        return self.wrap_resp(self.make_project_data(project_id))

    @staticmethod
    def wrap_resp(data):
        return {
            'code': 0,
            'data': data,
            'message': '',
            'request_id': uuid.uuid4().hex,
            'result': True,
        }

    @staticmethod
    def make_cluster_data(project_id: str, cluster_id: str):
        _stub_time = '2021-01-01T00:00:00+08:00'
        return {
            'area_id': 1,
            'artifactory': '',
            'capacity_updated_at': _stub_time,
            'cluster_id': cluster_id,
            'cluster_num': 1,
            'config_svr_count': 0,
            'created_at': _stub_time,
            'creator': 'unknown',
            'description': 'cluster description',
            'disabled': False,
            'environment': 'stag',
            'extra_cluster_id': '',
            'ip_resource_total': 0,
            'ip_resource_used': 0,
            'master_count': 0,
            'name': 'test-cluster',
            'need_nat': True,
            'node_count': 1,
            'project_id': project_id,
            'remain_cpu': 10,
            'remain_disk': 0,
            'remain_mem': 10,
            'status': 'normal',
            'total_cpu': 12,
            'total_disk': 0,
            'total_mem': 64,
            'type': 'k8s',
            'updated_at': _stub_time,
        }

    @staticmethod
    def make_cluster_data_by_id(cluster_id: str):
        _stub_time = '2021-01-01T00:00:00+08:00'
        return {
            'area_id': 1,
            'artifactory': '',
            'capacity_updated_at': _stub_time,
            'cluster_id': cluster_id,
            'cluster_num': 1,
            'config_svr_count': 0,
            'created_at': _stub_time,
            'creator': 'unknown',
            'description': 'cluster description',
            'disabled': False,
            'environment': 'stag',
            'extra_cluster_id': '',
            'ip_resource_total': 0,
            'ip_resource_used': 0,
            'master_count': 0,
            'name': 'test-cluster',
            'need_nat': True,
            'node_count': 1,
            'project_id': uuid.uuid4().hex,
            'remain_cpu': 10,
            'remain_disk': 0,
            'remain_mem': 10,
            'status': 'normal',
            'total_cpu': 12,
            'total_disk': 0,
            'total_mem': 64,
            'type': 'k8s',
            'updated_at': _stub_time,
        }

    @staticmethod
    def make_project_data(project_id: str):
        _stub_time = '2021-01-01T00:00:00+08:00'
        return {
            "approval_status": 2,
            "approval_time": "2020-01-01T00:00:00+08:00",
            "approver": "",
            "bg_id": -1,
            "bg_name": "",
            "cc_app_id": 100,
            "center_id": 100,
            "center_name": "",
            "created_at": "2020-01-01 00:00:00",
            "creator": "unknown",
            "data_id": 0,
            "deploy_type": "null",
            "dept_id": -1,
            "dept_name": "",
            "description": "",
            "english_name": "unittest-proj",
            "extra": {},
            "is_offlined": False,
            "is_secrecy": False,
            "kind": 1,
            "logo_addr": "",
            "project_id": project_id,
            "project_name": "unittest-proj",
            "project_type": 1,
            "remark": "",
            "updated_at": "2020-01-01 00:00:00",
            "use_bk": False,
            "cc_app_name": "demo-app",
            "can_edit": False,
            "project_code": "unittest-proj",
        }

    @staticmethod
    def make_cluster_namespace_data(project_id: str, cluster_id: str) -> Dict:
        _stub_time = '2021-06-30T11:13:00+08:00'
        return {
            "count": 1,
            "results": [
                {
                    "cluster_id": cluster_id,
                    "created_at": _stub_time,
                    "creator": "admin",
                    "description": "",
                    "env_type": "dev",
                    "has_image_secret": True,
                    "id": 1,
                    "name": "default",
                    "project_id": project_id,
                    "status": "",
                    "updated_at": _stub_time,
                }
            ],
        }
