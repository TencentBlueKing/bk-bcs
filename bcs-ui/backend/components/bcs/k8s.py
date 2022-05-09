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

from django.conf import settings
from django.utils.functional import cached_property
from kubernetes import client
from kubernetes.client.rest import ApiException

from backend.components.bcs import BCSClientBase
from backend.components.cluster_manager import get_shared_clusters
from backend.components.utils import http_delete, http_get, http_patch, http_post
from backend.utils.error_codes import error_codes

from . import resources
from .k8s_client import K8SProxyClient

logger = logging.getLogger(__name__)

REST_PREFIX = "{apigw_host}/rest/clusters"


class K8SClient(BCSClientBase):
    """K8S Client"""

    def __init__(self, access_token, project_id, cluster_id, env):
        super().__init__(access_token, project_id, cluster_id, env)
        self.proxy_client = K8SProxyClient(access_token, project_id, cluster_id, env)

    @cached_property
    def _context(self):
        server_address_path = f'/clusters/{self.cluster_id}'
        return {
            'server_address': f'{self._bcs_server_host}{server_address_path}',
            'server_address_path': server_address_path,
            'identifier': self.cluster_id,
            'user_token': settings.BCS_APIGW_TOKEN,
            'host': f'{self._bcs_server_host}{server_address_path}',
        }

    @cached_property
    def _context_for_shared_cluster(self):
        return {
            "host": f"{settings.BCS_APIGW_DOMAIN[self._bcs_server_stag]}/clusters/{self.cluster_id}",
            "user_token": settings.BCS_APIGW_TOKEN,
        }

    @cached_property
    def context(self):
        # 因为webcosole现在不能切换，所以先保留两个入口
        if self.cluster_id in [cluster["cluster_id"] for cluster in get_shared_clusters()]:
            return self._context_for_shared_cluster
        return self._context

    @cached_property
    def k8s_raw_client(self):
        configure = client.Configuration()
        configure.verify_ssl = False
        configure.host = self.context["host"]
        configure.api_key = {"authorization": f"Bearer {self.context['user_token']}"}
        api_client = client.ApiClient(
            configure,
            header_name='X-BKAPI-AUTHORIZATION',
            header_value=json.dumps({"access_token": self.access_token}),
        )
        return api_client

    @property
    def hpa_client(self):
        api_client = client.AutoscalingV2beta2Api(self.k8s_raw_client)
        return api_client

    @property
    def version(self):
        """获取k8s版本, 使用git_version字段"""
        _client = client.VersionApi(self.k8s_raw_client)
        code = _client.get_code()
        return code.git_version

    @property
    def rest_host(self):
        # NOTE: 切换回BKE
        # return REST_PREFIX.format(apigw_host=self._bke_server_host)
        return REST_PREFIX.format(apigw_host=self.api_host)

    def create_namespace(self, data):
        """创建namespaces"""
        return self.proxy_client.create_namespace(data)

    def delete_namespace(self, name):
        """删除namespaces"""
        return self.proxy_client.delete_namespace(name)

    def get_namespace(self, params=None):
        """获取namesapce，计算数量使用"""
        return self.proxy_client.get_namespace(params)

    def get_pod(self, host_ips=None, field=None, extra=None, params=None):
        """获取pod，获取docker列表使用
        根据label中不定key的过滤，需要使用extra字段
        比如要过滤data.metadata.labels.app为nginx，则需要如下操作
        filter_data = {"data.metadata.labels.app": "nginx"}
        data_json = json.dumps(filter_data)
        base64.b64encode(data_json)
        """
        return self.proxy_client.get_pod(host_ips, field, extra, params)

    def disable_agent(self, ip):
        """停用，禁止被调度"""
        return self.proxy_client.disable_agent(ip)

    def enable_agent(self, ip):
        """启用agent"""
        return self.proxy_client.enable_agent(ip)

    def create_service(self, namespace, data):
        """创建service"""
        return self.proxy_client.create_service(namespace, data)

    def delete_service(self, namespace, name):
        """删除service"""
        return self.proxy_client.delete_service(namespace, name)

    def update_service(self, namespace, name, data):
        """更新service"""
        return self.proxy_client.update_service(namespace, name, data)

    def get_service(self, params):
        return self.proxy_client.get_service(params)

    def get_endpoints(self, params):
        return self.proxy_client.get_endpoints(params)

    def create_configmap(self, namespace, data):
        """创建configmap"""
        return self.proxy_client.create_configmap(namespace, data)

    def delete_configmap(self, namespace, name):
        """删除configmap"""
        return self.proxy_client.delete_configmap(namespace, name)

    def update_configmap(self, namespace, name, data):
        """更新configmap"""
        return self.proxy_client.update_configmap(namespace, name, data)

    def get_configmap(self, params):
        return self.proxy_client.get_configmap(params)

    def create_secret(self, namespace, data):
        """创建secrets"""
        return self.proxy_client.create_secret(namespace, data)

    def delete_secret(self, namespace, name):
        """删除secrets"""
        return self.proxy_client.delete_secret(namespace, name)

    def update_secret(self, namespace, name, data):
        """更新secrets"""
        return self.proxy_client.update_secret(namespace, name, data)

    def get_secret(self, params):
        return self.proxy_client.get_secret(params)

    def create_ingress(self, namespace, data):
        """创建 ingress"""
        return self.proxy_client.create_ingress(namespace, data)

    def delete_ingress(self, namespace, name):
        """删除 ingress"""
        return self.proxy_client.delete_ingress(namespace, name)

    def update_ingress(self, namespace, name, data):
        """更新 ingress"""
        return self.proxy_client.update_ingress(namespace, name, data)

    def get_ingress(self, params):
        return self.proxy_client.get_ingress(params)

    def scale_instance(self, namespace, name, instance_num):
        """扩缩容"""
        return self.proxy_client.scale_instance(namespace, name, instance_num)

    def delete_pod(self, namespace, name):
        """删除pod"""
        return self.proxy_client.delete_pod(namespace, name)

    def get_rs(self, params):
        """查询rs"""
        return self.proxy_client.get_rs(params)

    def create_deployment(self, namespace, data):
        """创建deployment"""
        return self.proxy_client.create_deployment(namespace, data)

    def delete_deployment(self, namespace, deployment_name):
        """删除deployment"""
        return self.proxy_client.delete_deployment(namespace, deployment_name)

    def deep_delete_deployment(self, namespace, name):
        """删除Deployment，级联删除rs&pod"""
        return self.proxy_client.deep_delete_deployment(namespace, name)

    def update_deployment(self, namespace, deployment_name, data):
        """更新deployment
        包含滚动升级和扩缩容
        """
        return self.proxy_client.update_deployment(namespace, deployment_name, data)

    def patch_deployment(self, namespace, name, params):
        """针对deployment的patch操作"""
        return self.proxy_client.patch_deployment(namespace, name, params)

    def get_deployment(self, params):
        """查询deployment"""
        return self.proxy_client.get_deployment(params)

    def get_deployment_with_post(self, data):
        """通过post方法，查询deployment"""
        return self.proxy_client.get_deployment_with_post(data)

    def create_daemonset(self, namespace, data):
        """创建deamonset"""
        return self.proxy_client.create_daemonset(namespace, data)

    def delete_daemonset(self, namespace, name):
        """删除deamonset"""
        return self.proxy_client.delete_daemonset(namespace, name)

    def deep_delete_daemonset(self, namespace, name):
        return self.proxy_client.deep_delete_daemonset(namespace, name)

    def update_daemonset(self, namespace, name, data):
        """更新daemonset"""
        return self.proxy_client.update_daemonset(namespace, name, data)

    def patch_daemonset(self, namespace, name, params):
        """针对daemonset的patch操作"""
        return self.proxy_client.patch_daemonset(namespace, name, params)

    def get_daemonset(self, params):
        """查询daemonset"""
        return self.proxy_client.get_daemonset(params)

    def get_daemonset_with_post(self, data):
        """通过post方法，查询daemonset"""
        return self.proxy_client.get_daemonset_with_post(data)

    def create_statefulset(self, namespace, data):
        """创建statefulset"""
        return self.proxy_client.create_statefulset(namespace, data)

    def delete_statefulset(self, namespace, name):
        """删除statefulset"""
        return self.proxy_client.delete_statefulset(namespace, name)

    def deep_delete_statefulset(self, namespace, name):
        return self.proxy_client.deep_delete_statefulset(namespace, name)

    def update_statefulset(self, namespace, name, data):
        """更新statefulset"""
        return self.proxy_client.update_statefulset(namespace, name, data)

    def patch_statefulset(self, namespace, name, params):
        """针对statefulset的patch操作"""
        return self.proxy_client.patch_statefulset(namespace, name, params)

    def get_statefulset(self, params):
        """查询statefulset"""
        return self.proxy_client.get_statefulset(params)

    def get_statefulset_with_post(self, data):
        """通过post方法，查询statefulset"""
        return self.proxy_client.get_statefulset_with_post(data)

    def create_job(self, namespace, data):
        """创建job"""
        return self.proxy_client.create_job(namespace, data)

    def delete_job(self, namespace, name):
        """删除job"""
        return self.proxy_client.delete_job(namespace, name)

    def deep_delete_job(self, namespace, name):
        return self.proxy_client.deep_delete_job(namespace, name)

    def update_job(self, namespace, name, data):
        """更新job"""
        return self.proxy_client.update_job(namespace, name, data)

    def patch_job(self, namespace, name, params):
        """针对job的patch操作"""
        return self.proxy_client.patch_job(namespace, name, params)

    def get_job(self, params):
        """查询job"""
        return self.proxy_client.get_job(params)

    def get_job_with_post(self, data):
        """通过post方法，查询job"""
        return self.proxy_client.get_job_with_post(data)

    def create_node_labels(self, ip, labels):
        """添加节点标签"""
        return self.proxy_client.create_node_labels(ip, labels)

    def get_node_detail(self, ip):
        """获取节点详细配置"""
        return self.proxy_client.get_node_detail(ip)

    def create_serviceaccounts(self, namespace, data):
        """创建 serviceaccounts"""
        return self.proxy_client.create_serviceaccounts(namespace, data)

    def create_clusterrolebindings(self, namespace, data):
        """创建 ClusterRoleBinding"""
        return self.proxy_client.create_clusterrolebindings(namespace, data)

    def get_events(self, params):
        # storage可以获取比较长的event信息，因此，通过storage查询event
        url = f"{settings.BCS_APIGW_DOMAIN[self._bcs_server_stag]}/bcsapi/v4/storage/events"
        resp = http_get(url, params=params, headers=self.headers)
        return resp

    def get_used_namespace(self):
        """获取已经使用的命名空间名称"""
        params = {"used": 1}
        return self.get_namespace(params=params)

    @property
    def _headers_for_bcs_agent_api(self):
        return {
            "Authorization": f'Bearer {getattr(settings, "BCS_APIGW_TOKEN", "")}',
            "Content-Type": "application/json",
        }

    def query_cluster(self):
        """获取bke_cluster_id, identifier"""
        url = f"{self.rest_host}/bcs/query_by_id/"
        params = {"access_token": self.access_token, "project_id": self.project_id, "cluster_id": self.cluster_id}
        result = http_get(url, params=params, raise_for_status=False, headers=self._headers_for_bcs_agent_api)
        return result

    def register_cluster(self, data=None):
        url = f"{self.rest_host}/bcs/"

        req_data = {"id": self.cluster_id, "project_id": self.project_id, "access_token": self.access_token}
        if data:
            req_data.update(data)

        params = {"access_token": self.access_token}
        # 已经创建的会返回400, code_name: CLUSTER_ALREADY_EXISTS
        result = http_post(
            url, json=req_data, params=params, raise_for_status=False, headers=self._headers_for_bcs_agent_api
        )

        return result

    def get_client_credentials(self, bke_cluster_id: str) -> dict:
        """获取证书, user_token, server_address_path"""
        url = f"{self.rest_host}/{bke_cluster_id}/client_credentials"

        params = {"access_token": self.access_token}

        result = http_get(url, params=params, raise_for_status=False, headers=self._headers_for_bcs_agent_api)
        return result

    def get_register_tokens(self, bke_cluster_id: str) -> dict:
        url = f"{self.rest_host}/{bke_cluster_id}/register_tokens"

        params = {"access_token": self.access_token}

        result = http_get(url, params=params, raise_for_status=False, headers=self._headers_for_bcs_agent_api)
        return result

    def create_register_tokens(self, bke_cluster_id: str) -> dict:
        url = f"{self.rest_host}/{bke_cluster_id}/register_tokens"

        params = {"access_token": self.access_token}

        # 已经创建的会返回500, code_name: CANNOT_CREATE_RTOKEN
        result = http_post(url, params=params, headers=self._headers_for_bcs_agent_api, raise_for_status=False)
        return result

    def list_hpa(self, namespace=None):
        """获取hpa
        - namespace 为空则获取全部
        """
        try:
            # _preload_content 设置为True, 修复kubernetes condition 异常
            if namespace:
                resp = self.hpa_client.list_namespaced_horizontal_pod_autoscaler(namespace, _preload_content=False)
            else:
                resp = self.hpa_client.list_horizontal_pod_autoscaler_for_all_namespaces(_preload_content=False)
            data = json.loads(resp.data)
        except Exception as error:
            logger.exception("list hpa error, %s", error)
            data = {}
        return data

    def get_hpa(self, namespace, name):
        return self.hpa_client.read_namespaced_horizontal_pod_autoscaler(name, namespace)

    def create_hpa(self, namespace, spec):
        """创建HPA"""
        # _preload_content 设置为True, 修复kubernetes condition 异常
        return self.hpa_client.create_namespaced_horizontal_pod_autoscaler(namespace, spec, _preload_content=False)

    def update_hpa(self, namespace, name, spec):
        """修改HPA"""
        return self.hpa_client.patch_namespaced_horizontal_pod_autoscaler(name, namespace, spec)

    def delete_hpa(self, namespace, name):
        try:
            return self.hpa_client.delete_namespaced_horizontal_pod_autoscaler(name, namespace)
        except client.rest.ApiException as error:
            if error.status == 404:
                return

    def apply_hpa(self, namespace, spec):
        """部署HPA"""
        name = spec["metadata"]["name"]
        try:
            self.get_hpa(namespace, name)
        except client.rest.ApiException as error:
            if error.status == 404:
                result = self.create_hpa(namespace, spec)
                logger.info("hpa not found, create a new hpa, %s", result)
                return result
            else:
                logger.error("get hpa error: %s", error)
                raise error
        except Exception as error:
            logger.exception("get hpa exception: %s", error)
        else:
            logger.info("hpa found, create a new hpa, %s, %s", namespace, spec)
            return self.update_hpa(namespace, name, spec)

    def apply_cidr(self, ip_number: int, vpc: str) -> dict:
        url = f"{self.rest_host}/cidr/apply_cidr"
        params = {"access_token": self.access_token}
        data = {"ip_number": ip_number, "cluster": self.cluster_id, "vpc": vpc}
        return http_post(url, params=params, json=data, raise_for_status=False)

    @cached_property
    def _headers_for_service_monitor(self):
        return {
            "Authorization": f"Bearer {self.context['user_token']}",
            "X-BKAPI-AUTHORIZATION": json.dumps({"access_token": self.access_token, "project_id": self.project_id}),
        }

    def list_service_monitor(self, namespace=None):
        host = self.context["host"]
        if namespace:
            url = f"{host}/apis/monitoring.coreos.com/v1/namespaces/{namespace}/servicemonitors"
        else:
            url = f"{host}/apis/monitoring.coreos.com/v1/servicemonitors"

        return http_get(url, headers=self._headers_for_service_monitor, raise_for_status=False)

    def create_service_monitor(self, namespace, spec):
        url = f"{self.context['host']}/apis/monitoring.coreos.com/v1/namespaces/{namespace}/servicemonitors"
        return http_post(url, json=spec, headers=self._headers_for_service_monitor, raise_for_status=False)

    def get_service_monitor(self, namespace, name):
        url = f"{self.context['host']}/apis/monitoring.coreos.com/v1/namespaces/{namespace}/servicemonitors/{name}"
        return http_get(url, headers=self._headers_for_service_monitor, raise_for_status=False)

    def update_service_monitor(self, namespace, name, spec):
        headers = {
            "Authorization": f"Bearer {self.context['user_token']}",
            "Content-Type": "application/merge-patch+json",  # patch的特殊type
            "X-BKAPI-AUTHORIZATION": json.dumps({"access_token": self.access_token, "project_id": self.project_id}),
        }
        url = f"{self.context['host']}/apis/monitoring.coreos.com/v1/namespaces/{namespace}/servicemonitors/{name}"
        return http_patch(url, json=spec, headers=headers, raise_for_status=False)

    def delete_service_monitor(self, namespace, name):
        url = f"{self.context['host']}/apis/monitoring.coreos.com/v1/namespaces/{namespace}/servicemonitors/{name}"
        return http_delete(url, headers=self._headers_for_service_monitor, raise_for_status=False)

    @property
    def sc_client(self):
        api_client = resources.StorageClass(self.k8s_raw_client)
        return api_client

    def list_sc(self):
        return self.sc_client.list_storage_class()

    @property
    def pv_client(self):
        api_client = resources.PersistentVolume(self.k8s_raw_client)
        return api_client

    def list_pv(self):
        return self.pv_client.list_pv()

    @property
    def pvc_client(self):
        api_client = resources.PersistentVolumeClaim(self.k8s_raw_client)
        return api_client

    def list_pvc(self):
        return self.pvc_client.list_pvc()

    def get_prometheus(self, namespace, name):
        url = f"{self.context['host']}/apis/monitoring.coreos.com/v1/namespaces/{namespace}/prometheuses/{name}"
        return http_get(url, headers=self._headers_for_service_monitor, raise_for_status=False)

    def update_prometheus(self, namespace, name, spec):
        headers = {
            "Authorization": f"Bearer {self.context['user_token']}",
            "Content-Type": "application/merge-patch+json",  # patch的特殊type
            "X-BKAPI-AUTHORIZATION": json.dumps({"access_token": self.access_token, "project_id": self.project_id}),
        }
        url = f"{self.context['host']}/apis/monitoring.coreos.com/v1/namespaces/{namespace}/prometheuses/{name}"
        return http_patch(url, json=spec, headers=headers, raise_for_status=False)

    def list_node(self, label_selector=""):
        api_client = resources.Node(self.k8s_raw_client)
        return api_client.list_node(label_selector=label_selector)


class K8SClientWithJWT(K8SClient):
    def __init__(self, access_token, jwt, project_id, cluster_id, env):
        super().__init__(access_token, project_id, cluster_id, env)
        self.jwt = jwt

    @property
    def headers(self):
        _headers = {
            "BCS-ClusterID": self.cluster_id,
            "X-BKAPI-AUTHORIZATION": json.dumps(
                {"access_token": self.access_token, "project_id": self.project_id, "jwt": self.jwt}
            ),
        }
        return _headers
