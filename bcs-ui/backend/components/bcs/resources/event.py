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
import json
import logging
import re

from backend.utils.basic import getitems

from .resource import CoreAPIClassMixins, Resource

logger = logging.getLogger(__name__)


class Event(Resource, CoreAPIClassMixins):
    def get_res_name_list(self, params):
        res_name_list = params.get('extraInfo.name', '')
        if isinstance(res_name_list, str):
            res_name_list = re.split(r',|;', res_name_list)
        return res_name_list

    def get_events(self, params):
        namespace = params.get('extraInfo.namespace')
        resp = self.api_instance.list_namespaced_event(namespace, _preload_content=False)
        data = json.loads(resp.data)
        kind = params.get('kind')
        res_name_list = self.get_res_name_list(params)
        event_list = []
        for info in data.get('items') or []:
            res_name = info['involvedObject']['name']
            if res_name not in res_name_list:
                continue
            if not kind or kind != info['involvedObject']['kind']:
                continue
            item = {
                'extraInfo': {'namespace': info['involvedObject']['namespace'], 'name': res_name},
                'data': info,
                'kind': info['involvedObject']['kind'],
                'createTime': info['metadata']['creationTimestamp'],
                'type': info['reason'],
                'describe': info.get('message', ''),
                'clusterId': getitems(info, ['metadata', 'labels', 'io.tencent.bcs.clusterid'], default=''),
                'level': info['type'],
                'eventTime': info['metadata']['creationTimestamp'],
                'compoent': info['source']['component'],
                'env': 'k8s',
            }
            event_list.append(item)
        return {'code': 0, 'data': event_list, 'count': len(event_list)}
