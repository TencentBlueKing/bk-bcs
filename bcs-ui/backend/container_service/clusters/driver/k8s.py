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

from backend.components.bcs.k8s import K8SClient
from backend.container_service.clusters import constants as cluster_constants
from backend.container_service.clusters.constants import (
    K8S_SKIP_NS_LIST,
    DockerStatusDefaultOrder,
    DockerStatusOrdering,
)
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)


class K8SDriver:
    def __init__(self, request, project_id, cluster_id):
        self.request = request
        self.project_id = project_id
        self.cluster_id = cluster_id
        self.client = K8SClient(self.request.user.token.access_token, self.project_id, self.cluster_id, None)

    def host_container_map(self, resp):
        host_container_map = {}
        for info in resp.get('data') or []:
            if info.get('namespace') in K8S_SKIP_NS_LIST:
                continue
            host_ip = info.get('data', {}).get('status', {}).get('hostIP')
            container_count = len(info['data']['status']['containerStatuses'])
            if host_ip in host_container_map:
                host_container_map[host_ip] += container_count
            else:
                host_container_map[host_ip] = container_count
        return host_container_map

    def get_unit_info(self, inner_ip_list, field, raise_exception=True):
        """get unit info by inner_ip and field"""
        resp = self.client.get_pod(host_ips=inner_ip_list, field=field)
        if resp.get('code') != ErrorCode.NoError:
            logger.error("request pod api error, %s", resp.get("message"))
            if raise_exception:
                raise error_codes.APIError(resp.get('message'))

        return resp

    def get_host_container_count(self, host_ips):
        field_list = ['data.status.containerStatuses.containerID', 'data.status.hostIP', 'namespace']
        resp = self.get_unit_info(host_ips, ','.join(field_list))
        # compose the host container data
        return self.host_container_map(resp)

    def flatten_container_info(self, inner_ip):
        def container_info(pods, inner_ip):
            for p in pods:
                if p.get('namespace') in K8S_SKIP_NS_LIST:
                    continue
                container_status = p["data"]["status"]["containerStatuses"]
                container_spec = {info["name"]: info["image"] for info in p["data"]["spec"]["containers"]}
                for d in container_status:
                    last_status = d.get("state") or d.get("lastState")
                    if not last_status:
                        continue
                    status = list(last_status.keys())
                    status = status[0] if status else None
                    c = {
                        "container_id": d.get("containerID", "").split("docker://")[-1],
                        "status": status,
                        "name": d["name"],
                        "image": container_spec.get(d["name"]),
                    }
                    yield c

        pods = self.get_unit_info([inner_ip], field='data,namespace').get('data') or []
        containers = container_info(pods, inner_ip)

        containers = sorted(
            [i for i in containers],
            key=lambda x: DockerStatusOrdering.get(x['status'], DockerStatusDefaultOrder),
        )
        return containers

    def disable_node(self, ip):
        """stop scheduler"""
        node_resp = self.client.disable_agent(ip)
        if node_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(node_resp.get('message'))

    def enable_node(self, ip):
        node_resp = self.client.enable_agent(ip)
        if node_resp.get('code') != ErrorCode.NoError:
            raise error_codes.APIError(node_resp.get('message'))

    def get_host_unit_list(self, ip, raise_exception=True):
        """get exist pods on the node"""
        unit_list = []
        fields = 'namespace,resourceName,clusterId'
        resp = self.get_unit_info([ip], fields, raise_exception=raise_exception)
        for i in resp.get('data') or []:
            namespace = i.get('namespace')
            if namespace in cluster_constants.K8S_SKIP_NS_LIST:
                continue
            unit_list.append({'namespace': namespace, 'pod_name': i.get('resourceName')})
        return unit_list

    def reschedule_pod(self, pod_info, raise_exception=True):
        resp = self.client.delete_pod(pod_info['namespace'], pod_info['pod_name'])
        if resp.get('code') != ErrorCode.NoError:
            logger.error("request delete pod api error, %s", resp.get("message"))
            if raise_exception:
                raise error_codes.APIError(resp.get('message'))

        return resp

    def reschedule_host_pods(self, ip, raise_exception=True):
        unit_list = self.get_host_unit_list(ip, raise_exception=raise_exception)
        for info in unit_list:
            self.reschedule_pod(info, raise_exception=raise_exception)
        return
