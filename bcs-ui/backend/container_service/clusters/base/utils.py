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
from typing import List, Optional

from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.components import paas_cc
from backend.container_service.clusters.base.models import CtxCluster
from backend.container_service.clusters.constants import ClusterType
from backend.resources.namespace import Namespace
from backend.resources.namespace.constants import PROJ_CODE_ANNO_KEY
from backend.resources.node.client import Node
from backend.utils.basic import getitems
from backend.utils.cache import region
from backend.utils.decorators import parse_response_data
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


def get_clusters(access_token, project_id):
    resp = paas_cc.get_all_clusters(access_token, project_id, desire_all_data=True)
    if resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(f"get clusters error, {resp.get('message')}")
    return resp.get("data", {}).get("results") or []


def get_cluster_versions(access_token, kind="", ver_id="", env=""):
    resp = paas_cc.get_cluster_versions(access_token, kind=kind, ver_id=ver_id, env=env)
    if resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(f"get cluster version, {resp.get('message')}")
    data = resp.get("data") or []
    version_list = []
    # 以ID排序，稳定版本排在前面
    data.sort(key=lambda info: info["id"])
    for info in data:
        configure = json.loads(info.get("configure") or "{}")
        version_list.append(
            {"version_id": info["version"], "version_name": configure.get("version_name") or info["version"]}
        )
    return version_list


def get_cluster_masters(access_token, project_id, cluster_id):
    """获取集群下的master信息"""
    resp = paas_cc.get_master_node_list(access_token, project_id, cluster_id)
    if resp.get("code") != ErrorCode.NoError:
        raise error_codes.APIError(_("获取集群master ip失败，{}").format(resp.get("message")))
    results = resp.get("data", {}).get("results") or []
    if not results:
        raise error_codes.APIError(_("获取集群master ip为空"))
    return results


def get_cluster_nodes(access_token, project_id, cluster_id):
    """获取集群下的node信息
    NOTE: 节点数据通过集群中获取，避免数据不一致
    """
    ctx_cluster = CtxCluster.create(
        id=cluster_id,
        project_id=project_id,
        token=access_token,
    )
    try:
        cluster_nodes = Node(ctx_cluster).list(is_format=False)
    except Exception as e:
        logger.error("查询集群内节点数据异常, %s", e)
        return []
    return [{"inner_ip": node.inner_ip, "status": node.node_status} for node in cluster_nodes.items]


def get_cluster_snapshot(access_token, project_id, cluster_id):
    """获取集群快照"""
    resp = paas_cc.get_cluster_snapshot(access_token, project_id, cluster_id)
    if resp.get("code") != ErrorCode.NoError:
        raise error_codes.APIError(_("获取集群快照失败，{}").format(resp.get("message")))
    return resp.get("data") or {}


def get_cluster_info(access_token, project_id, cluster_id):
    resp = paas_cc.get_cluster(access_token, project_id, cluster_id)
    if resp.get("code") != ErrorCode.NoError:
        raise error_codes.APIError(_("获取集群信息失败，{}").format(resp.get("message")))
    return resp.get("data") or {}


def update_cluster_status(access_token, project_id, cluster_id, status):
    """更新集群状态"""
    data = {"status": status}
    resp = paas_cc.update_cluster(access_token, project_id, cluster_id, data)
    if resp.get("code") != ErrorCode.NoError:
        raise error_codes.APIError(_("更新集群状态失败，{}").format(resp.get("message")))
    return resp.get("data") or {}


@parse_response_data(default_data={})
def get_cluster(access_token, project_id, cluster_id):
    return paas_cc.get_cluster(access_token, project_id, cluster_id)


@region.cache_on_arguments(expiration_time=3600 * 24 * 7)
def get_cluster_coes(access_token, project_id, cluster_id):
    """获取集群类型，因为集群创建后，集群类型不允许修改
    TODO: 为减少调用接口耗时，是否需要缓存？
    """
    cluster = get_cluster(access_token, project_id, cluster_id)
    return cluster["type"]


@parse_response_data()
def delete_cluster(access_token, project_id, cluster_id):
    return paas_cc.delete_cluster(access_token, project_id, cluster_id)


def get_cc_zk_config(access_token, project_id, cluster_id):
    resp = paas_cc.get_zk_config(access_token, project_id, cluster_id)
    if resp.get("code") != ErrorCode.NoError:
        raise error_codes.APIError(_("通过cc获取zk信息出错，{}").format(resp.get("message")))
    data = resp.get("data")
    if not data:
        raise error_codes.APIError(_("通过cc获取zk信息为空"))
    return data[0]


def get_cc_repo_domain(access_token, project_id, cluster_id):
    return paas_cc.get_jfrog_domain(access_token, project_id, cluster_id)


@parse_response_data()
def update_cc_nodes_status(access_token, project_id, cluster_id, nodes):
    """更新记录的节点状态"""
    return paas_cc.update_node_list(access_token, project_id, cluster_id, data=nodes)


def append_shared_clusters(clusters: List) -> List:
    """"追加共享集群，返回包含共享集群的列表"""
    shared_clusters = settings.SHARED_CLUSTERS
    if not shared_clusters:
        return clusters

    # 追加到集群列表中
    # 转换为字典，方便进行匹配
    project_cluster_dict = {cluster["cluster_id"]: cluster for cluster in clusters}
    for cluster in shared_clusters:
        if cluster["cluster_id"] in project_cluster_dict:
            continue
        clusters.append(cluster)

    return clusters


def get_cluster_type(cluster_id: str) -> ClusterType:
    """ 根据集群 ID 获取集群类型（独立/联邦/共享） """
    for cluster in settings.SHARED_CLUSTERS:
        if cluster_id == cluster['cluster_id']:
            return ClusterType.SHARED
    return ClusterType.SINGLE


def is_proj_ns_in_shared_cluster(ctx_cluster: CtxCluster, namespace: Optional[str], project_code: str) -> bool:
    """
    检查命名空间是否在共享集群中且属于指定项目

    :param ctx_cluster: 集群 Context 信息
    :param namespace: 命名空间
    :param project_code: 项目英文名
    :return: True / False
    """
    if not namespace:
        return False
    ns = Namespace(ctx_cluster).get(name=namespace, is_format=False)
    return ns and getitems(ns.metadata, ['annotations', PROJ_CODE_ANNO_KEY]) == project_code


def get_shared_cluster_proj_namespaces(ctx_cluster: CtxCluster, project_code: str) -> List[str]:
    """
    获取指定项目在共享集群中拥有的命名空间

    :param ctx_cluster: 集群 Context 信息
    :param project_code: 项目英文名
    :return: 命名空间列表
    """
    return [
        getitems(ns, 'metadata.name')
        for ns in Namespace(ctx_cluster).list(
            is_format=False, cluster_type=ClusterType.SHARED, project_code=project_code
        )['items']
    ]
