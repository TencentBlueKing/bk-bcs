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

NOTE: 模板实例化和编辑模板都是跳转到先前的页面，接口废除
查询application的信息中
buildedInstance: 构建的数量(包含成功和失败Instance)
instance: 用户希望构建的数量
runningInstance: 构建成功的数量
"""
import copy
import json
import logging
from datetime import datetime

from django.utils.translation import ugettext_lazy as _
from rest_framework.renderers import BrowsableAPIRenderer

from backend.bcs_web.audit_log import client
from backend.celery_app.tasks.application import update_create_error_record
from backend.container_service.projects.base.constants import ProjectKindID
from backend.iam.permissions.decorators import response_perms
from backend.iam.permissions.resources.namespace import NamespaceRequest, calc_iam_ns_id
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedAction, NamespaceScopedPermission
from backend.iam.permissions.resources.templateset import TemplatesetAction, TemplatesetPermission, TemplatesetRequest
from backend.templatesets.legacy_apps.configuration.models import MODULE_DICT, ShowVersion, Template, VersionedEntity
from backend.templatesets.legacy_apps.instance import utils as inst_utils
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import (
    InstanceConfig,
    InstanceEvent,
    MetricConfig,
    VersionInstance,
)
from backend.utils.errcodes import ErrorCode
from backend.utils.renderers import BKAPIRenderer
from backend.utils.response import PermsResponse

from . import constants as app_constants
from . import utils
from .base_views import BaseAPI, InstanceAPI, error_codes
from .filters.base_metrics import BaseMusterMetric
from .other_views import k8s_views
from .utils import exclude_records

logger = logging.getLogger(__name__)

APPLICATION_CATEGORY = "application"
DEPLOYMENT_CATEGORY = "deployment"
K8SDEPLOYMENT_CATEGORY = "K8sDeployment"
K8SJOB_CATEGORY = "K8sJob"
K8SDAEMONSET_CATEGORY = "k8sDaemonSet"
K8SSTATEFULSET_CATEGORY = "K8sStatefulSet"
DEFAULT_ERROR_CODE = ErrorCode.UnknownError
ALL_LIMIT = 10000


class GetProjectMuster(BaseMusterMetric):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_muster(self, project_id, muster_id):
        """获取模板集"""
        all_muster_list = Template.objects.filter(is_deleted=False, project_id=project_id).order_by(
            "-updated", "-created"
        )
        if muster_id:
            all_muster_list = all_muster_list.filter(id=muster_id)
        all_muster_list = all_muster_list.values("id", "name")
        return all_muster_list

    def muster_tmpl_handler(self, muster_id_name_map, muster_num_map):
        ret_data = []
        for muster_id, info in muster_num_map.items():
            tmpl_num = len(info.get("deployment_name_list", [])) + len(info.get("application_name_list", []))
            if tmpl_num == 0:
                continue
            ret_data.append(
                {
                    "tmpl_muster_name": muster_id_name_map.get(muster_id, ""),
                    "tmpl_muster_id": muster_id,
                    "tmpl_num": tmpl_num,
                    "inst_num": info.get("inst_num", 0),
                }
            )
        return ret_data

    def get_version_instance(self, muster_id_list, cluster_env_map, cluster_type, app_name, ns_id):
        """查询version instance"""
        version_inst_info = (
            VersionInstance.objects.filter(template_id__in=muster_id_list, is_deleted=False)
            .values("id", "template_id", "version_id")
            .order_by("-updated", "-created")
        )
        # 数据匹配
        instance_id_list = [info["id"] for info in version_inst_info]
        # 过滤数据
        instance_info = (
            InstanceConfig.objects.filter(
                instance_id__in=instance_id_list,
                is_deleted=False,
                category__in=[APPLICATION_CATEGORY, DEPLOYMENT_CATEGORY],
            )
            .exclude(ins_state=InsState.NO_INS.value)
            .order_by("-updated", "-created")
        )
        if app_name:
            instance_info = instance_info.filter(name=app_name)
        if ns_id:
            instance_info = instance_info.filter(namespace=ns_id)
        instance_info = instance_info.values("instance_id", "name", "category", "config", "id")
        exist_inst_id_list = [info["instance_id"] for info in instance_info]
        ret_data = {}
        for info in version_inst_info:
            if info["id"] not in exist_inst_id_list:
                continue
            if info["template_id"] not in ret_data:
                ret_data[info["template_id"]] = {
                    "id_list": [info["id"]],
                    "inst_num": 0,
                    "application_name_list": set([]),
                    "deployment_name_list": set([]),
                }
            else:
                ret_data[info["template_id"]]["id_list"].append(info["id"])
        # 匹配数据
        for info in instance_info:
            config = json.loads(info["config"])
            cluster_id = ((config.get("metadata") or {}).get("labels") or {}).get("io.tencent.bcs.clusterid")
            muster_id = int(((config.get("metadata") or {}).get("labels") or {}).get("io.tencent.paas.templateid"))
            if not cluster_id:
                continue
            if str(cluster_env_map.get(cluster_id, {}).get("cluster_env")) != str(cluster_type):
                continue
            if ret_data.get(muster_id) and info["instance_id"] in ret_data[muster_id]["id_list"]:
                ret_data[muster_id]["inst_num"] += 1
                if info["category"] == APPLICATION_CATEGORY:
                    ret_data[muster_id]["application_name_list"] = ret_data[muster_id]["application_name_list"].union(
                        set([info.get("name")])
                    )
                else:
                    ret_data[muster_id]["deployment_name_list"] = ret_data[muster_id]["deployment_name_list"].union(
                        set([info.get("name")])
                    )
        return ret_data

    @response_perms(
        action_ids=[TemplatesetAction.INSTANTIATE],
        permission_cls=TemplatesetPermission,
        resource_id_key='tmpl_muster_id',
    )
    def get(self, request, project_id):
        """获取项目下的所有的模板集"""
        # 获取过滤参数
        cluster_type, app_status, muster_id, app_id, ns_id, request_cluster_id = self.get_filter_params(
            request, project_id
        )
        # 获取模板集
        all_muster_list = self.get_muster(project_id, muster_id)
        # 获取集群及环境
        cluster_env_map = self.get_cluster_id_env(request, project_id)
        # 获取实例名称
        app_name = self.get_inst_name(app_id)
        # 获取模板
        muster_id_list = [info["id"] for info in all_muster_list]
        category = request.GET.get("category")
        if not category:
            raise error_codes.CheckFailed(_("分类不能为空"))
        k8s_view_client = k8s_views.K8SMuster()
        ret_data = k8s_view_client.get(
            request,
            project_id,
            all_muster_list,
            muster_id_list,
            category,
            cluster_type,
            app_status,
            app_name,
            ns_id,
            cluster_env_map,
            request_cluster_id,
        )
        return PermsResponse(ret_data, TemplatesetRequest(project_id=project_id))


class GetMusterTemplate(BaseMusterMetric):
    def get_version_id_name_map(self, muster_id):
        """获取版本ID和名称的映射"""
        version_info = ShowVersion.objects.filter(template_id=muster_id, is_deleted=False).order_by("-updated")
        return {info.real_version_id: info.name for info in version_info}

    def get_version_map(self, muster_id):
        """获取版本映射，用于判断是否有更新"""
        version_info = ShowVersion.objects.filter(template_id=muster_id, is_deleted=False).order_by("-updated")

        real_version_id_list = [info.real_version_id for info in version_info]
        all_version_info = VersionedEntity.objects.filter(
            is_deleted=False, template_id=muster_id, id__in=real_version_id_list
        ).order_by("-updated")
        return {info.last_version_id: info.id for info in all_version_info if info.last_version_id != 0}

    def check_project_muster(self, project_id, muster_id):
        """检测项目和模板的关系"""
        if not Template.objects.filter(project_id=project_id, id=muster_id).exists():
            raise error_codes.RecordNotFound(_("模板集不属于项目!"))

    def get_tmpl_info(self, muster_id, category=None):
        """通过模板集获取模板"""
        muster_name = Template.objects.get(id=muster_id).name
        version_id_name_map = self.get_version_id_name_map(muster_id)
        version_info = VersionedEntity.objects.filter(
            is_deleted=False, template_id=muster_id, id__in=version_id_name_map.keys()
        ).order_by("-updated", "-created")
        version_id_map_list = []
        for info in version_info:
            # 获取application, deployment
            entity = info.get_entity() or {}
            application = entity.get("application")
            application_list = application.split(",") if application else []
            deployment = entity.get("deployment")
            deployment_list = deployment.split(",") if deployment else []
            other_list = []
            if category:
                other = entity.get(k8s_views.CATEGORY_MAP[category])
                other_list = other.split(",") if other else []
            version_id_map_list.append(
                {
                    "tmpl_muster_id": info.template_id,
                    "tmpl_muster_name": muster_name,
                    "version_id": info.id,
                    "version": version_id_name_map.get(info.id, ""),
                    "last_version_id": info.last_version_id,
                    "last_version": version_id_name_map.get(info.last_version_id, ""),
                    "application_list": application_list,
                    "deployment_list": deployment_list,
                    "other_list": other_list,
                }
            )
        return version_id_map_list

    def get_instance_conf(self, info):
        """获取配置"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析Instance config失败, ID: %s, 详情: %s" % (info.id, error))
            return {}
        return conf

    def compose_cluster_inst_ns(self, info, cluster_id, name, namespace, cluster_ns_inst, update_flag=False):
        if info.category in cluster_ns_inst.get(cluster_id, {}):
            cluster_ns_inst[cluster_id][info.category]["inst_list"].append(name)
            cluster_ns_inst[cluster_id][info.category]["ns_list"].append(namespace)
            cluster_ns_inst[cluster_id][info.category]["inst_ns_map"][namespace] = name
        else:
            item = {info.category: {"inst_list": [name], "ns_list": [namespace], "inst_ns_map": {namespace: name}}}
            if update_flag:
                cluster_ns_inst[cluster_id].update(item)
            else:
                cluster_ns_inst[cluster_id] = item

    def get_instances(self, muster_id, cluster_type, app_name, ns_id, cluster_env_map, request_cluster_id):
        """获取实例信息"""
        version_inst_id_info = (
            VersionInstance.objects.filter(is_deleted=False, template_id=muster_id)
            .values("id")
            .order_by("-updated", "-created")
        )
        version_inst_id_list = [info["id"] for info in version_inst_id_info]
        instance_info = (
            InstanceConfig.objects.filter(
                is_deleted=False, instance_id__in=version_inst_id_list, category__in=app_constants.ALL_CATEGORY_LIST
            )
            .exclude(ins_state=InsState.NO_INS.value)
            .order_by("-updated", "-created")
        )
        if app_name:
            instance_info = instance_info.filter(name=app_name)
        if ns_id:
            instance_info = instance_info.filter(namespace=ns_id)
        muster_tmpl_map = {}
        cluster_ns_inst = {}
        tmpl_create_error = {}
        exist_ns_name = []
        for info in instance_info:
            conf = self.get_instance_conf(info)
            metadata = conf.get("metadata", {})
            name = metadata.get("name", "")
            labels = metadata.get("labels", {})
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            namespace = metadata.get("namespace")
            if exclude_records(
                request_cluster_id,
                cluster_id,
                cluster_type,
                cluster_env_map.get(cluster_id, {}).get("cluster_env"),
            ):
                continue
            curr_key = "%s:%s" % (info.category, name)
            ns_name_key = "%s:%s" % (namespace, name)
            if curr_key not in muster_tmpl_map:
                muster_tmpl_map[curr_key] = 1
            else:
                muster_tmpl_map[curr_key] += 1
            # 注意去掉可能多个的情况(重复多次失败的情况)
            if not info.is_bcs_success:
                if ns_name_key not in exist_ns_name:
                    if curr_key in tmpl_create_error:
                        tmpl_create_error[curr_key] += 1
                    else:
                        exist_ns_name.append(ns_name_key)
                        tmpl_create_error[curr_key] = 1
                continue
            exist_ns_name.append(ns_name_key)
            # 组装集群、namespace、名称
            if cluster_id in cluster_ns_inst:
                self.compose_cluster_inst_ns(info, cluster_id, name, namespace, cluster_ns_inst, update_flag=True)
            else:
                self.compose_cluster_inst_ns(info, cluster_id, name, namespace, cluster_ns_inst)
        return muster_tmpl_map, cluster_ns_inst, tmpl_create_error

    def get_application_by_deployment(self, request, cluster_ns_inst, project_id, kind):
        """针对deployment获取对应的application"""
        ret_data = {}
        for cluster_id, val in cluster_ns_inst.items():
            if val.get("deployment"):
                flag, resp = self.get_app_deploy_with_post(
                    request,
                    project_id,
                    cluster_id,
                    ",".join(set(val["deployment"]["inst_list"])),
                    category=DEPLOYMENT_CATEGORY,
                    project_kind=kind,
                    namespace=",".join(set(val["deployment"]["ns_list"])),
                    field="data.application,data.application_ext,data.metadata",
                )
                if not flag:
                    raise error_codes.APIError.f(resp.data.get("message"))
                # 组装数据
                for val in resp.get("data") or []:
                    metadata = val.get("data", {}).get("metadata", {})
                    application = val.get("data", {}).get("application", {})
                    application_ext = val.get("data", {}).get("application_ext", {})
                    item_name = []
                    ns_list = []
                    if application:
                        item_name.append(application.get("name"))
                    if application_ext:
                        item_name.append(application_ext.get("name"))
                    ns_list.append(metadata.get("namespace"))
                    if cluster_ns_inst[cluster_id].get("application"):
                        cluster_ns_inst[cluster_id]["application"]["inst_list"].extend(item_name)
                        cluster_ns_inst[cluster_id]["application"]["ns_list"].extend(ns_list)
                    else:
                        cluster_ns_inst[cluster_id].update(
                            {"application": {"inst_list": item_name, "ns_list": ns_list}}
                        )
                    # 组装deployment和applcation的关系
                    if application:
                        key_name = "application:%s" % application.get("name")
                        ret_data[key_name] = "deployment:%s" % metadata.get("name")
                    if application_ext:
                        key_name = "application:%s" % application_ext.get("name")
                        ret_data[key_name] = "deployment:%s" % metadata.get("name")
        return ret_data

    def get_inst_status_for_tmpl(self, request, cluster_ns_inst, project_id, kind, application_deploy_map):
        """查询实例状态，以判断模板下应用是否存在异常"""
        # 根据application查询状态，如果有异常则计数异常数量
        ret_data = {}
        for cluster_id, info in cluster_ns_inst.items():
            if not info.get("application"):
                continue
            flag, resp = self.get_app_deploy_with_post(
                request,
                project_id,
                cluster_id,
                ",".join(set(info["application"]["inst_list"])),
                category=APPLICATION_CATEGORY,
                project_kind=kind,
                namespace=",".join(set(info["application"]["ns_list"])),
            )
            if not flag:
                raise error_codes.APIError.f(resp.data.get("message"))
            for val in resp.get("data", []):
                data = val.get("data", {})
                metadata = data.get("metadata", {})
                # 按照名称进行过滤
                key_name = "application:%s" % metadata.get("name")
                if data.get("status") in app_constants.UNNORMAL_STATUS:
                    if key_name not in ret_data:
                        ret_data[key_name] = 1
                    else:
                        ret_data[key_name] += 1
        # 处理deployment状态
        for cluster_id, info in cluster_ns_inst.items():
            if not info.get("deployment"):
                continue
            flag, resp = self.get_app_deploy_with_post(
                request,
                project_id,
                cluster_id,
                ",".join(set(info["deployment"]["inst_list"])),
                category=DEPLOYMENT_CATEGORY,
                project_kind=kind,
                namespace=",".join(set(info["deployment"]["ns_list"])),
            )
            if not flag:
                raise error_codes.APIError.f(resp.data.get("message"))
            for val in resp.get("data", []):
                data = val.get("data", {})
                metadata = data.get("metadata", {})
                # 按照名称进行过滤
                key_name = "deployment:%s" % metadata.get("name")
                if data.get("status") in app_constants.UNNORMAL_STATUS:
                    if key_name not in ret_data:
                        ret_data[key_name] = 1
                    else:
                        ret_data[key_name] += 1
        # 匹配状态
        tmpl_status_map = {}
        for key, val in ret_data.items():
            if key.startswith("application"):
                if key in application_deploy_map:
                    tmpl_status_map[application_deploy_map[key]] = val
                else:
                    tmpl_status_map[key] = val
            else:
                if key in tmpl_status_map:
                    tmpl_status_map[key] += val
                else:
                    tmpl_status_map[key] = val
        return tmpl_status_map

    def compose_status_count_data(self, muster_tmpl_map, tmpl_create_error, inst_status):
        """组装数量"""
        ret_data = {}
        for key, val in muster_tmpl_map.items():
            ret_data[key] = {"total_num": val, "error_num": 0}
            if key in tmpl_create_error:
                ret_data[key]["error_num"] += tmpl_create_error[key]

            if key in inst_status:
                ret_data[key]["error_num"] += inst_status[key]
        return ret_data

    def get_application_map(self, application_ids, deployment_ids):
        """获取application信息"""
        deployment_info = MODULE_DICT["deployment"].objects.filter(id__in=deployment_ids)
        app_ids = [info.app_id for info in deployment_info]
        application_info = (
            MODULE_DICT["application"].objects.filter(id__in=application_ids).exclude(app_id__in=app_ids)
        )
        return {info.id: info.name for info in application_info}

    def get_deployment_map(self, deployment_ids):
        """获取deployment信息"""
        deployment_info = MODULE_DICT["deployment"].objects.filter(id__in=deployment_ids)
        return {info.id: info.name for info in deployment_info}

    def compose_ret_data(self, version_tmpl_muster, version_map, tmpl_count_info, app_status):
        """拼装数据"""
        ret_data = {}
        exist_key = []

        for info in version_tmpl_muster:
            # 获取application和deployment
            application_detail = self.get_application_map(info["application_list"], info["deployment_list"])
            deployment_detail = self.get_deployment_map(info["deployment_list"])
            item = {
                "tmpl_muster_id": info["tmpl_muster_id"],
                "tmpl_muster_name": info["tmpl_muster_name"],
                "version": info["version"],
                "version_id": info["version_id"],
                "last_version": info["last_version"],
                "last_version_id": info["last_version_id"],
                "total_num": 0,
                "error_num": 0,
                "allow_edit": True,
            }
            for key, val in application_detail.items():
                curr_key = "application:%s" % val
                # 是否有前一个版本，并且不为1
                item_copy = item.copy()
                num_info = tmpl_count_info.get(curr_key) or {}
                if version_map.get(info["version_id"]) and not num_info.get("total_num"):
                    continue
                if curr_key not in exist_key:
                    exist_key.append(curr_key)
                    item_copy["category"] = "application"
                    item_copy["tmpl_app_id"] = key
                    item_copy["tmpl_app_name"] = val
                    item_copy["total_num"] = num_info.get("total_num") or 0
                    if item_copy["total_num"] == 0:
                        continue
                    item_copy["error_num"] = num_info.get("error_num") or 0
                    if app_status in [2, "2", None]:
                        if item_copy["error_num"]:
                            ret_data[curr_key] = item_copy
                    if app_status in [1, "1", None]:
                        if not item_copy["error_num"]:
                            ret_data[curr_key] = item_copy
                if curr_key in ret_data:
                    if ret_data[curr_key]["version_id"] < item_copy["version_id"]:
                        ret_data[curr_key].update(
                            {
                                "version": item_copy["version"],
                                "version_id": item_copy["version_id"],
                                "last_version": item_copy["last_version"],
                                "last_version_id": item_copy["last_version_id"],
                            }
                        )
                    else:
                        if version_map.get(item_copy["version_id"]):
                            ret_data[curr_key]["allow_edit"] = False

            for key, val in deployment_detail.items():
                curr_key = "deployment:%s" % val
                num_info = tmpl_count_info.get(curr_key) or {}
                item_copy = item.copy()
                if version_map.get(info["version_id"]) and not num_info.get("total_num"):
                    continue
                if curr_key not in exist_key:
                    exist_key.append(curr_key)
                    item_copy["category"] = "deployment"
                    item_copy["tmpl_app_id"] = key
                    item_copy["tmpl_app_name"] = val
                    item_copy["total_num"] = num_info.get("total_num") or 0
                    if item_copy["total_num"] == 0:
                        continue
                    item_copy["error_num"] = num_info.get("error_num") or 0
                    if app_status in [2, "2", None]:
                        if item_copy["error_num"]:
                            ret_data[curr_key] = item_copy
                    if app_status in [1, "1", None]:
                        if not item_copy["error_num"]:
                            ret_data[curr_key] = item_copy
                if curr_key in ret_data:
                    if ret_data[curr_key]["version_id"] < item_copy["version_id"]:
                        ret_data[curr_key].update(
                            {
                                "version": item_copy["version"],
                                "version_id": item_copy["version_id"],
                                "last_version": item_copy["last_version"],
                                "last_version_id": item_copy["last_version_id"],
                            }
                        )
                    else:
                        if version_map.get(item_copy["version_id"]):
                            ret_data[curr_key]["allow_edit"] = False

        # 处理最终版本是否允许编辑
        # for key, val in ret_data.items():
        #     if version_map.get(val["version_id"]):
        #         val["allow_edit"] = False

        return ret_data

    def refine_template(self, data):
        """过滤不包含实例的模板"""
        refine_data = []
        for key, val in data.items():
            if val["total_num"] != 0:
                refine_data.append(val)
        return refine_data

    def get(self, request, project_id, muster_id):
        """查询模板集下模板的信息"""
        # 获取过滤参数
        cluster_type, app_status, filter_muster_id, app_id, ns_id, request_cluster_id = self.get_filter_params(
            request, project_id
        )
        if filter_muster_id and str(filter_muster_id) != muster_id:
            return utils.APIResponse({"data": []})
        # 判断项目和模板集是否对应
        self.check_project_muster(project_id, muster_id)
        # 获取项目信息
        project_kind = self.project_kind(request)
        # 获取集群及环境
        cluster_env_map = self.get_cluster_id_env(request, project_id)
        # 获取实例名称
        filter_app_name = self.get_inst_name(app_id)
        # 获取版本映射
        version_map = self.get_version_map(muster_id)
        # 获取instance信息
        muster_tmpl_map, cluster_ns_inst, tmpl_create_error = self.get_instances(
            muster_id, cluster_type, filter_app_name, ns_id, cluster_env_map, request_cluster_id
        )

        category = request.GET.get("category")
        if not category or category not in k8s_views.CATEGORY_MAP:
            raise error_codes.CheckFailed(_("分类不正确"))
        # 根据模板集获取所有模板
        version_id_map_list = self.get_tmpl_info(muster_id, category=category)
        k8s_view_client = k8s_views.GetMusterTemplate()
        ret_data = k8s_view_client.get(
            request,
            cluster_ns_inst,
            project_id,
            project_kind,
            version_id_map_list,
            version_map,
            category,
            muster_tmpl_map,
            tmpl_create_error,
            cluster_type,
            app_status,
            filter_app_name,
            ns_id,
            cluster_env_map,
        )
        ret_data_values = self.refine_template(ret_data)
        return utils.APIResponse({"data": ret_data_values})


class AppInstance(BaseMusterMetric):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def get_version_instance(self, muster_id, ids, category, project_kind=ProjectKindID):
        """获取实例版本信息"""
        category = k8s_views.CATEGORY_MAP[category]
        all_version_info = VersionInstance.objects.filter(is_deleted=False, template_id=muster_id).order_by("-created")
        # 组装数据
        instance_version_ids = []
        for info in all_version_info:
            entity = info.get_entity
            category_ids = entity.get(category, [])
            for id in ids:
                if int(id) in category_ids:
                    instance_version_ids.append(info.id)
        return list(set(instance_version_ids))

    def get_inst_conf(self, info):
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析配置文件异常，当前实例ID: %s, 详情: %s" % info.id, error)
            conf = {}
        return conf

    def get_rolling_update_info(self, conf):
        """获取滚动升级信息"""
        strategy = conf.get("spec", {}).get("strategy", {})
        rolling_update_data = strategy.get("rollingupdate", {})
        return {
            "rolling_update_type": strategy.get("type", "RollingUpdate"),
            "rolling_max_unavilable": rolling_update_data.get("maxUnavilable", 0),
            "rolling_max_surge": rolling_update_data.get("maxSurge", 0),
            "rolling_duration": rolling_update_data.get("upgradeDuration", 0),
            "rolling_order": rolling_update_data.get("rollingOrder", "CreateFirst"),
        }

    def get_muster_info(self, tmpl_id):
        tmpl_info = Template.objects.filter(id=tmpl_id).first()
        if not tmpl_info:
            return None
        return tmpl_info.name

    def get_instance_info(
        self,
        instance_version_ids,
        category,
        tmpl_name,
        cluster_type,
        cluster_env_map,
        app_name,
        ns_id,
        project_kind=ProjectKindID,
        request_cluster_id=None,
    ):
        """获取实例信息"""
        category = k8s_views.CATEGORY_MAP[category]
        filter_category = [category]
        ret_data = {}
        inst_info = (
            InstanceConfig.objects.filter(instance_id__in=instance_version_ids, category=category, is_deleted=False)
            .exclude(ins_state=InsState.NO_INS.value)
            .order_by("-updated")
        )
        if app_name:
            inst_info = inst_info.filter(name=app_name)
        if ns_id:
            inst_info = inst_info.filter(namespace=ns_id)
        # 获取请求BCS接口失败的最新的一条事件
        all_events = InstanceEvent.objects.filter(
            instance_id__in=instance_version_ids,
            is_deleted=False,
            category__in=filter_category,
        ).order_by("-updated")
        inst_id_event_map = {}
        inst_id_event_map = {
            info.instance_config_id: info.msg
            for info in all_events
            if info.instance_config_id not in inst_id_event_map
        }
        # 匹配info
        for info in inst_info:
            conf = self.get_inst_conf(info)
            if not conf:
                continue
            metadata = conf.get("metadata", {})
            labels = metadata.get("labels")
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            if exclude_records(
                request_cluster_id,
                cluster_id,
                cluster_type,
                cluster_env_map.get(cluster_id, {}).get("cluster_env"),
            ):
                continue
            key_name = (cluster_id, metadata.get("namespace"), metadata.get("name"))
            if (not metadata.get("name", "").startswith(tmpl_name)) or (key_name in ret_data):
                continue
            backend_status = "BackendNormal"
            oper_type_flag = ""
            if not info.is_bcs_success:
                if info.oper_type == "create":
                    backend_status = "BackendError"
                else:
                    oper_type_flag = info.oper_type
            # 针对oper_type_flag的含义描述
            # backend_error && oper_type==create  oper_type_flag为空
            # 否则，oper_type_flag和oper_type同一个值，表示前端不显示重试+删除；
            # 显示的是这个值对应的操作，并且加上error的感叹号
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            muster_id = labels.get("io.tencent.paas.templateid")
            cluster_name_env_map = cluster_env_map.get(cluster_id) or {}
            item = {
                "id": info.id,
                "name": metadata.get("name"),
                "namespace": metadata.get("namespace"),
                "namespace_id": info.namespace,
                "create_at": info.created,
                "update_at": info.updated,
                "backend_status": backend_status,
                "backend_status_message": inst_id_event_map.get(info.id),
                "creator": info.creator,
                "oper_type": info.oper_type,
                "oper_type_flag": oper_type_flag,
                "cluster_id": cluster_id,
                "build_instance": 0,
                "instance": 0,
                "muster_id": muster_id,
                "cluster_name": cluster_name_env_map.get("cluster_name"),
                "cluster_env": cluster_name_env_map.get("cluster_env"),
                "environment": cluster_name_env_map.get("cluster_env_str"),
                "muster_name": self.get_muster_info(muster_id),
                "source_type": "模板集",
                "from_platform": True,
            }
            annotations = metadata.get("annotations") or {}
            item.update(utils.get_instance_version(annotations, labels))
            item.update(
                {
                    "status": "Unready",
                    "status_message": _("请点击查看详情"),
                    "pod_count": "0/0",
                    "category": k8s_views.REVERSE_CATEGORY_MAP[info.category],
                }
            )
            ret_data[key_name] = item
        return ret_data

    def get_cluster_namespace_inst1(self, instance_info):
        """获取集群、命名空间和deployment"""
        ret_data = {}
        for id, info in instance_info.items():
            if info["cluster_id"] in ret_data:
                ret_data[info["cluster_id"]]["inst_list"].append(info["name"])
                ret_data[info["cluster_id"]]["ns_list"].append(info["namespace"])
            else:
                ret_data[info["cluster_id"]] = {
                    "inst_list": [info["name"]],
                    "ns_list": [info["namespace"]],
                }
        return ret_data

    def get_cluster_namespace_inst(self, instance_info):
        ret_data = {}
        for id, info in instance_info.items():
            cluster_id = info["cluster_id"]
            app_id = "%s:%s" % (info["namespace"], info["name"])
            if cluster_id in ret_data:
                ret_data[cluster_id].append(app_id)
            else:
                ret_data[cluster_id] = [app_id]
        return ret_data

    def get_cluster_namespace_deployment(self, instance_info, deploy_app_info):
        """针对deployment组装请求taskgroup信息"""
        ret_data = {}
        for __, info in instance_info.items():
            # TODO: 后续重构时，注意匹配字段
            app_name_list = deploy_app_info.get((info['cluster_id'], info["namespace"], info["name"]), [])
            if not app_name_list:
                continue
            app_id = "%s:%s" % (info["namespace"], app_name_list[0])
            cluster_id = info["cluster_id"]
            if cluster_id in ret_data:
                ret_data[cluster_id].append(app_id)
            else:
                ret_data[cluster_id] = [app_id]
        return ret_data

    def get_application_by_deployment(self, request, query_info, kind=2, project_id=None):
        """通过deployment获取application"""
        ret_data = {}
        for cluster_id, info in query_info.items():
            flag, resp = self.get_app_deploy_with_post(
                request,
                project_id,
                cluster_id,
                "",
                category=DEPLOYMENT_CATEGORY,
                project_kind=kind,
                field="data.application,data.application_ext,data.metadata",
            )
            if not flag:
                raise error_codes.APIError.f(resp.data.get("message"))
            for val in resp.get("data") or []:
                metadata = val.get("data", {}).get("metadata", {})
                curr_app_id = "%s:%s" % (metadata.get("namespace"), metadata.get("name"))
                if curr_app_id not in info:
                    continue
                key_name = (cluster_id, metadata.get("namespace"), metadata.get("name"))
                application = val.get("data", {}).get("application", {})
                application_ext = val.get("data", {}).get("application_ext", {})
                if key_name not in ret_data:
                    ret_data[key_name] = []
                if application:
                    ret_data[key_name].append(application.get("name"))
                if application_ext:
                    ret_data[key_name].append(application_ext.get("name"))
        return ret_data

    def get_application_status(self, request, cluster_ns_inst, project_id=None, category=APPLICATION_CATEGORY, kind=2):
        """获取application的状态"""
        ret_data = {}
        for cluster_id, info in cluster_ns_inst.items():
            flag, resp = self.get_application_deploy_status(
                request,
                project_id,
                cluster_id,
                "",
                category=category,
                project_kind=kind,
                field="data.metadata.name,data.metadata.namespace,data.status,data.message,data.buildedInstance,data.instance",  # noqa
            )
            if not flag:
                raise error_codes.APIError.f(resp.data.get("message"))
            for val in resp.get("data", []):
                data = val.get("data", {})
                metadata = data.get("metadata", {})
                curr_app_id = "%s:%s" % (metadata.get("namespace"), metadata.get("name"))
                if curr_app_id not in info:
                    continue
                key_name = (cluster_id, metadata.get("namespace"), metadata.get("name"))
                build_instance = data.get("buildedInstance") or 0
                instance = data.get("instance") or 0
                ret_data[key_name] = {
                    "application_status": data.get("status"),
                    "application_status_message": data.get("message"),
                    "task_group_count": "%s/%s" % (build_instance, instance),
                    "build_instance": build_instance,
                    "instance": instance,
                }
        return ret_data

    def get_deployment_status(self, request, cluster_ns_inst, project_id=None, category=DEPLOYMENT_CATEGORY, kind=2):
        """获取deployment的状态"""
        ret_data = {}
        for cluster_id, info in cluster_ns_inst.items():
            flag, resp = self.get_application_deploy_status(
                request, project_id, cluster_id, "", category=category, project_kind=kind
            )
            if not flag:
                raise error_codes.APIError.f(resp.data.get("message"))
            for val in resp.get("data", []):
                data = val.get("data", {})
                metadata = data.get("metadata", {})
                curr_app_id = "%s:%s" % (metadata.get("namespace"), metadata.get("name"))
                if curr_app_id not in info:
                    continue
                key_name = (cluster_id, metadata.get("namespace"), metadata.get("name"))
                ret_data[key_name] = {
                    "deployment_status": data.get("status"),
                    "deployemnt_status_message": data.get("message"),
                }
        return ret_data

    def update_inst_label(self, inst_id_list):
        """更新删除的实例标识"""
        all_inst_conf = InstanceConfig.objects.filter(id__in=inst_id_list)
        all_inst_conf.update(is_deleted=True, deleted_time=datetime.now(), status="Deleted")
        # 更新metric config状态
        inst_ns = {info.instance_id: info.namespace for info in all_inst_conf}
        inst_ver_ids = inst_ns.keys()
        inst_ns_ids = inst_ns.values()
        try:
            MetricConfig.objects.filter(instance_id__in=inst_ver_ids, namespace__in=inst_ns_ids).update(
                ins_state=InsState.INS_DELETED.value
            )
        except Exception as err:
            logger.error(u"更新metric删除状态失败，详情: %s" % err)

    def compose_data(self, request, instance_info, all_status):
        """组装返回数据"""
        delete_key_name_list = []
        update_id_status_list = []
        need_update_status = []
        update_create_error_id_list = []
        for info in instance_info.values():
            key_name = (info['cluster_id'], info["namespace"], info["name"])
            if info["backend_status"] in ["BackendError"] and key_name in all_status:
                update_create_error_id_list.append(info["id"])
                info["backend_status"] = "BackendNormal"
            # info["task_group_count"] = task_group_count.get(key_name, 0)
            info.update(all_status.get(key_name, {}))
            if info["oper_type"] == app_constants.DELETE_INSTANCE:
                fail_status = app_constants.UNNORMAL_STATUS
                if not all_status.get(key_name):
                    delete_key_name_list.append(key_name)
                    update_id_status_list.append(info["id"])
                elif info["deployment_status"] in fail_status or info["application_status"] in fail_status:
                    # 更新实例的状态
                    need_update_status.append(info["id"])
        if update_create_error_id_list:
            update_create_error_record.delay(update_create_error_id_list)
        # 更新删除失败的实例状态
        InstanceConfig.objects.filter(id__in=need_update_status).update(oper_type=app_constants.RESUME_INSTANCE)
        # 更新状态
        if update_id_status_list:
            self.update_inst_label(update_id_status_list)

    def get_template_info(self, template_id, category, project_kind=ProjectKindID):
        """获取模板信息"""
        category = k8s_views.CATEGORY_MAP[category]
        info = MODULE_DICT[category].objects.filter(id=template_id, is_deleted=False)
        if not info:
            raise error_codes.RecordNotFound(_("没有查询到记录!"))
        name = info[0].name
        # 过滤整个模板表
        # 针对application获取整个表
        same_name_ids = []
        if category == APPLICATION_CATEGORY:
            all_info = MODULE_DICT[category].objects.all()
            for info in all_info:
                conf = info.get_config() or {}
                if name == conf.get("metadata", {}).get("name"):
                    same_name_ids.append(info.id)
        else:
            # 针对deployment根据名称过滤
            filter_info = MODULE_DICT[category].objects.filter(name=name)
            same_name_ids = [info.id for info in filter_info]
        # 返回
        return name, same_name_ids

    def inst_count_handler(self, instance_info, app_status):
        instance_list = instance_info.values()
        ret_data = {
            "error_num": 0,
        }
        instance_list = list(instance_list)
        inst_list_copy = copy.deepcopy(instance_list)
        for val in inst_list_copy:
            if (
                (val["backend_status"] in app_constants.UNNORMAL_STATUS)
                or (val["application_status"] in app_constants.UNNORMAL_STATUS)
                or (val["deployment_status"] in app_constants.UNNORMAL_STATUS)
            ):
                if app_status in [2, "2", None]:
                    ret_data["error_num"] += 1
                else:
                    instance_list.remove(val)
        ret_data.update({"total_num": len(instance_list), "instance_list": instance_list})
        return ret_data

    @response_perms(
        action_ids=[
            NamespaceScopedAction.VIEW,
            NamespaceScopedAction.UPDATE,
            NamespaceScopedAction.DELETE,
            NamespaceScopedAction.CREATE,
        ],
        permission_cls=NamespaceScopedPermission,
        resource_id_key='iam_ns_id',
    )
    def get(self, request, project_id, muster_id, template_id):
        # 获取过滤参数
        cluster_type, app_status, filter_muster_id, app_id, ns_id, request_cluster_id = self.get_filter_params(
            request, project_id
        )
        if filter_muster_id and str(filter_muster_id) != muster_id:
            return utils.APIResponse({"data": []})

        # 模板类型参数
        category = request.GET.get("category")
        if not category:
            return utils.APIResponse(
                {
                    "code": 400,
                    "message": _("参数[category]不能为空"),
                }
            )
        # 获取项目信息
        project_kind = self.project_kind(request)
        # 获取集群及环境
        cluster_env_map = self.get_cluster_id_env(request, project_id)
        # 获取实例名称
        filter_app_name = self.get_inst_name(app_id)
        # 获取模板名称
        tmpl_name, ids = self.get_template_info(template_id, category, project_kind)
        # 获取实例版本信息
        instance_version_ids = self.get_version_instance(muster_id, ids, category, project_kind)
        # 根据实例版本获取实例
        instance_info = self.get_instance_info(
            instance_version_ids,
            category,
            tmpl_name,
            cluster_type,
            cluster_env_map,
            filter_app_name,
            ns_id,
            project_kind=project_kind,
            request_cluster_id=request_cluster_id,
        )
        client = k8s_views.AppInstance()
        ret_data = client.get(request, project_id, instance_info, category, app_status)

        iam_ns_ids = set()
        for inst in ret_data["instance_list"]:
            iam_ns_id = calc_iam_ns_id(inst['cluster_id'], inst['namespace'])
            inst['iam_ns_id'] = iam_ns_id
            iam_ns_ids.add(iam_ns_id)

        return PermsResponse(
            ret_data,
            resource_request=NamespaceRequest(project_id=project_id, cluster_id=request_cluster_id),
            resource_data=[{'iam_ns_id': iam_ns_id} for iam_ns_id in iam_ns_ids],
        )


class CreateInstance(BaseAPI):
    def get_instance_conf(self, info):
        """获取instance conf"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError(_("Instance config解析异常"))
        return conf

    def update_inst_property(self, resp, curr_inst, request):
        """更新实例属性"""
        if resp.data.get("code") == ErrorCode.NoError:
            curr_inst.is_bcs_success = True
            curr_inst.updated = datetime.now()
            curr_inst.ins_state = InsState.UPDATE_SUCCESS.value
            curr_inst.oper_type = app_constants.CREATE_INSTANCE
            curr_inst.save()
        else:
            self.event_log_record(
                curr_inst.id,
                curr_inst.instance_id,
                curr_inst.category,
                resp.data.get("message") or "",
                resp.data,
                request.user.username,
            )

    def post(self, request, project_id, instance_id):
        """创建失败后，重试当前实例"""
        # 获取当前实例的类型
        inst_info = self.get_instance_info(instance_id)
        curr_inst = inst_info[0]
        if curr_inst.is_bcs_success:
            return utils.APIResponse({"code": 400, "message": _("实例已经创建，请勿重复操作!")})
        conf = self.get_instance_conf(curr_inst)
        metadata = conf.get("metadata", {})
        labels = metadata.get("labels", {})
        cluster_id = labels.get("io.tencent.bcs.clusterid")
        namespace = metadata.get("namespace")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
        )

        # 获取项目类型
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        try:
            with client.ContextActivityLogClient(
                project_id=project_id,
                user=request.user.username,
                resource_type="instance",
                resource=metadata.get("name"),
                resource_id=instance_id,
                extra=json.dumps({"config": conf}),
                description=_("重试应用实例化"),
            ).log_add():
                resp = self.create_instance(
                    request, project_id, cluster_id, namespace, conf, category=curr_inst.category, kind=project_kind
                )
        except Exception as error:
            logger.error(u"实例化应用失败，失败的实例ID: %s,详情: %s" % (curr_inst.id, error))
            return utils.APIResponse({"code": 500, "message": "%s" % error})

        # 更新属性信息
        self.update_inst_property(resp, curr_inst, request)
        return resp


class UpdateInstanceNew(InstanceAPI):
    """滚动更新实例"""

    def get_instance_info(self, instance_id):
        """获取实例信息"""
        info = InstanceConfig.objects.filter(id=instance_id)
        if not info:
            raise error_codes.CheckFailed(_("没有查询到实例信息"))
        return info[0]

    def get_tmpl_ids(self, version_id, category):
        """获取模板版本ID"""
        info = VersionedEntity.objects.filter(id=version_id)
        if not info:
            raise error_codes.CheckFailed(_("没有查询到相应的版本信息"))
        curr_info = info[0]
        entity = json.loads(curr_info.entity)
        return entity.get(category)

    def get_tmpl_entity(self, category, ids, name):
        """获取模板信息"""
        id_list = ids.split(",")
        tmpl_info = MODULE_DICT[category].objects.filter(id__in=id_list, name=name)
        if not tmpl_info:
            raise error_codes.CheckFailed(_("没有查询到模板信息"))
        return {category: [tmpl_info[0].id]}

    def generate_ns_config_info(self, request, ns_id, inst_entity, params, is_save=True):
        """生成某一个命名空间下的配置"""
        # 针对接口调用，跳过这一部分
        try:
            # 在数据平台创建项目信息
            cc_app_id = request.project.cc_app_id
            english_name = request.project.english_name
            project_id = request.project.project_id
        except Exception:
            pass

        inst_conf = inst_utils.generate_namespace_config(ns_id, inst_entity, is_save, **params)
        return inst_conf

    def get_show_version_info(self, version_id):
        """查询show version信息"""
        info = ShowVersion.objects.filter(id=version_id)
        if not info:
            raise error_codes.CheckFailed(_("没有查询展示版本信息"))
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
            logger.error(u"更新配置出现异常, 实例ID: %s, 详情: %s" % (inst_id, error))
            raise error_codes.DBOperError(_("更新实例配置异常!"))

    def get_params(self, request):
        """获取参数"""
        version_id = request.GET.get("version_id")
        if not version_id:
            raise error_codes.CheckFailed(_("升级的版本信息不能为空"))
        instance_num = request.GET.get("instance_num")
        category = request.GET.get("category")
        if category not in ["job", "daemonset"] and (not instance_num or not str(instance_num).isdigit()):
            raise error_codes.CheckFailed(_("实例数量必须为正数"))
        # 获取环境变量，如果有未赋值的，则认为参数不正确
        data = dict(request.data)
        variables = data.get("variable") or {}
        if variables:
            for key, val in variables.items():
                if not val:
                    raise error_codes.CheckFailed(_("环境变量 {} 对应的值不能为空").format(key))
        return version_id, instance_num, variables

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
        # 实例数需要转成整数
        try:
            instance_num = int(instance_num)
        except Exception:
            raise error_codes.CheckFailed(_("实例数 {} 不是整数").format(instance_num))

        if category in ["K8sDeployment", "K8sStatefulSet"]:
            new_conf["spec"]["replicas"] = instance_num
        return new_conf

    def check_selector(self, old_conf, new_conf):
        """检查selector是否一致，如果不一致，则提示用户处理"""
        old_item = ((old_conf.get("spec") or {}).get("selector") or {}).get("matchLabels") or {}
        new_item = ((new_conf.get("spec") or {}).get("selector") or {}).get("matchLabels") or {}
        if old_item != new_item:
            raise error_codes.CheckFailed(_("Selector不一致不能进行更新，请确认!"))

    def update_instance(
        self, request, project_id, project_kind, cluster_id, name, instance_id, namespace, category, conf
    ):
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=name,
            resource_id=instance_id,
            extra=json.dumps({"namespace": namespace}),
            description=_("应用滚动升级"),
        ).log_modify():
            resp = self.update_deployment(
                request, project_id, cluster_id, namespace, conf, kind=project_kind, category=category, app_name=name
            )
            # 为异常时，直接抛出
            if resp.data.get("code") != ErrorCode.NoError:
                raise error_codes.APIError(_("应用滚动升级失败，{}").format(resp.data.get("message")))
        return resp

    def update_online_app(self, request, project_id, project_kind):
        """滚动更新线上应用"""
        cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        if category == "job":
            raise error_codes.CheckFailed(_("JOB类型不允许滚动升级"))
        conf = dict(request.data).get("conf")
        if not conf:
            raise error_codes.CheckFailed(_("参数【conf】不能为空!"))
        conf = json.loads(conf)
        resp = self.update_instance(request, project_id, project_kind, cluster_id, name, 0, namespace, category, conf)

        return resp

    def put(self, request, project_id, instance_id):
        project_kind = self.project_kind(request)

        if str(instance_id) == "0":
            return self.update_online_app(request, project_id, project_kind)

        version_id, instance_num, variables = self.get_params(request)
        # 获取实例信息
        inst_info = self.get_instance_info(instance_id)

        category = inst_info.category
        if category == "K8sJob":
            raise error_codes.CheckFailed(_("JOB类型不允许滚动升级"))
        inst_name = inst_info.name
        # 获取namespace
        inst_conf = json.loads(inst_info.config)
        metadata = inst_conf.get("metadata", {})
        labels = metadata.get("labels", {})
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), inst_info.namespace
        )

        cluster_id = labels.get("io.tencent.bcs.clusterid")
        namespace = metadata.get("namespace")
        pre_instance_num = 0
        if category in ["K8sDeployment", "K8sStatefulSet"]:
            pre_instance_num = inst_conf["spec"]["replicas"]
        # 获取版本配置
        show_version_info = self.get_show_version_info(version_id)
        ids = self.get_tmpl_ids(show_version_info.real_version_id, category)
        tmpl_entity = self.get_tmpl_entity(category, ids, inst_name)
        real_instance_num = pre_instance_num
        if instance_num:
            real_instance_num = int(instance_num)
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
            variables,
        )
        self.check_selector(inst_conf, new_inst_conf)
        # self.render_instance_conf(new_inst_conf, params)
        resp = self.update_instance(
            request, project_id, project_kind, cluster_id, inst_name, instance_id, namespace, category, new_inst_conf
        )
        inst_state = InsState.UPDATE_SUCCESS.value
        if resp.data.get("code") == ErrorCode.NoError:
            inst_state = InsState.UPDATE_FAILED.value
            self.update_instance_record_status(
                inst_info, oper_type=app_constants.ROLLING_UPDATE_INSTANCE, is_bcs_success=True
            )
        # 更新conf
        self.update_conf_version(
            instance_id,
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
            variables,
        )

        return resp


class ScaleInstance(InstanceAPI):
    def update_inst_conf(self, curr_inst, instance_num, inst_state):
        """更新配置"""
        try:
            conf = json.loads(curr_inst.config)
        except Exception as error:
            logger.error("解析出现异常，当前实例ID: %s, 详情: %s" % (curr_inst.id, error))
            return
        if curr_inst.category in k8s_views.REVERSE_CATEGORY_MAP:
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

    def scale_inst(
        self, request, project_id, project_kind, cluster_id, instance_id, name, namespace, instance_num, category, conf
    ):
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=name,
            resource_id=instance_id,
            extra=json.dumps({"config": conf, "namespace": namespace}),
            description=_("应用扩缩容"),
        ).log_modify():
            resp = self.scale_instance(
                request,
                project_id,
                cluster_id,
                namespace,
                name,
                instance_num,
                kind=project_kind,
                category=category,
                data=conf,
            )
        return resp

    def scale_online_app(self, request, project_id, project_kind, inst_count):
        """扩缩容线上应用"""
        cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        online_app_conf = None
        online_app_conf = self.online_app_conf(
            request, project_id, project_kind, cluster_id, name, namespace, category
        )
        return self.scale_inst(
            request, project_id, project_kind, cluster_id, 0, name, namespace, inst_count, category, online_app_conf
        )

    def put(self, request, project_id, instance_id, instance_name):
        """扩缩容"""
        # 获取数量
        instance_num = request.GET.get("instance_num")
        if not instance_num and instance_num != 0:
            return utils.APIResponse({"code": 400, "message": _("参数[instance_num]不能为空")})
        if not str(instance_num).isdigit():
            return utils.APIResponse({"code": 400, "message": _("参数[instance_num]必须为整数")})
        project_kind = self.project_kind(request)
        # 针对非模板的操作
        if str(instance_id) == "0":
            return self.scale_online_app(request, project_id, project_kind, instance_num)

        # 获取instance info
        inst_info = self.get_instance_info(instance_id)
        # 获取namespace
        curr_inst = inst_info[0]

        conf = self.get_common_instance_conf(curr_inst)
        metadata = conf.get("metadata", {})
        labels = metadata.get("labels", {})
        cluster_id = labels.get("io.tencent.bcs.clusterid")
        namespace = metadata.get("namespace")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
        )
        # 检测资源是否满足扩容条件
        # self.check_resource(request, project_id, conf, cluster_id, instance_num)
        name = metadata.get("name")
        if curr_inst.category == DEPLOYMENT_CATEGORY:
            app_name = self.get_rc_name_by_deployment(
                request, cluster_id, name, project_id=project_id, project_kind=project_kind, namespace=namespace
            )
            name = app_name[0]
        resp = self.scale_inst(
            request,
            project_id,
            project_kind,
            cluster_id,
            instance_id,
            name,
            namespace,
            instance_num,
            curr_inst.category,
            conf,
        )
        inst_state = InsState.UPDATE_SUCCESS.value

        if resp.data.get("code") == ErrorCode.NoError:
            self.update_instance_record_status(curr_inst, oper_type=app_constants.SCALE_INSTANCE, is_bcs_success=True)
        # 更新配置
        self.update_inst_conf(curr_inst, instance_num, inst_state)
        return resp


class CancelUpdateInstance(InstanceAPI):
    def update_conf(self, inst_id, category, inst_version_id):
        """更新配置"""
        try:
            old_inst_info = InstanceConfig.objects.filter(id=inst_id)
            # 更改版本
            inst_version_info = VersionInstance.objects.filter(id=inst_version_id)
            save_conf = {
                "old_conf": json.loads(old_inst_info[0].config),
                "old_version_id": inst_version_info[0].version_id,
                "old_show_version_id": inst_version_info[0].show_version_id,
                "old_show_version_name": inst_version_info[0].show_version_name,
                "old_variables": json.loads(old_inst_info[0].variables),
            }
            entity = json.loads(inst_version_info[0].instance_entity)
            category_content = entity[category]
            save_conf["old_category_entity"] = category_content
            old_conf = json.loads(old_inst_info[0].last_config)
            # 获取模板交集
            entity.update({category: old_conf["old_category_entity"]})
            inst_version_info.update(
                version_id=old_conf["old_version_id"],
                instance_entity=json.dumps(entity),
                show_version_id=old_conf["old_show_version_id"],
                show_version_name=old_conf["old_show_version_name"],
            )
            inst_state = InsState.UPDATE_SUCCESS.value
            old_inst_info.update(
                config=json.dumps(old_conf["old_conf"]),
                ins_state=inst_state,
                last_config=json.dumps(save_conf),
                variables=json.dumps(old_conf.get("old_variables") or {}),
            )
        except Exception as error:
            logger.error(u"更新配置出现异常, 实例ID: %s, 详情: %s" % (inst_id, error))
            raise error_codes.DBOperError(_("更新实例配置异常!"))

    def put(self, request, project_id, instance_id, instance_name):
        """取消更新
        针对k8s采用上一个版本进行撤销
        """
        if not self._from_template(instance_id):
            cluster_id, namespace, name, _ = self.get_instance_resource(request, project_id)
            return self.cancel_update_deployment(request, project_id, cluster_id, namespace, name)
        # 获取instance info
        inst_info = self.get_instance_info(instance_id)
        # 获取kind
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        # 获取namespace
        curr_inst = inst_info[0]

        conf = self.get_common_instance_conf(curr_inst)
        metadata = conf.get("metadata", {})
        labels = metadata.get("labels", {})
        cluster_id = labels.get("io.tencent.bcs.clusterid")
        namespace = metadata.get("namespace")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
        )
        name = metadata.get("name")
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=name,
            resource_id=instance_id,
            extra=json.dumps({"config": conf, "namespace": "namespace"}),
            description=_("应用取消滚动升级"),
        ).log_modify():
            last_config = json.loads(curr_inst.last_config)
            resp = self.update_deployment(
                request,
                project_id,
                cluster_id,
                namespace,
                last_config["old_conf"],
                kind=project_kind,
                category=curr_inst.category,
                app_name=curr_inst.name,
            )

        if resp.data.get("code") == ErrorCode.NoError:
            self.update_instance_record_status(curr_inst, oper_type=app_constants.CANCEL_INSTANCE, is_bcs_success=True)
            # 更新信息
            self.update_conf(curr_inst.id, curr_inst.category, curr_inst.instance_id)

        return resp


class PauseUpdateInstance(InstanceAPI):
    def pause_inst(
        self, request, project_id, project_kind, cluster_id, instance_id, name, namespace, category, conf=None
    ):

        self.can_operate(category)
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=name,
            resource_id=instance_id,
            extra=json.dumps({"config": conf, "namespace": namespace}),
            description=_("应用暂停更新"),
        ).log_modify():
            resp = self.pause_update_deployment(
                request, project_id, cluster_id, namespace, name, kind=project_kind, category=category
            )
        return resp

    def pause_online_app(self, request, project_id, project_kind):
        cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        return self.pause_inst(request, project_id, project_kind, cluster_id, 0, name, namespace, category)

    def put(self, request, project_id, instance_id, instance_name):
        """暂停更新"""
        project_kind = self.project_kind(request)
        if str(instance_id) == "0":
            return self.pause_online_app(request, project_id, project_kind)
        # 获取instance info
        inst_info = self.get_instance_info(instance_id)
        # 获取namespace
        curr_inst = inst_info[0]

        conf = self.get_common_instance_conf(curr_inst)
        metadata = conf.get("metadata", {})
        labels = metadata.get("labels", {})
        cluster_id = labels.get("io.tencent.bcs.clusterid")
        namespace = metadata.get("namespace")
        name = metadata.get("name")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
        )

        resp = self.pause_inst(
            request, project_id, project_kind, cluster_id, instance_id, name, namespace, curr_inst.category, conf=conf
        )
        if resp.data.get("code") == ErrorCode.NoError:
            self.update_instance_record_status(
                curr_inst, oper_type=app_constants.PAUSE_INSTANCE, status="Normal", is_bcs_success=True
            )
        return resp


class ResumeUpdateInstance(InstanceAPI):
    def resume_inst(
        self, request, project_id, project_kind, cluster_id, instance_id, name, namespace, category, conf=None
    ):
        self.can_operate(category)
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=name,
            resource_id=instance_id,
            extra=json.dumps({"config": conf, "namespace": namespace}),
            description=_("应用恢复滚动升级"),
        ).log_modify():
            resp = self.resume_update_deployment(
                request, project_id, cluster_id, namespace, name, kind=project_kind, category=category
            )
        return resp

    def resume_online_app(self, request, project_id, project_kind):
        cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        return self.resume_inst(request, project_id, project_kind, cluster_id, 0, name, namespace, category)

    def put(self, request, project_id, instance_id, instance_name):
        """恢复更新"""
        project_kind = self.project_kind(request)
        if str(instance_id) == "0":
            return self.resume_online_app(request, project_id, project_kind)
        # 获取instance info
        inst_info = self.get_instance_info(instance_id)
        # 获取namespace
        curr_inst = inst_info[0]

        conf = self.get_common_instance_conf(curr_inst)
        metadata = conf.get("metadata", {})
        labels = metadata.get("labels", {})
        cluster_id = labels.get("io.tencent.bcs.clusterid")
        namespace = metadata.get("namespace")
        name = metadata.get("name")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
        )
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=name,
            resource_id=instance_id,
            extra=json.dumps({"config": conf, "namespace": namespace}),
            description=_("应用恢复滚动升级"),
        ).log_modify():
            resp = self.resume_update_deployment(
                request, project_id, cluster_id, namespace, name, kind=project_kind, category=curr_inst.category
            )
        if resp.data.get("code") == ErrorCode.NoError:
            self.update_instance_record_status(
                curr_inst, oper_type=app_constants.RESUME_INSTANCE, status="Normal", is_bcs_success=True
            )
        return resp


class DeleteInstance(InstanceAPI):
    def get_enforce(self, request):
        # 是否强制删除
        enforce = request.GET.get("enforce")
        if enforce not in [None, "0"]:
            enforce = 1
        else:
            enforce = 0
        return enforce

    def delete(self, request, project_id, instance_id, instance_name):
        """删除实例"""
        # 是否强制删除
        enforce = self.get_enforce(request)
        # 获取kind
        project_kind = self.project_kind(request)
        if str(instance_id) == "0":
            cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
            return self.delete_instance(
                request, project_id, cluster_id, namespace, name, category=category, kind=project_kind, enforce=enforce
            )

        # 获取instance info
        inst_info = self.get_instance_info(instance_id)
        # 所有状态都可以强删
        if not enforce:
            if inst_info[0].is_deleted:
                return utils.APIResponse({"code": 400, "message": _("实例已经删除，请勿重复操作!")})
            if inst_info[0].oper_type == "delete":
                return utils.APIResponse({"code": 400, "message": _("已执行删除，请勿重复操作!")})
        # 如果实例化时，如果失败的话，直接标记为删除
        curr_inst = inst_info[0]

        conf = self.get_common_instance_conf(curr_inst)
        metadata = conf.get("metadata", {})
        labels = metadata.get("labels", {})
        cluster_id = labels.get("io.tencent.bcs.clusterid")
        namespace = metadata.get("namespace")
        name = metadata.get("name")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
        )
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=name,
            resource_id=instance_id,
            extra=json.dumps({"config": conf, "namespace": namespace}),
            description=_("应用删除操作"),
        ).log_delete():
            resp = self.delete_instance(
                request,
                project_id,
                cluster_id,
                namespace,
                name,
                category=curr_inst.category,
                kind=project_kind,
                inst_id_list=[curr_inst.id],
                enforce=enforce,
            )
        if resp.data.get("code") == ErrorCode.NoError:
            self.update_instance_record_status(
                inst_info[0], oper_type=app_constants.DELETE_INSTANCE, status="Deleting", is_bcs_success=True
            )

        return resp


class ReCreateInstance(InstanceAPI):
    """重新创建实例"""

    def delete_instance_oper(
        self, request, cluster_id, ns_name, instance_name, project_id=None, category=APPLICATION_CATEGORY, kind=2
    ):
        """删除实例"""
        resp = self.delete_instance(
            request, project_id, cluster_id, ns_name, instance_name, category=category, kind=kind
        )
        logger.error("curr_error: %s" % resp.data)
        if resp.data.get("code") != ErrorCode.NoError:
            raise error_codes.APIError(resp.data.get("message"))

        return resp

    def get_category_info(self, request, project_id, cluster_id, project_kind, inst_name, namespace, category):
        """
        针对这种特殊的情况，查询一遍category信息
        """
        flag, resp = self.get_application_deploy_info(
            request,
            project_id,
            cluster_id,
            inst_name,
            category=category,
            project_kind=project_kind,
            namespace=namespace,
            field="data",
        )
        if not flag:
            return resp
        return resp.get("data")

    def get_online_app_conf(self, request, project_id, project_kind):
        cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        # get the online yaml
        online_app_conf = self.online_app_conf(
            request, project_id, project_kind, cluster_id, name, namespace, category
        )
        return online_app_conf, name, namespace, category, cluster_id

    def post(self, request, project_id, instance_id, instance_name):
        project_kind = self.project_kind(request)
        if str(instance_id) == "0":
            conf, name, namespace, category, cluster_id = self.get_online_app_conf(request, project_id, project_kind)
            ns_name_id = self.get_namespace_name_id(request, project_id)
            curr_inst_ns_id = ns_name_id.get(namespace)
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, "", curr_inst_ns_id, source_type=app_constants.NOT_TMPL_SOURCE_TYPE
            )
        else:
            # 获取instance info
            inst_info = self.get_instance_info(instance_id)
            # 获取namespace
            curr_inst = inst_info[0]

            conf = self.get_common_instance_conf(curr_inst)
            metadata = conf.get("metadata", {})
            labels = metadata.get("labels", {})
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            namespace = metadata.get("namespace")
            name = metadata.get("name")
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
            )
            category = curr_inst.category
            # set the oper type
            inst_info.update(oper_type=app_constants.REBUILD_INSTANCE)
        # 删除节点
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=name,
            resource_id=instance_id,
            extra=json.dumps({"config": conf, "namespace": namespace}),
            description=_("应用重新创建"),
        ).log_add():
            # if exist_inst:
            self.delete_instance_oper(
                request, cluster_id, namespace, name, project_id=project_id, category=category, kind=project_kind
            )
            if str(instance_id) != "0":
                # 更新instance 操作
                try:
                    self.update_instance_record_status(
                        curr_inst, app_constants.REBUILD_INSTANCE, created=datetime.now(), is_bcs_success=True
                    )
                except Exception as error:
                    logger.error(u"更新重建操作实例状态失败，详情: %s" % error)

            from backend.celery_app.tasks import application as app_model

            # 启动任务
            app_model.application_polling_task.delay(
                request.user.token.access_token,
                instance_id,
                cluster_id,
                name,
                category,
                project_kind,
                namespace,
                project_id,
                username=request.user.username,
                conf=conf,
            )
        return utils.APIResponse(
            {
                "message": _("下发任务成功"),
            }
        )
