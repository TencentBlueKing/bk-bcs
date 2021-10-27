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
fake_create_project_resp = {"code": 0, "data": {"name": "test"}, "message": "ok"}

fake_create_repo_resp = {"code": 0, "data": {"name": "test"}, "message": "ok"}

fake_set_auth_resp = True

fake_list_charts_resp = {
    "test": [
        {
            "apiVersion": "v1",
            "appVersion": "1.0",
            "created": "2020-06-04T15:14:39.866Z",
            "description": "A Helm chart for Kubernetes",
            "digest": "0c1f206f5f5c80ef8fa7574d68b22c91544c9ebe3f0ce12a3c67f26aa86b442e",
            "keywords": [],
            "maintainers": [],
            "name": "test",
            "sources": [],
            "urls": ["http://repo.example.com/charts/test-1.0.1.tgz"],
            "version": "1.0.1",
        }
    ],
    "test1": [
        {
            "apiVersion": "v1",
            "appVersion": "1.0",
            "created": "2020-06-04T15:12:39.866Z",
            "description": "A Helm chart for Kubernetes",
            "digest": "0c1f206f5f5c80ef8fa7574d68b22c91544c9ebe3f0ce12a3c67f26aa86b4421",
            "keywords": [],
            "maintainers": [],
            "name": "test1",
            "sources": [],
            "urls": ["http://repo.example.com/charts/test1-1.0.0.tgz"],
            "version": "1.0.0",
        },
    ],
}

fake_chart_versions_resp = [
    {
        "apiVersion": "v1",
        "appVersion": "1.0",
        "created": "2020-06-04T15:14:39.866Z",
        "description": "A Helm chart for Kubernetes",
        "digest": "0c1f206f5f5c80ef8fa7574d68b22c91544c9ebe3f0ce12a3c67f26aa86b442e",
        "keywords": [],
        "maintainers": [],
        "name": "test",
        "sources": [],
        "urls": ["http://repo.example.com/charts/test-1.0.0.tgz"],
        "version": "1.0.0",
    },
    {
        "apiVersion": "v1",
        "appVersion": "1.0",
        "created": "2020-06-04T12:14:39.866Z",
        "description": "A Helm chart for Kubernetes",
        "digest": "0c1f206f5f5c80ef8fa7574d68b22c91544c9ebe3f0ce12a3c67f26aa86b4421",
        "keywords": [],
        "maintainers": [],
        "name": "test",
        "sources": [],
        "urls": ["http://repo.example.com/charts/test-1.0.1.tgz"],
        "version": "1.0.1",
    },
]

fake_chart_versions_detail_resp = {
    "apiVersion": "v1",
    "appVersion": "1.0",
    "created": "2020-06-04T15:14:39.866Z",
    "description": "A Helm chart for Kubernetes",
    "digest": "0c1f206f5f5c80ef8fa7574d68b22c91544c9ebe3f0ce12a3c67f26aa86b442e",
    "keywords": [],
    "maintainers": [],
    "name": "test",
    "sources": [],
    "urls": ["http://repo.example.com/charts/test-1.0.0.tgz"],
    "version": "1.0.0",
}

fake_delete_chart_version_resp = {"deleted": True}
