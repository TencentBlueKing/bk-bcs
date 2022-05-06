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
import logging
from typing import Dict, List

from backend.components import cc, gse, paas_cc
from backend.container_service.clusters.models import CommonStatus
from backend.utils.basic import get_with_placeholder

logger = logging.getLogger(__name__)


def fetch_cc_app_hosts(username, bk_biz_id, bk_module_id=None, bk_set_id=None) -> List[Dict]:
    """
    拉取 业务 下机器列表（业务/集群/模块全量）
    TODO 当前场景只需要支持单模块/集群，后续有需要可扩展

    :return: CMDB 业务下机器列表
    """
    bk_module_ids = [bk_module_id] if bk_module_id else None
    bk_set_ids = [bk_set_id] if bk_set_id else None
    return cc.HostQueryService(username, bk_biz_id, bk_module_ids, bk_set_ids).fetch_all()


def fetch_all_cluster_nodes(access_token: str) -> Dict:
    """
    获取所有集群中使用的主机信息

    :return: {'ip': node_info}
    """
    nodes = paas_cc.get_all_cluster_hosts(access_token, exclude_status_list=[CommonStatus.Removed])
    return {n['inner_ip']: n for n in nodes}


def fetch_project_cluster_info(access_token: str) -> Dict:
    """
    获取 项目 & 集群信息

    :return: {cluster_id: {'project_name': p_name, 'cluster_name': c_name}
    """
    resources = paas_cc.get_project_cluster_resource(access_token)
    return {
        cluster['id']: {'project_name': project['name'], 'cluster_name': cluster['name']}
        for project in resources
        if project
        for cluster in project['cluster_list']
        if cluster
    }


def gen_used_info(host: Dict, all_cluster_nodes: Dict, project_cluster_info: Dict) -> Dict:
    """
    生成主机被使用的项目，集群信息

    :param host: 原始主机信息
    :param all_cluster_nodes: 全集群节点信息
    :param project_cluster_info: 项目 & 集群信息
    :return: 主机的使用信息
    """
    info = {'project_name': '', 'cluster_name': '', 'cluster_id': '', 'is_used': False}

    for ip in host['bk_host_innerip'].split(','):
        node_info = all_cluster_nodes.get(ip)
        if not node_info:
            continue
        info['is_used'] = True
        info['cluster_id'] = node_info.get('cluster_id')
        name_dict = project_cluster_info.get(info['cluster_id']) or {}
        info['project_name'] = get_with_placeholder(name_dict, 'project_name')
        info['cluster_name'] = get_with_placeholder(name_dict, 'cluster_name')
        break

    return info


def is_valid_machine(*args, **kwargs) -> bool:
    """判断是否为机器类型是否可用"""
    # NOTE ce 版本不做判断
    return True


def attach_project_cluster_info(host_list: List, all_cluster_nodes: Dict, project_cluster_info: Dict) -> List:
    """
    更新节点使用状态 & 是否类型可用，补充项目，集群等信息

    :param host_list: 原始主机列表
    :param all_cluster_nodes: 全集群节点信息
    :param project_cluster_info: 项目 & 集群信息
    :return: 包含使用信息的主机列表
    """
    new_host_list, used_host_list = [], []
    for host in host_list:
        # 没有 内网 IP 的主机直接忽略
        if 'bk_host_innerip' not in host or not host['bk_host_innerip']:
            continue

        host.update(gen_used_info(host, all_cluster_nodes, project_cluster_info))
        host['is_valid'] = is_valid_machine(host)

        if host['is_used']:
            used_host_list.append(host)
        else:
            new_host_list.append(host)

    # 已被使用的 Host 放列表最后
    new_host_list.extend(used_host_list)
    return new_host_list


def is_host_selectable(host: Dict) -> bool:
    """
    判断机器是否可被选择为 Master / Node
    已经被使用 / 机器类型不可用 则不能选择
    Agent 异常也可以选择（初始化会重装）

    :param host: 主机配置信息
    :return: 是否可选择
    """
    return not host['is_used'] and host['is_valid']


def update_gse_agent_status(username, host_list: List) -> List:
    """更新 GSE Agent 状态信息"""
    gse_params = []
    for info in host_list:
        bk_cloud_id = info.get('bk_cloud_id') or 0
        gse_params.extend(
            [
                {'plat_id': bk_cloud_id, 'bk_cloud_id': bk_cloud_id, 'ip': ip}
                for ip in info.get('bk_host_innerip', '').split(',')
            ]
        )
    gse_host_status_map = {info['ip']: info for info in gse.get_agent_status(username, gse_params)}
    # 根据 IP 匹配更新 Agent 信息
    cc_host_map = {host['bk_host_innerip']: host for host in host_list}
    for ips in cc_host_map:
        # 同主机可能存在多个 IP，任一 IP Agent 正常即可
        for ip in ips.split(','):
            if ip not in gse_host_status_map:
                continue
            ip_status = gse_host_status_map[ip]
            cc_host_map[ips]['agent_alive'] = ip_status.get('bk_agent_alive')
            break

    return list(cc_host_map.values())


try:
    from .utils_ext import *  # noqa
except ImportError as e:
    logger.debug("Load extension failed: %s", e)
