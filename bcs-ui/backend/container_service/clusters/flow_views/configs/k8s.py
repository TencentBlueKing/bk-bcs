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

K8S 获取相关配置
NOTE: 现阶段还没有kube agent的相关配置，需要bowei处理下后面的流程
"""
import json
import logging
import socket
from urllib.parse import urlparse

from django.conf import settings

from backend.container_service.clusters import constants
from backend.container_service.clusters.utils import gen_hostname
from backend.container_service.misc.bke_client import BCSClusterClient
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)

BCS_SERVER_HOST = settings.BCS_SERVER_HOST['prod']


class ClusterConfig(object):
    def __init__(self, base_cluster_config, area_info, cluster_name=""):
        self.k8s_config = base_cluster_config
        self.area_config = json.loads(area_info.get('configuration', '{}'))

    def _split_ip_by_role(self, ip_list):
        """get master/etcd by role
        NOTE: master and etcd are same in the current stage
        """
        return ip_list, ip_list

    def _get_clusters_vars(self, cluster_id, kube_master_list, etcd_list):
        masters, etcdpeers, clusters = {}, {}, {}

        for etcd_ip in etcd_list:
            host_name = gen_hostname(etcd_ip, cluster_id, is_master=True)
            etcdpeers[host_name] = etcd_ip
            clusters[host_name] = etcd_ip

        for master_ip in kube_master_list:
            host_name = gen_hostname(master_ip, cluster_id, is_master=True)
            masters[host_name] = master_ip
            clusters[host_name] = master_ip

        return masters, etcdpeers, clusters

    def _get_common_vars(self, cluster_id, masters, etcdpeers, clusters, cluster_state):
        self.k8s_config['common'].update(
            {
                'cluster_id': cluster_id,
                'etcd_peers': etcdpeers,
                'cluster_masters': masters,
                'clusters': clusters,
                'bk_registry': self.area_config['jfrog_registry'],
                'dns_host': self.area_config['dns_host'],
                'zk_urls': ','.join(self.area_config["zk_hosts"]),
            }
        )

        if cluster_state == constants.ClusterState.BCSNew.value:
            return
        # NOTE: 针对非bcs平台创建集群，配置中common支持websvr，这里websvr是字符串
        web_svr = self.k8s_config.get("websvr")
        if web_svr:
            self.k8s_config["common"]["websvr"] = web_svr[0]

    def _get_node_vars(self, master_legal_host):
        zk_urls = ','.join(self.area_config["zk_hosts"])
        # 为防止key对应内容的变动，单独更新key
        self.k8s_config['kubernetes.node'].update(
            {'legal_hosts': master_legal_host, 'is_kube_master': True, 'standalone_kubelet': True}
        )
        self.k8s_config['kubernetes.master'].update({'legal_hosts': master_legal_host})
        self.k8s_config['docker'].update({'legal_hosts': master_legal_host})
        self.k8s_config['bcs.driver'].update({'legal_hosts': master_legal_host, 'zk_urls': zk_urls})
        self.k8s_config['bcs.datawatch'].update({'legal_hosts': master_legal_host, 'zk_urls': zk_urls})

    def _get_etcd_vars(self, etcd_legal_host):
        self.k8s_config['etcd'].update({'legal_hosts': etcd_legal_host})

    def _add_kube_agent_config(self, cluster_id, params):
        """针对纳管集群，需要在创建集群时，传递kube client组件需要的配置信息"""
        if params.get("cluster_state") == constants.ClusterState.BCSNew.value:
            return
        # get bcs agent info
        bcs_client = BCSClusterClient(
            host=BCS_SERVER_HOST,
            access_token=params["access_token"],
            project_id=params["project_id"],
            cluster_id=cluster_id,
        )
        bcs_cluster_info = bcs_client.get_or_register_bcs_cluster()
        if not bcs_cluster_info.get("result"):
            err_msg = bcs_cluster_info.get("message", "request bcs agent api error")
            raise error_codes.APIError(err_msg)
        bcs_cluster_data = bcs_cluster_info.get("data", {})
        if not bcs_cluster_data:
            raise error_codes.APIError("bcs agent api response is null")

        self.k8s_config["bcs.kube_agent"].update(
            {
                "register_token": bcs_cluster_data["token"],
                "bcs_api_server": BCS_SERVER_HOST,
                "register_cluster_id": bcs_cluster_data["bcs_cluster_id"],
            }
        )

    def get_request_config(self, cluster_id, master_ips, need_nat=True, **kwargs):
        # 获取master和etcd ip列表
        kube_master_list, etcd_list = self._split_ip_by_role(master_ips)
        # 组装name: ip map
        masters, etcdpeers, clusters = self._get_clusters_vars(cluster_id, kube_master_list, etcd_list)
        # 更新组装参数
        self._get_common_vars(cluster_id, masters, etcdpeers, clusters, kwargs.get("cluster_state"))
        master_legal_host, etcd_legal_host = list(masters.keys()), list(etcdpeers.keys())
        self._get_node_vars(master_legal_host)
        self._get_etcd_vars(etcd_legal_host)
        self._add_kube_agent_config(cluster_id, kwargs)

        return self.k8s_config


class NodeConfig(object):
    def __init__(self, snapshot_config, op_type):
        self.k8s_config = snapshot_config
        self.op_type = op_type

    def _get_clusters_vars(self, cluster_id, node_ip_list, master_ip_list):
        clusters, masters, node_legals = {}, {}, {}

        for node_ip in node_ip_list:
            host_name = gen_hostname(node_ip, cluster_id, is_master=False)
            node_legals[host_name] = node_ip
            clusters[host_name] = node_ip

        for master_ip in master_ip_list:
            host_name = gen_hostname(master_ip, cluster_id, is_master=True)
            clusters[host_name] = master_ip
            masters[host_name] = master_ip

        return clusters, masters, node_legals

    def _get_common_vars(self, cluster_id, masters, clusters):
        self.k8s_config['common'].update({'cluster_id': cluster_id, 'cluster_masters': masters, 'clusters': clusters})

    def _get_network_vars(self, node_legals, kubeapps_master_legal_host):
        self.k8s_config['network_plugin'].update({'legal_hosts': list(node_legals.keys()), 'plugin_type': 'flannel'})
        if self.op_type == constants.OpType.ADD_NODE.value:
            self.k8s_config['network_plugin']['legal_hosts'] = kubeapps_master_legal_host
            self.k8s_config['kubeapps.network_plugin'] = self.k8s_config['network_plugin']
            self.k8s_config['dns'].update({'legal_hosts': kubeapps_master_legal_host, 'dns_type': 'kubedns'})

    def _get_node_vars(self, node_legals, kubeapps_master_legal_host, access_token, project_id, cluster_id):
        legal_hosts = list(node_legals.keys())
        self.k8s_config['kubernetes.node'].update(
            {'legal_hosts': legal_hosts, 'is_kube_master': False, 'standalone_kubelet': False}
        )
        self.k8s_config['docker'].update({'legal_hosts': legal_hosts})
        self.k8s_config['agent.cadvisorbeat'].update({'legal_hosts': legal_hosts})
        self.k8s_config['agent.logbeat'].update({'legal_hosts': legal_hosts})
        if self.op_type == constants.OpType.ADD_NODE.value:
            # get bcs agent info
            bcs_client = BCSClusterClient(
                host=BCS_SERVER_HOST, access_token=access_token, project_id=project_id, cluster_id=cluster_id
            )
            bcs_cluster_info = bcs_client.get_or_register_bcs_cluster()
            if not bcs_cluster_info.get('result'):
                err_msg = bcs_cluster_info.get('message', 'request bcs agent api error')
                raise error_codes.APIError(err_msg)
            bcs_cluster_data = bcs_cluster_info.get('data', {})
            if not bcs_cluster_data:
                raise error_codes.APIError('bcs agent api response is null')

            self.k8s_config['bcs.kube_agent'].update(
                {
                    'legal_hosts': kubeapps_master_legal_host,
                    'register_token': bcs_cluster_data['token'],
                    'bcs_api_server': BCS_SERVER_HOST,
                    'register_cluster_id': bcs_cluster_data['bcs_cluster_id'],
                }
            )
            self.k8s_config['kubeapps.kube_agent'].update({'legal_hosts': kubeapps_master_legal_host})
        # 根据操作
        if self.op_type == constants.OpType.DELETE_NODE.value:
            self.k8s_config['kubeapps.node'].update({'legal_hosts': kubeapps_master_legal_host, 'nodes': node_legals})

    def _get_secrets_vars(self):
        self.k8s_config['secrets.kubernetes'].update({'legal_hosts': []})

    def _get_prometheus(self, kubeapps_master_legal_host):
        """获取prometheus"""
        self.k8s_config.update(
            {
                'kubeapps.prometheus': {'legal_hosts': kubeapps_master_legal_host},
                'agent.prometheus': {'legal_hosts': kubeapps_master_legal_host},
            }
        )

    def get_request_config(self, access_token, project_id, cluster_id, master_ip_list, ip_list):
        # 获取master、node
        clusters, masters, node_legals = self._get_clusters_vars(cluster_id, ip_list, master_ip_list)
        kubeapps_master_legal_host = [
            gen_hostname(master_ip_list[0], cluster_id, True),
        ]
        self._get_common_vars(cluster_id, masters, clusters)
        self._get_node_vars(node_legals, kubeapps_master_legal_host, access_token, project_id, cluster_id)
        self._get_network_vars(node_legals, kubeapps_master_legal_host)
        self._get_secrets_vars()
        self._get_prometheus(kubeapps_master_legal_host)

        return self.k8s_config
