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
import re
from datetime import datetime

import yaml
from django.conf import settings
from django.db.models import Q
from django.http import JsonResponse

from backend.bcs_web.audit_log import client
from backend.components import paas_cc
from backend.components.bcs.k8s import K8SClient
from backend.container_service.projects.base.constants import ProjectKindID
from backend.iam.permissions.resources.templateset import TemplatesetAction, TemplatesetPermCtx, TemplatesetPermission
from backend.templatesets.legacy_apps.configuration.models import MODULE_DICT, ShowVersion, Template, VersionedEntity
from backend.templatesets.legacy_apps.configuration.utils import check_var_by_config
from backend.templatesets.legacy_apps.instance import utils as inst_utils
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.templatesets.legacy_apps.instance.serializers import (
    VariableNamespaceSLZ,
    VersionInstanceCreateOrUpdateSLZ,
)
from backend.templatesets.legacy_apps.instance.utils import (
    handle_all_config,
    validate_instance_entity,
    validate_ns_by_tempalte_id,
    validate_version_id,
)
from backend.templatesets.var_mgmt.models import NameSpaceVariable, Variable
from backend.uniapps.apis.applications import serializers
from backend.uniapps.apis.base_views import APIUser, BaseAPIViews
from backend.uniapps.apis.constants import CATEGORY_MODULE_MAP
from backend.uniapps.apis.utils import check_user_project, skip_authentication
from backend.uniapps.application import constants as app_constants
from backend.uniapps.application import utils
from backend.uniapps.application import views as app_views
from backend.uniapps.application.constants import FUNC_MAP, REVERSE_CATEGORY_MAP
from backend.utils import FancyDict
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

logger = logging.getLogger(__name__)
PIPELINE_DEFAULT_USER = settings.PIPELINE_DEFAULT_USER
GCLOUD_DEFAULT_USER = settings.GCLOUD_DEFAULT_USER


class ProjectApplicationInfo(app_views.BaseAPI, BaseAPIViews):
    def get_template(self, project_id):
        """获取模板集"""
        return Template.objects.filter(project_id=project_id, is_deleted=False).values_list("id", flat=True)

    def get_version_instance(self, template_id_list):
        """获取实例对应的版本信息"""
        version_tmpl_map = VersionInstance.objects.filter(template_id__in=template_id_list, is_deleted=False)
        return {info.id: info.template_id for info in version_tmpl_map}

    def get_namespace_info(self, access_token, project_id):
        namespace_res = paas_cc.get_namespace_list(access_token, project_id, desire_all_data=True)
        if namespace_res.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(namespace_res.get("message"))
        namespace_data = namespace_res.get("data", {}).get("results") or []
        namespace_dict = {i["name"]: i for i in namespace_data}
        return namespace_dict

    def get_inst(self, request, category, project_id, namespace=None, access_token=None):
        """获取项目下的实例信息"""
        tmpl_id_list = self.get_template(project_id)
        version_tmpl_map = self.get_version_instance(tmpl_id_list)
        version_inst_id_list = version_tmpl_map.keys()
        inst_info = InstanceConfig.objects.filter(
            instance_id__in=version_inst_id_list, is_deleted=False, category=category
        ).exclude(ins_state=0)
        if namespace:
            namespace_info = self.get_namespace_info(access_token, project_id)
            curr_ns = namespace_info.get(namespace)
            if not curr_ns:
                raise error_codes.CheckFailed.f("命名空间【%s】没有实例化" % namespace)
            inst_info = inst_info.filter(namespace=curr_ns["id"])
        ret_data = []
        cluster_map = self.get_cluster_id_env(request, project_id)
        for info in inst_info:
            conf = json.loads(info.config)
            metadata = conf.get("metadata") or {}
            spec = conf.get("spec") or {}
            labels = metadata.get("labels") or {}
            instance_num = spec.get("instance") or spec.get("replicas")
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            ret_data.append(
                {
                    "id": info.id,
                    "creator": info.creator,
                    "created": info.created,
                    "updated": info.updated,
                    "name": info.name,
                    "cluster_id": cluster_id,
                    "namespace": labels.get("io.tencent.bcs.namespace"),
                    "namespace_id": info.namespace,
                    "instance_num": instance_num,
                    "muster_id": version_tmpl_map.get(info.instance_id),
                    "environment": cluster_map.get(cluster_id, {}).get("cluster_env_str"),
                }
            )
        self.bcs_perm_handler(request, project_id, ret_data, filter_use=False)
        return ret_data

    def get(self, request, cc_app_id, project_id):
        """获取项目下的application信息"""
        params = request.query_params
        params_slz = serializers.ProjectAPPParamsSLZ(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data

        project_kind, app_code = check_user_project(
            params_slz["access_token"], project_id, cc_app_id, self.jwt_info(request)
        )
        category = params_slz["category"]
        if category not in CATEGORY_MODULE_MAP[project_kind]:
            raise error_codes.CheckFailed.f("请求类型必须为%s中的一种，请确认" % "/".join(CATEGORY_MODULE_MAP[project_kind].keys()))
        curr_category = CATEGORY_MODULE_MAP[project_kind][category]
        self.get_request_user(request, params_slz["access_token"], project_id)
        return JsonResponse(
            {
                "code": 0,
                "data": self.get_inst(
                    request,
                    curr_category,
                    project_id,
                    namespace=params_slz.get("namespace"),
                    access_token=params_slz["access_token"],
                ),
            }
        )

    def get_instance_id_list(self, entity, ns_info):
        name_list = []
        category_list = []
        for key, val in entity.items():
            item = [info["name"] for info in val]
            name_list.extend(item)
            category_list.append(key)
        ns_list = ns_info.split(",")
        inst_info = (
            InstanceConfig.objects.filter(name__in=name_list, namespace__in=ns_list, category__in=category_list)
            .exclude(ins_state=InsState.NO_INS.value)
            .values("id")
        )
        return [info["id"] for info in inst_info]

    def get_resource_variables(self, request, req_data, version_id, project_id, project_kind, namespace_res):
        version_entity = validate_version_id(project_id, version_id, is_version_entity_retrun=True)

        self.slz = VariableNamespaceSLZ(data=req_data, context={"project_kind": project_kind})
        self.slz.is_valid(raise_exception=True)
        slz_data = self.slz.data

        instance_entity = slz_data["instance_entity"]

        lb_services = []
        key_list = []
        for cate in instance_entity:
            cate_data = instance_entity[cate]
            cate_id_list = [i.get("id") for i in cate_data if i.get("id")]
            # 查询这些配置文件的变量名
            for _id in cate_id_list:
                if cate in ["metric"]:
                    config = self.get_metric_confg(_id)
                else:
                    try:
                        resource = MODULE_DICT.get(cate).objects.get(id=_id)
                    except Exception:
                        continue
                    config = resource.config
                search_list = check_var_by_config(config)
                key_list.extend(search_list)

        key_list = list(set(key_list))
        variable_dict = {}
        if key_list:
            # 验证变量名是否符合规范，不符合抛出异常，否则后续用 django 模板渲染变量也会抛出异常

            var_objects = Variable.objects.filter(Q(project_id=project_id) | Q(project_id=0))

            # access_token = request.user.token.access_token
            # namespace_res = paas_cc.get_namespace_list(
            #     access_token, project_id, desire_all_data=True)
            namespace_data = namespace_res.get("data", {}).get("results") or []
            namespace_dict = {str(i["id"]): i["cluster_id"] for i in namespace_data}

            ns_list = slz_data["namespaces"].split(",") if slz_data["namespaces"] else []
            for ns_id in ns_list:
                _v_list = []
                for _key in key_list:
                    key_obj = var_objects.filter(key=_key)
                    if key_obj.exists():
                        _obj = key_obj.first()
                        # 只显示自定义变量
                        if _obj.category == "custom":
                            cluster_id = namespace_dict.get(ns_id, 0)
                            _v_list.append(
                                {"key": _obj.key, "name": _obj.name, "value": _obj.get_show_value(cluster_id, ns_id)}
                            )
                    else:
                        _v_list.append({"key": _key, "name": _key, "value": ""})
                variable_dict[ns_id] = _v_list
        return variable_dict, lb_services

    def compose_ns_vars(self, req_vars, default_vars):
        """组装最终的变量数据"""
        for ns_id, val in default_vars.items():
            for item in val:
                item_key = item.get("key", "")
                if item_key not in req_vars.get(ns_id) or []:
                    req_vars[ns_id][item_key] = item.get("value", "")

    def get_version_entity(self, version_id):
        version_info = VersionedEntity.objects.filter(id=version_id)
        if not version_info:
            raise error_codes.CheckFailed.f("没有查询到版本信息", replace=True)
        return version_info[0]

    def get_category_tmpl_name_id(self, category_entity):
        ret_data = {}
        for category, tmpl_info in category_entity.items():
            for tmpl in tmpl_info:
                if category in ret_data:
                    ret_data[category][tmpl["name"]] = tmpl["id"]
                else:
                    ret_data[category] = {tmpl["name"]: tmpl["id"]}
        return ret_data

    def compose_req_instance_entity_info(self, req_entity, all_tmpl):
        for category, tmpl_info in req_entity.items():
            entity = all_tmpl.get(category)
            if not entity:
                raise error_codes.CheckFailed.f("输入模板不存在，请确认后重试", replace=True)
            for info in tmpl_info:
                if info.get("id"):
                    continue
                info["id"] = entity.get(info["name"])

    def render_instance_entity(self, project_kind, instance_entity, real_version_id):
        """渲染entity
        # 因为允许用户只传递模板的名称，因此，需要通过名称查找到模板ID
        """
        version_entity = VersionedEntity.objects.get(id=real_version_id)
        entity = json.loads(version_entity.entity)

        for cate, info in instance_entity.items():
            ids = entity.get(cate) or ""
            if not ids:
                continue
            id_list = ids.split(",")
            category_name_list = [name["name"] for name in info]
            category_tmpl_info = MODULE_DICT[cate].objects.filter(id__in=id_list, name__in=category_name_list)
            tmpl_name_id_map = {tmpl.name: tmpl.id for tmpl in category_tmpl_info}
            for item in info:
                item["id"] = tmpl_name_id_map.get(item["name"])

    def allow_render_tmpl(self, instance_entity):
        """如果没有指定模板ID，需要渲染模板信息"""
        for cate, info in instance_entity.items():
            for item in info:
                if not item.get("id"):
                    return True
        return False

    def api_post(self, request, cc_app_id, project_id):
        """实例化模板"""
        params = request.query_params
        params_slz = serializers.ProjectTemplateSetParamsSLZ(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data
        project_kind, app_code, english_name = check_user_project(
            params_slz["access_token"], project_id, cc_app_id, self.jwt_info(request), project_code_flag=True
        )
        # 获取用户
        self.get_request_user(request, params_slz["access_token"], project_id)
        # 参数验证
        self.project_id = project_id
        req_data = request.data
        # show_version_id和show_version_name不能同时为空
        show_version_name = req_data.get("show_version_name")
        show_version_id = req_data.get("show_version_id")
        template_name = req_data.get("template_name")
        # 通过 show_version_id 查询 version_id，不通过API传参数
        show_ver = self.get_show_version_detail(show_version_name, show_version_id, template_name=template_name)
        version_id = show_ver.real_version_id
        req_data["version_id"] = version_id
        show_version_id = show_ver.id
        req_data["show_version_id"] = show_version_id

        template, version_entity = validate_version_id(
            project_id, version_id, is_return_all=True, show_version_id=show_version_id
        )

        self.template_id = version_entity.template_id
        tem_instance_entity = version_entity.get_version_instance_resource_ids

        project_kind = ProjectKindID
        # 默认为is_start为True
        req_data["is_start"] = True
        # 转换ns名称为ns ID
        cluster_ns_info = req_data.get("cluster_ns_info")
        if not cluster_ns_info:
            raise error_codes.CheckFailed.f("集群ID或命名空间，请确认!")
        ns_id_list, variable_info, all_ns_info = self.get_ns_id_by_ns_name(
            params_slz["access_token"], project_id, cluster_ns_info
        )
        req_data["namespaces"] = ",".join(ns_id_list)
        req_data["variable_info"] = variable_info

        default_ns_vars, lb_service = self.get_resource_variables(
            request, req_data, version_id, project_id, project_kind, all_ns_info
        )
        # 匹配数据
        self.compose_ns_vars(req_data["variable_info"], default_ns_vars)

        self.slz = VersionInstanceCreateOrUpdateSLZ(data=req_data, context={"project_kind": project_kind})
        self.slz.is_valid(raise_exception=True)
        slz_data = self.slz.data

        version_info = self.get_version_entity(version_id)
        category_tmpl_info = version_info.get_version_instance_resources()
        category_tmpl_map = self.get_category_tmpl_name_id(category_tmpl_info)
        # 验证前端传过了的预览资源是否在该版本的资源
        req_instance_entity = slz_data.get("instance_entity") or {}
        if self.allow_render_tmpl(req_instance_entity):
            self.render_instance_entity(project_kind, req_instance_entity, version_id)
        self.compose_req_instance_entity_info(req_instance_entity, category_tmpl_map)
        try:
            self.instance_entity = validate_instance_entity(req_instance_entity, tem_instance_entity)
        except Exception as err:
            return JsonResponse({"code": 400, "message": ";".join(err.detail)})

        access_token = params_slz["access_token"]

        # 验证关联lb情况下，lb 是否都已经选中
        service_id_list = self.instance_entity.get("service") or []

        # 判断 template 下 前台传过来的 namespace 是否已经实例化过
        res, ns_name_list, namespace_dict = validate_ns_by_tempalte_id(
            self.template_id, ns_id_list, access_token, project_id, req_instance_entity
        )
        if not res:
            return JsonResponse(
                {"code": 400, "message": "以下命名空间已经实例化过，不能再实例化\n%s" % "\n".join(ns_name_list), "data": ns_name_list}
            )
        username = request.user.username
        if not username:
            username = GCLOUD_DEFAULT_USER if app_code == "gcloud" else PIPELINE_DEFAULT_USER

        slz_data["ns_list"] = ns_id_list
        slz_data["instance_entity"] = self.instance_entity
        slz_data["template_id"] = self.template_id
        slz_data["project_id"] = project_id
        slz_data["version_id"] = version_id
        slz_data["show_version_id"] = show_version_id

        result = handle_all_config(slz_data, access_token, username, project_kind=project_kind)
        # 添加操作记录
        temp_name = version_entity.get_template_name()

        for i in result["success"]:
            client.ContextActivityLogClient(
                project_id=project_id,
                user=app_code,
                resource_type="template",
                resource=temp_name,
                resource_id=self.template_id,
                extra=json.dumps(self.instance_entity),
                description="实例化模板集[%s]命名空间[%s]" % (temp_name, i["ns_name"]),
            ).log_add(activity_status="succeed")

        failed_ns_name_list = []
        failed_msg = []
        is_show_failed_msg = False
        for i in result["failed"]:
            if i["res_type"]:
                description = "实例化模板集[%s]命名空间[%s]，在实例化%s时失败，错误消息：%s" % (
                    temp_name,
                    i["ns_name"],
                    i["res_type"],
                    i["err_msg"],
                )
                failed_ns_name_list.append("%s(实例化%s时)" % (i["ns_name"], i["res_type"]))
            else:
                description = "实例化模板集[%s]命名空间[%s]失败，错误消息：%s" % (temp_name, i["ns_name"], i["err_msg"])
                failed_ns_name_list.append(i["ns_name"])
                if i.get("show_err_msg"):
                    failed_msg.append(i["err_msg"])
                    is_show_failed_msg = True

            client.ContextActivityLogClient(
                project_id=project_id,
                user=username,
                resource_type="template",
                resource=temp_name,
                resource_id=self.template_id,
                extra=json.dumps(self.instance_entity),
                description=description,
            ).log_add(activity_status="failed")

            if is_show_failed_msg:
                msg = "\n".join(failed_msg)
            else:
                msg = "以下命名空间实例化失败，\n%s，请联系集群管理员解决" % "\n".join(failed_ns_name_list)
            if failed_ns_name_list:
                return JsonResponse({"code": 400, "message": msg, "data": failed_ns_name_list})
        inst_id_list = self.get_instance_id_list(
            self.slz.data.get("instance_entity") or {}, self.slz.data.get("namespaces") or ""
        )
        return JsonResponse(
            {
                "code": 0,
                "message": "OK",
                "data": {"version_id": version_id, "template_id": self.template_id, "inst_id_list": inst_id_list},
            }
        )


class InstanceInfo(app_views.BaseAPI, BaseAPIViews):
    def get_inst(self, request, category, params, project_id):
        """获取项目下的实例信息"""
        _, name_id_map = self.get_ns_id_name_map(params["access_token"], project_id)
        namespace = name_id_map.get(params["namespace"])
        if not namespace:
            return {}
        inst_infos = InstanceConfig.objects.filter(
            name=params["name"], namespace=namespace, category=category, is_deleted=False
        )
        if not inst_infos:
            return {}
        inst_info = inst_infos[0]
        return {
            "id": inst_info.id,
            "name": inst_info.name,
            "namespace": inst_info.namespace,
            "config": inst_info.config,
        }

    def get(self, request, cc_app_id, project_id):
        """获取项目下的application信息
        注: 此接口现阶段只供pipeline使用
        """
        params = request.query_params

        project_kind, app_code = check_user_project(
            params["access_token"], project_id, cc_app_id, self.jwt_info(request)
        )
        if not skip_authentication(app_code):
            raise error_codes.CheckFailed(f"应用编码[{app_code}]没有权限调用，请联系管理员处理")
        category = params["category"]
        if category not in CATEGORY_MODULE_MAP[project_kind]:
            raise error_codes.CheckFailed.f("请求类型必须为%s中的一种，请确认" % "/".join(CATEGORY_MODULE_MAP[project_kind].keys()))
        curr_category = CATEGORY_MODULE_MAP[project_kind][category]
        return JsonResponse({"code": 0, "data": self.get_inst(request, curr_category, params, project_id)})


class InstanceNamespace(BaseAPIViews, app_views.BaseAPI):
    def get_namespace(self, request, project_id, inst_name, category):
        inst_info = InstanceConfig.objects.filter(name=inst_name, is_deleted=False, category=category)
        project_id_list = []
        namespace_data = {}
        cluster_map = self.get_cluster_id_env(request, project_id)
        for info in inst_info:
            conf = json.loads(info.config)
            metadata = conf.get("metadata") or {}
            labels = metadata.get("labels") or {}
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            namespace_data[info.namespace] = {
                "namespace": labels.get("io.tencent.bcs.namespace"),
                "namespace_id": info.namespace,
                "cluster_id": cluster_id,
                "environment": cluster_map.get(cluster_id, {}).get("cluster_env_str"),
            }
            project_id_list.append(labels.get("io.tencent.paas.projectid"))
        return project_id_list, namespace_data

    def get(self, request, cc_app_id, project_id, instance_name):
        params = request.query_params
        params_slz = serializers.ProjectAPPParamsSLZ(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data
        project_kind, app_code = check_user_project(
            params_slz["access_token"], project_id, cc_app_id, self.jwt_info(request)
        )
        self.get_request_user(request, params_slz["access_token"], project_id)
        category = params_slz["category"]
        if category not in CATEGORY_MODULE_MAP[project_kind]:
            raise error_codes.CheckFailed.f("请求类型必须为%s中的一种，请确认" % "/".join(CATEGORY_MODULE_MAP[project_kind].keys()))
        curr_category = CATEGORY_MODULE_MAP[project_kind][category]
        project_id_list, namespace_data = self.get_namespace(request, project_id, instance_name, curr_category)
        project_id_list = [info for info in project_id_list if info]
        if not (len(set(project_id_list)) == 1 and project_id_list[0] == project_id):
            raise error_codes.CheckFailed.f("实例不属于项目，请确认")
        ret_data = list(namespace_data.values())
        return JsonResponse({"code": 0, "data": ret_data})


class InstanceStatus(BaseAPIViews):
    def get_cluster_ns(self, inst_id):
        inst_info = InstanceConfig.objects.filter(id=inst_id)
        if not inst_info:
            raise error_codes.CheckFailed.f("获取应用实例为空，请确认", replace=True)
        curr_inst = inst_info[0]
        try:
            config = json.loads(curr_inst.config)
        except Exception as err:
            logger.error("Parse error, detail: %s" % err)
            raise error_codes.CheckFailed.f("解析实例配置异常!")
        metadata = config.get("metadata") or {}
        labels = metadata.get("labels") or {}
        cluster_id = labels.get("io.tencent.bcs.clusterid")
        namespace = labels.get("io.tencent.bcs.namespace")
        return cluster_id, namespace, curr_inst.category, curr_inst.name

    def get_k8s_category_status(self, access_token, project_id, cluster_id, category, instance_name, namespace):
        """获取k8s状态"""
        client = K8SClient(access_token, project_id, cluster_id, None)
        curr_func = getattr(client, FUNC_MAP[category] % "get")
        params = {
            "name": instance_name,
            "namespace": namespace,
            "field": "data.status,resourceName,namespace,data.spec.parallelism,data.spec.paused,data.spec.replicas",
        }
        resp = curr_func(params)
        if resp.get("code") != 0:
            raise error_codes.CheckFailed.f(resp.get("message"), replace=True)

        return resp.get("data")

    def get(self, request, cc_app_id, project_id, instance_id):
        params = request.query_params
        params_slz = serializers.InstanceStatusParamsSLZ(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data
        access_token = params_slz["access_token"]
        # 获取namespace及集群信息
        cluster_id, namespace, category, inst_name = self.get_cluster_ns(instance_id)
        # 根据类型请求k8s状态
        data = self.get_k8s_category_status(access_token, project_id, cluster_id, category, inst_name, namespace)
        if not data:
            return JsonResponse({"code": 0, "data": {}, "message": "查询数据为空"})
        for info in data:
            replicas, available = utils.get_k8s_desired_ready_instance_count(info, REVERSE_CATEGORY_MAP[category])
            if available != replicas or available == 0:
                return JsonResponse({"code": 0, "data": {"status": "unready"}})

        return JsonResponse({"code": 0, "data": {"status": "running"}})


class BaseHandleInstance(BaseAPIViews):
    def init_handler(self, request, cc_app_id, project_id, instance_id, slz_func):
        params = request.query_params
        # params_slz = serializers.ScaleInstanceParamsSLZ(data=params)
        params_slz = slz_func(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data
        project_kind, app_code, project_info = check_user_project(
            params_slz["access_token"], project_id, cc_app_id, self.jwt_info(request), is_orgin_project=True
        )
        request.user = APIUser
        request.user.token.access_token = params_slz["access_token"]
        request.user.username = app_code
        request.user.project_kind = project_kind
        # 添加project信息，方便同时处理提供给apigw和前台页面使用
        request.project = FancyDict(project_info)

    def get_instance_id(self, request, instance_id):
        """通过名称+命名空间+类型获取实例ID"""
        if instance_id != 0:
            return instance_id
        params = request.query_params
        name = params.get("name")
        namespace = params.get("namespace")
        category = params.get("category")
        k8s_resource = CATEGORY_MODULE_MAP[str(request.user.project_kind)]
        instance_info = MODULE_DICT[k8s_resource[category]].objects.filter(name=name, namespace=namespace)
        if not instance_info:
            raise error_codes.CheckFailed.f("没有查询到[%s]-[%s]-[%s]对应的实例" % (category, name, namespace))
        return instance_info[0].id


class BaseBatchHandleInstance(BaseAPIViews):
    def init_handler(self, request, cc_app_id, project_id, slz_func):
        params = request.query_params
        params_slz = slz_func(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data
        project_kind, app_code = check_user_project(
            params_slz["access_token"], project_id, cc_app_id, self.jwt_info(request)
        )
        self.get_request_user(request, params_slz["access_token"], project_id)
        request.user.project_kind = project_kind

    def get_inst_info(self, inst_id):
        inst_info = InstanceConfig.objects.filter(id=inst_id, is_deleted=False)
        if not inst_info:
            raise error_codes.CheckFailed.f("没有查询到相应的记录", replace=True)
        return inst_info

    def get_project_info(self, access_token, project_id):
        project_info = paas_cc.get_project(access_token, project_id)
        if project_info.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(project_info.get("message"))
        return project_info["data"]

    def get_batch_inst_info(self, inst_id_list):
        inst_info = InstanceConfig.objects.filter(id__in=inst_id_list, is_deleted=False)
        if not inst_info:
            raise error_codes.CheckFailed.f("没有查询到相应的记录", replace=True)
        return {str(info.id): info for info in inst_info}

    def get_ns_variables(self, project_id, ns_id_data):
        default_variables = {}
        project_var = NameSpaceVariable.get_project_ns_vars(project_id)
        for ns_id in ns_id_data:
            ns_vars = []
            for _var in project_var:
                _ns_values = _var["ns_values"]
                _ns_value_ids = _ns_values.keys()
                ns_vars.append(
                    {
                        "id": _var["id"],
                        "key": _var["key"],
                        "name": _var["name"],
                        "value": _ns_values.get(ns_id) if ns_id in _ns_value_ids else _var["default_value"],
                    }
                )
            default_variables[ns_id] = ns_vars
        return default_variables

    def category_map(self, kind, category):
        map_info = CATEGORY_MODULE_MAP["k8s"]
        real_category = map_info.get(category)
        if not real_category:
            raise error_codes.CheckFailed.f("类型[%s]不正确，请确认后重试" % category, replace=True)
        return real_category


class ScaleInstance(app_views.ScaleInstance, BaseHandleInstance):
    def api_post(self, request, cc_app_id, project_id, instance_id):
        self.init_handler(request, cc_app_id, project_id, instance_id, serializers.ScaleInstanceParamsSLZ)
        instance_id = self.get_instance_id(request, instance_id)
        return self.put(request, project_id, instance_id, "")


class BatchScaleInstance(BaseBatchHandleInstance, app_views.BaseAPI):
    def update_inst_conf(self, curr_inst, instance_num, inst_state):
        """更新配置"""
        try:
            conf = json.loads(curr_inst.config)
        except Exception as error:
            logger.error("解析出现异常，当前实例ID: %s, 详情: %s" % (curr_inst.id, error))
            return
        if curr_inst.category in app_constants.REVERSE_CATEGORY_MAP:
            conf["spec"]["replicas"] = int(instance_num)
        else:
            conf["spec"]["instance"] = int(instance_num)
        curr_inst.config = json.dumps(conf)
        curr_inst.ins_state = inst_state
        curr_inst.save()

    def get_rc_name_by_deployment(
        self, request, cluster_id, instance_name, project_id=None, project_kind=ProjectKindID, namespace=None
    ):
        """如果是deployment，需要现根据deployment获取到application name"""
        flag, resp = self.get_application_deploy_info(
            request,
            project_id,
            cluster_id,
            instance_name,
            category="deployment",
            project_kind=project_kind,
            field="data.application,data.application_ext",
            namespace=namespace,
        )
        if not flag:
            raise error_codes.APIError.f(resp.data.get("message"))
        ret_data = []
        for info in resp.get("data") or []:
            application = (info.get("data") or {}).get("application") or {}
            application_ext = (info.get("data") or {}).get("application") or {}
            if application:
                ret_data.append(application.get("name"))
            if application_ext:
                ret_data.append(application_ext.get("name"))
        # ret_data.append(instance_name)
        return ret_data

    def api_put(self, request, cc_app_id, project_id):
        self.init_handler(request, cc_app_id, project_id, serializers.BatchScaleInstanceParamsSLZ)
        # 查询相应的实例信息
        inst_id_info = dict(request.data)
        inst_id_list = inst_id_info.get("inst_id_list") or []
        if not inst_id_list:
            raise error_codes.CheckFailed.f("实例ID不能为空", replace=True)
        instance_num = request.GET.get("instance_num")
        if not instance_num or not str(instance_num).isdigit():
            raise error_codes.CheckFailed.f("参数instance_num必须为整数")
        project_kind = ProjectKindID
        batch_inst_info = self.get_batch_inst_info(inst_id_list)
        for inst_id in inst_id_list:
            curr_inst = batch_inst_info[str(inst_id)]

            conf = self.get_common_instance_conf(curr_inst)
            metadata = conf.get("metadata", {})
            labels = metadata.get("labels", {})
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            namespace = metadata.get("namespace")
            # 检测资源是否满足扩容条件
            # self.check_resource(request, project_id, conf, cluster_id, instance_num)
            name = metadata.get("name")
            if curr_inst.category == "deployment":
                app_name = self.get_rc_name_by_deployment(
                    request, cluster_id, name, project_id=project_id, project_kind=project_kind, namespace=namespace
                )
                name = app_name[0]
            resp = self.scale_instance(
                request,
                project_id,
                cluster_id,
                namespace,
                name,
                instance_num,
                kind=project_kind,
                category=curr_inst.category,
                data=conf,
            )
            inst_state = InsState.UPDATE_SUCCESS.value

            if resp.data.get("code") == ErrorCode.NoError:
                self.update_instance_record_status(
                    curr_inst,
                    oper_type=app_constants.SCALE_INSTANCE,
                    is_bcs_success=True,
                )
            else:
                raise error_codes.APIError.f(resp.get("message"))
            # 更新配置
            self.update_inst_conf(curr_inst, instance_num, inst_state)
        return JsonResponse({"message": "下发任务成功!", "code": 0})


class CreateInstance(app_views.CreateInstance, BaseHandleInstance):
    def api_put(self, request, cc_app_id, project_id, instance_id):
        self.init_handler(request, cc_app_id, project_id, instance_id, serializers.CreateInstanceParamsSLZ)
        return self.post(request, project_id, instance_id)


class UpdateInstance(app_views.UpdateInstanceNew, BaseHandleInstance):
    def api_post(self, request, cc_app_id, project_id, instance_id):
        self.init_handler(request, cc_app_id, project_id, instance_id, serializers.UpdateInstanceParamsSLZ)
        return self.put(request, project_id, instance_id)


class BatchUpdateInstance(BaseBatchHandleInstance, app_views.BaseAPI):
    def get_params_from_gcloud(self, request, project_id, category):
        """标准运维请求来的参数格式"""
        ns_id_name_map, ns_name_id_map = self.get_ns_id_name_map(request.user.token.access_token, project_id)
        all_data = dict(request.data)
        inst_variables = all_data.get("inst_variables")

        if not isinstance(inst_variables, dict):
            raise error_codes.CheckFailed.f("变量[inst_variables]必须为DICT类型", replace=True)
        inst_id_list = []
        all_ns_id_list = []
        for inst_name, inst_info in inst_variables.items():
            ns_id_list = []
            ns_list = inst_info.get("namespace_list")
            if not inst_info.get("version_name"):
                raise error_codes.CheckFailed.f("参数[version_name]不能为空")
            if not ns_list:
                raise error_codes.CheckFailed.f("参数[namespace_list]不能为空", replace=True)
            for info in inst_info["namespace_list"]:
                ns_id = ns_name_id_map.get(info)
                if not ns_id:
                    raise error_codes.CheckFailed.f("命名空间[%s]不存在" % info, replace=True)
                all_ns_id_list.append(ns_id)
                ns_id_list.append(ns_id)
            # 进行适配名称->ID
            inst_info = InstanceConfig.objects.filter(namespace__in=ns_id_list, name=inst_name, category=category)
            inst_id_list.extend([info.id for info in inst_info])

        ns_var_map = self.get_ns_variables(project_id, all_ns_id_list)
        ns_var_map_new = {}
        for ns_id, var_info in ns_var_map.items():
            for info in var_info:
                item = {info["key"]: info["value"]}
                if ns_id in ns_var_map_new:
                    ns_var_map_new[ns_id].update(item)
                else:
                    ns_var_map_new[ns_id] = item
        return inst_id_list, inst_variables, ns_id_name_map, ns_var_map_new, None

    def get_params(self, request):
        """获取参数"""
        version_id = request.GET.get("version_id")
        if not version_id:
            raise error_codes.CheckFailed.f("升级的版本信息不能为空", replace=True)
        instance_num = request.GET.get("instance_num")
        category = request.GET.get("category")
        if category not in ["K8sDeployment", "K8sStatefulSet"] and (
            not instance_num or not str(instance_num).isdigit()
        ):
            raise error_codes.CheckFailed.f("实例数量必须为正数")
        # 获取环境变量，如果有未赋值的，则认为参数不正确
        data = dict(request.data)
        variables = data.get("variable") or {}
        inst_variables = data.get("inst_variables") or {}
        # if variables:
        #     for key, val in variables.items():
        #         if not val:
        #             raise error_codes.CheckFailed.f("环境变量 %s 对应的值不能为空" % key)
        inst_id_list = data.get("inst_id_list") or []
        if not inst_id_list:
            raise error_codes.CheckFailed.f("实例ID不能为空", replace=True)
        return version_id, instance_num, variables, inst_id_list, inst_variables

    def check_selector(self, old_conf, new_conf):
        """检查selector是否一致，如果不一致，则提示用户处理"""
        old_item = ((old_conf.get("spec") or {}).get("selector") or {}).get("matchLabels") or {}
        new_item = ((new_conf.get("spec") or {}).get("selector") or {}).get("matchLabels") or {}
        if old_item != new_item:
            raise error_codes.CheckFailed.f("Selector不一致不能进行更新，请确认!", replace=True)

    def generate_conf(
        self,
        request,
        project_id,
        inst_info,
        show_version_info,
        ns,
        tmpl_entity,
        category,
        instance_num,
        kind,
        variables,
    ):
        """生成配置文件"""
        params = {
            "instance_id": inst_info.instance_id,
            "version_id": show_version_info.real_version_id,
            "show_version_id": show_version_info.id,
            "template_id": show_version_info.template_id,
            "project_id": project_id,
            "access_token": request.user.token.access_token,
            "username": request.user.username,
            "lb_info": {},
            "variable_dict": variables,
            "is_preview": False,
        }
        conf = self.generate_ns_config_info(request, inst_info.namespace, tmpl_entity, params, is_save=False)
        new_conf = conf[category][0]["config"]
        if category in ["K8sDeployment", "K8sStatefulSet"]:
            new_conf["spec"]["replicas"] = instance_num
        return new_conf

    def get_tmpl_ids(self, version_id, category):
        """获取模板版本ID"""
        info = VersionedEntity.objects.filter(id=version_id)
        if not info:
            raise error_codes.CheckFailed.f("没有查询到相应的版本信息", replace=True)
        curr_info = info[0]
        entity = json.loads(curr_info.entity)
        return entity.get(category)

    def get_tmpl_entity(self, category, ids, name):
        """获取模板信息"""
        id_list = ids.split(",")
        tmpl_info = MODULE_DICT[category].objects.filter(id__in=id_list, name=name)
        if not tmpl_info:
            raise error_codes.CheckFailed.f("没有查询到模板信息", replace=True)
        return {category: [tmpl_info[0].id]}

    def generate_ns_config_info(self, request, ns_id, inst_entity, params, is_save=True):
        """生成某一个命名空间下的配置"""
        inst_conf = inst_utils.generate_namespace_config(ns_id, inst_entity, is_save, **params)
        return inst_conf

    def get_show_version_info(self, version_id):
        """查询show version信息"""
        info = ShowVersion.objects.filter(id=version_id)
        if not info:
            raise error_codes.CheckFailed.f("没有查询到展示版本信息", replace=True)
        return info[0]

    def update_conf_version(
        self,
        inst_id,
        inst_conf,
        inst_state,
        last_config,
        inst_version_id,
        show_ver_id,
        show_ver_name,
        new_tmpl_id,
        inst_name,
        category,
        real_ver_id,
        old_variables,
        new_variables,
    ):
        """更新配置"""
        try:
            # 更改版本
            tmpl_id_list = MODULE_DICT[category].objects.filter(name=inst_name).values_list("id", flat=True)
            inst_version_info = VersionInstance.objects.filter(id=inst_version_id)
            save_conf = {
                "old_conf": json.loads(last_config),
                "old_version_id": inst_version_info[0].version_id,
                "old_show_version_id": inst_version_info[0].show_version_id,
                "old_show_version_name": inst_version_info[0].show_version_name,
                "old_variables": json.loads(old_variables),
            }
            entity = json.loads(inst_version_info[0].instance_entity)
            category_content = entity[category]
            save_conf["old_category_entity"] = category_content
            # 获取模板交集
            new_category_content = list(set(category_content) - (set(tmpl_id_list) & set(category_content)))
            new_category_content.extend(new_tmpl_id)
            entity.update({category: new_category_content})
            inst_version_info.update(
                version_id=real_ver_id,
                instance_entity=json.dumps(entity),
                show_version_id=show_ver_id,
                show_version_name=show_ver_name,
            )
            InstanceConfig.objects.filter(id=inst_id).update(
                config=json.dumps(inst_conf),
                ins_state=inst_state,
                last_config=json.dumps(save_conf),
                variables=json.dumps(new_variables),
            )
        except Exception as error:
            logger.error("更新配置出现异常, 实例ID: %s, 详情: %s" % (inst_id, error))
            raise error_codes.DBOperError.f("更新实例配置异常!")

    def api_put(self, request, cc_app_id, project_id):
        self.init_handler(request, cc_app_id, project_id, serializers.BatchUpdateInstanceParamsSLZ)
        all_category_list = ["Deployment", "DaemonSet", "Job", "StatefulSet"]
        req_category = request.GET.get("category")
        if req_category not in all_category_list:
            raise error_codes.CheckFailed.f("类型[%s]不存在，请确认后重试" % req_category)
        project_kind = ProjectKindID
        real_category = self.category_map(project_kind, req_category)
        # 查询相应的实例信息
        version_id = None
        ns_var_map = {}
        if request.user.app_code in ["gcloud", "workbench"]:
            inst_id_list, inst_variables, ns_id_name_map, ns_var_map, instance_num = self.get_params_from_gcloud(
                request, project_id, real_category
            )
        else:
            version_id, instance_num, variables, inst_id_list, inst_variables = self.get_params(request)
        batch_inst_info = self.get_batch_inst_info(inst_id_list)
        for inst_id in inst_id_list:
            # 获取实例信息
            inst_info = batch_inst_info[str(inst_id)]

            category = inst_info.category
            inst_name = inst_info.name
            # 获取namespace
            inst_conf = json.loads(inst_info.config)
            metadata = inst_conf.get("metadata", {})
            labels = metadata.get("labels", {})

            cluster_id = labels.get("io.tencent.bcs.clusterid")
            template_id = labels.get("io.tencent.paas.templateid")
            if request.user.app_code in ["gcloud", "workbench"]:
                curr_inst_info_detail = inst_variables.get(inst_name)
                if not curr_inst_info_detail:
                    continue
                show_version_info = ShowVersion.objects.filter(
                    name=curr_inst_info_detail["version_name"], template_id=template_id
                )
                if not show_version_info:
                    raise error_codes.CheckFailed.f("版本信息[%s]不存在" % curr_inst_info_detail["version_name"])
                version_id = show_version_info[0].id
            namespace = metadata.get("namespace")
            pre_instance_num = 0
            if category in ["Deployment", "StatefulSet"]:
                pre_instance_num = inst_conf["spec"]["replicas"]
            else:
                pre_instance_num = inst_conf["spec"]["instance"]
            # 获取版本配置
            show_version_info = self.get_show_version_info(version_id)
            ids = self.get_tmpl_ids(show_version_info.real_version_id, category)
            tmpl_entity = self.get_tmpl_entity(category, ids, inst_name)
            real_instance_num = pre_instance_num
            # 获取当前实例对应的变量
            curr_variables = json.loads(inst_info.variables) if inst_info.variables else {}
            if request.user.app_code in ["gcloud", "workbench"]:
                ns_var_info = ns_var_map.get(int(inst_info.namespace)) or {}
                curr_variables.update(ns_var_info)
            if inst_variables:
                if request.user.app_code in ["gcloud", "workbench"]:
                    curr_inst_info_detail = inst_variables[inst_name]
                    if namespace in curr_inst_info_detail.get("namespace_list") or []:
                        curr_variables.update(curr_inst_info_detail.get("variables") or {})
                else:
                    if inst_id in inst_variables:
                        curr_variables.update(inst_variables.get(inst_id) or {})
            if instance_num:
                real_instance_num = int(instance_num)
            if instance_num is None and curr_variables.get("instance_num"):
                real_instance_num = curr_variables["instance_num"]
            new_inst_conf = self.generate_conf(
                request,
                project_id,
                inst_info,
                show_version_info,
                namespace,
                tmpl_entity,
                category,
                real_instance_num,
                project_kind,
                curr_variables,
            )
            self.check_selector(inst_conf, new_inst_conf)
            resp = self.update_deployment(
                request,
                project_id,
                cluster_id,
                namespace,
                new_inst_conf,
                kind=project_kind,
                category=category,
                app_name=inst_name,
            )
            inst_state = InsState.UPDATE_SUCCESS.value
            if resp.data.get("code") == ErrorCode.NoError:
                inst_state = InsState.UPDATE_FAILED.value
                self.update_instance_record_status(
                    inst_info, oper_type=app_constants.ROLLING_UPDATE_INSTANCE, is_bcs_success=True
                )
            # 更新conf
            self.update_conf_version(
                inst_id,
                new_inst_conf,
                inst_state,
                inst_info.config,
                inst_info.instance_id,
                version_id,
                show_version_info.name,
                tmpl_entity[category],
                inst_info.name,
                inst_info.category,
                show_version_info.real_version_id,
                inst_info.variables,
                curr_variables,
            )
        return JsonResponse({"message": "下发任务成功!", "code": 0})


class RecreateInstance(app_views.ReCreateInstance, BaseHandleInstance):
    def api_put(self, request, cc_app_id, project_id, instance_id):
        self.init_handler(request, cc_app_id, project_id, instance_id, serializers.UpdateInstanceParamsSLZ)
        return self.post(request, project_id, instance_id, None)


class BatchDeleteInstance(BaseBatchHandleInstance, app_views.BaseAPI):
    def delete_instance_oper(
        self,
        request,
        cluster_id,
        ns_name,
        instance_name,
        project_id=None,
        category="application",
        kind=2,
        inst_id_list=[],
    ):
        """删除实例"""
        resp = self.delete_instance(
            request,
            project_id,
            cluster_id,
            ns_name,
            instance_name,
            category=category,
            kind=kind,
            inst_id_list=inst_id_list,
        )
        logger.error("curr_error: %s" % resp.data)
        if resp.data.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.data.get("message"))

        return resp

    def get_inst_id_list(self, access_token, project_id, inst_info, project_kind):
        """通过名称获取实例ID"""
        inst_id_list = []
        id_name_map, name_id_map = self.get_ns_id_name_map(access_token, project_id)
        for info in inst_info:
            ns_id = name_id_map.get(info["namespace"])
            if not ns_id:
                raise error_codes.CheckFailed.f("命名空间[%s]不存在，请确认后重试" % info["namespace"])
            category = self.category_map(project_kind, info["category"])
            infos = InstanceConfig.objects.filter(category=category, namespace=ns_id, name=info["inst_name"])
            if not infos:
                raise error_codes.CheckFailed.f(
                    "应用(%s-%s-%s)不存在，请确认后重试" % (info["category"], info["namespace"], info["inst_name"])
                )
            inst_id_list.append(infos[0].id)
        return inst_id_list

    def api_delete(self, request, cc_app_id, project_id):
        self.init_handler(request, cc_app_id, project_id, serializers.BatchDeleteInstanceParamsSLZ)
        project_kind = ProjectKindID
        data = dict(request.data)
        inst_id_list = data.get("inst_id_list") or []
        inst_info = data.get("inst_info")
        if not inst_id_list and inst_info:
            inst_id_list = self.get_inst_id_list(request.user.token.access_token, project_id, inst_info, project_kind)
        if not inst_id_list:
            raise error_codes.CheckFailed.f("实例参数不能为空", replace=True)
        batch_inst_info = self.get_batch_inst_info(inst_id_list)
        for inst_id in inst_id_list:
            # 获取实例信息
            curr_inst = batch_inst_info[str(inst_id)]

            conf = self.get_common_instance_conf(curr_inst)
            metadata = conf.get("metadata", {})
            labels = metadata.get("labels", {})
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            namespace = metadata.get("namespace")
            name = metadata.get("name")
            category = curr_inst.category
            self.delete_instance_oper(
                request,
                cluster_id,
                namespace,
                name,
                project_id=project_id,
                category=category,
                kind=project_kind,
                inst_id_list=[inst_id],
            )

        return JsonResponse({"message": "下发任务成功!", "code": 0})


class BatchRecreateInstance(BaseBatchHandleInstance, app_views.BaseAPI):
    def delete_instance_oper(
        self, request, cluster_id, ns_name, instance_name, project_id=None, category="application", kind=2
    ):
        """删除实例"""
        resp = self.delete_instance(
            request, project_id, cluster_id, ns_name, instance_name, category=category, kind=kind
        )
        logger.error("curr_error: %s" % resp.data)
        if resp.data.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.data.get("message"))

        return resp

    def api_post(self, request, cc_app_id, project_id):
        """批量重建实例"""
        self.init_handler(request, cc_app_id, project_id, serializers.BatchReCreateInstanceParamsSLZ)
        # 查询相应的实例信息
        inst_id_info = dict(request.data)
        inst_id_list = inst_id_info.get("inst_id_list") or []
        if not inst_id_list:
            raise error_codes.CheckFailed.f("实例ID不能为空", replace=True)
        project_kind = ProjectKindID
        for inst_id in inst_id_list:
            inst_info = self.get_inst_info(inst_id)
            # 获取namespace
            curr_inst = inst_info[0]

            conf = self.get_common_instance_conf(curr_inst)
            metadata = conf.get("metadata", {})
            labels = metadata.get("labels", {})
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            namespace = metadata.get("namespace")
            name = metadata.get("name")
            category = curr_inst.category
            if not curr_inst.is_deleted:
                self.delete_instance_oper(
                    request, cluster_id, namespace, name, project_id=project_id, category=category, kind=project_kind
                )
            # 更新instance 操作
            try:
                self.update_instance_record_status(
                    curr_inst, app_constants.REBUILD_INSTANCE, created=datetime.now(), is_bcs_success=True
                )
            except Exception as error:
                logger.error("更新重建操作实例状态失败，详情: %s" % error)

            from backend.celery_app.tasks import application as app_model

            # 启动任务
            app_model.application_polling_task.delay(
                request.user.token.access_token,
                curr_inst.id,
                cluster_id,
                name,
                category,
                project_kind,
                namespace,
                project_id,
                username=request.user.username,
            )

        return JsonResponse({"message": "下发任务成功!", "code": 0})


class CancelInstance(app_views.CancelUpdateInstance, BaseHandleInstance):
    def api_post(self, request, cc_app_id, project_id, instance_id):
        self.init_handler(request, cc_app_id, project_id, instance_id, serializers.UpdateInstanceParamsSLZ)
        return self.put(request, project_id, instance_id, None)


class PauseInstance(app_views.PauseUpdateInstance, BaseHandleInstance):
    def api_post(self, request, cc_app_id, project_id, instance_id):
        self.init_handler(request, cc_app_id, project_id, instance_id, serializers.UpdateInstanceParamsSLZ)
        return self.put(request, project_id, instance_id, None)


class ResumeInstance(app_views.ResumeUpdateInstance, BaseHandleInstance):
    def api_post(self, request, cc_app_id, project_id, instance_id):
        self.init_handler(request, cc_app_id, project_id, instance_id, serializers.UpdateInstanceParamsSLZ)
        return self.put(request, project_id, instance_id, None)


class GetInstanceVersions(BaseAPIViews):
    def get_inst_version_info(self, inst_version_id):
        """获取实例版本信息"""
        inst_version_info = VersionInstance.objects.filter(id=inst_version_id)
        if not inst_version_info:
            raise error_codes.APIError.f("没有查询到实例版本", replace=True)
        return inst_version_info[0]

    def get_show_version_info(self, id_list):
        """获取展示的版本信息"""
        show_version_info = ShowVersion.objects.filter(real_version_id__in=id_list, is_deleted=False).order_by(
            "-updated"
        )
        if not show_version_info:
            raise error_codes.APIError.f("没有查询到实例版本", replace=True)
        ret_data = []
        for info in show_version_info:
            ret_data.append(
                {
                    "id": info.id,
                    "real_version_id": info.template_id,
                    "name": info.name,
                    "template_id": info.template_id,
                    "updated_time": info.updated,
                }
            )
        return ret_data

    def get_version_entity(self, tmpl_id, category, tmpl_id_list):
        """获取所有对应的版本"""
        ret_data = []
        version_entity = VersionedEntity.objects.filter(template_id=tmpl_id)
        tmpl_id_list = [str(id) for id in tmpl_id_list]
        for info in version_entity:
            entity = json.loads(info.entity)
            category_entity = entity.get(category) or ""
            category_id_list = category_entity.split(",")
            if set(tmpl_id_list) & set(category_id_list):
                ret_data.append(info.id)
        return ret_data

    def get_category_tmpl(self, category, name):
        """获取相同名称的模板"""
        return MODULE_DICT[category].objects.filter(name=name).values_list("id", flat=True)

    def get(self, request, cc_app_id, project_id, instance_id):
        """获取实例对应的版本"""
        inst_info = self.get_instance_info(instance_id)[0]
        category = inst_info.category
        inst_name = inst_info.name
        inst_version_id = inst_info.instance_id
        inst_version_info = self.get_inst_version_info(inst_version_id)
        muster_id = inst_version_info.template_id
        category_tmpl_id_list = self.get_category_tmpl(category, inst_name)
        version_id_list = self.get_version_entity(muster_id, category, category_tmpl_id_list)
        # 查询显示的版本
        ret_data = self.get_show_version_info(version_id_list)

        return JsonResponse({"data": ret_data, "code": 0})


class GetInstanceVersionConf(BaseAPIViews):
    def json2yaml(self, conf):
        """json转yaml"""
        yaml_profile = yaml.safe_dump(conf)
        return yaml_profile

    def get_show_version_info(self, show_version_id):
        """获取展示版本ID"""
        show_version = ShowVersion.objects.filter(id=show_version_id)
        if not show_version:
            raise error_codes.CheckFailed.f("没有查询到展示版本信息")
        return show_version[0].real_version_id

    def get_version_info(self, category, version_id):
        """获取真正的版本信息"""
        info = VersionedEntity.objects.filter(id=version_id)
        if not info:
            raise error_codes.CheckFailed.f("没有查询到展示版本信息")
        curr_info = info[0]
        entity = json.loads(curr_info.entity)
        category_id_list = (entity.get(category) or "").split(",")
        return category_id_list

    def get_tmpl_info(self, category, category_id_list, name):
        """根据名称获取相应的配置"""
        return MODULE_DICT[category].objects.filter(id__in=category_id_list, name=name)

    def get_variables(self, request, project_id, config, cluster_id, ns_id):
        """获取变量"""
        key_list = check_var_by_config(config)
        key_list = list(set(key_list))
        v_list = []
        if key_list:
            # 验证变量名是否符合规范，不符合抛出异常，否则后续用 django 模板渲染变量也会抛出异常

            var_objects = Variable.objects.filter(Q(project_id=project_id) | Q(project_id=0))

            for _key in key_list:
                key_obj = var_objects.filter(key=_key)
                if key_obj.exists():
                    _obj = key_obj.first()
                    # 只显示自定义变量
                    if _obj.category == "custom":
                        v_list.append(
                            {
                                "key": _obj.key,
                                # "name": _obj.name,
                                "value": _obj.get_show_value(cluster_id, ns_id),
                            }
                        )
                else:
                    v_list.append(
                        {
                            "key": _key,
                            # "name": _key,
                            "value": "",
                        }
                    )
        return v_list

    def get_variables_new(self, request, project_id, config, cluster_id, ns_id_list):
        """根据命名空间获取变量"""
        key_list = check_var_by_config(config)
        key_list = list(set(key_list))
        variable_dict = {}
        if key_list:
            # 验证变量名是否符合规范，不符合抛出异常，否则后续用 django 模板渲染变量也会抛出异常

            var_objects = Variable.objects.filter(Q(project_id=project_id) | Q(project_id=0))

            access_token = request.user.token.access_token
            namespace_res = paas_cc.get_namespace_list(access_token, project_id, limit=10000)
            namespace_data = namespace_res.get("data", {}).get("results") or []
            namespace_dict = {str(i["id"]): i["cluster_id"] for i in namespace_data}
            ns_list = ns_id_list
            for ns_id in ns_list:
                _v_list = []
                for _key in key_list:
                    key_obj = var_objects.filter(key=_key)
                    if key_obj.exists():
                        _obj = key_obj.first()
                        # 只显示自定义变量
                        if _obj.category == "custom":
                            cluster_id = namespace_dict.get(ns_id, 0)
                            _v_list.append({"key": _obj.key, "value": _obj.get_show_value(cluster_id, ns_id)})
                    else:
                        _v_list.append({"key": _key, "value": ""})
                variable_dict[ns_id] = _v_list
        return variable_dict

    def get_default_instance_value(self, variable_list, key):
        ret_val = ""
        for info in variable_list:
            if info["key"] == key:
                return info.get("value") or ""
        return ret_val

    def init_request_obj(self, request):
        request.user = APIUser
        request.user.token.access_token = request.GET.get("access_token")
        request.user.username = "gcloud"

    def get_ns_id_list(self, request, add_inst_ns=False, inst_ns_id=None):
        ns_id_list = re.findall(r"[^,;]+", request.GET.get("ns_ids", ""))
        if add_inst_ns and inst_ns_id:
            ns_id_list.append(inst_ns_id)
        return ns_id_list

    def get(self, request, cc_app_id, project_id, instance_id):
        """获取当前实例的版本配置信息
        添加针对不同namespace的处理
        """
        # 初始化request
        self.init_request_obj(request)

        access_token = request.GET.get("access_token")
        project_info = paas_cc.get_project(access_token, project_id)
        if project_info.get("code") != 0:
            raise error_codes.APIError.f(project_info.get("message"))
        inst_info = self.get_instance_info(instance_id)[0]
        category = inst_info.category
        # inst_name = inst_info.name
        inst_version_id = inst_info.instance_id
        inst_conf = json.loads(inst_info.config)
        cluster_id = inst_conf.get("metadata", {}).get("labels", {}).get("io.tencent.bcs.clusterid")
        inst_version_info = VersionInstance.objects.filter(id=inst_version_id)
        if not inst_version_info:
            raise error_codes.APIError.f("没有查询到实例版本", replace=True)
        show_version_id = request.GET.get("show_version_id")
        app_category_id_list = []
        if show_version_id:
            real_version_id = self.get_show_version_info(show_version_id)
            category_id_list = self.get_version_info(category, real_version_id)
        else:
            instance_entity = json.loads(inst_version_info[0].instance_entity)
            # 获取模板
            category_id_list = instance_entity.get(category) or []
        all_info = self.get_tmpl_info(category, category_id_list, inst_info.name)

        if not all_info:
            raise error_codes.CheckFailed.f("没有查询到版本信息", replace=True)
        # 通过展示版本获取真正版本
        curr_info = all_info[0]
        version_conf = curr_info.config
        instance_num_var_flag = False
        variable_map = []
        if show_version_id:
            # 获取namespace id列表
            ns_id_list = self.get_ns_id_list(request)
            if ns_id_list:
                variable_map = self.get_variables_new(request, project_id, version_conf, cluster_id, ns_id_list)
            variables = self.get_variables(request, project_id, version_conf, cluster_id, inst_info.namespace)
            spec_info = json.loads(version_conf).get("spec") or {}
            instance_num = spec_info.get("replicas") or spec_info.get("instance")
        else:
            spec_info = inst_conf.get("spec") or {}
            instance_num = spec_info.get("replicas") or spec_info.get("instance")
            variables = []
            variables_dict = json.loads(inst_info.variables)
            if isinstance(variables_dict, dict):
                variables = [{"key": key, "value": val} for key, val in variables_dict.items()]
        instance_num_key = ""
        if not str(instance_num).isdigit():
            instance_num_var_flag = True
            instance_num_key = re.findall(r"[^\{\}]+", instance_num)[-1]
            instance_num = self.get_default_instance_value(variables, instance_num_key)
        config_profile = json.loads(version_conf)
        # yaml_conf = self.json2yaml(config_profile)
        return JsonResponse(
            {
                "code": 0,
                "data": {
                    "json": config_profile,
                    # "yaml": yaml_conf,
                    "variable": variables,
                    "instance_num": instance_num,
                    "instance_num_key": instance_num_key,
                    "instance_num_var_flag": instance_num_var_flag,
                    "variable_map": variable_map,
                },
            }
        )


class BaseProjectMuster(BaseAPIViews):
    def init_handler(self, request, cc_app_id, project_id):
        params = request.query_params
        params_slz = serializers.BaseProjectParamsSLZ(data=params)
        params_slz.is_valid(raise_exception=True)
        params_slz = params_slz.data
        project_kind, app_code = check_user_project(
            params_slz["access_token"], project_id, cc_app_id, self.jwt_info(request)
        )
        request.user = APIUser
        request.user.token.access_token = params_slz["access_token"]
        request.user.username = app_code
        request.user.project_kind = project_kind

    def split_res(self, res):
        return re.findall(r"[^,;]+", res)

    def template_handler(self, project_kind, entity):
        deployment = self.split_res(entity.get("K8sDeployment") or "")
        daemonset = self.split_res(entity.get("K8sDaemonSet") or "")
        job = self.split_res(entity.get("K8sJob") or "")
        statefulset = self.split_res(entity.get("K8sStatefulSet") or "")
        return {
            "K8sDeployment": deployment,
            "K8sDaemonSet": daemonset,
            "K8sJob": job,
            "K8sStatefulSet": statefulset,
        }


class ProjectMuster(BaseProjectMuster):
    def get_project_tmpl_set(self, project_id):
        try:
            data = Template.objects.filter(project_id=project_id).values("id", "name")
        except Exception as err:
            logger.error("Query db error, detail: %s" % err)
            data = []
        return data

    def get(self, request, cc_app_id, project_id):
        """获取项目下的所有模板集"""
        self.init_handler(request, cc_app_id, project_id)
        self.get_request_user(request, request.GET.get("access_token"), project_id)

        ret_data = list(self.get_project_tmpl_set(project_id))
        # TODO 调整为带 web_annotations 的新协议
        resources_actions_allowed = TemplatesetPermission().resources_actions_allowed(
            res=[tpl['id'] for tpl in ret_data],
            action_ids=[
                TemplatesetAction.CREATE,
                TemplatesetAction.DELETE,
                TemplatesetAction.VIEW,
                TemplatesetAction.UPDATE,
                TemplatesetAction.INSTANTIATE,
            ],
            perm_ctx=TemplatesetPermCtx(username=request.user.username, project_id=project_id),
        )

        for tpl in ret_data:
            action_allowed = resources_actions_allowed[tpl['id']]
            tpl['permissions'] = {
                'create': action_allowed[TemplatesetAction.CREATE],
                'delete': action_allowed[TemplatesetAction.DELETE],
                'view': action_allowed[TemplatesetAction.VIEW],
                'edit': action_allowed[TemplatesetAction.UPDATE],
                'use': action_allowed[TemplatesetAction.INSTANTIATE],
            }

        return JsonResponse({"code": 0, "data": ret_data})


class ProjectMusterVersion(BaseProjectMuster):
    def get(self, request, cc_app_id, project_id, muster_id):
        self.init_handler(request, cc_app_id, project_id)
        show_version_list = []
        show_sets = ShowVersion.objects.filter(template_id=muster_id)
        for _s in show_sets:
            show_version_list.append(
                {"id": _s.real_version_id, "show_version_id": _s.id, "show_version_name": _s.name, "version": _s.name}
            )

        return JsonResponse({"code": 0, "data": show_version_list})


class ProjectMusterTemplate(BaseProjectMuster):
    def get_version_entity(self):
        version_info = VersionedEntity.objects.filter(id=self.pk)
        if not version_info:
            raise error_codes.CheckFailed.f("没有查询到版本信息", replace=True)
        return version_info[0]

    def get_template_info(self, tmpl_info):
        ret_res_info = {}
        for tmpl_type, id_list_str in tmpl_info.items():
            id_list = self.split_res(id_list_str)
            ret_res_info[tmpl_type] = list(MODULE_DICT[tmpl_type].objects.filter(id__in=id_list).values("id", "name"))
        return ret_res_info

    def get(self, request, cc_app_id, project_id, version_id):
        self.init_handler(request, cc_app_id, project_id)
        self.pk = version_id
        version_info = self.get_version_entity()
        return JsonResponse({"code": 0, "data": version_info.get_version_instance_resources()})


class ProjectNamespace(BaseProjectMuster, app_views.BaseAPI):
    def get_project_namespace(self, request, project_id):
        resp = paas_cc.get_namespace_list(request.user.token.access_token, project_id, desire_all_data=1)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"), replace=True)
        data = resp.get("data") or {}
        results = data.get("results") or []
        if not results:
            raise error_codes.APIError.f("当前项目下没有命名空间，请确认")
        return results

    def get_used_namespace(self, request):
        req_data = dict(request.data)
        resource_name_info = req_data.get("resource_name_info") or {}
        # 过滤掉占用的命名空间
        ret_data = []
        for category, name_list in resource_name_info.items():
            obj = InstanceConfig.objects.filter(category=category, name__in=name_list, is_deleted=False).exclude(
                ins_state=InsState.NO_INS.value
            )
            ret_data.extend([int(info.namespace) for info in obj])
        return ret_data

    def compose_cluster_env(self, request, project_id, ns_data):
        cluster_map = self.get_cluster_id_env(request, project_id)
        for info in ns_data:
            info["environment"] = cluster_map.get(info["cluster_id"], {}).get("cluster_env_str")

    def post(self, request, cc_app_id, project_id):
        """获取项目下，集群+命名空间
        注: 只返回没有使用的命名空间
        """
        self.init_handler(request, cc_app_id, project_id)
        self.get_request_user(request, request.GET.get("access_token"), project_id)
        ns_data = self.get_project_namespace(request, project_id)
        used_namespace = self.get_used_namespace(request)
        ns_data_new = []
        if ns_data:
            project_var = NameSpaceVariable.get_project_ns_vars(project_id)
            for i in ns_data:
                ns_id = i["id"]
                if ns_id in used_namespace:
                    continue
                ns_vars = []
                for _var in project_var:
                    _ns_values = _var["ns_values"]
                    _ns_value_ids = _ns_values.keys()
                    ns_vars.append(
                        {
                            "id": _var["id"],
                            "key": _var["key"],
                            "name": _var["name"],
                            "value": _ns_values.get(ns_id) if ns_id in _ns_value_ids else _var["default_value"],
                        }
                    )
                i["ns_vars"] = ns_vars
                ns_data_new.append(i)

        self.compose_cluster_env(request, project_id, ns_data_new)
        return JsonResponse({"code": 0, "data": ns_data_new})
