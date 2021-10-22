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

凭证管理系统
"""
import logging

from django.conf import settings

from backend.apps.ticket.bkdh import shortcuts
from backend.components.utils import http_get

logger = logging.getLogger(__name__)


# Ticket(凭证管理)API地址
TICKET_API_PREFIX = "%s/ticket/api/" % settings.DEVOPS_CI_API_HOST


class TicketClient(object):

    url_cert_list = '{prefix}user/certs/{project_code}/hasPermissionList'

    def __init__(self, project_id, project_code, request=None):
        self.project_id = project_id
        self.project_code = project_code

        cookies = request.COOKIES if request else None
        self.kwargs = {'cookies': cookies}

    def handle_error_msg(self, resp):
        """
        1.code 统一返回 0
        2.API 返回错误信息时，记录 error 日志，项目统一的日志记录只记录info 日志
        """
        if resp.get('code') != '00':
            logger.error(
                u'''curl -X {methord} -d "{data}" {url}\nresp:{resp}'''.format(
                    methord=self.methord, data=self.query, url=self.url, resp=resp
                )
            )
        else:
            resp['code'] = 0
        return resp

    def get_tls_cert_list(self):
        """获取tls证书列表"""
        self.url = self.url_cert_list.format(prefix=TICKET_API_PREFIX, project_code=self.project_code)

        self.query = {'permission': 'USE', 'certType': 'tls', 'pageSize': '10000'}
        resp = http_get(self.url, params=self.query, **self.kwargs)
        self.methord = 'GET'
        self.handle_error_msg(resp)
        return resp

    def get_tls_crt_content(self, cert_id):
        """"""
        url = '{prefix}service/certs/{project_code}/tls/{cert_id}/'.format(
            prefix=TICKET_API_PREFIX, project_code=self.project_code, cert_id=cert_id
        )
        crt_content = shortcuts(url, 'serverCrtFile', 'serverCrtSha1')
        return crt_content

    def get_tls_key_content(self, cert_id):
        """"""
        url = '{prefix}service/certs/{project_code}/tls/{cert_id}/'.format(
            prefix=TICKET_API_PREFIX, project_code=self.project_code, cert_id=cert_id
        )
        key_content = shortcuts(url, 'serverKeyFile', 'serverKeySha1')
        return key_content


# 尝试加载并应用 TicketClient 的补丁函数
try:
    from .ticket_ext import patch_ticket_client

    TicketClient = patch_ticket_client(TicketClient)
except ImportError:
    logger.debug('`patch_ticket_client` hook not found, will skip')
