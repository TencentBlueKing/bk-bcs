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

from backend.resources.networks.common.formatter import NetworkFormatter
from backend.utils.basic import getitems


class ServiceFormatter(NetworkFormatter):
    """Service 格式化"""

    def parse_external_ip(self, resource_dict: Dict) -> List:
        """解析 Service external_ip"""
        external_ips = []
        for ingress in getitems(resource_dict, 'status.loadBalancer.ingress', []):
            if ingress.get('ip'):
                external_ips.append(ingress['ip'])
            elif ingress.get('hostname'):
                external_ips.append(ingress['hostname'])
        return external_ips

    def parse_ports(self, resource_dict: Dict) -> List:
        """解析 Service ports"""
        origin_ports = getitems(resource_dict, 'spec.ports', [])
        # 若 nodePort 存在则需要展示，否则隐藏
        return [
            f"{p['port']}:{p['nodePort']}/{p['protocol']}" if p.get('nodePort') else f"{p['port']}/{p['protocol']}"
            for p in origin_ports
        ]

    def format_dict(self, resource_dict: Dict) -> Dict:
        res = self.format_common_dict(resource_dict)
        res.update(
            {
                'externalIP': self.parse_external_ip(resource_dict),
                'ports': self.parse_ports(resource_dict),
            }
        )
        return res
