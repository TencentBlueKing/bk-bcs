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
from typing import Dict, List, Union

from backend.resources.networks.common.formatter import NetworkFormatter
from backend.utils.basic import get_with_placeholder, getitems


class IngressFormatter(NetworkFormatter):
    """Ingress 格式化"""

    def parse_hosts(self, resource_dict: Dict) -> List:
        """解析 Ingress hosts"""
        rules = getitems(resource_dict, 'spec.rules', [])
        return [r['host'] for r in rules if r.get('host')]

    def parse_addresses(self, resource_dict: Dict) -> List:
        """解析 Ingress address"""
        addresses = []
        for ingress in getitems(resource_dict, 'status.loadBalancer.ingress', []):
            if ingress.get('ip'):
                addresses.append(ingress['ip'])
            elif ingress.get('hostname'):
                addresses.append(ingress['hostname'])
        return addresses

    def parse_default_ports(self, resource_dict: Dict) -> Union[str, int]:
        """
        解析 Ingress 默认 port
        默认是 HTTP 端口，如果有TLS配置则为 HTTP + HTTPS 端口
        """
        return '80, 443' if 'tls' in resource_dict['spec'] else '80'

    def parse_v1_rules(self, resource_dict: Dict) -> List:
        """解析 networking.k8s.io/v1 版本 Ingress Rules"""
        rules = []
        for r in getitems(resource_dict, 'spec.rules', []):
            sub_rules = [
                {
                    'host': get_with_placeholder(r, 'host'),
                    'path': get_with_placeholder(p, 'path'),
                    'pathType': get_with_placeholder(p, 'pathType'),
                    'serviceName': get_with_placeholder(p, 'backend.service.name'),
                    'port': get_with_placeholder(p, 'backend.service.port.number'),
                }
                for p in getitems(r, 'http.paths', [])
            ]
            rules.extend(sub_rules)
        return rules

    def parse_v1beta1_rules(self, resource_dict: Dict) -> List:
        """解析 extensions/v1beta1 版本 Ingress Rules"""
        rules = []
        for r in getitems(resource_dict, 'spec.rules', []):
            sub_rules = [
                {
                    'host': get_with_placeholder(r, 'host'),
                    'path': get_with_placeholder(p, 'path'),
                    'serviceName': get_with_placeholder(p, 'backend.serviceName'),
                    'port': get_with_placeholder(p, 'backend.servicePort'),
                }
                for p in getitems(r, 'http.paths', [])
            ]
            rules.extend(sub_rules)
        return rules

    def format_dict(self, resource_dict: Dict) -> Dict:
        res = self.format_common_dict(resource_dict)
        # 根据不同 api 版本，选择不同的解析 Rules 方法
        parse_rules_func = {
            'networking.k8s.io/v1': self.parse_v1_rules,
            'extensions/v1beta1': self.parse_v1beta1_rules,
        }[resource_dict['apiVersion']]
        res.update(
            {
                'hosts': self.parse_hosts(resource_dict),
                'addresses': self.parse_addresses(resource_dict),
                'default_ports': self.parse_default_ports(resource_dict),
                'rules': parse_rules_func(resource_dict),
            }
        )
        return res
