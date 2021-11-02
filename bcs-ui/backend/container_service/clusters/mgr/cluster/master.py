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

import attr
from django.conf import settings

from backend.components import cc, gse
from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.constants import K8S_NODE_ROLE_MASTER
from backend.resources.node.client import Node


@attr.dataclass
class BcsClusterMaster:
    ctx_cluster: CtxCluster
    biz_id: str
    admin_username: str = settings.ADMIN_USERNAME

    def get_masters(self) -> List[Dict]:
        """获取master信息
        1. 查询集群中的ip和name
        2. 通过ip查询主机所在的机房、机架、机型
        3. 通过ip查询主机的agent信息
        """
        masters = self._get_cluster_masters()
        cc_hosts = self._get_cc_hosts_by_ip(list(masters.keys()))
        gse_agents = self.get_agent_status_by_ip(list(cc_hosts.values()))
        # 组装数据
        for inner_ip, master in masters.items():
            master.update(cc_hosts.get(inner_ip, {}), **gse_agents.get(inner_ip, {}))

        return list(masters.values())

    def _get_cluster_masters(self) -> Dict[str, str]:
        """查询集群中的master ip和name"""
        node_client = Node(self.ctx_cluster)
        # NOTE: 返回节点出现异常，直接报错
        cluster_nodes = node_client.list(is_format=False)
        # 过滤 master 信息
        masters = {}
        for node in cluster_nodes.items:
            labels = node.labels
            # 排除非master节点
            if labels.get(K8S_NODE_ROLE_MASTER) != "true":
                continue
            masters[node.inner_ip] = {"inner_ip": node.inner_ip, "host_name": node.name}
        return masters

    def _get_cc_hosts_by_ip(self, inner_ips: List[str]) -> Dict[str, Dict]:
        """通过 IP 查询主机信息
        包含: 机房、机架、机型
        """
        host_property_filter = {
            "condition": "OR",
            "rules": [{"field": "bk_host_innerip", "operator": "equal", "value": inner_ip} for inner_ip in inner_ips],
        }
        try:
            hosts = cc.HostQueryService(
                self.admin_username, self.biz_id, host_property_filter=host_property_filter
            ).fetch_all()
        except Exception:
            # 忽略异常，直接返回为空
            return {}
        # 组装机房、机架、机型数据
        default_cloud_id = 0
        return {
            host["bk_host_innerip"]: {
                "inner_ip": host["bk_host_innerip"],
                "idc": host.get("idc_name"),
                "rack": host.get("rack"),
                "device_class": host.get("svr_device_class"),
                "bk_cloud_id": host.get("bk_cloud_id", default_cloud_id),
            }
            for host in hosts
        }

    def get_agent_status_by_ip(self, hosts: List) -> Dict[str, int]:
        """通过 IP 查询主机 agent 状态"""
        # 主机为空时，直接返回
        if not hosts:
            return {}
        params = [{"ip": host["inner_ip"], "bk_cloud_id": host["bk_cloud_id"]} for host in hosts]
        try:
            agents = gse.get_agent_status(self.admin_username, params)
        except Exception:
            return {}
        # 如果返回状态字段缺失，则认为agent状态异常，其中0表示agent不在线
        default_agent_alive = 0
        return {agent["ip"]: {"agent": agent.get("bk_agent_alive", default_agent_alive)} for agent in agents}
