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
import base64
import json
import logging
from datetime import datetime

import yaml
from django.utils.translation import ugettext_lazy as _
from rest_framework import views
from rest_framework.exceptions import ValidationError

from backend.accounts import bcs_perm
from backend.celery_app.tasks.application import delete_instance_task
from backend.components import paas_cc
from backend.components.bcs.k8s import K8SClient
from backend.container_service.projects.base.constants import ProjectKindID
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.iam.permissions.resources.templateset import TemplatesetPermCtx, TemplatesetPermission
from backend.templatesets.legacy_apps.configuration.constants import K8sResourceName
from backend.templatesets.legacy_apps.configuration.models import Template
from backend.templatesets.legacy_apps.instance.constants import EventType
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, InstanceEvent
from backend.utils.basic import getitems
from backend.utils.errcodes import ErrorCode
from backend.utils.error_codes import error_codes

from .common_views.serializers import BaseNotTemplateInstanceParamsSLZ
from .constants import FUNC_MAP, NOT_TMPL_SOURCE_TYPE, OWENER_REFERENCE_MAP
from .drivers import BCSDriver
from .utils import APIResponse, cluster_env, image_handler, retry_requests

logger = logging.getLogger(__name__)

DEFAULT_RESPONSE = {"code": 0}
DEFAULT_ERROR_CODE = ErrorCode.UnknownError

# 模板集删除时, 需要一起删除的资源
CASCADE_DELETE_RESOURCES = [
    "service",
    "secret",
    "configmap",
    "K8sSecret",
    "K8sConfigMap",
    "K8sService",
    "K8sIngress",
    "K8sHPA",
]


class BaseAPI(views.APIView):
    def get_params_for_client(self, request):
        name = request.GET.get("name")
        namespace = request.GET.get("namespace")
        category = request.GET.get("category")
        if not (name and namespace and category):
            raise error_codes.CheckFailed(_("参数[name]、[namespace]、[category]不能为空"))
        return name, namespace, category

    def get_instance_info(self, id):
        """获取instance info"""
        inst_info = InstanceConfig.objects.filter(id=id, is_deleted=False)
        if not inst_info:
            raise error_codes.CheckFailed(_("没有查询到相应的记录"))
        return inst_info

    def get_common_instance_conf(self, info):
        """获取instance conf"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError(_("Instance config解析异常"))
        return conf

    def project_kind(self, request):
        """
        处理项目project info
        """
        kind = request.project.get("kind")
        if kind not in [1]:
            raise error_codes.CheckFailed(_("项目类型必须为k8s, 请确认后重试!"))
        return kind

    def get_project_kind(self, request, project_id):
        """获取项目类型"""
        return True, request.project.kind

    def get_project_clusters(self, request, project_id):
        """查询项目下所有的集群"""
        resp = paas_cc.get_all_clusters(request.user.token.access_token, project_id, limit=1000, offset=0)
        if resp.get("code") != ErrorCode.NoError:
            return False, APIResponse({"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message")})
        # 获取集群id list
        cluster_id_list = []
        data = resp.get("data") or {}
        for info in data.get("results") or []:
            cluster_id_list.append(info["cluster_id"])
        return True, cluster_id_list

    def get_project_cluster_info(self, request, project_id):
        """获取项目下所有集群信息"""
        resp = paas_cc.get_all_clusters(request.user.token.access_token, project_id, desire_all_data=1)
        if resp.get("code") != ErrorCode.NoError:
            raise error_codes.APIError.f(resp.get("message"))
        return resp.get("data") or {}

    def get_cluster_id_env(self, request, project_id):
        """获取集群和环境"""
        data = self.get_project_cluster_info(request, project_id)
        if not data.get("results"):
            # raise error_codes.APIError.f("没有查询到集群信息")
            return {}
        cluster_results = data.get("results") or []
        return {
            info["cluster_id"]: {
                "cluster_name": info["name"],
                "cluster_env": cluster_env(info["environment"]),
                "cluster_env_str": cluster_env(info["environment"], ret_num_flag=False),
            }
            for info in cluster_results
            if not info["disabled"]
        }

    def get_namespaces(self, request, project_id):
        """获取namespace"""
        resp = paas_cc.get_namespace_list(request.user.token.access_token, project_id, limit=10000, offset=0)
        if resp.get("code") != ErrorCode.NoError:
            return False, APIResponse({"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message")})
        return True, resp["data"]

    def get_namespace_name_id(self, request, project_id):
        resp = paas_cc.get_namespace_list(request.user.token.access_token, project_id, limit=10000, offset=0)
        if resp.get('code') != ErrorCode.NoError:
            raise error_codes.CheckFailed.f(resp.get('message'))
        data = resp.get('data', {}).get('results', [])
        return {info.get('name'): info.get('id') for info in data if info}

    def get_namespace_info(self, request, project_id, ns_id):
        """获取单个namespace"""
        resp = paas_cc.get_namespace(request.user.token.access_token, project_id, ns_id)
        if resp.get("code") != ErrorCode.NoError:
            return False, APIResponse({"code": resp.get("code") or DEFAULT_ERROR_CODE, "message": resp.get("message")})
        if not resp.get("data"):
            return False, APIResponse({"code": 400, "message": _("查询记录为空!")})

        return True, resp["data"]

    def get_cluster_by_ns_name(self, request, project_id, ns_name):
        """通过命名空间获取集群"""
        ok, all_ns_info = self.get_namespaces(request, project_id)
        if not ok:
            raise error_codes.APIError.f(all_ns_info.data.get("message"))
        for info in all_ns_info.get("results") or []:
            if info["name"] == ns_name:
                return info["cluster_id"]
        raise error_codes.CheckFailed(_("没有查询到命名空间对应的集群ID"))

    def get_k8s_rs_info(self, request, project_id, cluster_id, ns_name, resource_name):
        """获取k8s deployment副本信息"""
        ret_data = {}
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        extra = {"data.metadata.ownerReferences.name": resource_name}
        extra_encode = base64_encode_params(extra)
        resp = client.get_rs({"extra": extra_encode, "namespace": ns_name, "field": "resourceName,data.status"})
        if resp.get("code") != 0:
            raise error_codes.APIError.f(resp.get("message") or _("查询出现异常"), replace=True)
        data = resp.get("data") or []
        if not data:
            return ret_data
        # NOTE: 因为线上存在revision history，需要忽略掉replica为0的rs
        rs_name_list = [
            info["resourceName"]
            for info in data
            if info.get("resourceName") and getitems(info, ["data", "status", "replicas"], 0) > 0
        ]

        return rs_name_list

    def get_k8s_pod_info(
        self,
        request,
        project_id,
        cluster_id,
        ns_name,
        owner_ref_name=None,
        field=None,
        pod_name=None,
        owner_ref_kind=None,
    ):
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        extra_encode = None
        if owner_ref_name:
            extra = {"data.metadata.ownerReferences.name": owner_ref_name}
            if owner_ref_kind:
                extra["data.metadata.ownerReferences.kind"] = owner_ref_kind
            extra_encode = base64_encode_params(extra)
        field_list = field or [
            "resourceName,createTime",
            "data.status.podIP",
            "data.status.hostIP",
            "data.status.phase",
            "data.status.containerStatuses",
        ]
        params = {"namespace": ns_name}
        if pod_name:
            params["name"] = pod_name
        resp = client.get_pod(extra=extra_encode, field=",".join(field_list), params=params)
        if resp.get("code") != 0:
            raise error_codes.APIError.f(resp.get("message") or _("查询出现异常"), replace=True)

        return resp

    def get_pod_or_taskgroup(
        self,
        request,
        project_id,
        cluster_id,
        field=None,
        app_name=None,
        taskgroup_name=None,
        ns_name=None,
        kind=2,
        category=None,
    ):
        """获取taskgroup或者pod"""
        # 添加owner reference的kind属性，用来过滤pod关联的deployment/sts等类型
        owner_ref_kind = OWENER_REFERENCE_MAP.get(category)
        if taskgroup_name:
            pod_name = taskgroup_name
            rs_name = None
        elif app_name:
            pod_name = None
            rs_name = app_name
            if category in ["deployment", "K8sDeployment"]:
                rs_name = self.get_k8s_rs_info(request, project_id, cluster_id, ns_name, app_name)
                if not rs_name:
                    return True, []
        else:
            pod_name = None
            rs_name = None
        resp = self.get_k8s_pod_info(
            request,
            project_id,
            cluster_id,
            ns_name,
            owner_ref_name=rs_name,
            field=field,
            pod_name=pod_name,
            owner_ref_kind=owner_ref_kind,
        )

        if resp.get("code") != ErrorCode.NoError:
            return False, APIResponse({"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message")})
        return True, resp["data"]

    def create_instance(self, request, project_id, cluster_id, ns, data, category="application", kind=2):
        """创建实例"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        curr_func = getattr(client, FUNC_MAP[category] % "create")
        resp = curr_func(ns, data)
        if resp.get("code") != ErrorCode.NoError:
            return APIResponse({"code": resp.get("code") or DEFAULT_ERROR_CODE, "message": resp.get("message")})
        return APIResponse({"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message")})

    def delete_instance(
        self,
        request,
        project_id,
        cluster_id,
        ns,
        instance_name,
        category="application",
        kind=2,
        inst_id_list=None,
        enforce=0,
    ):
        """删除instance"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        # deployment 需要级联删除 res\pod; daemonset/job/statefulset 需要级联删除 pod
        if FUNC_MAP[category] in ['%s_deployment', '%s_daemonset', '%s_job', '%s_statefulset']:
            fun_prefix = 'deep_delete'
        else:
            fun_prefix = 'delete'
        curr_func = getattr(client, FUNC_MAP[category] % fun_prefix)

        resp = curr_func(ns, instance_name)
        # 级联删除，会返回空
        if resp is None:
            # 启动后台任务，轮训任务状态
            if inst_id_list:
                delete_instance_task.delay(request.user.token.access_token, inst_id_list, kind)
            return APIResponse({"code": ErrorCode.NoError, "message": _("删除成功")})

        # response
        msg = resp.get("message")
        # message中有not found或者node does not exist时，认为已经删除成功
        # 状态码为正常或者满足不存在条件时，认为已经删除成功
        if (resp.get("code") in [ErrorCode.NoError]) or ("not found" in msg) or ("node does not exist" in msg):
            return APIResponse({"code": ErrorCode.NoError, "message": _("删除成功")})

        return APIResponse({"code": resp.get("code"), "message": msg})

    def get_k8s_application_deploy_status(
        self, request, project_id, cluster_id, instance_name, category="application", namespace=None, field=None
    ):
        """获取k8s下application和deployment状态"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        curr_func = getattr(client, FUNC_MAP[category] % "get")
        params = {"name": instance_name, "namespace": namespace, "field": ",".join(field)}
        resp = retry_requests(curr_func, params=params)
        if resp.get("code") != ErrorCode.NoError:
            return False, APIResponse({"code": resp.get("code") or DEFAULT_ERROR_CODE, "message": resp.get("message")})

        return True, resp

    def get_application_deploy_info(
        self,
        request,
        project_id,
        cluster_id,
        instance_name,
        category="application",
        project_kind=ProjectKindID,
        namespace=None,
        field=None,
    ):
        """获取详情"""
        return self.get_k8s_application_deploy_status(
            request,
            project_id,
            cluster_id,
            instance_name,
            category=category,
            namespace=namespace,
            field=field,
        )

    def get_k8s_app_deploy_with_post(
        self, request, project_id, cluster_id, instance_name=None, category="application", namespace=None, field=None
    ):
        """获取k8s下application和deployment状态"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        curr_func = getattr(client, "%s_with_post" % (FUNC_MAP[category] % "get"))
        params = {"name": instance_name, "namespace": namespace, "field": ",".join(field)}
        resp = retry_requests(curr_func, params=params)
        if resp.get("code") != ErrorCode.NoError:
            return False, APIResponse({"code": resp.get("code") or DEFAULT_ERROR_CODE, "message": resp.get("message")})

        return True, resp

    def get_app_deploy_with_post(
        self,
        request,
        project_id,
        cluster_id,
        instance_name=None,
        category="application",
        project_kind=ProjectKindID,
        namespace=None,
        field=None,
    ):
        """获取详情"""
        return self.get_k8s_app_deploy_with_post(
            request,
            project_id,
            cluster_id,
            instance_name=instance_name,
            category=category,
            namespace=namespace,
            field=field,
        )

    def get_application_deploy_status(
        self,
        request,
        project_id,
        cluster_id,
        instance_name,
        category="application",
        project_kind=ProjectKindID,
        namespace=None,
        field=None,
    ):
        """获取application和deployment"""
        result, resp = self.get_k8s_application_deploy_status(
            request, project_id, cluster_id, instance_name, category=category, namespace=namespace, field=field
        )

        return result, resp

    def get_instances(self, request, project_id, cluster_id, kind=2):
        """拉取项目下的所有Instance"""
        resp = DEFAULT_RESPONSE
        if resp.get("code") != ErrorCode.NoError:
            return False, APIResponse({"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message")})
        return True, resp["data"]

    def update_instance(
        self, request, project_id, cluster_id, ns, instance_num, conf, kind=2, category=None, name=None
    ):  # noqa
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        curr_func = getattr(client, FUNC_MAP[category] % "update")
        resp = curr_func(ns, name, conf)

        if resp.get("code") != ErrorCode.NoError:
            return False, APIResponse(
                {"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message", _("请求出现异常!"))}
            )
        return True, APIResponse({"message": _("更新成功!")})

    def scale_instance(
        self, request, project_id, cluster_id, ns, app_name, instance_num, kind=2, category=None, data=None
    ):  # noqa
        """扩缩容"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        data["spec"]["replicas"] = int(instance_num)
        curr_func = getattr(client, FUNC_MAP[category] % "update")
        resp = curr_func(ns, app_name, data)

        if resp.get("code") != ErrorCode.NoError:
            return APIResponse(
                {"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message", _("请求出现异常!"))}
            )
        return APIResponse({"message": _("更新成功!")})

    def update_deployment(self, request, project_id, cluster_id, ns, data, kind=2, category=None, app_name=None):
        """滚动升级"""
        access_token = request.user.token.access_token
        client = K8SClient(access_token, project_id, cluster_id, None)
        curr_func = getattr(client, FUNC_MAP[category] % "update")
        resp = curr_func(ns, app_name, data)

        if resp.get("code") != ErrorCode.NoError:
            return APIResponse(
                {"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message", _("请求出现异常!"))}
            )
        return APIResponse({"message": _("更新成功!")})

    def cancel_update_deployment(self, request, project_id, cluster_id, ns, deployment_name, kind=2):
        """取消更新"""
        return APIResponse({"message": _("取消更新成功!")})

    def pause_update_deployment(self, request, project_id, cluster_id, ns, deployment_name, kind=1, category=None):
        """暂停更新"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        curr_func = getattr(client, FUNC_MAP[category] % "patch")
        params = {"spec": {"paused": True}}
        resp = curr_func(ns, deployment_name, params)

        if resp.get("code") != ErrorCode.NoError:
            return APIResponse(
                {"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message", _("请求出现异常!"))}
            )
        return APIResponse({"message": _("暂停更新成功!")})

    def resume_update_deployment(self, request, project_id, cluster_id, ns, deployment_name, kind=2, category=None):
        """重启更新"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        curr_func = getattr(client, FUNC_MAP[category] % "patch")
        params = {"spec": {"paused": False}}
        resp = curr_func(ns, deployment_name, params)
        if resp.get("code") != ErrorCode.NoError:
            return APIResponse(
                {"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message", _("请求出现异常!"))}
            )
        return APIResponse({"message": _("重启更新成功!")})

    def rescheduler_taskgroup(self, request, project_id, cluster_id, ns, instance_name, taskgroup_name, kind=2):
        """重启更新"""
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        resp = client.delete_pod(ns, taskgroup_name)

        if resp.get("code") != ErrorCode.NoError:
            return APIResponse(
                {"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message", _("请求出现异常!"))}
            )
        return APIResponse({"message": _("重新调度成功!")})

    def query_events(self, request, project_id, cluster_id, params):
        """查询事件"""
        client = BCSDriver(request, project_id, cluster_id)
        resp = client.get_events(params)
        if resp.get("code") != ErrorCode.NoError:
            return APIResponse(
                {"code": resp.get("code", DEFAULT_ERROR_CODE), "message": resp.get("message", _("请求出现异常!"))}
            )
        return APIResponse(
            {
                "data": {
                    "data": resp.get("data", []),
                    "total": 100 if resp.get("total", 0) > 100 else resp.get("total", 0),
                }
            }
        )

    def update_instance_record_status(
        self,
        info,
        oper_type,
        status="Running",
        deleted_time=None,
        category=None,
        created=None,
        inst_state=None,
        is_deleted=None,
        is_bcs_success=None,
    ):
        """更新单条记录状态"""
        info.oper_type = oper_type
        info.status = status
        if deleted_time:
            info.deleted_time = deleted_time
        # 删除模板集是级联删除的resource，其他更新时为none
        if category in CASCADE_DELETE_RESOURCES:
            info.is_deleted = True
            info.deleted_time = datetime.now()
        if created:
            info.created = created
        if inst_state:
            info.ins_state = inst_state
        if is_deleted is not None:
            info.is_deleted = is_deleted
        if is_bcs_success is not None:
            info.is_bcs_success = is_bcs_success
        info.save()

    def get_task_group_info_base(self, data, namespace=None):
        """获取pod信息"""
        ret_data = {}
        for info in data:
            info_data = info.get("data", {})
            task_group_name = info_data.get("metadata", {}).get("name")

            for info in info_data.get("containerStatuses", []):
                info["image"] = image_handler(info.get("image", ""))
                # 处理状态
            ret_data[task_group_name] = {
                "name": task_group_name,
                "namespace": namespace,
                "status": info_data.get("status"),
                "type": "taskgroup",
                "container_list": info_data.get("containerStatuses", []),
                "message": info_data.get("message", ""),
                "start_time": info_data.get("startTime", ""),
                "pod_ip": info_data.get("podIP", ""),
                "host_ip": info_data.get("hostIP", ""),
            }
        return ret_data

    def check_resource(self, request, project_id, config, cluster_id, req_instance_num):
        """校验资源是否满足"""
        conf = config.copy()
        pre_instance_num = int(conf["spec"]["instance"])
        if int(req_instance_num) > pre_instance_num:
            result = paas_cc.get_cluster(request.user.token.access_token, project_id, cluster_id)
            if result.get("code") != 0:
                raise error_codes.APIError(_("获取资源失败，请联系蓝鲸管理员解决"))

            data = result.get('data') or {}
            remain_cpu = data.get('remain_cpu') or 0
            remain_mem = data.get('remain_mem') or 0

            cpu_require = 0
            mem_require = 0
            # CPU单位是核心，有小数点，内存单位是M，和paas-cc返回一致
            for info in conf["spec"]["template"]["spec"]["containers"]:
                cpu_require += float(info["resources"]["limits"]["cpu"])
                mem_require += float(info["resources"]["limits"]["memory"])
            diff_instance_num = int(req_instance_num) - pre_instance_num
            if remain_cpu < (diff_instance_num * cpu_require):
                raise error_codes.CheckFailed(_("没有足够的CPU资源，请添加node或释放资源!"))
            if remain_mem < (diff_instance_num * mem_require):
                raise error_codes.CheckFailed(_("没有足够的内存资源，请添加node或释放资源!"))

    def event_log_record(self, inst_id, conf_inst_id, category, err_msg, resp_snap, username):
        """记录事件"""
        try:
            InstanceEvent(
                instance_config_id=inst_id,
                category=category,
                msg_type=EventType.REQ_FAILED.value,
                instance_id=conf_inst_id,
                msg=err_msg,
                creator=username,
                updator=username,
                resp_snapshot=json.dumps(resp_snap),
            ).save()
        except Exception as error:
            logger.error(u"存储实例化失败消息失败，详情: %s" % error)

    def json2yaml(self, conf):
        """json转yaml"""
        yaml_profile = yaml.safe_dump(conf)
        return yaml_profile

    def get_category_info(self, request, project_id, cluster_id, project_kind, inst_name, namespace, category):
        """
        针对0/0这种特殊的情况，查询一遍category信息
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
            raise error_codes.APIError.f(resp.data.get("message"))
        return resp.get("data")

    def bcs_perm_handler(
        self,
        request,
        project_id,
        data,
        filter_use=False,
        ns_id_flag="namespace_id",
        ns_name_flag='namespace',
        tmpl_view=True,
    ):  # noqa
        return data

    def validate_view_perms(self, request, project_id, muster_id, ns_id, source_type="模板集"):
        """查询资源时, 校验用户的查询权限"""
        resp = paas_cc.get_namespace(request.user.token.access_token, project_id, ns_id)
        data = resp.get('data')
        perm_ctx = NamespaceScopedPermCtx(
            username=request.user.username,
            project_id=project_id,
            cluster_id=data.get('cluster_id'),
            name=data.get('name'),
        )
        NamespaceScopedPermission().can_view(perm_ctx)

        if source_type == "模板集":
            muster_info = Template.objects.filter(id=muster_id, is_deleted=False).first()
            if not muster_info:
                raise error_codes.CheckFailed(_("没有查询到模板集信息"))
            perm_ctx = TemplatesetPermCtx(username=request.user.username, project_id=project_id, template_id=muster_id)
            TemplatesetPermission().can_view(perm_ctx)

    def bcs_single_app_perm_handler(self, request, project_id, muster_id, ns_id, source_type="模板集"):
        """针对具体资源的权限处理"""
        # 继承命名空间的权限
        resp = paas_cc.get_namespace(request.user.token.access_token, project_id, ns_id)
        data = resp.get('data')
        perm_ctx = NamespaceScopedPermCtx(
            username=request.user.username,
            project_id=project_id,
            cluster_id=data.get('cluster_id'),
            name=data.get('name'),
        )
        NamespaceScopedPermission().can_use(perm_ctx)

        if source_type == "模板集":
            muster_info = Template.objects.filter(id=muster_id, is_deleted=False).first()
            if not muster_info:
                raise error_codes.CheckFailed(_("没有查询到模板集信息"))
            # 继承模板集的权限
            perm_ctx = TemplatesetPermCtx(username=request.user.username, project_id=project_id, template_id=muster_id)
            TemplatesetPermission().can_instantiate(perm_ctx)

    def online_app_conf(self, request, project_id, project_kind, cluster_id, name, namespace, category):
        """针对非模板创建的应用，获取线上的配置"""
        conf = {}
        client = K8SClient(request.user.token.access_token, project_id, cluster_id, None)
        curr_func = FUNC_MAP[category] % "get"
        resp = getattr(client, curr_func)({"name": name, "namespace": namespace})
        if resp.get("code") != 0:
            raise error_codes.APIError.f(resp.get("message", _("获取应用线上配置异常，请联系管理员处理!")))
        data = resp.get("data") or []
        if not data:
            return {}
        data = data[0]
        # 组装数据
        conf["kind"] = data["resourceType"]
        conf["metadata"] = {}
        conf["spec"] = data["data"]["spec"]
        conf["metadata"]["name"] = data["data"]["metadata"]["name"]
        conf["metadata"]["namespace"] = data["data"]["metadata"]["namespace"]
        conf["metadata"]["labels"] = data["data"]["metadata"]["labels"]
        conf["metadata"]["annotations"] = data["data"]["metadata"]["annotations"]
        return conf


def base64_encode_params(info):
    """base64编码"""
    extra_json = bytes(json.dumps(info), "utf-8")
    extra_encode = base64.b64encode(extra_json)
    return extra_encode


class BaseInstanceView:
    def get_instance_info(self, inst_id):
        inst_info = InstanceConfig.objects.filter(id=inst_id, is_deleted=False).first()
        if not inst_info:
            raise error_codes.CheckFailed(f"instance({inst_id}) not found")
        return inst_info


class InstanceAPI(BaseAPI):
    def can_use_instance(self, request, project_id, ns_id):
        # TODO: 调整函数名，并且注意替换调用的地方
        self.bcs_single_app_perm_handler(request, project_id, None, ns_id, source_type=NOT_TMPL_SOURCE_TYPE)

    def get_instance_resource(self, request, project_id):
        """return cluster id，namespace name， instance name, instance category"""
        slz = BaseNotTemplateInstanceParamsSLZ(data=request.query_params)
        slz.is_valid(raise_exception=True)
        data = slz.validated_data
        ns_name_id_map = self.get_namespace_name_id(request, project_id)
        ns_id = ns_name_id_map.get(data['namespace'])
        # check perm
        self.validate_view_perms(request, project_id, None, ns_id, source_type=NOT_TMPL_SOURCE_TYPE)
        return data['cluster_id'], data['namespace'], data['name'], data['category']

    def _from_template(self, instance_id):
        """判断是否来源于表单模式的模板集
        现阶段，只有表单模式的模板集会写入instance configure表，生成instance id；
        针对非表单模式，没有具体的id，统一使用0标识
        """
        if str(instance_id) == "0":
            return False
        return True

    def can_operate(self, resource_kind):
        # 为避免类型的大小写不一致，采用包含的方式处理
        if resource_kind.lower() in K8sResourceName.K8sStatefulSet.value.lower():
            raise ValidationError(_("StatefulSet类型不允许此操作"))
        return True
