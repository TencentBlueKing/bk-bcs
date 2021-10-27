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
import re

from backend.components import paas_cc
from backend.uniapps.network.constants import K8S_NGINX_INGRESS_CONTROLLER_CHART_VALUES
from backend.utils.basic import getitems

from .constants import BACKEND_IMAGE_PATH, CONTROLLER_IMAGE_PATH

try:
    from backend.container_service.observability.datalog.utils import get_data_id_by_project_id
except ImportError:
    from backend.container_service.observability.datalog_ce.utils import get_data_id_by_project_id

logger = logging.getLogger(__name__)
DEFAULT_HTTP_PORT = "80"
DEFAULT_HTTPS_PORT = "443"


def render_helm_values(access_token, project_id, cluster_id, protocol_type, replica_count, namespace):
    """渲染helm values配置文件"""
    # check protocol exist
    http_enabled = "false"
    https_enabled = "false"
    http_port = DEFAULT_HTTP_PORT
    https_port = DEFAULT_HTTPS_PORT
    protocol_type_list = re.findall(r"[^,; ]+", protocol_type)
    for info in protocol_type_list:
        protocol_port = info.split(":")
        if "http" in protocol_port:
            http_enabled = "true"
            http_port = protocol_port[-1] if len(protocol_port) == 2 and protocol_port[-1] else DEFAULT_HTTP_PORT
        if "https" in protocol_port:
            https_enabled = "true"
            https_port = protocol_port[-1] if len(protocol_port) == 2 and protocol_port[-1] else DEFAULT_HTTPS_PORT
    jfrog_domain = paas_cc.get_jfrog_domain(access_token=access_token, project_id=project_id, cluster_id=cluster_id)
    # render
    template = K8S_NGINX_INGRESS_CONTROLLER_CHART_VALUES
    template = template.replace("__REPO_ADDR__", jfrog_domain)
    template = template.replace("__CONTROLLER_IMAGE_PATH__", CONTROLLER_IMAGE_PATH)
    # TODO: 先调整为固定版本，后续允许用户在前端选择相应的版本
    template = template.replace("__TAG__", "0.35.0")
    template = template.replace("__CONTROLLER_REPLICA_COUNT__", str(replica_count))
    template = template.replace("__BACKEND_IMAGE_PATH__", BACKEND_IMAGE_PATH)
    template = template.replace("__HTTP_ENABLED__", http_enabled)
    template = template.replace("__HTTP_PORT__", http_port)
    template = template.replace("__HTTPS_ENABLED__", https_enabled)
    template = template.replace("__HTTPS_PORT__", https_port)
    template = template.replace("__NAMESPACE__", namespace)

    return template


def get_svc_access_info(manifest, cluster_id, extended_routes):
    """
    {
        'external': {
            'NodePort': ['node_ip:{node_port}'],
        },
        'internal': {
            'ClusterIP': [':{port} {Protocol}']
        }
    }
    """
    access_info = {'external': {}, 'internal': {}}
    svc_type = getitems(manifest, ['spec', 'type'])
    ports = getitems(manifest, ['spec', 'ports'])

    if not ports:
        return access_info

    if svc_type == 'ClusterIP':
        cluster_ip = getitems(manifest, ['spec', 'clusterIP'])
        if not cluster_ip or cluster_ip == 'None':
            cluster_ip = '--'
        access_info['internal'] = {'ClusterIP': [f"{cluster_ip}:{p['port']} {p['protocol']}" for p in ports]}
    elif svc_type == 'NodePort':
        access_info['external'] = {'NodePort': [f":{p['nodePort']}" for p in ports]}

    return access_info


try:
    from .utils_ext import get_svc_access_info  # noqa
except ImportError as e:
    logger.debug('Load extension failed: %s', e)
