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
from rest_framework.response import Response

from backend.components import paas_cc
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer

from ..base_serializers import BaseParamsSLZ
from ..base_views import BaseAPIViews


class ClustersViewSet(BaseAPIViews):
    renderer_classes = (BKAPIRenderer,)

    def list_clusters(self, request, project_id):
        params = request.query_params
        params_slz = BaseParamsSLZ(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data

        resp = paas_cc.get_all_clusters(params_slz["access_token"], project_id, desire_all_data=True)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"))
        data = resp.get("data") or {}
        results = data.get("results") or []
        if not results:
            raise error_codes.CheckFailed.f("查询项目下集群失败，请稍后重试")
        return Response(results)
