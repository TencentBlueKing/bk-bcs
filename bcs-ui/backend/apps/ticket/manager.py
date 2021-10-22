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
import abc
import logging

from django.conf import settings

from backend.apps.ticket.models import TlsCert
from backend.components.ticket import TicketClient

logger = logging.getLogger(__name__)


class TLSCertManagerFactory:
    def __init__(self):
        self._tls_certs = {}

    def register(self, tls_cert: str, tls_cert_cls):
        self._tls_certs[tls_cert] = tls_cert_cls

    def create(self, request, project_id: str):
        if settings.IS_USE_BCS_TLS:
            tls_cert = 'bcs'
        else:
            tls_cert = 'ticket'

        tls_cert_cls = self._tls_certs.get(tls_cert)
        if not tls_cert_cls:
            raise ValueError(f'{tls_cert} not in {self._tls_certs}')
        return tls_cert_cls(request, project_id)


class TLSCertManagerBase(abc.ABC):
    def __init__(self, request, project_id):
        self.request = request
        self.project_id = project_id

    @abc.abstractmethod
    def get_certs(self):
        """获取证书列表"""

    @property
    def cert_list_url(self):
        """跳转链接"""
        return ''


class BCSTLSCertManager(TLSCertManagerBase):
    def get_certs(self):
        tls_sets = TlsCert.objects.filter(project_id=self.project_id)
        tls_records = []
        for _tls in tls_sets:
            tls_records.append({'certId': _tls.id, 'certName': _tls.name, 'certType': 'bcstls'})

        data = {'records': tls_records, 'cert_list_url': self.cert_list_url}
        return data


class TicketTLSCertManager(TLSCertManagerBase):
    def get_certs(self):
        project_code = self.request.project.english_name
        client = TicketClient(self.project_id, project_code, self.request)

        # 出现异常时，不弹出错误，并且允许跳转连接
        try:
            resp = client.get_tls_cert_list()
            if resp['status'] != 0:
                raise ValueError(resp)
            data = resp['data']
        except Exception as error:
            logger.error('Request ticket api error, detail: %s', error)
            data = {}

        # 为方便前端同一显示，都添加 certName 字段
        records = data.get('records') or []
        for _r in records:
            _r['certName'] = _r['certId']

        data['cert_list_url'] = self.cert_list_url

        return data

    @property
    def cert_list_url(self):
        project_code = self.request.project.english_name
        url = f'{settings.PAAS_HOST}/console/ticket/{project_code}/certList'
        return url


factory = TLSCertManagerFactory()
factory.register('bcs', BCSTLSCertManager)
factory.register('ticket', TicketTLSCertManager)
