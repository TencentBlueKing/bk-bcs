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

resp_get_project_ok = {
    "data": {
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
        "project_id": uuid.uuid4().hex,
        "project_name": "unittest-proj",
        "project_type": 1,
        "remark": "",
        "updated_at": "2020-01-01 00:00:00",
        "use_bk": False,
        "cc_app_name": "demo-app",
        "can_edit": False,
    },
    "code": 0,
    "message": "OK",
    "request_id": uuid.uuid4().hex,
}

resp_filter_projects_ok = {
    "data": {
        "count": 2,
        "results": [
            {
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
                "project_id": uuid.uuid4().hex,
                "project_name": "unittest-proj",
                "project_type": 1,
                "remark": "",
                "updated_at": "2020-01-01 00:00:00",
                "use_bk": False,
                "cc_app_name": "demo-app",
                "can_edit": False,
            },
            {
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
                "english_name": "unittest-proj-a",
                "extra": {},
                "is_offlined": False,
                "is_secrecy": False,
                "kind": 1,
                "logo_addr": "",
                "project_id": uuid.uuid4().hex,
                "project_name": "unittest-proj-a",
                "project_type": 1,
                "remark": "",
                "updated_at": "2020-01-01 00:00:00",
                "use_bk": False,
                "cc_app_name": "demo-app",
                "can_edit": False,
            },
        ],
    },
    "code": 0,
    "message": "OK",
    "request_id": uuid.uuid4().hex,
}

resp_get_clusters_ok = {
    "code": 0,
    "data": {
        "count": 1,
        "results": [
            {
                "area_id": -1,
                "artifactory": "",
                "capacity_updated_at": "2020-01-01T00:00:00+08:00",
                "cluster_id": "BCS-K8S-10000",
                "cluster_num": 10000,
                "config_svr_count": 0,
                "created_at": "2020-01-01T00:00:00+08:00",
                "creator": "unknown",
                "description": "Fake cluster info for unittests",
                "disabled": False,
                "environment": "stag",
                "extra_cluster_id": "",
                "ip_resource_total": 0,
                "ip_resource_used": 0,
                "master_count": 0,
                "name": "unitests-cluster",
                "need_nat": True,
                "node_count": 1,
                "project_id": None,
                "remain_cpu": 100,
                "remain_disk": 100,
                "remain_mem": 100,
                "status": "normal",
                "total_cpu": 14,
                "total_disk": 0,
                "total_mem": 100,
                "type": "k8s",
                "updated_at": "2020-01-01T00:00:00+08:00",
            }
        ],
    },
    "message": "获取集群成功",
    "result": True,
}

resp_get_namespaces_ok = {
    "code": 0,
    "data": {
        "count": 1,
        "results": [
            {
                "cluster_id": "BCS-K8S-10000",
                "created_at": "2020-01-01T00:00:00+08:00",
                "creator": "unknown",
                "description": "",
                "env_type": "dev",
                "has_image_secret": False,
                "id": 1,
                "name": "unittests-ns-1",
                "project_id": None,
                "status": "",
                "updated_at": "2020-01-01T00:00:00+08:00",
            }
        ],
    },
    "message": "获取Namespace成功",
    "result": True,
}

resp_get_namespace_ok = {
    "code": 0,
    "data": {
        "cluster_id": "BCS-K8S-10000",
        "created_at": "2020-01-01T00:00:00+08:00",
        "creator": "unknown",
        "description": "",
        "env_type": "dev",
        "has_image_secret": False,
        "id": 1,
        "name": "test",
        "project_id": None,
        "status": "",
        "updated_at": "2020-01-01T00:00:00+08:00",
    },
    "message": "获取Namespace成功",
    "result": True,
}
