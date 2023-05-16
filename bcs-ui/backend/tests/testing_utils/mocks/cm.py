# -*- coding: utf-8 -*-
#
# Tencent is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS Community Edition) available.
# Copyright (C) 2017-2019 THL A29 Limited, a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://opensource.org/licenses/MIT
#
# Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
# an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
# specific language governing permissions and limitations under the License.
#
from typing import Dict, List

from django.utils.crypto import get_random_string


class StubClusterManagerClient:
    def __init__(self, *args, **kwargs):
        pass

    def list_cloud_accounts(self, project_id: str) -> List[Dict]:
        """查询云账号"""
        return [
            {
                "projectID": project_id,
                "accountID": f"BCS-tencentCloud-{get_random_string(8)}",
                "accountName": "operator测试凭证",
                "account": {"secretID": get_random_string(8), "secretKey": get_random_string(8)},
                "creator": get_random_string(8),
                "updater": get_random_string(8),
                "creatTime": "2022-06-11T13:05:29+08:00",
                "updateTime": "2022-06-11T13:05:29+08:00",
            },
            {
                "projectID": project_id,
                "accountID": f"BCS-tencentCloud-{get_random_string(8)}",
                "accountName": "ingress-controller测试凭证",
                "account": {"secretID": get_random_string(8), "secretKey": get_random_string(8)},
                "creator": get_random_string(8),
                "updater": get_random_string(8),
                "creatTime": "2022-06-12T13:05:29+08:00",
                "updateTime": "2022-06-12T13:05:29+08:00",
            },
        ]

    def list_cloud_accounts_by_ids(self, accounts_ids: List[str]) -> List[Dict]:
        """根据 ID 查询账号信息"""
        return [
            {
                "projectID": get_random_string(32),
                "accountID": acct_id,
                "accountName": "operator测试凭证",
                "account": {"secretID": get_random_string(8), "secretKey": get_random_string(8)},
                "creator": get_random_string(8),
                "updater": get_random_string(8),
                "creatTime": "2022-06-11T13:05:29+08:00",
                "updateTime": "2022-06-11T13:05:29+08:00",
            }
            for acct_id in accounts_ids
        ]
