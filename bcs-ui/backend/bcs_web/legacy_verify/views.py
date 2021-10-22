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
import logging

from django.utils.translation import ugettext_lazy as _
from rest_framework import viewsets

from backend.accounts import bcs_perm
from backend.components import paas_cc
from backend.utils import FancyDict
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.exceptions import APIError, NoAuthPermError, VerifyAuthPermError, VerifyAuthPermErrorWithNoRaise
from backend.utils.response import APIResult

from . import constants, serializers

logger = logging.getLogger(__name__)


class Perm(viewsets.ViewSet):
    def get_project_info(self, request, project_id):
        """获取项目信息"""
        resp = paas_cc.get_project(request.user.token.access_token, project_id)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get('message'))
        request.project = FancyDict(resp.get('data', {}))

    def verify(self, request):
        """权限判断接口，前端使用"""
        # serializer在设置字段required为false时，如果参数为None，会提示字段为null
        if not request.data.get('resource_code'):
            request.data.pop('resource_code', '')
        serializer = serializers.PermVerifySLZ(data=request.data, context={'request': request})
        serializer.is_valid(raise_exception=True)

        data = serializer.data

        resource_code = data.get('resource_code')
        resource_name = data.get('resource_name')
        is_raise = data.get('is_raise')
        self.get_project_info(request, data['project_id'])

        try:
            perm = bcs_perm.get_perm_cls(
                data['resource_type'], request, data['project_id'], resource_code, resource_name
            )
        except (AttributeError, TypeError):
            raise APIError(_("resource_code不合法"))

        handler = getattr(perm, 'can_%s' % data['policy_code'])
        if handler:
            try:
                handler(raise_exception=True)
            except NoAuthPermError as error:
                if is_raise:
                    raise VerifyAuthPermError(error.args[0], error.args[1])
                else:
                    raise VerifyAuthPermErrorWithNoRaise(error.args[0], error.args[1])

        return APIResult({}, _("验证权限成功"))

    def verify_multi(self, request):
        """权限判断接口，前端使用, 批量接口"""
        serializer = serializers.PermMultiVerifySLZ(data=request.data, context={'request': request})
        serializer.is_valid(raise_exception=True)

        data = serializer.data
        operator = data['operator']
        msg = ''
        err_data = []
        self.get_project_info(request, data['project_id'])

        for res in data['resource_list']:
            resource_code = res.get('resource_code')
            resource_name = res.get('resource_name')
            try:
                perm = bcs_perm.get_perm_cls(
                    res['resource_type'], request, data['project_id'], resource_code, resource_name
                )
            except (AttributeError, TypeError):
                raise APIError(_("resource_code不合法"))

            handler = getattr(perm, 'can_%s' % res['policy_code'])
            try:
                if handler:
                    handler(raise_exception=True)
            except bcs_perm.NoAuthPermError as error:
                msg = msg or error.args[0]
                err_data.extend(error.args[1])

        if len(data['resource_list']) == 1 and err_data:
            raise VerifyAuthPermError(msg, err_data)
        elif operator == constants.PermMultiOperator.AND.value and err_data:
            raise VerifyAuthPermError(msg, err_data)
        elif operator == constants.PermMultiOperator.OR.value and len(data['resource_list']) == len(err_data):
            raise VerifyAuthPermError(msg, err_data)

        return APIResult({}, _("验证权限成功"))
