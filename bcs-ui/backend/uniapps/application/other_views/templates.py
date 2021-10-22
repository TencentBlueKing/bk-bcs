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

针对模板集的操作
"""
import json
import logging
from datetime import datetime
from itertools import groupby

from django.db.models import Q
from django.utils.translation import ugettext_lazy as _

from backend.bcs_web.audit_log import client
from backend.components import paas_cc
from backend.container_service.projects.base.constants import LIMIT_FOR_ALL_DATA
from backend.templatesets.legacy_apps.configuration.models import MODULE_DICT, Template
from backend.templatesets.legacy_apps.configuration.utils import to_bcs_res_name
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.utils.errcodes import ErrorCode

from .. import constants as app_constants
from ..base_views import BaseAPI, error_codes
from ..utils import APIResponse

logger = logging.getLogger(__name__)
SKIP_CATEGORY = ["application", "deployment"]


class TemplateNamespace(BaseAPI):
    def get_cluster_env_map(self, request, project_id):
        return self.get_cluster_id_env(request, project_id)

    def get_entity(self, info, category):
        """"""
        try:
            entity = json.loads(info.instance_entity)
        except Exception as error:
            logger.error(u"解析entity出现异常，ID：%s, 详情: %s" % (info.id, error))
            return []
        return entity.get(category) or []

    def get_active_ns(self, muster_id, show_version_name, category, res_name):
        """获取正在使用的实例"""
        tmpl_version_info = VersionInstance.objects.filter(
            is_deleted=False, show_version_name=show_version_name, template_id=muster_id
        )
        instance_id_list = [info.id for info in tmpl_version_info]

        ins_set = InstanceConfig.objects.filter(
            instance_id__in=instance_id_list, is_deleted=False, is_bcs_success=True
        ).exclude(Q(oper_type=app_constants.DELETE_INSTANCE) | Q(ins_state=InsState.NO_INS.value))
        if category != "ALL":
            ins_set = ins_set.filter(category=category, name=res_name)

        inst_info = ins_set.values("instance_id")
        inst_active_id_list = [info["instance_id"] for info in inst_info]
        # 获取namespace
        ret_data = []
        for info in tmpl_version_info:
            if info.id in inst_active_id_list:
                ret_data.append(info.ns_id)
        return ret_data

    def get_ns_info(self, request, project_id, ns_id_list):
        """获取ns信息"""
        resp = paas_cc.get_namespace_list(request.user.token.access_token, project_id, limit=LIMIT_FOR_ALL_DATA)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"))
        results = (resp.get("data") or {}).get("results") or []
        return results

    def get(self, request, project_id, muster_id):
        """获取命名空间"""
        # 获取参数
        group_by = request.GET.get('group_by') or "env_type"
        category = request.GET.get("category")
        show_version_name = request.GET.get('show_version_name')
        res_name = request.GET.get('res_name')
        perm_can_use = request.GET.get('perm_can_use')
        if perm_can_use == '1':
            perm_can_use = True
        else:
            perm_can_use = False

        # 前端的category转换为后台需要的类型
        if category != 'ALL':
            project_kind = request.project.kind
            category = to_bcs_res_name(project_kind, category)

        if category != 'ALL' and category not in MODULE_DICT:
            raise error_codes.CheckFailed(f'category: {category} does not exist')
        # 获取被占用的ns，没有处于删除中和已删除
        ns_id_list = self.get_active_ns(muster_id, show_version_name, category, res_name)
        # 查询ns信息
        results = self.get_ns_info(request, project_id, ns_id_list)
        # 解析&排序
        cluster_env_map = self.get_cluster_env_map(request, project_id)
        results = filter(lambda x: x["id"] in ns_id_list, results)
        results = [
            {
                'name': k,
                'cluster_name': cluster_env_map.get(k, {}).get('cluster_name', k),
                'environment_name': _("正式")
                if cluster_env_map.get(k, {}).get('cluster_env_str', '') == 'prod'
                else _("测试"),
                'results': sorted(list(v), key=lambda x: x['id'], reverse=True),
            }
            for k, v in groupby(sorted(results, key=lambda x: x[group_by]), key=lambda x: x[group_by])
        ]
        # ordering = [i.value for i in constants.EnvType]
        # results = sorted(results, key=lambda x: ordering.index(x['name']))
        ret_data = []
        for info in results:
            for item in info["results"] or []:
                item["muster_id"] = muster_id
                item["environment"] = cluster_env_map.get(item["cluster_id"], {}).get("cluster_env_str")
            info["results"] = self.bcs_perm_handler(
                request, project_id, info["results"], filter_use=perm_can_use, ns_id_flag="id", ns_name_flag="name"
            )
            if info["results"]:
                ret_data.append(info)
        return APIResponse({"data": ret_data})


class DeleteTemplateInstance(BaseAPI):
    def get_muster_name(self, muster_id):
        """获取模板名称"""
        muster_name_list = Template.objects.filter(id=muster_id).values("name")
        if not muster_name_list:
            raise error_codes.CheckFailed(_("模板集ID: {} 不存在").format(muster_id))
        return muster_name_list[0]["name"]

    def get_template_name(self, template_id, category):
        """获取模板名称"""
        info = MODULE_DICT[category].objects.filter(id=template_id).values("name")
        if not info:
            raise error_codes.CheckFailed(_("模板ID: {} 不存在").format(template_id))
        return info[0]["name"]

    def check_project_muster(self, project_id, muster_id):
        """判断项目和集群"""
        if not Template.objects.filter(project_id=project_id, id=muster_id).exists():
            raise error_codes.CheckFailed(_("项目:{},模板集: {} 不存在!").format(project_id, muster_id))

    def get_instance_info(self, ns_id_list, name, category=None):
        """获取实例信息"""
        inst_info = InstanceConfig.objects.filter(name__in=name, namespace__in=ns_id_list, is_deleted=False).exclude(
            oper_type=app_constants.DELETE_INSTANCE
        )
        if category:
            inst_info = inst_info.filter(category=category)
        ret_data = {}
        for info in inst_info:
            try:
                conf = self.get_common_instance_conf(info)
            except Exception:
                continue
            metadata = conf.get("metadata") or {}
            namespace = metadata.get("namespace")
            labels = metadata.get("labels")
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            ret_data[info.id] = {
                "cluster_id": cluster_id,
                "namespace": namespace,
                "instance_name": info.name,
                "muster_id": labels.get("io.tencent.paas.templateid"),
                "info": info,
            }
        return ret_data

    def delete_single(self, request, data, project_id, muster_id, show_version_name, res_name, muster_name):
        """删除单个类型"""
        ns_id_list = data.get("namespace_list")
        category = data.get("category")
        if not (ns_id_list and category):
            raise error_codes.CheckFailed(_("参数不能为空!"))

        project_kind = request.project.kind
        category = to_bcs_res_name(project_kind, category)

        if category not in MODULE_DICT:
            raise error_codes.CheckFailed(f'category: {category} does not exist')
        # 获取要删除的实例的信息
        inst_info = self.get_instance_info(ns_id_list, [res_name], category=category)
        # 获取项目信息
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=res_name,
            resource_id=res_name,
            extra=json.dumps(
                {
                    "muster_id": muster_id,
                    "show_version_name": show_version_name,
                    "res_name": res_name,
                    "category": category,
                }
            ),
            description=_("删除模板集实例"),
        ).log_delete():
            return self.delete_handler(request, inst_info, project_id, project_kind)

    def delete_all(self, request, data, project_id, muster_id, show_version_name, res_name, muster_name):
        """删除所有类型"""
        ns_id_list = data.get("namespace_list")
        id_list = data.get("id_list")
        if not (ns_id_list and id_list):
            raise error_codes.CheckFailed(_("参数不能为空!"))
        # 获取项目信息
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        # 获取所有模板名称
        tmpl_category_name_map = {}
        for info in id_list:
            category_name = info["id"]
            tmpl_category_name_map[category_name] = info["category"]
        # 获取要删除的实例信息
        inst_info = self.get_instance_info(ns_id_list, tmpl_category_name_map.keys())
        # instance_version_ids = [val["info"].instance_id for key, val in inst_info.items()]
        with client.ContextActivityLogClient(
            project_id=project_id,
            user=request.user.username,
            resource_type="instance",
            resource=muster_name,
            resource_id=muster_id,
            extra=json.dumps(
                {
                    "muster_id": muster_id,
                    "show_version_name": show_version_name,
                    "tmpl_category_name_map": tmpl_category_name_map,
                }
            ),
            description=_("删除模板集实例"),
        ).log_delete():
            resp = self.delete_handler(request, inst_info, project_id, project_kind)
            return resp
            # if resp.data.get("code") != ErrorCode.NoError:
            #     return resp
            # else:
            #     # 更新version instance表记录
            #     try:
            #         VersionInstance.objects.filter(id__in=instance_version_ids).update(
            #             is_deleted=True, deleted_time=datetime.now()
            #         )
            #     except Exception as error:
            #         logger.error(u"删除失败，详情: %s" % error)
            #         return APIResponse({
            #             "code": 400,
            #             "message": u"更新实例状态失败，已通知管理员"
            #         })
            #     return resp

    def delete_handler(self, request, inst_info, project_id, project_kind):
        """删除操作"""
        oper_error_inst = []
        deleted_id_list = []
        for inst_id, info in inst_info.items():
            # 判断权限
            self.bcs_single_app_perm_handler(request, project_id, info["muster_id"], info["info"].namespace)
            # 针对0/0的情况先查询一次
            if info["info"].category in app_constants.ALL_CATEGORY_LIST:
                if not self.get_category_info(
                    request,
                    project_id,
                    info["cluster_id"],
                    project_kind,
                    info["instance_name"],
                    info["namespace"],
                    info["info"].category,
                ):
                    deleted_id_list.append(inst_id)
                    continue
            resp = self.delete_instance(
                request,
                project_id,
                info["cluster_id"],
                info["namespace"],
                info["instance_name"],
                category=info["info"].category,
                kind=project_kind,
                inst_id_list=[info["info"].id],
            )
            if resp.data.get("code") != ErrorCode.NoError:
                logger.error("删除实例ID: %s 失败, 详情: %s" % (inst_id, resp.data.get("message")))
                oper_error_inst.append(
                    "%s::%s: %s" % (info["namespace"], info["instance_name"], resp.data.get("message"))
                )
                continue
            # 更新状态
            if info["info"].category in SKIP_CATEGORY:
                self.update_instance_record_status(info["info"], app_constants.DELETE_INSTANCE, status="Deleting")
            else:
                self.update_instance_record_status(
                    info["info"],
                    app_constants.DELETE_INSTANCE,
                    status="Deleted",
                    category=info["info"].category,
                    deleted_time=datetime.now(),
                )
        # 如果存在0/0的情况
        if deleted_id_list:
            InstanceConfig.objects.filter(id__in=deleted_id_list).update(is_deleted=True, deleted_time=datetime.now())
        if oper_error_inst:
            return APIResponse(
                {"code": 400, "message": _("存在删除失败情况，(命名空间:实例名称)详情: {}").format(";".join(oper_error_inst))}
            )
        return APIResponse({"message": _("操作成功!")})

    def delete(self, request, project_id, muster_id):
        """删除某一个版本下命名空间和instance信息"""
        show_version_name = request.GET.get('show_version_name')
        res_name = request.GET.get('res_name')
        # 判断项目和模板集
        self.check_project_muster(project_id, muster_id)
        data = dict(request.data)
        muster_name = self.get_muster_name(muster_id)
        if data.get("id_list") and show_version_name:
            return self.delete_all(request, data, project_id, muster_id, show_version_name, res_name, muster_name)
        else:
            return self.delete_single(request, data, project_id, muster_id, show_version_name, res_name, muster_name)
