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
import copy

from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.components import paas_cc
from backend.utils.errcodes import ErrorCode

from ..base_views import error_codes
from ..constants import CATEGORY_MAP
from ..filters.base_metrics import BaseNamespaceMetric
from ..utils import APIResponse, cluster_env, exclude_records
from . import k8s_views

CLUSTER_ENV_MAP = settings.CLUSTER_ENV_FOR_FRONT


class GetProjectNamespace(BaseNamespaceMetric):
    def get_namespace(self, request, project_id):
        """获取namespace"""
        resp = paas_cc.get_namespace_list(request.user.token.access_token, project_id, desire_all_data=True)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"))
        data = resp.get("data") or {}
        return data.get("results") or []

    def get_cluster_list(self, request, project_id, cluster_ids):
        """根据cluster_id获取集群信息"""
        resp = paas_cc.get_cluster_list(request.user.token.access_token, project_id, cluster_ids)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"))
        data = resp.get("data") or []
        # if not data:
        #     raise error_codes.APIError.f("查询集群信息为空")
        return data

    def compose_data(
        self, ns_map, cluster_env, ns_app, exist_app, ns_inst_error_count, create_error, app_status, all_ns_inst_count
    ):
        """组装返回数据"""
        ns_map_copy = copy.deepcopy(ns_map)
        for key, val in ns_map_copy.items():
            cluster_id = val["cluster_id"]
            error_num = (ns_inst_error_count.get((cluster_id, key[1])) or 0) + (create_error.get(int(key[0])) or 0)
            ns_map[key]["total_num"] = all_ns_inst_count.get((cluster_id, key[1])) or ns_app.get(str(key[0])) or 0
            if app_status in [1, "1"]:
                ns_map[key]["total_num"] = (ns_map[key]["total_num"]) - error_num
                if error_num > 0 and ns_app.get(str(key[0])) == error_num:
                    ns_map.pop(key)
                    continue
            if app_status in [2, "2"]:
                ns_map[key]["total_num"] = error_num
                if error_num == 0:
                    ns_map.pop(key)
                    continue
            ns_map[key]["env_type"] = cluster_env.get(cluster_id, {}).get('env_type')
            ns_map[key]["error_num"] = error_num
            ns_map[key]["cluster_name"] = cluster_env.get(cluster_id, {}).get('name')
            if exist_app == "1":
                if not ns_map[key]["total_num"]:
                    ns_map.pop(key, None)

    def get_cluster_ns(self, ns_list, cluster_type, ns_id, cluster_env_map, request_cluster_id):
        """组装集群、命名空间等信息"""
        ns_map = {}
        ns_id_list = []
        cluster_id_list = []
        ns_name_list = []
        for info in ns_list:
            if exclude_records(
                request_cluster_id,
                info["cluster_id"],
                cluster_type,
                cluster_env_map.get(info["cluster_id"], {}).get("cluster_env"),
            ):
                continue
            if ns_id and str(info["id"]) != str(ns_id):
                continue
            ns_map[(info["id"], info["name"])] = {
                "cluster_id": info["cluster_id"],
                "id": info["id"],
                "name": info["name"],
                "project_id": info["project_id"],
            }
            ns_name_list.append(info["name"])
            # id和cluster_id肯定存在
            ns_id_list.append(info["id"])
            cluster_id_list.append(info["cluster_id"])
        return ns_map, ns_id_list, cluster_id_list, ns_name_list

    def get_cluster_id_env(self, request, project_id):
        data = self.get_project_cluster_info(request, project_id)
        if not data.get("results"):
            return {}, {}
        cluster_results = data.get("results") or []
        cluster_env_map = {
            info["cluster_id"]: {
                "cluster_name": info["name"],
                "cluster_env": cluster_env(info["environment"]),
                "cluster_env_str": cluster_env(info["environment"], ret_num_flag=False),
            }
            for info in cluster_results
            if not info["disabled"]
        }
        return cluster_results, cluster_env_map

    def get(self, request, project_id):
        """获取项目下的所有命名空间"""
        # 获取过滤参数
        cluster_type, app_status, app_id, ns_id, request_cluster_id = self.get_filter_params(request, project_id)
        exist_app = request.GET.get("exist_app")
        # 获取项目类型
        project_kind = self.project_kind(request)
        # 获取项目下集群类型
        cluster_list, cluster_env_map = self.get_cluster_id_env(request, project_id)
        if not cluster_list:
            return APIResponse({"data": []})
        # 获取项目下面的namespace
        ns_list = self.get_namespace(request, project_id)
        if not ns_list:
            return APIResponse({"data": []})
        # 组装命名空间数据、命名空间ID、项目下集群信息
        ns_map, ns_id_list, cluster_id_list, ns_name_list = self.get_cluster_ns(
            ns_list, cluster_type, ns_id, cluster_env_map, request_cluster_id
        )
        # 匹配集群的环境
        cluster_env = {
            info["cluster_id"]: {'env_type': CLUSTER_ENV_MAP.get(info["environment"], "stag"), 'name': info['name']}
            for info in cluster_list
        }
        inst_name = None
        if app_id:
            inst_name = self.get_inst_name(app_id)
        category = request.GET.get("category")
        if not category or category not in CATEGORY_MAP.keys():
            raise error_codes.CheckFailed(_("类型不正确"))
        client = k8s_views.GetNamespace()
        ns_app, ns_inst_error_count, create_error, all_ns_inst_count = client.get(
            request,
            ns_id_list,
            category,
            ns_map,
            project_id,
            project_kind,
            self.get_app_deploy_with_post,
            inst_name,
            ns_name_list,
            cluster_id_list,
        )
        # 匹配数据
        self.compose_data(
            ns_map, cluster_env, ns_app, exist_app, ns_inst_error_count, create_error, app_status, all_ns_inst_count
        )
        ret_data = ns_map.values()
        return APIResponse({"data": ret_data})


class GetInstances(BaseNamespaceMetric):
    def check_ns_with_project(self, request, project_id, ns_id, cluster_type, cluster_env_map):
        """判断命名空间属于项目"""
        resp = paas_cc.get_namespace_list(request.user.token.access_token, project_id, desire_all_data=True)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"))
        data = resp.get("data") or {}
        if not data.get("results"):
            raise error_codes.APIError(_("查询命名空间为空"))
        ns_list = data["results"]
        cluster_id = None
        ns_name = None
        for info in ns_list:
            if str(info["id"]) != str(ns_id):
                continue
            else:
                cluster_id = info["cluster_id"]
                ns_name = info["name"]

        return cluster_id, ns_name

    def get(self, request, project_id, ns_id):
        cluster_type, app_status, app_id, filter_ns_id, request_cluster_id = self.get_filter_params(
            request, project_id
        )
        if filter_ns_id and str(ns_id) != str(filter_ns_id):
            return APIResponse({"data": {}})
        # 获取项目下集群类型
        cluster_env_map = self.get_cluster_id_env(request, project_id)
        # 检查命名空间属于项目
        cluster_id, ns_name = self.check_ns_with_project(request, project_id, ns_id, cluster_type, cluster_env_map)
        inst_name = None
        if app_id:
            inst_name = self.get_inst_name(app_id)
        ns_name_id = self.get_namespace_name_id(request, project_id)
        # 根据类型进行过滤数据
        category = request.GET.get("category")
        if not category or category not in CATEGORY_MAP.keys():
            raise error_codes.CheckFailed(_("类型不正确"))
        client = k8s_views.GetInstances()
        ret_data = client.get(
            request,
            project_id,
            ns_id,
            category,
            inst_name,
            app_status,
            cluster_env_map,
            cluster_id,
            ns_name,
            ns_name_id,
        )
        # 拆分需要处理权限和默认权限
        auth_instance_list = []
        default_auth_instance_list = []
        for info in ret_data["instance_list"]:
            if info["from_platform"]:
                auth_instance_list.append(info)
            else:
                # info["permissions"] = {"create": True, "delete": True, "view": True, "edit": True, "use": True}
                default_auth_instance_list.append(info)
        # 添加权限
        ret_instance_list = self.bcs_perm_handler(request, project_id, auth_instance_list)
        # 针对非模板集的权限解析
        default_ret_instance_list = self.bcs_perm_handler(
            request, project_id, default_auth_instance_list, tmpl_view=False
        )
        ret_instance_list.extend(default_ret_instance_list)
        ret_data["instance_list"] = ret_instance_list
        return APIResponse({"data": ret_data})
