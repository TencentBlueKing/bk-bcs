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
from backend.templatesets.legacy_apps.instance.constants import APPLICATION_ID_SEPARATOR, LOG_CONFIG_MAP_SUFFIX


class ResourceMixin:
    @classmethod
    def get_resources_config(cls, resource_id_list, is_simple):
        config_list = []
        robj_qsets = cls.objects.filter(id__in=resource_id_list)
        for robj in robj_qsets:
            config_list.append(robj.get_res_config(is_simple))
        return config_list

    @classmethod
    def get_resources_info(cls, resource_id_list):
        return []


class ConfigMapAndSecretBase(ResourceMixin):
    @classmethod
    def get_resources_info(cls, resource_id_list):
        resource_data = []
        robj_qsets = cls.objects.filter(id__in=resource_id_list)
        for robj in robj_qsets:
            resource_data.append({"name": robj.name, "keys": robj.get_data_keys()})
        return resource_data


class ConfigMapAndSecretMixin(ConfigMapAndSecretBase):
    def get_data_keys(self):
        config = self.get_config()
        if config:
            return config.get("data", {}).keys()
        return []


class MConfigMapAndSecretMixin(ConfigMapAndSecretBase):
    def get_data_keys(self):
        config = self.get_config()
        if config:
            return config.get("datas", {}).keys()
        return []


class PodMixin(ResourceMixin):
    def get_containers(self):
        config = self.get_config()
        containers = config.get("spec", {}).get("template", {}).get("spec", {}).get("containers", [])
        return containers

    def get_extra_configmaps_and_metrics(self, resource_name):
        """
        configmap used for nonstandard logbeat
        """
        configmaps, metrics = [], []
        for container in self.get_containers():
            container_name = container.get("name")
            log_path_list = container.get("logPathList")
            if log_path_list:
                configmaps.append(
                    {
                        "id": f"{self.id}{APPLICATION_ID_SEPARATOR}{container_name}"
                        f"{APPLICATION_ID_SEPARATOR}{resource_name}",
                        "name": f"{self.name}-{container_name}{LOG_CONFIG_MAP_SUFFIX}",
                    }
                )

        web_cache = self.get_config().get("webCache") or {}
        is_metric = web_cache.get("isMetric")
        metric_id_list = web_cache.get("metricIdList")
        if not is_metric or not metric_id_list:
            return configmaps, metrics

        return configmaps, metrics
