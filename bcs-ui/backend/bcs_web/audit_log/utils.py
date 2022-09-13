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
from dataclasses import dataclass
from typing import Tuple

from django.conf import settings

from backend.components import paas_cc
from backend.components.bcs.bcs_common_api import BCSClient
from backend.container_service.clusters.models import CommonStatus
from backend.utils.errcodes import ErrorCode
from backend.utils.funutils import convert_mappings
from backend.utils.response import ComponentData

EVENT_RESULT_MAPPINGS = {
    "id": "_id",
    "env": "env",
    "kind": "kind",
    "level": "level",
    "component": "component",
    "type": "type",
    "cluster_id": "clusterId",
    "event_time": "eventTime",
    "describe": "describe",
    "data": "data",
    "create_time": "createTime",
    "begin_time": "timeBegin",
    "end_time": "timeEnd",
    "offset": "offset",
    "limit": "length",
}


@dataclass
class EventComponentData(ComponentData):
    count: int = 0


def get_project_clusters(access_token: str, project_id: str) -> ComponentData:
    """获取集群信息"""
    cluster_info = paas_cc.get_all_clusters(access_token, project_id)

    result = cluster_info.get('code') == ErrorCode.NoError
    message = cluster_info.get('message', '')
    data = {}

    clusters = cluster_info.get('data')
    for i in clusters.get('results') or ():
        if not i.get('name', '').startswith('deleted'):
            data[i["cluster_id"]] = i

    return ComponentData(result, data, message)


def _get_event(access_token: str, project_id: str, env: str, query) -> EventComponentData:
    client = BCSClient(access_token, project_id, None, env)
    resp = client.get_events(query)
    result = resp.get('code') == ErrorCode.NoError
    message = resp.get('message', '')
    data = resp.get('data') or []
    count = int(resp.get('total') or 0)
    return EventComponentData(result, data, message, count)


def get_event(access_token: str, project_id: str, validated_data: dict, clusters: dict) -> Tuple[list, int]:
    results = []
    count = 0

    query = convert_mappings(EVENT_RESULT_MAPPINGS, validated_data, reversed=True)
    for env in settings.BCS_EVENT_ENV:
        result = _get_event(access_token, project_id, env, query)
        for res in result.data:
            # 转换变量
            res = convert_mappings(EVENT_RESULT_MAPPINGS, res)

            # 补充集群名称
            cluster_id = res.get('cluster_id')
            cluster_info = clusters.get(cluster_id) or {}
            res['cluster_name'] = cluster_info.get('name', '')

            results.append(res)
        count += result.count

    return results, count
