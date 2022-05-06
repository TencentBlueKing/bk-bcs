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
from itertools import product
from typing import Dict, List

from backend.resources.networks.common.formatter import NetworkFormatter


class EndpointsFormatter(NetworkFormatter):
    """Endpoints 格式化"""

    def parse_endpoints(self, resource_dict: Dict) -> List:
        """解析 endpoints 信息"""
        endpoints = []
        for subset in resource_dict.get('subsets', []):
            # endpoints 为 subsets ips 与 ports 的笛卡儿积
            ips = [addr['ip'] for addr in subset.get('addresses', []) if addr.get('ip')]
            ports = [p['port'] for p in subset.get('ports', []) if p.get('port')]
            endpoints.extend([f'{ip}:{port}' for ip, port in product(ips, ports)])
        return endpoints

    def format_dict(self, resource_dict: Dict) -> Dict:
        res = self.format_common_dict(resource_dict)
        res.update({'endpoints': self.parse_endpoints(resource_dict)})
        return res
