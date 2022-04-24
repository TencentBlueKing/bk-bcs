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
from typing import Dict, List, Optional

from django.conf import settings
from iam import IAM, OP, Action, Request, Subject
from iam.exceptions import AuthInvalidRequest

from .resources.project import ProjectAction


class ProjectFilter:
    """
    项目权限过滤器.
    note: 用于查询用户具有 project_view 权限的所有项目(iam v3 本身不支持这种查询)
    """

    iam = IAM(
        settings.APP_CODE,
        settings.SECRET_KEY,
        settings.BK_IAM_HOST,
        settings.BK_PAAS_INNER_HOST,
        settings.BK_IAM_APIGATEWAY_URL,
    )

    def make_view_perm_filter(self, username: str) -> Dict:
        request = Request(
            settings.BK_IAM_SYSTEM_ID,
            Subject('user', username),
            Action(ProjectAction.VIEW),
            None,
            None,
        )
        policies = self._do_policy_query(request)
        if not policies:
            return {}

        return self._make_dict_filter(policies)

    @staticmethod
    def op_is_any(dict_filter: Dict[str, List[str]]) -> bool:
        if not dict_filter:
            return False
        if dict_filter.get('op') == OP.ANY:
            return True
        return False

    def _do_policy_query(self, request) -> Optional[Dict]:
        # 1. validate
        if not isinstance(request, Request):
            raise AuthInvalidRequest('request should be instance of iam.auth.models.Request')

        request.validate()

        # 2. _client.policy_query
        policies = self.iam._do_policy_query(request)

        # the polices maybe none
        if not policies:
            return None

        return policies

    @staticmethod
    def _make_dict_filter(policies: Dict) -> Dict:
        """
        基于策略规则, 生成 project_id 过滤器

        :param policies: 权限中心返回的策略规则, 如 {'op': OP.IN, 'value': [2, 1], 'field': 'project.id'}
        :return: project_id 过滤器, 如 {'value': [1, 2], 'op': OP.IN}
        """
        op = policies['op']
        if op not in [OP.IN, OP.EQ, OP.ANY, OP.OR, OP.AND]:
            raise AuthInvalidRequest(f'make_dict_filter does not support op:{op}')

        if op == OP.EQ:
            return {'value': [policies['value']], 'op': OP.IN}

        if op in [OP.IN, OP.ANY]:
            return {'value': policies['value'] or [], 'op': op}

        # 如果 op 是 OP.OR 或 OP.AND，只处理一级，不考虑嵌套的情况
        value_list = []
        for policy in policies['content']:
            if policy['field'] != 'project.id':
                continue

            op = policy['op']
            if op == OP.ANY:
                return {'value': policy['value'] or [], 'op': op}

            value = policy['value']
            if op == OP.IN:
                value_list.extend(value)
            elif op == OP.EQ:
                value_list.append(value)

        return {'value': list(set(value_list)), 'op': OP.IN}
