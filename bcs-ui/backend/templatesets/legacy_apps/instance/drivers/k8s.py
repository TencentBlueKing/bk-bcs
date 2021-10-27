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

K8S底层调用
"""
import logging

from django.utils.translation import ugettext_lazy as _

from backend.components.bcs.k8s import K8SClient
from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.constants import DEFAULT_HPA_API_VERSION
from backend.resources.hpa.client import HPA
from backend.utils.exceptions import ComponentError, ConfigError, Rollback

from .base import SchedulerBase

logger = logging.getLogger(__name__)


class Scheduler(SchedulerBase):
    """"""

    def handler_k8sdeployment(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        result = client.create_deployment(ns, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sDeployment失败，{}").format(result.get('message')))

    def handler_update_k8sdeployment(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        deployment_name = spec['metadata']['name']
        result = client.update_deployment(ns, deployment_name, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("更新K8sDeployment失败，{}").format(result.get('message')))

    def rollback_k8sdeployment(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        deployment_name = spec['metadata']['name']
        client.deep_delete_deployment(ns, deployment_name)

    def handler_k8sservice(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        result = client.create_service(ns, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sService失败，{}").format(result.get('message')))

    def handler_update_k8sservice(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        deployment_name = spec['metadata']['name']
        result = client.update_service(ns, deployment_name, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("更新K8sService失败，{}").format(result.get('message')))

    def rollback_k8sservice(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        name = spec['metadata']['name']
        result = client.delete_service(ns, name)
        if result.get('code') != 0:
            raise ComponentError(result.get('message', ''))

    def handler_k8sconfigmap(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        result = client.create_configmap(ns, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sConfigMap失败，{}").format(result.get('message')))

    def handler_update_k8sconfigmap(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        name = spec['metadata']['name']
        result = client.update_configmap(ns, name, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("更新K8sConfigMap失败，{}").format(result.get('message')))

    def rollback_k8sconfigmap(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        name = spec['metadata']['name']
        result = client.delete_configmap(ns, name)
        if result.get('code') != 0:
            raise ComponentError(result.get('message', ''))

    def handler_k8ssecret(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        result = client.create_secret(ns, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sSecret失败，{}").format(result.get('message')))

    def handler_update_k8ssecret(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        deployment_name = spec['metadata']['name']
        result = client.update_secret(ns, deployment_name, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sSecret失败，{}").format(result.get('message')))

    def rollback_k8ssecret(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        name = spec['metadata']['name']
        result = client.delete_secret(ns, name)
        if result.get('code') != 0:
            raise ComponentError(result.get('message', ''))

    def handler_k8sdaemonset(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        result = client.create_daemonset(ns, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sDaemonSet失败，{}").format(result.get('message')))

    def handler_update_k8sdaemonset(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        deployment_name = spec['metadata']['name']
        result = client.update_daemonset(ns, deployment_name, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("更新K8sDaemonSet失败，{}").format(result.get('message')))

    def rollback_k8sdaemonset(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        name = spec['metadata']['name']
        client.deep_delete_daemonset(ns, name)

    def handler_k8sjob(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        result = client.create_job(ns, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sJob失败，{}").format(result.get('message')))

    def handler_update_k8sjob(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        deployment_name = spec['metadata']['name']
        result = client.update_job(ns, deployment_name, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("更新K8sJob失败，{}").format(result.get('message')))

    def rollback_k8sjob(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        name = spec['metadata']['name']
        client.deep_delete_job(ns, name)

    def handler_k8sstatefulset(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        result = client.create_statefulset(ns, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sStatefulSet失败，{}").format(result.get('message')))

    def handler_update_k8sstatefulset(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        deployment_name = spec['metadata']['name']
        result = client.update_statefulset(ns, deployment_name, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("更新K8sStatefulSet失败，{}").format(result.get('message')))

    def rollback_k8sstatefulset(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        name = spec['metadata']['name']
        client.deep_delete_statefulset(ns, name)

    def handler_k8singress(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        result = client.create_ingress(ns, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("创建K8sIngress失败，{}").format(result.get('message')))

    def handler_update_k8singress(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        deployment_name = spec['metadata']['name']
        result = client.update_ingress(ns, deployment_name, spec)
        if result.get('code') != 0:
            if result.get('code') == 4001:
                raise ConfigError(_("配置文件格式错误:{}").format(result.get('message')))
            raise ComponentError(_("更新K8sIngress失败，{}").format(result.get('message')))

    def rollback_k8singress(self, ns, cluster_id, spec):
        client = K8SClient(self.access_token, self.project_id, cluster_id, env=None)
        name = spec['metadata']['name']
        result = client.delete_ingress(ns, name)
        if result.get('code') != 0:
            raise ComponentError(result.get('message', ''))

    def handler_k8shpa(self, ns, cluster_id, spec):
        """下发HPA配置"""
        ctx_cluster = CtxCluster.create(token=self.access_token, project_id=self.project_id, id=cluster_id)
        client = HPA(ctx_cluster)

        name = spec["metadata"]["name"]
        spec['apiVersion'] = DEFAULT_HPA_API_VERSION

        try:
            result = client.update_or_create(spec, name, ns)
            logger.debug("deploy hpa success, %s", result)
        except Exception as error:
            logger.exception('deploy hpa error, %s', error)
            raise Rollback({})

    def rollback_k8shpa(self, ns, cluster_id, spec):
        """回滚HPA"""
