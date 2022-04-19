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

from rest_framework import response, viewsets
from rest_framework.renderers import BrowsableAPIRenderer

from backend.utils.basic import getitems
from backend.utils.error_codes import error_codes
from backend.utils.renderers import BKAPIRenderer
from backend.web_console.api import exec_command

from .. import base_perm_views, constants, drivers, utils
from ..base_views import BaseInstanceView
from ..common_views import serializers as common_serializers

logger = logging.getLogger(__name__)


class K8SContainerInfo(BaseInstanceView, viewsets.ViewSet):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_params_from_no_tmpl_type(self, request, project_id):
        """获取非模板集"""
        slz = common_serializers.BaseNotTemplateInstanceParamsSLZ(data=request.query_params)
        slz.is_valid(raise_exception=True)
        params = dict(slz.validated_data)

        # get cluster id and namespace id
        ns_name = params['namespace']
        access_token = request.user.token.access_token
        ns_name_map = utils.get_namespace_name_map(access_token, project_id)
        ns_info_by_name = ns_name_map.get(ns_name)
        if not ns_info_by_name:
            raise error_codes.CheckFailed(f'namespace({ns_name}) not found')

        ns_id = ns_info_by_name['id']
        # 校验权限
        base_perm_views.InstancePerm.can_use_instance(request, project_id, ns_id, source_type=None)
        return params

    def get_params_from_tmpl_type(self, request, project_id, instance_id):
        inst_info = self.get_instance_info(instance_id)
        inst_config = json.loads(inst_info.config)
        metadata = inst_config.get('metadata') or {}
        params = {
            'name': metadata['name'],
            'namespace': metadata['namespace'],
            'category': inst_info.category,
        }
        labels = metadata.get('labels') or {}
        # 校验权限
        base_perm_views.InstancePerm.can_use_instance(
            request, project_id, inst_info.namespace, labels.get(constants.LABEL_TEMPLATE_ID)
        )
        params['cluster_id'] = labels.get(constants.LABEL_CLUSTER_ID)
        return params

    def compose_container_data(self, container_id, container_spec, container_status, spec, status, labels):
        # TODO:是否有更好方式处理下面数据
        return {
            'volumes': [
                {
                    'hostPath': info.get('name', ''),
                    'mountPath': info.get('mountPath', ''),
                    'readOnly': info.get('readOnly', ''),
                }
                for info in container_spec.get('volumeMounts') or []
            ],
            'ports': container_spec.get('ports') or [],
            'command': {
                'command': container_spec.get('command', ''),
                'args': ' '.join(container_spec.get('args', '')),
            },
            'network_mode': spec.get('dnsPolicy', ''),
            'labels': [{'key': key, 'val': val} for key, val in labels.items()],
            'resources': container_spec.get('resources', {}),
            'health_check': container_spec.get('livenessProbe', {}),
            'readiness_check': container_spec.get('readinessProbe', {}),
            'host_ip': status.get('hostIP', ''),
            'container_ip': status.get('podIP', ''),
            'host_name': spec.get('nodeName', ''),
            'container_name': container_status.get('name', ''),
            'container_id': container_id,
            'image': utils.image_handler(container_status.get('image', '')),
            # 先保留env, 因为前端还没有适配
            'env_args': container_spec.get('env', []),
        }

    def compose_data(self, pod_info, container_id):
        """根据container id处理数据"""
        ret_data = {}
        if not (pod_info and container_id):
            return ret_data
        # pod_info type is list, and only one
        pod_info = pod_info[0]
        container_statuses = getitems(pod_info, ['data', 'status', 'containerStatuses'], default=[])
        status = getitems(pod_info, ['data', 'status'], default={})
        spec = getitems(pod_info, ['data', 'spec'], default={})
        labels = getitems(pod_info, ['data', 'metadata', 'labels'], default={})
        # 开始数据匹配
        container_data = {info.get("name"): info for info in spec.get('containers') or []}
        for info in container_statuses:
            curr_container_id = info.get('containerID')
            # container_id format: "docker://ad7034695ae7f911babf771447b65e1cb97f3f1987ad214c22decba0dd3fa121"
            if container_id not in curr_container_id:
                continue
            container_spec = container_data.get(info.get('name'), {})
            ret_data = self.compose_container_data(container_id, container_spec, info, spec, status, labels)
            break

        ret_data['namespace'] = getitems(pod_info, ['data', 'metadata', 'namespace'])
        return ret_data

    def compose_pod_params(self, request, project_id, instance_id):
        if str(instance_id) == constants.NOT_TMPL_IDENTIFICATION:
            return self.get_params_from_no_tmpl_type(request, project_id)
        return self.get_params_from_tmpl_type(request, project_id, instance_id)

    def get_container_id(self, request):
        slz = common_serializers.K8SContainerInfoSLZ(data=request.data)
        slz.is_valid(raise_exception=True)
        return slz.validated_data['container_id']

    def container_info(self, request, project_id, instance_id, name):
        container_id = self.get_container_id(request)

        # if instance_id is zero, the source of instance is not template
        params = self.compose_pod_params(request, project_id, instance_id)

        params.update({'unit_name': name, 'field': ['data']})
        k8s_driver = drivers.BCSDriver(request, project_id, params['cluster_id'])
        pod_info = k8s_driver.get_unit_info_by_name(params=params)
        ret_data = self.compose_data(pod_info, container_id)

        return response.Response(ret_data)

    def env_info(self, request, project_id, instance_id, name):
        container_id = self.get_container_id(request)

        params = self.compose_pod_params(request, project_id, instance_id)
        env_resp = exec_command(request.user.token.access_token, project_id, params['cluster_id'], container_id, 'env')

        # parse and compose the return data
        try:
            # docker env format: key=val
            data = [dict(zip(['name', 'value'], info.split('=', 1))) for info in env_resp.splitlines() if info]
        except Exception as err:
            # not raise error, record log
            logger.error('parse the env data error, detial: %s', err)
            data = []

        return response.Response(data)
