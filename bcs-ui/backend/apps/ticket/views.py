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
import time

from django.conf import settings
from django.utils import timezone
from rest_framework import generics, views
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.apps.ticket.models import TlsCert
from backend.apps.ticket.serializers import TlsCertModelSLZ, TlsCertSlZ
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import BKAPIResponse
from backend.utils.views import FinalizeResponseMixin

from . import manager

logger = logging.getLogger(__name__)


class TLSCertView(views.APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get(self, request, project_id):
        cert_mgr = manager.factory.create(request, project_id)
        data = cert_mgr.get_certs()
        return BKAPIResponse(data)


class BCSTLSCertView(generics.CreateAPIView):
    serializer_class = TlsCertModelSLZ
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def perform_create(self, serializer):
        slz = TlsCertSlZ(data=self.request.data, context={'project_id': self.kwargs['project_id']})
        slz.is_valid(raise_exception=True)

        serializer.save(creator=self.request.user.username, project_id=self.kwargs['project_id'])


class SingleBCSTLSCertView(generics.RetrieveUpdateDestroyAPIView):
    serializer_class = TlsCertModelSLZ
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_queryset(self):
        return TlsCert.objects.filter(project_id=self.kwargs['project_id'])

    def perform_update(self, serializer):
        slz = TlsCertSlZ(
            data=self.request.data, context={"pk": self.kwargs['pk'], 'project_id': self.kwargs['project_id']}
        )
        slz.is_valid(raise_exception=True)

        serializer.save(updator=self.request.user.username, project_id=self.kwargs['project_id'])

    def post(self, request, project_id, pk):
        """前端是post请求，直接alias put方法"""
        return super().put(request, project_id, pk)

    def perform_destroy(self, instance):
        _del_prefix = f'[deleted_{int(time.time())}]'
        del_name = f"{_del_prefix}{instance.name}"
        instance.name = del_name
        instance.is_deleted = True
        instance.deleted_time = timezone.now()
        instance.save()
