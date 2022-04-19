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
from dataclasses import asdict, dataclass
from typing import Dict, List, Union

from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.components.base import BaseHttpClient, BkApiClient, ComponentAuth, response_handler
from backend.components.utils import http_delete, http_get, http_patch, http_post, http_put
from backend.container_service.clusters.models import CommonStatus
from backend.utils.basic import getitems
from backend.utils.decorators import parse_response_data
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

from .cluster_manager import get_shared_clusters

logger = logging.getLogger(__name__)

BCS_CC_API_PRE_URL = settings.BCS_CC_API_PRE_URL

DEFAULT_TIMEOUT = 20


def get_project(access_token: str, project_id: str) -> Dict:
    """获取项目信息"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/"
    params = {"access_token": access_token}
    project = http_get(url, params=params, timeout=20)
    return project


def get_projects(access_token, query_params=None):
    url = f"{BCS_CC_API_PRE_URL}/projects/"
    params = query_params or {}
    params["access_token"] = access_token
    return http_get(url, params=params)


def create_project(access_token, data):
    url = f"{BCS_CC_API_PRE_URL}/projects/"
    return http_post(url, params={"access_token": access_token}, json=data)


def update_project_new(access_token, project_id, data):
    """更新项目信息"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/"
    params = {"access_token": access_token}
    project = http_put(url, params=params, json=data)
    return project


def create_cluster(access_token, project_id, data):
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/"
    # 判断环境
    env_name = data.get("environment")
    data["environment"] = settings.CLUSTER_ENV.get(env_name)
    params = {"access_token": access_token}
    return http_post(url, params=params, json=data)


def update_cluster(access_token, project_id, cluster_id, data):
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}"
    params = {"access_token": access_token}
    return http_put(url, params=params, json=data)


def get_cluster(access_token, project_id, cluster_id):
    if cluster_id in [cluster["cluster_id"] for cluster in get_shared_clusters()]:
        return get_cluster_by_id(access_token, cluster_id)
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}"
    params = {"access_token": access_token}
    return http_get(url, params=params)


def get_cluster_by_id(access_token: str, cluster_id: str) -> Dict:
    """通过集群ID获取集群信息"""
    url = f"{BCS_CC_API_PRE_URL}/clusters/{cluster_id}/"
    params = {"access_token": access_token}
    return http_get(url, params=params)


def get_all_clusters(access_token, project_id, limit=None, offset=None, desire_all_data=0):
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/"
    params = {"access_token": access_token}
    if limit:
        params["limit"] = limit
    if offset:
        params["offset"] = offset
    # NOTE: 现阶段都是查询全量集群的场景
    params["desire_all_data"] = 1
    return http_get(url, params=params)


def get_cluster_list(access_token, project_id, cluster_ids):
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters_list/"
    params = {"access_token": access_token}
    data = {"cluster_ids": cluster_ids} if cluster_ids else None
    return http_post(url, params=params, json=data)


def verify_cluster_exist(access_token, project_id, cluster_name):
    """校验cluster name是否存在"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/"
    params = {"name": cluster_name, "access_token": access_token}
    return http_get(url, params=params)


def get_cluster_by_name(access_token, project_id, cluster_name):
    """get cluster info by cluster name"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/"
    params = {"name": cluster_name, "access_token": access_token}
    return http_get(url, params=params)


def get_cluster_snapshot(access_token, project_id, cluster_id):
    """获取集群快照"""
    if getattr(settings, "BCS_CC_CLUSTER_CONFIG", None):
        path = settings.BCS_CC_CLUSTER_CONFIG.format(cluster_id=cluster_id)
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/v1/clusters/{cluster_id}/cluster_config"
    params = {"access_token": access_token}
    return http_get(url, params=params)


def get_area_list(access_token):
    url = f"{BCS_CC_API_PRE_URL}/areas/"
    params = {"access_token": access_token}
    return http_get(url, params=params)


def get_area_info(access_token, area_id):
    """查询指定区域的信息"""
    url = f"{BCS_CC_API_PRE_URL}/areas/{area_id}/"
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token})}
    return http_get(url, headers=headers)


def get_master_node_list(access_token, project_id, cluster_id):
    if getattr(settings, "BCS_CC_GET_CLUSTER_MASTERS", None):
        path = settings.BCS_CC_GET_CLUSTER_MASTERS.format(project_id=project_id, cluster_id=cluster_id)
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/masters/"
    params = {"access_token": access_token}
    return http_get(url, params=params)


def get_project_master_list(access_token, project_id):
    if getattr(settings, "BCS_CC_GET_PROJECT_MASTERS", None):
        path = settings.BCS_CC_GET_PROJECT_MASTERS.format(project_id=project_id)
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/masters/"
    params = {"access_token": access_token}
    return http_get(url, params=params)


def create_node(access_token, project_id, cluster_id, data):
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/nodes/"
    params = {"access_token": access_token}
    return http_patch(url, params=params, json=data)


def get_node_list(access_token, project_id, cluster_id, params=()):
    if getattr(settings, "BCS_CC_GET_PROJECT_NODES", None):
        path = settings.BCS_CC_GET_PROJECT_NODES.format(project_id=project_id)
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/nodes/"
    params = dict(params)
    # 默认拉取项目或集群下的所有节点，防止view层出现分页查询问题
    if "desire_all_data" not in params:
        params["desire_all_data"] = 1
    params.update({"access_token": access_token, "cluster_id": cluster_id})
    return http_get(url, params=params)


def get_node(access_token, project_id, node_id, cluster_id=""):
    if getattr(settings, "BCS_CC_OPER_PROJECT_NODE", None):
        path = settings.BCS_CC_OPER_PROJECT_NODE.format(project_id=project_id, node_id=node_id)
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/nodes/{node_id}/"
    params = {"access_token": access_token, "cluster_id": cluster_id}
    return http_get(url, params=params)


def update_node(access_token, project_id, node_id, data):
    if getattr(settings, "BCS_CC_OPER_PROJECT_NODE", None):
        path = settings.BCS_CC_OPER_PROJECT_NODE.format(project_id=project_id, node_id=node_id)
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/nodes/{node_id}/"
    params = {"access_token": access_token}
    return http_put(url, params=params, json=data)


def update_node_list(access_token, project_id, cluster_id, data):
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/nodes/"
    params = {"access_token": access_token}
    req_data = {"updates": data}
    return http_patch(url, params=params, json=req_data)


def get_cluster_history_data(access_token, project_id, cluster_id, metric, start_at, end_at):
    """获取集群概览历史数据"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/history_data/"
    params = {"access_token": access_token, "start_at": start_at, "end_at": end_at, "metric": metric}
    return http_get(url, params=params)


def get_all_masters(access_token):
    """获取配置中心所有Master"""
    url = f"{BCS_CC_API_PRE_URL}/v1/masters/all_master_list/"
    params = {"desire_all_data": 1}
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token})}
    return http_get(url, params=params, headers=headers)


def get_all_nodes(access_token):
    """获取配置中心所有Node"""
    url = f"{BCS_CC_API_PRE_URL}/v1/nodes/all_node_list/"
    params = {"desire_all_data": 1}
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token})}
    return http_get(url, params=params, headers=headers)


def get_all_cluster_hosts(access_token, exclude_status_list=None):
    node_list_info = get_all_nodes(access_token)
    if node_list_info.get("code") != ErrorCode.NoError:
        raise error_codes.APIError(_("查询所有集群的node节点失败，已经通知管理员，请稍后!"))
    else:
        data = node_list_info.get("data") or []
    master_list_info = get_all_masters(access_token)
    if master_list_info.get("code") != ErrorCode.NoError:
        raise error_codes.APIError(_("查询所有集群的master节点失败，已经通知管理员，请稍后!"))
    data.extend(master_list_info.get("data") or [])
    # 在component层过滤掉状态为removed的host，便于上层直接使用
    if exclude_status_list:
        return [info for info in data if info["status"] not in exclude_status_list]
    return data


def get_project_nodes(access_token, project_id, is_master=False):
    """获取项目下已经添加的Master和Node"""
    # add filter for master or node
    # node filter
    # filter_status = [CommonStatus.InitialCheckFailed, CommonStatus.InitialFailed, CommonStatus.Removed]
    # if is_master:
    filter_status = [CommonStatus.Removed]

    data = []
    # 获取Node
    node_list_info = get_all_nodes(access_token)
    if node_list_info["code"] == ErrorCode.NoError:
        if node_list_info.get("data", []):
            # 过滤掉removed和initial_failed
            data.extend([info for info in node_list_info["data"] if info.get("status") not in filter_status])
    else:
        raise error_codes.APIError(_("查询项目下node节点失败，已经通知管理员，请稍后!"))
    # 获取master
    master_list_info = get_all_masters(access_token)
    if master_list_info["code"] == ErrorCode.NoError:
        if master_list_info.get("data", []):
            # 过滤掉removed和initial_failed
            data.extend([info for info in master_list_info["data"] if info.get("status") not in filter_status])
    else:
        raise error_codes.APIError(_("查询项目下master节点失败，已经通知管理员，请稍后!"))
    return {info["inner_ip"]: True for info in data}


def get_namespace_list(access_token, project_id, with_lb=None, limit=None, offset=None, desire_all_data=None):
    """获取namespace列表"""
    if getattr(settings, "BCS_CC_OPER_PROJECT_NAMESPACES", None):
        path = settings.BCS_CC_OPER_PROJECT_NAMESPACES.format(project_id=project_id)
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/namespaces/"
    params = {"access_token": access_token}
    if desire_all_data:
        params["desire_all_data"] = 1
    if limit:
        params["limit"] = limit
    if offset:
        params["offset"] = offset
    if with_lb:
        params["with_lb"] = with_lb
    return http_get(url, params=params)


def get_cluster_namespace_list(
    access_token, project_id, cluster_id, with_lb=None, limit=None, offset=None, desire_all_data=None
):
    """查询集群下命名空间的信息"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/namespaces/"
    params = {"access_token": access_token}
    if desire_all_data:
        params["desire_all_data"] = 1
    if limit:
        params["limit"] = limit
    if offset:
        params["offset"] = offset
    if with_lb:
        params["with_lb"] = with_lb

    return http_get(url, params=params)


def create_namespace(
    access_token, project_id, cluster_id, name, description, creator, env_type, has_image_secret=None
):
    """创建namespace"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/namespaces/"
    params = {"access_token": access_token}
    payload = {"name": name, "description": description, "creator": creator, "env_type": env_type}
    if has_image_secret is not None:
        payload["has_image_secret"] = has_image_secret
    return http_post(url, params=params, json=payload)


def get_namespace(access_token, project_id, namespace_id, with_lb=None):
    """获取单个namespace"""
    if getattr(settings, "BCS_CC_OPER_PROJECT_NAMESPACE", None):
        path = settings.BCS_CC_OPER_PROJECT_NAMESPACE.format(project_id=project_id, namespace_id=namespace_id)
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/namespaces/{namespace_id}/"
    params = {"access_token": access_token}
    if with_lb:
        params["with_lb"] = with_lb
    return http_get(url, params=params)


def update_node_with_cluster(access_token, project_id, data):
    """批量更新节点所属集群及状态"""
    url = "{host}/projects/{project_id}/nodes/".format(host=BCS_CC_API_PRE_URL, project_id=project_id)

    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token})}
    return http_put(url, json=data, headers=headers)


def get_zk_config(access_token, project_id, cluster_id, environment=None):
    if not environment:
        cluster = get_cluster(access_token, project_id, cluster_id)
        if cluster.get("code") != 0:
            raise error_codes.APIError(cluster.get("message"))
        environment = cluster["data"]["environment"]

    url = f"{BCS_CC_API_PRE_URL}/zk_config/"
    params = {"access_token": access_token, "environment": environment}
    zk_config = http_get(url, params=params, timeout=20)
    return zk_config


def get_jfrog_domain(access_token, project_id, cluster_id):
    """"""
    url = f"{BCS_CC_API_PRE_URL}/clusters/{cluster_id}/related/areas/info/"
    params = {"access_token": access_token}
    res = http_get(url, params=params)
    jfrog_registry = ""
    if res.get("code") == ErrorCode.NoError:
        data = res.get("data") or {}
        configuration = data.get("configuration") or {}
        # jfrog_registry = configuration.get('httpsJfrogRegistry')
        env_type = data.get("env_type")
        # 按集群环境获取仓库地址
        if env_type == "prod":
            jfrog_registry = configuration.get("httpsJfrogRegistry")
        else:
            jfrog_registry = configuration.get("testHttpsJfrogRegistry")
    else:
        logger.error("get jfrog domain error:%s\nurl:%s", res.get("message"), url)
    return jfrog_registry


def get_jfrog_domain_list(access_token, project_id, cluster_id):
    url = f"{BCS_CC_API_PRE_URL}/clusters/{cluster_id}/related/areas/info/"
    params = {"access_token": access_token}
    res = http_get(url, params=params)
    if res.get("code") != ErrorCode.NoError:
        logger.error("get jfrog domain error:%s\nurl:%s", res.get("message"), url)
        return []
    data = res.get("data") or {}
    configuration = data.get("configuration") or {}
    domain_list = []
    for key in ["httpsJfrogRegistry", "testHttpsJfrogRegistry"]:
        if configuration.get(key):
            domain_list.append(configuration.get(key))
    return domain_list


def get_image_registry_list(access_token, cluster_id):
    resp = http_get(
        f"{BCS_CC_API_PRE_URL}/clusters/{cluster_id}/related/areas/info/", params={"access_token": access_token}
    )
    if resp.get("code") != ErrorCode.NoError:
        logger.error("get image registry domains error:%s", resp.get("message"))
        return []

    area_cfg = getitems(resp, ["data", "configuration"], [])
    image_registry_keys = ["httpsJfrogRegistry", "testHttpsJfrogRegistry", "jfrogRegistry", "testJfrogRegistry"]
    return [area_cfg[r_key] for r_key in image_registry_keys if area_cfg.get(r_key)]


def delete_cluster(access_token, project_id, cluster_id):
    """删除集群"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/"
    params = {"access_token": access_token}
    return http_delete(url, params=params)


def delete_cluster_namespace(access_token, project_id, cluster_id):
    """删除集群下的命名空间"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/batch_delete_namespaces/"
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token, "project_id": project_id})}
    return http_delete(url, headers=headers)


def get_base_cluster_config(access_token, project_id, params):
    """获取集群基本配置"""
    if getattr(settings, "BCS_CC_CLUSTER_CONFIG", None):
        path = settings.BCS_CC_CLUSTER_CONFIG.format(cluster_id="null")
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f"{BCS_CC_API_PRE_URL}/v1/clusters/version_config/"
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token, "project_id": project_id})}
    return http_get(url, params=params, headers=headers)


def save_cluster_snapshot(access_token, data):
    """存储集群快照"""
    if getattr(settings, "BCS_CC_CLUSTER_CONFIG", None):
        path = settings.BCS_CC_CLUSTER_CONFIG.format(cluster_id=data["cluster_id"])
        url = f"{BCS_CC_API_PRE_URL}{path}"
    else:
        url = f'{BCS_CC_API_PRE_URL}/v1/clusters/{data["cluster_id"]}/cluster_config/'
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token, "project_id": data["project_id"]})}
    return http_post(url, json=data, headers=headers)


def _get_project_cluster_resource(access_token):
    """获取所有项目、集群信息"""
    url = f"{BCS_CC_API_PRE_URL}/v1/projects/resource/"
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token})}
    return http_get(url, headers=headers)


def get_project_cluster_resource(access_token):
    """获取所有项目 & 集群信息，异常情况 raise_exception"""
    resp = _get_project_cluster_resource(access_token)
    if resp.get('code') != ErrorCode.NoError:
        raise error_codes.APIError(resp.get('message'))
    return resp.get('data') or []


def update_master(access_token, project_id, cluster_id, data):
    """更新master信息"""
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/masters/"
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token})}
    return http_put(url, json=data, headers=headers)


def delete_namespace(access_token, project_id, cluster_id, ns_id):
    url = f"{BCS_CC_API_PRE_URL}/projects/{project_id}/clusters/{cluster_id}/namespaces/{ns_id}/"
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token})}
    return http_delete(url, headers=headers)


def get_cluster_versions(access_token, ver_id="", env="", kind=""):
    url = f"{BCS_CC_API_PRE_URL}/v1/all/clusters/version_config/"
    headers = {"X-BKAPI-AUTHORIZATION": json.dumps({"access_token": access_token})}
    params = {"ver_id": ver_id, "environment": env, "kind": kind}
    return http_get(url, params=params, headers=headers)


def get_all_cluster_host_ips(access_token):
    data = get_all_cluster_hosts(access_token, exclude_status_list=[CommonStatus.Removed])
    return [info["inner_ip"] for info in data]


# new client module start


class PaaSCCConfig:
    """PaaSCC 系统配置对象，为 Client 提供地址等信息"""

    def __init__(self, host: str):
        self.host = host

        # PaaSCC 系统接口地址
        self.get_cluster_url = f"{host}/projects/{{project_id}}/clusters/{{cluster_id}}"
        self.get_cluster_by_id_url = f"{host}/clusters/{{cluster_id}}/"
        self.get_project_url = f"{host}/projects/{{project_id}}/"
        self.update_cluster_url = f"{host}/projects/{{project_id}}/clusters/{{cluster_id}}/"
        self.delete_cluster_url = f"{host}/projects/{{project_id}}/clusters/{{cluster_id}}/"
        self.list_clusters_url = f"{host}/cluster_list/"
        self.update_node_list_url = f"{host}/projects/{{project_id}}/clusters/{{cluster_id}}/nodes/"
        self.get_cluster_namespace_list_url = f"{host}/projects/{{project_id}}/clusters/{{cluster_id}}/namespaces/"
        self.get_node_list_url = f"{host}/projects/{{project_id}}/nodes/"
        self.list_projects_by_ids = f"{host}/project_list/"
        self.list_namespaces_in_shared_cluster = f"{host}/shared_clusters/{{cluster_id}}/"


@dataclass
class UpdateNodesData:
    inner_ip: str
    status: str


class PaaSCCClient(BkApiClient):
    """访问 PaaSCC 服务的 Client 对象

    :param auth: 包含校验信息的对象
    """

    def __init__(self, auth: ComponentAuth):
        self._config = PaaSCCConfig(host=BCS_CC_API_PRE_URL)
        self._client = BaseHttpClient(auth.to_header_api_auth())

    def get_cluster(self, project_id: str, cluster_id: str) -> Dict:
        """获取集群信息"""
        if cluster_id in [cluster["cluster_id"] for cluster in get_shared_clusters()]:
            url = self._config.get_cluster_by_id_url.format(cluster_id=cluster_id)
            return self._client.request_json('GET', url)
        url = self._config.get_cluster_url.format(project_id=project_id, cluster_id=cluster_id)
        return self._client.request_json('GET', url)

    @response_handler()
    def get_cluster_by_id(self, cluster_id: str) -> Dict:
        """根据集群ID获取集群信息"""
        url = self._config.get_cluster_by_id_url.format(cluster_id=cluster_id)
        return self._client.request_json('GET', url)

    @response_handler()
    def list_clusters(self, cluster_ids: List[str]) -> Dict:
        """根据集群ID列表批量获取集群信息"""
        url = self._config.list_clusters_url
        data = {"cluster_ids": cluster_ids}
        # ugly: search_project 的设置才能使接口生效
        return self._client.request_json('POST', url, params={'search_project': 1}, json=data)

    @parse_response_data()
    def get_project(self, project_id: str) -> Dict:
        """获取项目信息"""
        url = self._config.get_project_url.format(project_id=project_id)
        return self._client.request_json('GET', url)

    @response_handler()
    def update_cluster(self, project_id: str, cluster_id: str, data: Dict) -> Dict:
        """更新集群信息

        :param project_id: 项目32为长度 ID
        :param cluster_id: 集群ID
        :param data: 更新的集群属性，包含status，名称、描述等
        """
        url = self._config.update_cluster_url.format(project_id=project_id, cluster_id=cluster_id)
        return self._client.request_json("PUT", url, json=data)

    @response_handler()
    def delete_cluster(self, project_id: str, cluster_id: str):
        """删除集群

        :param project_id: 项目32为长度 ID
        :param cluster_id: 集群ID
        """
        url = self._config.delete_cluster_url.format(project_id=project_id, cluster_id=cluster_id)
        return self._client.request_json("DELETE", url)

    @response_handler()
    def update_node_list(self, project_id: str, cluster_id: str, nodes: List[UpdateNodesData]) -> List:
        """更新节点信息

        :param project_id: 项目32为长度 ID
        :param cluster_id: 集群ID
        :param nodes: 更新的节点属性，包含IP和状态
        :returns: 返回更新的节点信息
        """
        url = self._config.update_node_list_url.format(project_id=project_id, cluster_id=cluster_id)
        return self._client.request_json("PATCH", url, json={"updates": [asdict(node) for node in nodes]})

    @response_handler()
    def get_cluster_namespace_list(
        self,
        project_id: str,
        cluster_id: str,
        limit=None,
        offset=None,
        with_lb: Union[bool, int] = False,
        desire_all_data: Union[bool, int] = None,
    ) -> Dict[str, Union[int, List[Dict]]]:
        """获取集群下命名空间列表

        :param project_id: 项目ID
        :param cluster_id: 集群ID
        :param limit: 每页的数量
        :param offset: 第几页
        :param with_lb: 是否返回lb，兼容了bool和int型
        :param desire_all_data: 是否拉取集群下全量命名空间，兼容bool和int型
        :returns: 返回集群下的命名空间
        """
        url = self._config.get_cluster_namespace_list_url.format(project_id=project_id, cluster_id=cluster_id)
        req_params = {"desire_all_data": desire_all_data}
        # NOTE: 根据上层调用，希望获取的是集群下的所有命名空间，因此，当desire_all_data为None时，设置为拉取全量
        if desire_all_data is None:
            req_params["desire_all_data"] = 1
        if limit:
            req_params["limit"] = limit
        if offset:
            req_params["offset"] = offset
        if with_lb:
            req_params["with_lb"] = with_lb

        return self._client.request_json("GET", url, params=req_params)

    @response_handler()
    def get_node_list(self, project_id: str, cluster_id: str, params: Dict = None) -> Dict:
        """获取节点列表
        :param project_id: 项目ID
        :param cluster_id: 集群ID
        :param params: 额外的参数
        :returns: 返回节点列表，格式: {"count": 1, "results": [node的详情信息]}
        """
        url = self._config.get_node_list_url.format(project_id=project_id)
        req_params = {"cluster_id": cluster_id}
        if params and isinstance(params, dict):
            req_params.update(params)
        # 默认拉取项目或集群下的所有节点，防止返回分页数据，导致数据不准确
        req_params.setdefault("desire_all_data", 1)
        return self._client.request_json("GET", url, params=req_params)

    @response_handler()
    def list_projects_by_ids(self, project_ids: List[str]) -> Dict:
        """获取项目列表
        :param project_ids: 查询项目的 project_id 列表
        """
        return self._client.request_json("POST", self._config.list_projects_by_ids, json={'project_ids': project_ids})

    @response_handler()
    def list_namespaces_in_shared_cluster(self, cluster_id: str) -> Dict:
        url = self._config.list_namespaces_in_shared_cluster.format(cluster_id=cluster_id)
        # TODO 支持分页查询
        return self._client.request_json("GET", url, params={'offset': 0, 'limit': 1000})
