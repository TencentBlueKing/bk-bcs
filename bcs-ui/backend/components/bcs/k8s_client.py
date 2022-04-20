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
from functools import lru_cache
from typing import Dict

from django.conf import settings
from django.utils.functional import cached_property
from kubernetes import client

from backend.components.bcs import BCSClientBase, resources
from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.client import BcsKubeConfigurationService

logger = logging.getLogger(__name__)


@lru_cache(maxsize=32)
def make_cluster_configuration(access_token: str, project_id: str, cluster_id: str) -> Dict:
    ctx_cluster = CtxCluster.create(
        id=cluster_id,
        project_id=project_id,
        token=access_token,
    )
    return BcsKubeConfigurationService(ctx_cluster).make_configuration()


class K8SAPIClient(BCSClientBase):
    @property
    def rest_host(self):
        return f"{self.api_host}/rest/clusters"

    @property
    def _headers_for_bcs_agent_api(self):
        return {
            "Authorization": f'Bearer {getattr(settings, "BCS_APIGW_TOKEN", "")}',
            "Content-Type": "application/json",
        }

    @cached_property
    def api_client(self):
        configure = make_cluster_configuration(self.access_token, self.project_id, self.cluster_id)
        api_client = client.ApiClient(
            configure,
            header_name='X-BKAPI-AUTHORIZATION',
            header_value=json.dumps({"access_token": self.access_token}),
        )
        return api_client


class K8SProxyClient(K8SAPIClient):
    def create_namespace(self, data):
        namespace = resources.Namespace(self.api_client)
        return namespace.create_namespace(data)

    def delete_namespace(self, name):
        namespace = resources.Namespace(self.api_client)
        return namespace.delete_namespace(name)

    def get_namespace(self, params=None):
        """获取namesapce，计算数量使用"""
        if not params:
            params = {"cluster_id": self.cluster_id}
        if isinstance(params, dict):
            params["cluster_id"] = self.cluster_id
        namespace = resources.Namespace(self.api_client)
        return namespace.get_namespace(params)

    def disable_agent(self, ip):
        node = resources.Node(self.api_client)
        return node.disable_agent(ip)

    def enable_agent(self, ip):
        node = resources.Node(self.api_client)
        return node.enable_agent(ip)

    def create_service(self, namespace, data):
        service = resources.Service(self.api_client)
        return service.create_service(namespace, data)

    def delete_service(self, namespace, name):
        service = resources.Service(self.api_client)
        return service.delete_serivce(namespace, name)

    def update_service(self, namespace, name, data):
        service = resources.Service(self.api_client)
        return service.update_service(namespace, name, data)

    def get_service(self, params):
        service = resources.Service(self.api_client)
        return service.get_service(params)

    def get_endpoints(self, params):
        endpoints = resources.Endpoints(self.api_client)
        return endpoints.get_endpoints(params)

    def create_configmap(self, namespace, data):
        configmap = resources.ConfigMap(self.api_client)
        return configmap.create_configmap(namespace, data)

    def delete_configmap(self, namespace, name):
        configmap = resources.ConfigMap(self.api_client)
        return configmap.delete_configmap(namespace, name)

    def update_configmap(self, namespace, name, data):
        configmap = resources.ConfigMap(self.api_client)
        return configmap.update_configmap(namespace, name, data)

    def get_configmap(self, params):
        configmap = resources.ConfigMap(self.api_client)
        return configmap.get_configmap(params)

    def create_secret(self, namespace, data):
        secret = resources.Secret(self.api_client)
        return secret.create_secret(namespace, data)

    def delete_secret(self, namespace, name):
        secret = resources.Secret(self.api_client)
        return secret.delete_secret(namespace, name)

    def update_secret(self, namespace, name, data):
        secret = resources.Secret(self.api_client)
        return secret.update_secret(namespace, name, data)

    def get_secret(self, params):
        secret = resources.Secret(self.api_client)
        return secret.get_secret(params)

    def create_ingress(self, namespace, data):
        ingress = resources.Ingress(self.api_client)
        return ingress.create_ingress(namespace, data)

    def delete_ingress(self, namespace, name):
        ingress = resources.Ingress(self.api_client)
        return ingress.delete_ingress(namespace, name)

    def update_ingress(self, namespace, name, data):
        ingress = resources.Ingress(self.api_client)
        return ingress.update_ingress(namespace, name, data)

    def get_ingress(self, params):
        ingress = resources.Ingress(self.api_client)
        return ingress.get_ingress(params)

    def delete_pod(self, namespace, name):
        """删除pod"""
        pod = resources.Pod(self.api_client)
        return pod.delete_pod(namespace, name)

    def get_pod(self, host_ips=None, field=None, extra=None, params=None):
        try:
            pod = resources.Pod(self.api_client)
            return pod.get_pod(host_ips, field, extra, params)
        except Exception:
            return {"code": 0, "data": []}

    def scale_instance(self, namespace, name, instance_num):
        pass

    def get_rs(self, params):
        """查询rs"""
        replicaset = resources.ReplicaSet(self.api_client)
        return replicaset.get_replicaset(params)

    def create_deployment(self, namespace, data):
        deployment = resources.Deployment(self.api_client)
        return deployment.create_deployment(namespace, data)

    def delete_deployment(self, namespace, deployment_name):
        deployment = resources.Deployment(self.api_client)
        return deployment.delete_deployment(namespace, deployment_name)

    def deep_delete_deployment(self, namespace, name):
        deployment = resources.Deployment(self.api_client)
        return deployment.delete_deployment(namespace, name)

    def update_deployment(self, namespace, deployment_name, data):
        deployment = resources.Deployment(self.api_client)
        return deployment.update_deployment(namespace, deployment_name, data)

    def patch_deployment(self, namespace, name, params):
        deployment = resources.Deployment(self.api_client)
        return deployment.patch_deployment(namespace, name, params)

    def get_deployment(self, params):
        deployment = resources.Deployment(self.api_client)
        return deployment.get_deployment(params)

    def get_deployment_with_post(self, data):
        deployment = resources.Deployment(self.api_client)
        return deployment.get_deployment(data)

    def create_daemonset(self, namespace, data):
        daemonset = resources.DaemonSet(self.api_client)
        return daemonset.create_daemonset(namespace, data)

    def delete_daemonset(self, namespace, name):
        daemonset = resources.DaemonSet(self.api_client)
        return daemonset.delete_daemonset(namespace, name)

    def deep_delete_daemonset(self, namespace, name):
        daemonset = resources.DaemonSet(self.api_client)
        return daemonset.delete_daemonset(namespace, name)

    def update_daemonset(self, namespace, name, data):
        daemonset = resources.DaemonSet(self.api_client)
        return daemonset.update_daemonset(namespace, name, data)

    def patch_daemonset(self, namespace, name, params):
        daemonset = resources.DaemonSet(self.api_client)
        return daemonset.patch_daemonset(namespace, name, params)

    def get_daemonset(self, params):
        daemonset = resources.DaemonSet(self.api_client)
        return daemonset.get_daemonset(params)

    def get_daemonset_with_post(self, data):
        daemonset = resources.DaemonSet(self.api_client)
        return daemonset.get_daemonset(data)

    def create_statefulset(self, namespace, data):
        statefulset = resources.StatefulSet(self.api_client)
        return statefulset.create_statefulset(namespace, data)

    def delete_statefulset(self, namespace, name):
        statefulset = resources.StatefulSet(self.api_client)
        return statefulset.delete_statefulset(namespace, name)

    def deep_delete_statefulset(self, namespace, name):
        statefulset = resources.StatefulSet(self.api_client)
        return statefulset.delete_statefulset(namespace, name)

    def update_statefulset(self, namespace, name, data):
        statefulset = resources.StatefulSet(self.api_client)
        return statefulset.update_statefulset(namespace, name, data)

    def patch_statefulset(self, namespace, name, params):
        statefulset = resources.StatefulSet(self.api_client)
        return statefulset.patch_statefulset(namespace, name, params)

    def get_statefulset(self, params):
        statefulset = resources.StatefulSet(self.api_client)
        return statefulset.get_statefulset(params)

    def get_statefulset_with_post(self, data):
        statefulset = resources.StatefulSet(self.api_client)
        return statefulset.get_statefulset(data)

    def create_job(self, namespace, data):
        job = resources.Job(self.api_client)
        return job.create_job(namespace, data)

    def delete_job(self, namespace, name):
        job = resources.Job(self.api_client)
        return job.delete_job(namespace, name)

    def deep_delete_job(self, namespace, name):
        job = resources.Job(self.api_client)
        return job.delete_job(namespace, name)

    def update_job(self, namespace, name, data):
        job = resources.Job(self.api_client)
        return job.update_job(namespace, name, data)

    def patch_job(self, namespace, name, params):
        job = resources.Job(self.api_client)
        return job.patch_job(namespace, name, params)

    def get_job(self, params):
        job = resources.Job(self.api_client)
        return job.get_job(params)

    def get_job_with_post(self, data):
        job = resources.Job(self.api_client)
        return job.get_job(data)

    def create_node_labels(self, ip, labels):
        node = resources.Node(self.api_client)
        return node.create_node_labels(ip, labels)

    def get_node_detail(self, ip):
        node = resources.Node(self.api_client)
        return node.get_node_detail(ip)

    def create_serviceaccounts(self, namespace, data):
        pass

    def create_clusterrolebindings(self, namespace, data):
        pass

    def get_events(self, params):
        event = resources.Event(self.api_client)
        return event.get_events(params)
