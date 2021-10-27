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

from django.utils.translation import ugettext_lazy as _

from backend.templatesets.legacy_apps.configuration.models import Template
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance

from ..base_views import BaseAPI, error_codes
from ..constants import CATEGORY_MAP
from ..utils import APIResponse, cluster_env, exclude_records


class BaseFilter(BaseAPI):
    def get_muster(self, project_id):
        """获取模板集"""
        all_musters = Template.objects.filter(project_id=project_id, is_deleted=False)
        muster_id_name_map = {info.id: info.name for info in all_musters}
        return muster_id_name_map

    def get_version_instances(self, muster_id_list):
        """根据模板集获取相应的版本实例"""
        version_inst = VersionInstance.objects.filter(template_id__in=muster_id_list, is_deleted=False)
        version_inst_id_muster_id_map = {info.id: info.template_id for info in version_inst}
        return version_inst_id_muster_id_map

    def get_insts(self, version_inst_ids, category=None):
        """获取实例"""
        # TODO: 在storage支持批量集群后，再添加过滤所有实例
        # 因为现阶段查询实例时，请求路径中必须包含集群ID，导致多次请求，耗时较长
        all_insts = InstanceConfig.objects.filter(instance_id__in=version_inst_ids, is_deleted=False).exclude(
            ins_state=InsState.NO_INS.value
        )
        if category:
            all_insts = all_insts.filter(category__in=category)
        ret_data = []
        for info in all_insts:
            conf = json.loads(info.config)
            metadata = conf.get("metadata") or {}
            labels = metadata.get("labels") or {}
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            if not cluster_id:
                continue
            ret_data.append(
                {"id": info.id, "version_inst_id": info.instance_id, "cluster_id": cluster_id, "name": info.name}
            )
        return ret_data

    def get_cluster_id_env(self, request, project_id):
        """获取集群和环境"""
        data = self.get_project_cluster_info(request, project_id)
        if not data.get("results"):
            return {}
        cluster_results = data.get("results") or []
        return {info["cluster_id"]: info["environment"] for info in cluster_results if not info["disabled"]}

    def get_cluster_category(self, request, kind):
        """获取类型"""
        cluster_type = request.GET.get("cluster_type")
        category = request.GET.get("category")
        if kind == 1:
            if not category:
                raise error_codes.CheckFailed(_("应用类型不能为空"))
            else:
                if category not in CATEGORY_MAP:
                    raise error_codes.CheckFailed(_("类型不正确，请确认"))
                category = [CATEGORY_MAP[category]]
        else:
            category = ["application", "deployment"]
        return cluster_type, category


class GetAllMusters(BaseFilter):
    def compose_data(
        self, all_musters, version_inst_map, version_inst_cluster, cluster_env_map, cluster_type, request_cluster_id
    ):
        """组装返回数据"""
        ret_data = {}
        for info in version_inst_cluster:
            cluster_id = info["cluster_id"]
            version_inst_id = info["version_inst_id"]
            template_id = version_inst_map[version_inst_id]
            curr_env = cluster_env_map.get(cluster_id)
            if not exclude_records(request_cluster_id, cluster_id, cluster_type, cluster_env(curr_env)):
                ret_data[template_id] = all_musters[template_id]
        return ret_data

    def get(self, request, project_id):
        """查询项目下不同集群类型的模板集"""
        flag, kind = self.get_project_kind(request, project_id)
        if not flag:
            return kind
        cluster_type, category = self.get_cluster_category(request, kind)
        all_musters = self.get_muster(project_id)
        version_inst_map = self.get_version_instances(all_musters.keys())
        version_inst_cluster = self.get_insts(version_inst_map.keys(), category=category)
        cluster_env_map = self.get_cluster_id_env(request, project_id)
        # 组装数据
        ret_data = self.compose_data(
            all_musters,
            version_inst_map,
            version_inst_cluster,
            cluster_env_map,
            cluster_type,
            request.query_params.get("cluster_id"),
        )
        ret_data = [{"muster_id": key, "muster_name": val} for key, val in ret_data.items()]
        return APIResponse({"data": ret_data})


class GetAllInstances(BaseFilter):
    def compose_data(self, version_inst_cluster, cluster_env_map, cluster_type, request_cluster_id):
        """组装数据"""
        ret_data = {}
        for info in version_inst_cluster:
            cluster_id = info["cluster_id"]
            curr_env = cluster_env_map.get(cluster_id)
            if not exclude_records(request_cluster_id, cluster_id, cluster_type, cluster_env(curr_env)):
                ret_data[info["name"]] = info["id"]
        return ret_data

    def get(self, request, project_id):
        """获取所有实例"""
        kind = self.project_kind(request)
        cluster_type, category = self.get_cluster_category(request, kind)
        all_musters = self.get_muster(project_id)
        version_inst_map = self.get_version_instances(all_musters.keys())
        version_inst_cluster = self.get_insts(version_inst_map.keys(), category=category)
        cluster_env_map = self.get_cluster_id_env(request, project_id)

        # 组装返回数据
        ret_data = self.compose_data(
            version_inst_cluster, cluster_env_map, cluster_type, request.query_params.get("cluster_id")
        )
        ret_data = [{"app_id": val, "app_name": key} for key, val in ret_data.items()]
        return APIResponse({"data": ret_data})


class GetAllNamespaces(BaseFilter):
    def get_all_namespace(self, request, project_id):
        """获取所有命名空间"""
        flag, all_data = self.get_namespaces(request, project_id)
        if not flag:
            raise error_codes.APIError.f(all_data.data.get("message"))
        results = all_data.get("results") or []
        return {(info["cluster_id"], info["id"]): info["name"] for info in results}

    def compose_data(self, all_namespaces, cluster_env_map, cluster_type):
        """组装数据"""
        ret_data = []
        for (cluster_id, ns_id), ns_name in all_namespaces.items():
            curr_env = cluster_env_map.get(cluster_id)
            if curr_env and str(cluster_env(curr_env)) == str(cluster_type):
                ret_data.append(
                    {
                        "ns_id": ns_id,
                        "ns_name": ns_name,
                        "cluster_id": cluster_id,
                    }
                )
        return ret_data

    def get(self, request, project_id):
        """获取所有的命名空间"""
        kind = self.project_kind(request)
        cluster_type, category = self.get_cluster_category(request, kind)
        all_namespaces = self.get_all_namespace(request, project_id)
        cluster_env_map = self.get_cluster_id_env(request, project_id)
        ret_data = self.compose_data(all_namespaces, cluster_env_map, cluster_type)
        return APIResponse({"data": ret_data})
