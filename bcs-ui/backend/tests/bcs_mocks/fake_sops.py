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
from typing import Dict

from backend.components.sops import CreateTaskParams

from .data import sops_json


class FakeSopsMod:
    """A fake object for replacing the real components.sops module"""

    def create_task(self, bk_biz_id: str, template_id: str, data: CreateTaskParams) -> Dict:
        return sops_json.create_task_ok

    def start_task(self, bk_biz_id: str, task_id: str) -> Dict:
        return sops_json.start_task_ok

    def get_task_status(self, bk_biz_id: str, task_id: str) -> Dict:
        return sops_json.get_task_status_ok
