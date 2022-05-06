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
from typing import Dict, List

from backend.resources.workloads.common.formatter import WorkloadFormatter
from backend.resources.workloads.pod.utils import PodStatusParser
from backend.utils.basic import getitems


class PodFormatter(WorkloadFormatter):
    """Pod 格式化"""

    def parse_container_images(self, resource_dict: Dict) -> List:
        """pod 配置格式与其它 工作负载类 资源不一致，需要重写解析逻辑"""
        containers = getitems(resource_dict, 'spec.containers', [])
        return list({c['image'] for c in containers if 'image' in c})

    def format_dict(self, resource_dict: Dict) -> Dict:
        res = self.format_common_dict(resource_dict)
        status = resource_dict['status']

        container_statuses = status.get('containerStatuses', [])
        res.update(
            {
                'status': PodStatusParser(resource_dict).parse(),
                'readyCnt': len([s for s in container_statuses if s['ready']]),
                'totalCnt': len(container_statuses),
                'restartCnt': sum([s['restartCount'] for s in container_statuses]),
                'hostIP': status.get('hostIP', ''),
                'podIP': status.get('podIP', ''),
                'name': getitems(resource_dict, ["metadata", "name"], ""),
                'namespace': getitems(resource_dict, ["metadata", "namespace"], ""),
            }
        )
        return res
