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
from backend.utils.decorators import parse_response_data


@parse_response_data(default_data={})
def get_project(access_token, project_id):
    return paas_cc.get_project(access_token, project_id)


@parse_response_data(default_data={})
def query_projects(access_token, query_params=None):
    return paas_cc.get_projects(access_token, query_params)


def list_projects(access_token, query_params=None):
    data = query_projects(access_token, query_params)
    projects = data.get("results") or []
    # 为了兼容导航的参数要求, 增加了project_code字段
    for p in projects:
        p["project_code"] = p["english_name"]
    return projects


@parse_response_data()
def update_project(access_token, project_id, data):
    return paas_cc.update_project_new(access_token, project_id, data)


@parse_response_data()
def create_project(access_token, data):
    return paas_cc.create_project(access_token, data)
