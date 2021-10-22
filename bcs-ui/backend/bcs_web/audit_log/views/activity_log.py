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
import arrow
from django.utils.translation import ugettext_lazy as _
from rest_framework import generics, views
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import BKAPIResponse

from .. import constants, serializers
from ..models import UserActivityLog


class LogView(generics.ListAPIView):
    queryset = UserActivityLog.objects.all().order_by('-activity_time')
    serializer_class = serializers.ActivityLogSLZ
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def filter_queryset(self, qs):
        params = {k: v for k, v in self.slz.validated_data.items() if v}
        begin_time = params.pop("begin_time", None)
        if begin_time:
            params["activity_time__gte"] = arrow.get(begin_time).datetime

        end_time = params.pop("end_time", None)
        if end_time:
            params["activity_time__lt"] = arrow.get(end_time).datetime

        return qs.filter(project_id=self.project_id, **params)

    def list(self, request, project_id):
        self.project_id = project_id
        self.slz = serializers.ActivityLogGetSLZ(data=request.GET)
        self.slz.is_valid(raise_exception=True)
        return super().list(self, request)


class ResourceTypesView(views.APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get(self, request):
        """返回资源类型"""
        return Response(constants.ResourceTypes)


class MetaDataView(views.APIView):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get(self, request):
        """获取操作审计元数据"""
        return BKAPIResponse(constants.MetaMap, message=_('获取操作审计元数据成功'))
