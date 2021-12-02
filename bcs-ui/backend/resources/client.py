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
from typing import Dict

from django.conf import settings
from kubernetes import client
from kubernetes.client.configuration import Configuration

from backend.components.bcs import k8s
from backend.container_service.clusters.base import CtxCluster
from backend.utils import exceptions
from backend.utils.errcodes import ErrorCode

logger = logging.getLogger(__name__)


class K8SClient:
    def __init__(self, access_token, project_id, cluster_id):
        self.access_token = access_token
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.client = k8s.K8SClient(access_token, project_id, cluster_id, None)


class BcsAPIEnvironmentQuerier:
    """查询 BCS 集群的 API 网关环境名称"""

    def __init__(self, cluster: CtxCluster):
        self.cluster = cluster

    def do(self) -> str:
        """通过 PaaS-CC 服务查询查询集群环境名称，然后找到其对应的 API 网关环境名称

        :raises ComponentError: 从 PaaS-CC 返回了错误响应
        """
        # Cache
        if hasattr(self, '_api_env_name'):
            return self._api_env_name

        cluster_resp = self.cluster.comps.paas_cc.get_cluster(self.cluster.project_id, self.cluster.id)
        # TODO: 封装异常，不使用模糊的 ComponentError
        if cluster_resp.get('code') != ErrorCode.NoError:
            raise exceptions.ComponentError(cluster_resp.get('message'))

        environment = cluster_resp['data']['environment']
        self._api_env_name = settings.BCS_API_ENV[environment]
        return self._api_env_name


class BcsKubeConfigurationService:
    """生成用于连接 Kubernetes 集群的配置对象 Configuration
    通过查询 BCS 服务，获取集群 apiserver 地址与 token 等信息

    :param cluster: 集群对象
    """

    def __init__(self, cluster: CtxCluster):
        self.cluster = cluster
        self.env_querier = BcsAPIEnvironmentQuerier(cluster)

    def make_configuration(self) -> Configuration:
        """生成 Kubernetes SDK 所需的 Configuration 对象"""
        env_name = self.env_querier.do()

        config = client.Configuration()
        config.verify_ssl = False
        # Get credentials
        credentials = self.get_client_credentials(env_name)
        config.host = credentials['host']
        config.api_key = {"authorization": f"Bearer {credentials['user_token']}"}
        return config

    def get_client_credentials(self, env_name: str) -> Dict[str, str]:
        """获取通过 bcs api gateway 网关访问集群 apiserver 所需的鉴权信息

        :param env_name: 集群所属环境，包含正式环境和测试环境
        """
        host = f'{self._get_apiservers_host(env_name)}/clusters/{self.cluster.id}'
        return {'host': host, 'user_token': settings.BCS_API_GW_AUTH_TOKEN}

    @staticmethod
    def _get_apiservers_host(api_env_name: str) -> str:
        """获取 Kubernetes 集群 apiserver 基础地址"""
        return settings.BCS_SERVER_HOST[api_env_name]
