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
from typing import Dict, List

from .data import bk_repo_json


class FakeBkRepoMod:
    """A fake object for replacing the real components.bkrepo.BkRepoClient module"""

    def __init__(self, username: str = None, access_token: str = None, password: str = None):
        pass

    def get_chart_version_detail(self, project_name: str, repo_name: str, chart_name: str, version: str) -> Dict:
        return bk_repo_json.fake_chart_versions_detail_resp

    def list_charts(self, project_name: str, repo_name: str, start_time: str = None) -> Dict:
        return bk_repo_json.fake_list_charts_resp

    def create_project(self, project_code: str, project_name: str, description: str) -> Dict:
        return bk_repo_json.fake_create_project_resp

    def create_repo(self, project_code: str) -> Dict:
        return bk_repo_json.fake_create_repo_resp

    def set_auth(self, project_code: str, repo_admin_user: str, repo_admin_pwd: str) -> bool:
        return bk_repo_json.fake_set_auth_resp

    def get_chart_versions(self, project_name: str, repo_name: str, chart_name: str) -> List:
        return bk_repo_json.fake_chart_versions_resp

    def delete_chart_version(self, project_name: str, repo_name: str, chart_name: str, version: str) -> Dict:
        return bk_repo_json.fake_delete_chart_version_resp
