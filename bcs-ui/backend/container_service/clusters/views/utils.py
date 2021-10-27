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
from backend.components import paas_cc
from backend.container_service.clusters.base.constants import ClusterCOES
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes


def filter_areas(request, data):
    """
    1. 如果容器编排类型为TKE，并且区域中文名称以TKE开始，则认为是来源TKE
    2. 如果区域不以TKE开始，则认为来源BCS
    """
    areas = {'TKE': [], 'BCS': []}
    for area in data['results'] or []:
        if area['chinese_name'].startswith(ClusterCOES.TKE.name):
            areas['TKE'].append(area)
        else:
            areas['BCS'].append(area)

    # 通过集群类型获取区域配置
    if request.query_params.get("coes") == ClusterCOES.TKE.value:
        return areas["TKE"]

    return areas["BCS"]


def get_areas(request):
    areas = paas_cc.get_area_list(request.user.token.access_token)
    if areas.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(areas.get('message'))

    data = areas.get('data') or {}
    if not data:
        return data

    # 处理区域来源
    area_list = filter_areas(request, data)

    return {'results': area_list, 'count': len(area_list)}
