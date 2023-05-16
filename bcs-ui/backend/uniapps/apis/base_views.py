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
from django.conf import settings
from rest_framework import viewsets

from backend.components.paas_cc import get_all_clusters, get_namespace_list, get_project
from backend.templatesets.legacy_apps.configuration.models import ShowVersion, Template
from backend.templatesets.legacy_apps.instance.models import InstanceConfig
from backend.uniapps.apis.utils import parse_jwt_info
from backend.utils import FancyDict
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

DEFAULT_USER = settings.DEFAULT_API_TEST_USER
JWT_KEY_NAME = "HTTP_X_BKAPI_JWT"
DEFAULT_APP_CODE = "workbench"


class BaseAPIViews(viewsets.ViewSet):
    authentication_classes = ()
    permission_classes = ()

    def get_instance_info(self, instance_id):
        return InstanceConfig.objects.filter(id=instance_id)

    def jwt_info(self, request):
        return request.META.get(JWT_KEY_NAME, "")

    def get_cluster_name_id_map(self, access_token, project_id):
        resp = get_all_clusters(access_token, project_id, desire_all_data=True)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"))
        data = resp.get("data") or {}
        results = data.get("results") or []
        if not results:
            raise error_codes.CheckFailed.f("查询项目下集群失败，请稍后重试")
        return {info["cluster_id"]: info["name"] for info in results}

    def get_namespace_data(self, access_token, project_id, mult_return=False):
        all_info = get_namespace_list(access_token, project_id, desire_all_data=True)
        if all_info.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(all_info.get("message"), replace=True)
        # 处理命名空间名称和ID
        data = all_info.get("data") or {}
        results = data.get("results") or []
        if not results:
            raise error_codes.APIError.f("获取命名空间信息为空", replace=True)
        if mult_return:
            return results, all_info
        return results

    def get_ns_id_by_ns_name(self, access_token, project_id, cluster_ns_info):
        """通过ns名称获取ns id"""
        cluster_name_id_map = self.get_cluster_name_id_map(access_token, project_id)
        results, all_info = self.get_namespace_data(access_token, project_id, mult_return=True)
        # 针对不同集群和命名空间名称匹配命名空间ID
        cluster_ns_map = {}
        cluster_id_ns_map = {}
        for info in results:
            # 共享集群命名空间过滤
            if info['name'].startswith(settings.SHARED_CLUSTER_NS_PREFIX):
                continue

            curr_cluster_name = cluster_name_id_map.get(info["cluster_id"])
            if not curr_cluster_name:
                raise error_codes.CheckFailed.f("没有查询到集群ID对应的集群名称，请联系管理员处理!")
            cluster_ns_map["%s:%s" % (curr_cluster_name, info["name"])] = info["id"]
            cluster_id_ns_map["%s:%s" % (info["cluster_id"], info["name"])] = info["id"]
        # 记录不存在的命名空间
        message_list = []
        ns_id_list = []
        variable_info = {}
        for cluster_name, ns_name_var in cluster_ns_info.items():
            if not ns_name_var:
                raise error_codes.APIError.f("命名空间不能为空，请确认!", replace=True)
            for item in ns_name_var.keys():
                key = "%s:%s" % (cluster_name, item)
                if not ((key in cluster_ns_map) or (key in cluster_id_ns_map)):
                    message_list.append("集群:%s，命名空间:%s" % (cluster_name, item))
                    continue
                if key in cluster_ns_map:
                    ns_id_list.append(str(cluster_ns_map[key]))
                    variable_info[str(cluster_ns_map[key])] = ns_name_var[item]
                else:
                    ns_id_list.append(str(cluster_id_ns_map[key]))
                    variable_info[str(cluster_id_ns_map[key])] = ns_name_var[item]
        if message_list:
            err_message = ";".join(message_list)
            raise error_codes.CheckFailed.f("%s，不存在，请确认" % err_message)
        return ns_id_list, variable_info, all_info

    def get_request_user(self, request, access_token, project_id):
        if settings.DEBUG:
            app_code, username = DEFAULT_APP_CODE, DEFAULT_USER
        else:
            app_code, username = parse_jwt_info(self.jwt_info(request))
        request.user = APIUser
        request.user.token.access_token = access_token
        request.user.username = DEFAULT_USER if settings.DEBUG else username
        request.user.app_code = app_code

        result = get_project(access_token, project_id)
        project = result.get("data") or {}
        project = FancyDict(project)
        request.project = project

    def get_show_version_detail(self, version_name, version_id, template_name=None):
        """查询show version相应信息
        如果两者都存在则以version id为准
        """
        if version_id:
            info = ShowVersion.objects.filter(id=version_id)
        else:
            info = ShowVersion.objects.filter(name=version_name)
        if template_name:
            tmpl_info = Template.objects.filter(name=template_name)
            if not tmpl_info:
                raise error_codes.CheckFailed.f("没有查询到模板集信息", replace=True)
            tmpl_id = tmpl_info[0].id
            info = info.filter(template_id=tmpl_id)
        if not info:
            raise error_codes.CheckFailed.f("没有查询到展示版本信息", replace=True)
        return info[0]

    def get_ns_id_name_map(self, access_token, project_id):
        """获取项目下所有的命名空间信息"""
        namespace_results = self.get_namespace_data(access_token, project_id)
        id_name_map = {}
        name_id_map = {}
        for info in namespace_results:
            id_name_map.update({info["id"]: info["name"]})
            name_id_map.update({info["name"]: info["id"]})
        return id_name_map, name_id_map


class APIUserToken(object):
    access_token = None


class APIUser(object):
    token = APIUserToken
    username = None
