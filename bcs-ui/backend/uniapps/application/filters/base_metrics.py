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
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance

from ..base_views import BaseAPI, error_codes
from ..constants import CATEGORY_MAP

CLUSTER_TYPE = [1, 2, "1", "2"]
APP_STATUS = [1, 2, "1", "2"]


class BaseMetricAPI(BaseAPI):
    def get_namespace_id(self, request, project_id, ns_id, ns_name):
        if ns_id:
            return ns_id
        elif ns_name:
            name_id_map = self.get_namespace_name_id(request, project_id)
            ns_id = name_id_map.get(ns_name)
            if not ns_id:
                raise error_codes.CheckFailed(_("命名空间: {} 不存在").format(ns_name))
            return ns_id
        else:
            return None

    def get_app_id(self, request, project_id, app_id, app_name):
        """获取应用ID"""
        if app_id:
            return app_id
        elif app_name:
            tmpl_id_list = Template.objects.filter(project_id=project_id).values_list("id", flat=True)
            version_inst_list = VersionInstance.objects.filter(template_id__in=tmpl_id_list).values_list(
                "id", flat=True
            )
            insts = InstanceConfig.objects.filter(name=app_name, instance_id__in=version_inst_list)
            category = request.GET.get('category')
            if category:
                category_list = self.get_category(request, request.project['kind'])
                insts = insts.filter(category__in=category_list)
            if not insts:
                raise error_codes.CheckFailed(_("应用: {} 不存在").format(app_name))
            return insts[0].id
        return None

    def get_category(self, request, kind):
        """获取类型"""
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
        return category


class BaseMusterMetric(BaseMetricAPI):
    def get_muster_id(self, project_id, muster_id, muster_name):
        """获取模板集ID"""
        if muster_id:
            return muster_id
        elif muster_name:
            musters = Template.objects.filter(project_id=project_id, is_deleted=False)
            musters = musters.filter(name=muster_name)
            if not musters:
                raise error_codes.CheckFailed(_("模板集: {} 不存在").format(muster_name))
            return musters[0].id
        else:
            return None

    def get_filter_params(self, request, project_id):
        """获取过滤参数"""
        cluster_type = request.GET.get("cluster_type")
        app_status = request.GET.get("app_status")
        if app_status and app_status not in APP_STATUS:
            raise error_codes.CheckFailed(_("应用状态不正确，请确认"))
        muster_id = self.get_muster_id(project_id, request.GET.get("muster_id"), request.GET.get("muster_name"))
        app_id = self.get_app_id(request, project_id, request.GET.get("app_id"), request.GET.get("app_name"))
        ns_id = self.get_namespace_id(request, project_id, request.GET.get("ns_id"), request.GET.get("ns_name"))

        # 兼容cluster_id，用以过滤集群下的资源
        request_cluster_id = request.query_params.get("cluster_id")
        return cluster_type, app_status, muster_id, app_id, ns_id, request_cluster_id

    def get_filter_muster(self, project_id, muster_id):
        """获取模板集"""
        all_musters = Template.objects.filter(project_id=project_id, is_deleted=False)
        if muster_id:
            all_musters = all_musters.filter(id=muster_id)
        muster_id_name_map = {info.id: info.name for info in all_musters}
        return muster_id_name_map

    def get_filter_version_instances(self, muster_id_list):
        """根据模板集获取相应的版本实例"""
        version_inst = VersionInstance.objects.filter(template_id__in=muster_id_list, is_deleted=False)
        version_inst_id_muster_id_map = {info.id: info.template_id for info in version_inst}
        return version_inst_id_muster_id_map

    def get_filter_insts(self, version_inst_ids, category=None):
        """获取实例"""
        all_insts = InstanceConfig.objects.filter(instance_id__in=version_inst_ids, is_deleted=False)
        if category:
            all_insts.filter(category__in=category)
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

    def get_inst_name(self, inst_id):
        """根据实例ID获取实例名称"""
        if not inst_id:
            return None
        inst_info = InstanceConfig.objects.filter(id=inst_id, is_deleted=False)
        if inst_info:
            return inst_info[0].name
        else:
            return None


class BaseNamespaceMetric(BaseMetricAPI):
    def get_filter_params(self, request, project_id):
        """获取过滤参数"""
        cluster_type = request.GET.get("cluster_type")
        app_status = request.GET.get("app_status")
        if app_status and app_status not in APP_STATUS:
            raise error_codes.CheckFailed(_("应用状态不正确，请确认"))
        app_id = self.get_app_id(request, project_id, request.GET.get("app_id"), request.GET.get("app_name"))
        ns_id = self.get_namespace_id(request, project_id, request.GET.get("ns_id"), request.GET.get("ns_name"))

        # 兼容cluster_id，用以过滤集群下的资源
        request_cluster_id = request.query_params.get("cluster_id")
        return cluster_type, app_status, app_id, ns_id, request_cluster_id

    def get_inst_name(self, inst_id):
        """获取实例名称"""
        inst_info = InstanceConfig.objects.filter(id=inst_id, is_deleted=False)
        if not inst_info:
            raise error_codes.CheckFailed(_("实例不存在，请确认!"))
        return inst_info[0].name
