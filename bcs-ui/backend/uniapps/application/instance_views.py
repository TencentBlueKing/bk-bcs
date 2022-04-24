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
import time
from datetime import datetime

from django.conf import settings
from django.db.models import Q
from django.utils.translation import ugettext_lazy as _
from rest_framework.renderers import BrowsableAPIRenderer
from rest_framework.response import Response

from backend.bcs_web.audit_log import client
from backend.components import data, paas_cc
from backend.container_service.projects.base.constants import ProjectKindID
from backend.iam.permissions.resources.namespace_scoped import NamespaceScopedPermCtx, NamespaceScopedPermission
from backend.templatesets.legacy_apps.configuration.models import MODULE_DICT, ShowVersion, VersionedEntity
from backend.templatesets.legacy_apps.configuration.utils import check_var_by_config
from backend.templatesets.legacy_apps.instance import utils as inst_utils
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, InstanceEvent, VersionInstance
from backend.templatesets.var_mgmt.models import Variable
from backend.utils.basic import getitems
from backend.utils.errcodes import ErrorCode
from backend.utils.renderers import BKAPIRenderer

from . import constants as app_constants
from . import utils
from .base_views import BaseAPI, InstanceAPI, error_codes
from .serializers import BatchDeleteResourceSLZ
from .utils import APIResponse, image_handler
from .views import UpdateInstanceNew

try:
    from backend.container_service.observability.datalog import utils as datalog_utils
except ImportError:
    from backend.container_service.observability.datalog_ce import utils as datalog_utils

logger = logging.getLogger(__name__)

K8SDEPLOYMENT_CATEGORY = "K8sDeployment"
K8SJOB_CATEGORY = "K8sJob"
K8SDAEMONSET_CATEGORY = "k8sDaemonSet"
K8SSTATEFULSET_CATEGORY = "K8sStatefulSet"
DEFAULT_ERROR_CODE = ErrorCode.UnknownError
ALL_LIMIT = 10000
CLUSTER_TYPE = [1, 2, "1", "2"]
APP_STATUS = [1, 2, "1", "2"]
DEFAULT_INSTANCE_NUM = 0


class BaseTaskgroupCls(InstanceAPI):
    def common_handler_for_platform(self, request, project_id, instance_id, project_kind, field=None):
        """公共信息的处理"""
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
        self.validate_view_perms(request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace)
        return cluster_id, namespace, [name], curr_inst.category

    def common_handler_for_client(self, request, project_id):
        """针对非平台创建的应用的公共处理"""
        cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        return cluster_id, namespace, [name], category

    def common_handler(self, request, project_id, instance_id, project_kind, field=None):
        if not self._from_template(instance_id):
            return self.common_handler_for_client(request, project_id)
        return self.common_handler_for_platform(request, project_id, instance_id, project_kind, field=field)


class QueryAllTaskgroups(BaseTaskgroupCls):
    def get_taskgroup_info(self, data):
        """获取taskgroup信息"""
        ret_data = []
        for info in data:
            info_data = info.get("data", {})
            metadata = info_data.pop("metadata", {})
            task_group_name = metadata.get("name")
            info_data["name"] = task_group_name
            info_data["current_time"] = datetime.now()
            info_data["host_ip"] = info_data.pop("hostIP", "")
            info_data["start_time"] = info_data.pop("startTime")
            info_data["reason"] = ""
            ret_data.append(info_data)
        return ret_data

    def get_pod_conditions(self, data):
        """处理pod conditions"""
        if not data:
            return "", ""
        info = data[0]
        return info.get("message") or "", info.get("reason") or ""

    def is_pod_normal(self, status):
        """获取pod的状态是否正常
        1. 如果pod phase不处于Succeeded或者Running则肯定是异常的
        2. 如果pod下的container中包含reason和message，则认为处于了异常
        """
        # 查询pod状态
        pod_status = status.get("phase")
        if pod_status not in ["Succeeded", "Running"]:
            return False
        # 查询container状态
        container_statuses = status.get("containerStatuses") or []
        for info in container_statuses:
            state = info.get("state") or {}
            for item in state.values():
                # 如果包含reason和message字段，认为container是不正常的
                if "reason" in item and "message" in item:
                    return False
        return True

    def get_pod_info(self, data):
        """获取pod信息
        TODO: message信息怎样获取
        """
        ret_data = []
        for info in data:
            item = {
                "start_time": info.get("createTime"),
                "name": info.get("resourceName"),
                "current_time": datetime.now(),
            }
            info_data = info.get("data") or {}
            status = info_data.get("status") or {}
            condition = status.get("conditions") or []
            message, reason = self.get_pod_conditions(condition)
            item.update(
                {
                    "status": status.get("phase"),
                    "podIP": status.get("podIP"),
                    "host_ip": status.get("hostIP"),
                    "message": message,
                    "reason": reason,
                    "is_normal": self.is_pod_normal(status),
                }
            )
            ret_data.append(item)
        return ret_data

    def get(self, request, project_id, instance_id):
        """查询所有taskgroup信息"""
        # 获取kind
        project_kind = request.project.kind
        cluster_id, namespace, rc_names, category = self.common_handler(request, project_id, instance_id, project_kind)
        field = [
            "resourceName,createTime",
            "data.status.podIP",
            "data.status.hostIP",
            "data.status.phase",
            "data.status.containerStatuses",
            "data.status.conditions",
        ]
        if not rc_names:
            return APIResponse({"data": []})
        flag, resp = self.get_pod_or_taskgroup(
            request,
            project_id,
            cluster_id,
            field=field,
            app_name=",".join(rc_names),
            ns_name=namespace,
            category=category,
            kind=project_kind,
        )
        if not flag:
            return resp
        ret_data = self.get_pod_info(resp)
        return APIResponse({"data": ret_data})


class QueryContainersByTaskgroup(BaseTaskgroupCls):
    def get_container_info(self, data):
        """"""
        if not data:
            return []
        ret_data = []
        data = data[0].get("data") or {}
        container_info = data.get("containerStatuses") or []
        for info in container_info:
            ret_data.append(
                {
                    "container_id": info.get("containerID"),
                    "image": image_handler(info.get("image", "")),
                    "message": info.get("message"),
                    "name": info.get("name"),
                    "status": info.get("status"),
                    "reason": info.get("reason") or "",
                }
            )
        return ret_data

    def get_k8s_container_info(self, data):
        """获取k8s下容器信息"""
        ret_data = []
        if not data:
            return ret_data
        data = data[0].get("data") or {}
        status = data.get("status") or {}
        container_info = status.get("containerStatuses") or []
        for info in container_info:
            item = {
                "name": info.get("name"),
                "container_id": (info.get("containerID") or "").split("//")[-1],
                "image": image_handler(info.get("image", "")),
            }
            state = info.get("state") or {}
            for key, val in state.items():
                item["status"] = key
                item["messsage"] = val.get("message") or key
                item["reason"] = val.get("reason") or key
            ret_data.append(item)
        return ret_data

    def get(self, request, project_id, instance_id, taskgroup_name):
        """获取某一个taskgroup下的容器信息"""
        # 获取kind
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        cluster_id, namespace, rc_names, category = self.common_handler(request, project_id, instance_id, project_kind)
        field = ["data.status.containerStatuses", "data.status.phase"]

        flag, resp = self.get_pod_or_taskgroup(
            request,
            project_id,
            cluster_id,
            field=field,
            taskgroup_name=taskgroup_name,
            app_name=",".join(rc_names),
            ns_name=namespace,
            kind=project_kind,
            category=category,
        )
        if not flag:
            return resp
        ret_data = self.get_k8s_container_info(resp)

        return APIResponse({"data": ret_data})


class QueryTaskgroupInfo(BaseTaskgroupCls):
    def get_category_rollupdate_info(self, request, cluster_id, ns, inst_name, project_id, kind, category):
        """通过名称查询滚动升级信息"""
        field = ["data.spec.strategy"]
        flag, resp = self.get_application_deploy_info(
            request, project_id, cluster_id, inst_name, category=category, project_kind=kind, namespace=ns, field=field
        )
        if not flag:
            raise error_codes.APIError.f(resp.data.get("message"))
        data = resp.get("data") or []
        ret_data = {}
        if not data:
            return ret_data
        data = data[0].get("data") or {}
        spec = data.get("spec") or {}
        strategy = spec.get("strategy") or {}
        type = strategy.get("type")
        detail = strategy.get("rollingUpdate") or {}
        ret_data = {
            "type": type,
        }
        ret_data.update(detail)
        return ret_data

    def get_pod_info(self, data):
        """组装pod信息"""
        ret_data = {}
        if not data:
            return ret_data
        data = data[0]
        other = data.get("data") or {}
        status_info = other.get("status") or {}
        ret_data["base_info"] = {
            "last_status": "",
            "last_update_time": data.get("updateTime"),
            "message": "",
            "rc_name": data.get("resourceName"),
            "start_time": data.get("createTime"),
            "current_time": datetime.now(),
            "namespace": data.get("namespace"),
            "pod_ip": status_info.get("podIP"),
            "status": status_info.get("phase"),
        }
        spec_info = other.get("spec") or {}
        ret_data["kill_policy"] = spec_info.get("terminationGracePeriodSeconds") or ""
        ret_data["restart_policy"] = spec_info.get("restartPolicy") or ""
        return ret_data

    def get(self, request, project_id, instance_id, taskgroup_name):
        """获取某一个taskgroup详细信息"""
        # 获取kind
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        cluster_id, namespace, rc_names, category = self.common_handler(request, project_id, instance_id, project_kind)
        rolling_strategy = {}
        if self._from_template(instance_id):
            # 获取instnace info
            inst_info = self.get_instance_info(instance_id)
            curr_inst = inst_info[0]
            category = curr_inst.category
        field = [
            "updateTime",
            "createTime",
            "namespace",
            "resourceName",
            "data.status.phase",
            "data.status.podIP",
            "data.spec.terminationGracePeriodSeconds",
            "data.spec.restartPolicy",
        ]

        flag, resp = self.get_pod_or_taskgroup(
            request,
            project_id,
            cluster_id,
            field=field,
            taskgroup_name=taskgroup_name,
            app_name=",".join(rc_names),
            ns_name=namespace,
            kind=project_kind,
            category=category,
        )
        if not flag:
            return resp

        ret_data = self.get_pod_info(resp)
        rolling_strategy = self.get_category_rollupdate_info(
            request, cluster_id, namespace, rc_names[0], project_id, project_kind, category
        )
        ret_data["update_strategy"] = rolling_strategy

        return APIResponse({"data": ret_data})


class QueryApplicationContainers(BaseTaskgroupCls):
    def get_container_info(self, data):
        """组装container ID信息"""
        ret_data = []
        for info in data:
            item = info.get("data") or {}
            container_status = item.get("containerStatuses") or []
            item_id_list = [
                {"container_id": info["containerID"], "container_name": info["name"]}
                for info in container_status
                if info.get("containerID")
            ]
            ret_data.extend(item_id_list)
        return ret_data

    def get_k8s_container_info(self, data):
        """组装k8s id信息"""
        ret_data = []
        for info in data:
            item = info.get("data") or {}
            status = item.get("status") or {}
            container_status = status.get("containerStatuses") or []
            item_id_list = [
                {"container_id": info["containerID"].split("//")[-1], "container_name": info["name"]}
                for info in container_status
                if info.get("containerID")
            ]
            ret_data.extend(item_id_list)
        return ret_data

    def get(self, request, project_id, instance_id):
        """查询应用下所有的容器信息"""
        project_kind = request.project.kind
        cluster_id, namespace, rc_names, category = self.common_handler(request, project_id, instance_id, project_kind)
        field = ["data.status.containerStatuses.containerID", "data.status.containerStatuses.name"]

        flag, resp = self.get_pod_or_taskgroup(
            request,
            project_id,
            cluster_id,
            field=field,
            app_name=",".join(rc_names),
            ns_name=namespace,
            kind=project_kind,
            category=category,
        )
        if not flag:
            return resp
        ret_data = self.get_k8s_container_info(resp)

        return APIResponse({"data": ret_data})


class GetInstanceLabels(InstanceAPI):
    def get_instance_labels(self, info):
        """获取instance labels"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError.f(u"Instance config解析异常")

        return conf.get("metadata", {}).get("labels", {})

    def get(self, request, project_id, instance_id, instance_name):
        """获取instance 信息"""
        if not self._from_template(instance_id):
            cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
            project_kind = request.project.kind
            # 请求storage获取时间
            field = ["data.metadata.labels"]
            ok, resp = self.get_application_deploy_info(
                request,
                project_id,
                cluster_id,
                name,
                category=category,
                project_kind=project_kind,
                namespace=namespace,
                field=field,
            )
            if not ok:
                return resp
            resp_data = resp.get("data") or []
            if resp_data:
                resp_data = resp_data[0]
            else:
                resp_data = {}
            data = resp_data.get("data", {}).get("metadata", {}).get("labels") or {}
        else:
            # 获取instance info
            inst_info = self.get_instance_info(instance_id)

            # 获取labels
            data = self.get_instance_labels(inst_info[0])
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, data.get("io.tencent.paas.templateid"), inst_info[0].namespace
            )
        ret_data = [{"key": key, "val": val} for key, val in data.items()]
        return APIResponse({"data": ret_data})


class GetInstanceAnnotations(InstanceAPI):
    def get_instance_annotations(self, info):
        """获取instance annotations"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError(_("Instance config解析异常"))
        metadata = conf.get("metadata") or {}
        return metadata.get("annotations") or {}, metadata.get("labels") or {}

    def get(self, request, project_id, instance_id, instance_name):
        """获取注解信息"""
        if not self._from_template(instance_id):
            cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
            project_kind = request.project.kind
            # 请求storage获取时间
            field = ["data.metadata.annotations"]
            ok, resp = self.get_application_deploy_info(
                request,
                project_id,
                cluster_id,
                name,
                category=category,
                project_kind=project_kind,
                namespace=namespace,
                field=field,
            )
            if not ok:
                return resp
            resp_data = resp.get("data") or []
            if resp_data:
                resp_data = resp_data[0]
            else:
                resp_data = {}
            data = resp_data.get("data", {}).get("metadata", {}).get("annotations") or {}
        else:
            # 获取instance info
            inst_info = self.get_instance_info(instance_id)

            # 获取labels
            data, labels = self.get_instance_annotations(inst_info[0])
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, labels.get("io.tencent.paas.templateid"), inst_info[0].namespace
            )
        ret_data = [{"key": key, "val": val} for key, val in data.items()]
        return APIResponse({"data": ret_data})


class GetInstanceStatus(BaseAPI):
    def get(self, request, project_id, instance_id, instance_name):
        """获取实例状态"""
        # 类别不能为空
        category = request.GET.get("category")
        if not category:
            return APIResponse(
                {
                    "code": 400,
                    "message": _("参数[category]不能为空"),
                }
            )
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
        name = metadata.get("name")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
        )
        flag, resp = self.get_application_deploy_status(
            request,
            project_id,
            cluster_id,
            name,
            category=category,
            project_kind=project_kind,
            namespace=namespace,
        )
        if not flag:
            return resp
        data = resp.get("data")
        if not data:
            ret_data = {"status": "none"}
        else:
            ret_data = {"status": resp["data"][0].get("data", {}).get("status")}
        return APIResponse({"data": ret_data})


class GetInstanceInfo(InstanceAPI):
    def get_cluster_info(self, request, project_id, cluster_id):
        resp = paas_cc.get_cluster(request.user.token.access_token, project_id, cluster_id)
        if resp.get("code") != ErrorCode.NoError:
            return cluster_id
        return resp.get("data", {}).get("name")

    def get_instance_conf(self, info):
        """获取instance conf"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError(_("Instance config解析异常"))
        return conf

    def get_instance_name(self, conf):
        """通过conf获取name"""
        return conf.get("metadata", {}).get("name")

    def get_template_id(self, conf):
        """通过conf获取template id"""
        return conf.get("metadata", {}).get("labels", {}).get("io.tencent.paas.templateid")

    def get(self, request, project_id, instance_id):
        """获取instance信息"""
        # 获取instance info
        if not self._from_template(instance_id):
            cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
            # 获取kind
            project_kind = self.project_kind(request)
            # 请求storage获取时间
            field = ["updateTime", "createTime"]
            ok, resp = self.get_application_deploy_info(
                request,
                project_id,
                cluster_id,
                name,
                category=category,
                project_kind=project_kind,
                namespace=namespace,
                field=field,
            )
            if not ok:
                return resp
            resp_data = resp.get("data") or []
            if resp_data:
                resp_data = resp_data[0]
            else:
                resp_data = {}
            create_time = resp_data.get("createTime", "")
            update_time = resp_data.get("updateTime", "")
            template_id = ""
        else:
            inst_info = self.get_instance_info(instance_id)
            instance_conf = self.get_instance_conf(inst_info[0])
            # 获取namespace
            curr_inst = inst_info[0]

            conf = self.get_common_instance_conf(curr_inst)
            metadata = conf.get("metadata", {})
            labels = metadata.get("labels", {})
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            namespace = metadata.get("namespace")
            # 添加权限
            self.validate_view_perms(
                request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
            )
            name = self.get_instance_name(instance_conf)
            create_time = curr_inst.created
            update_time = curr_inst.updated
            template_id = self.get_template_id(instance_conf)

        all_cluster_info = self.get_cluster_id_env(request, project_id)
        cluster_name = all_cluster_info.get(cluster_id).get("cluster_name") or cluster_id
        return APIResponse(
            {
                "data": {
                    "name": name,
                    "create_time": create_time,
                    "update_time": update_time,
                    "namespace_name": namespace,
                    "template_id": template_id,
                    "cluster_id": cluster_id,
                    "cluster_name": cluster_name,
                }
            }
        )


class ReschedulerTaskgroup(InstanceAPI):
    """
    针对k8s确认重新调度是删除pod
    """

    def get_instance_conf(self, info):
        """获取instance conf"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError(_("Instance config解析异常"))
        return conf

    def put(self, request, project_id, instance_id):
        data = dict(request.data)
        taskgroup = data.get("taskgroup")
        if not taskgroup:
            return APIResponse({"code": 400, "message": _("参数【taskgroup】不能为空")})
        # 获取kind
        project_kind = self.project_kind(request)
        if not self._from_template(instance_id):
            cluster_id, namespace_name, instance_name, category = self.get_instance_resource(request, project_id)
            # 增加命名空间域的权限校验
            perm_ctx = NamespaceScopedPermCtx(
                username=request.user.username,
                project_id=project_id,
                cluster_id=cluster_id,
                name=namespace_name,
            )
            NamespaceScopedPermission().can_use(perm_ctx)
        else:
            # 获取instance info
            inst_info = self.get_instance_info(instance_id)

            instance_conf = self.get_instance_conf(inst_info[0])
            # 获取instance_name
            metadata = instance_conf.get("metadata", {})
            instance_name = metadata.get("name")
            namespace_name = metadata.get("namespace")
            labels = metadata.get("labels") or {}
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, labels.get("io.tencent.paas.templateid"), inst_info[0].namespace
            )
        resp = self.rescheduler_taskgroup(
            request, project_id, cluster_id, namespace_name, instance_name, taskgroup, kind=project_kind
        )
        return resp


class TaskgroupEvents(InstanceAPI):
    def get_instance_conf(self, info):
        """获取instance conf"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError(_("Instance config解析异常"))
        return conf

    def get_rc_name_by_deployment(
        self, request, project_id, cluster_id, instance_name, project_kind=ProjectKindID, namespace=None
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

    def get_backend_record(self, cluster_id, instance_namespace, instance_id, instance_name, kind):
        """获取backend记录的create error信息"""
        all_events = InstanceEvent.objects.filter(
            instance_config_id=instance_id,
            category__in=[
                K8SJOB_CATEGORY,
                K8SDAEMONSET_CATEGORY,
                K8SSTATEFULSET_CATEGORY,
                K8SDEPLOYMENT_CATEGORY,
            ],
            is_deleted=False,
        ).order_by("-updated")
        return [
            {
                "eventTime": info.created,
                "component": "Scheduler",
                "describe": info.msg,
                "clusterId": cluster_id,
                "extraInfo": {"kind": "", "name": instance_name, "namespace": instance_namespace},
                "type": "CreateError",
            }
            for info in all_events
        ]

    def get_k8s_pod_events(self, data):
        """拼装k8s pod事件"""
        pass

    def get(self, request, project_id, instance_id):
        data = dict(request.GET.items())
        offset = data.get("offset") or 0
        limit = data.get("limit") or 10
        if not (str(offset).isdigit() and str(limit).isdigit()):
            return APIResponse({"code": 400, "message": _("参数[offset]和[limit]必须为整数!")})
        offset = int(offset)
        limit = int(limit)
        if not self._from_template(instance_id):
            cluster_id, inst_namespace, inst_name, category = self.get_instance_resource(request, project_id)
        else:
            # 通过instance id获取instance信息
            inst_info = self.get_instance_info(instance_id)
            curr_inst = inst_info[0]
            # 添加权限
            # perm = bcs_perm.Namespace(request, project_id, curr_inst.namespace)
            # perm.can_view(raise_exception=True)

            instance_conf = self.get_instance_conf(curr_inst)
            metadata = instance_conf.get("metadata") or {}
            labels = metadata.get("labels", {})
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            inst_namespace = metadata.get("namespace")
            inst_name = metadata.get("name")
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
            )
            category = curr_inst.category
        # 获取kind
        project_kind = request.project.kind
        application_info = [inst_name]

        params = {"offset": offset, "length": limit, "clusterId": cluster_id}
        field = ["resourceName"]
        flag, pod_resp = self.get_pod_or_taskgroup(
            request,
            project_id,
            cluster_id,
            kind=project_kind,
            category=category,
            ns_name=inst_namespace,
            app_name=application_info,
            field=field,
        )
        if not flag:
            return pod_resp
        pod_name = ",".join([info["resourceName"] for info in pod_resp if info.get("resourceName")])
        params.update({"env": "k8s", "kind": "Pod", "extraInfo.name": pod_name, "extraInfo.namespace": inst_namespace})
        # 添加创建失败的instance消息后，查询返回
        return self.query_events(request, project_id, cluster_id, params)


class ContainerInfo(InstanceAPI):
    def get_instance_conf(self, info):
        """获取instance conf"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError(_("Instance config解析异常"))
        return conf

    def get_rc_name_by_deployment(
        self, request, project_id, cluster_id, instance_name, project_kind=ProjectKindID, namespace=None
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

    def data_process(self, data, container_id):
        """根据container id处理数据"""
        ret_data = {}
        if not data:
            return ret_data
        data = data[0].get("data") or {}
        container_status = data.get("containerStatuses") or []
        for info in container_status:
            if info.get("containerID") == container_id:
                image = info.get("image", "")
                image_split_str = utils.image_handler(image)
                ret_data = {
                    "volumes": info.get("volumes", []),
                    "ports": info.get("containerPort", []),
                    "commands": {
                        "command": info.get("command", ""),
                        "args": ' '.join(info.get("args", "")),
                    },
                    "network_mode": info.get("networkMode", ""),
                    "labels": [{"key": key, "val": val} for key, val in info.get("labels", {}).items()],
                    "resources": info.get("resources", {}),
                    "health_check": info.get("healCheckStatus", []),
                    "env_args": [{"name": key, "value": val} for key, val in info.get("env", {}).items()],
                    "container_id": info.get("containerID", ""),
                    "host_ip": data.get("hostIP", ""),
                    "image": image_split_str,
                    "container_ip": data.get("pod_ip", ""),
                    "host_name": data.get("hostName", ""),
                    "container_name": info.get("name", ""),
                }
        return ret_data

    def get(self, request, project_id, instance_id, taskgroup_name, container_id):
        """获取容器信息"""
        # 获取kind
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        if not self._from_template(instance_id):
            cluster_id, namespace, instance_name, category = self.get_instance_resource(request, project_id)
        else:
            # 获取instance info
            inst_info = self.get_instance_info(instance_id)

            # 获取信息
            instance_conf = self.get_instance_conf(inst_info[0])
            metadata = instance_conf.get("metadata", {})
            instance_name = metadata.get("name")
            namespace = metadata.get("namespace")
            labels = metadata.get("labels") or {}
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, labels.get("io.tencent.paas.templateid"), inst_info[0].namespace
            )
            # name = metadata.get("name")
            namespace = metadata.get("namespace")
            category = inst_info[0].category
        flag, taskgroup = self.get_pod_or_taskgroup(
            request,
            project_id,
            cluster_id,
            field="data",
            app_name=instance_name,
            taskgroup_name=taskgroup_name,
            ns_name=namespace,
            category=category,
        )
        if not flag:
            return taskgroup
        # 处理数据
        data = self.data_process(taskgroup, container_id)
        return APIResponse({"data": data})


class K8sContainerInfo(InstanceAPI):
    def get_instance_conf(self, info):
        """获取instance conf"""
        try:
            conf = json.loads(info.config)
        except Exception as error:
            logger.error(u"解析instance config异常，id为 %s, 详情: %s" % (info.id, error))
            raise error_codes.JSONParseError(_("Instance config解析异常"))
        return conf

    def match_container_info(self, container_data):
        """匹配容器信息"""
        return {info.get("name"): info for info in container_data}

    def data_process(self, data, container_id):
        """根据container id处理数据"""
        ret_data = {}
        if not data:
            return ret_data
        data = data[0].get("data") or {}
        status = data.get("status") or {}
        spec = data.get("spec") or {}
        container_status = status.get("containerStatuses") or []
        metadata = data.get("metadata") or {}
        # 匹配容器
        container_data = self.match_container_info(spec.get("containers") or [])
        for info in container_status:
            if not info.get("containerID"):
                continue
            if info.get("containerID").split("//")[-1] == container_id:
                container_info = container_data.get(info.get("name")) or {}
                image = info.get("image", "")
                image_split_str = utils.image_handler(image)
                ret_data = {
                    "volumes": container_info.get("volumeMounts"),
                    "ports": container_info.get("ports", []),
                    "commands": {
                        "command": container_info.get("command", ""),
                        "args": ' '.join(container_info.get("args", "")),
                    },
                    "network_mode": spec.get("dnsPolicy", ""),
                    "labels": [{"key": key, "val": val} for key, val in metadata.get("labels", {}).items()],
                    "resources": container_info.get("resources", {}),
                    "health_check": container_info.get("livenessProbe", {}),
                    "readiness_check": container_info.get("readinessProbe", {}),
                    "env_args": container_info.get("env", []),
                    "container_id": info.get("containerID", "").split("//")[-1],
                    "host_ip": status.get("hostIP", ""),
                    "image": image_split_str,
                    "container_ip": status.get("podIP", ""),
                    "host_name": spec.get("nodeName", ""),
                    "container_name": info.get("name", ""),
                }
                break
        return ret_data

    def post(self, request, project_id, instance_id, taskgroup_name):
        """获取容器信息"""
        container_id = request.data.get("container_id")
        if not container_id:
            raise error_codes.CheckFailed(_("容器ID不能为空"))
        # 获取kind
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        if not self._from_template(instance_id):
            cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        else:
            # 获取instance info
            inst_info = self.get_instance_info(instance_id)

            # 获取信息
            instance_conf = self.get_instance_conf(inst_info[0])
            metadata = instance_conf.get("metadata", {})
            name = metadata.get("name")
            namespace = metadata.get("namespace")
            labels = metadata.get("labels") or {}
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, labels.get("io.tencent.paas.templateid"), inst_info[0].namespace
            )
            curr_inst = inst_info[0]
            category = curr_inst.category
        # 获取container信息
        flag, taskgroup = self.get_pod_or_taskgroup(
            request,
            project_id,
            cluster_id,
            field=["data"],
            app_name=name,
            taskgroup_name=taskgroup_name,
            ns_name=namespace,
            category=category,
            kind=project_kind,
        )
        if not flag:
            return taskgroup
        # 处理数据
        data = self.data_process(taskgroup, container_id)
        return APIResponse({"data": data})


class ContainerLogs(BaseAPI):
    renderer_classes = (BKAPIRenderer, BrowsableAPIRenderer)

    def clean_log(self, log_data):
        all_data = log_data.get("data") or {}
        hits_data = all_data.get("list", {}).get("hits", {}).get("hits", [])
        return hits_data

    def get(self, request, project_id, container_id):
        """获取日志
        注意: 获取最新的100条日志
        """
        std_log_index = datalog_utils.get_std_log_index(project_id)
        standard_log = []
        username = request.user.username
        if std_log_index:
            standard_log = data.get_container_logs(username, container_id=container_id, index=std_log_index)
            standard_log = self.clean_log(standard_log)

        # 没有数据则再查询一下默认的index，兼容使用默认dataid实例化的容器
        if not standard_log:
            standard_log = data.get_container_logs(username, container_id=container_id)
            standard_log = self.clean_log(standard_log)

        log_content = []
        for info in standard_log:
            source = info.get('_source') or {}
            log = source.get('log') or ''
            timestamp = (
                int(source.get('dtEventTimeStamp')) / 1000 if str(source.get('dtEventTimeStamp')).isdigit() else None
            )
            localtime = time.strftime(settings.REST_FRAMEWORK['DATETIME_FORMAT'], time.localtime(timestamp))
            log_content.append({'log': log, 'localtime': localtime})

        return Response(log_content)


class InstanceConfigInfo(InstanceAPI):
    """获取应用实例配置文件信息"""

    def get_online_app_conf(self, request, project_id, project_kind):
        cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        # get the online yaml
        online_app_conf = self.online_app_conf(
            request, project_id, project_kind, cluster_id, name, namespace, category
        )
        return online_app_conf

    def get(self, request, project_id, instance_id):
        # 获取项目类型
        project_kind = self.project_kind(request)
        if not self._from_template(instance_id):
            return APIResponse({"data": self.get_online_app_conf(request, project_id, project_kind)})
        try:
            instance_info = self.get_instance_info(instance_id)
            conf = self.get_common_instance_conf(instance_info[0])
            labels = (conf.get("metadata") or {}).get("labels") or {}
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, labels.get("io.tencent.paas.templateid"), instance_info[0].namespace
            )
        except Exception as error:
            return APIResponse({"code": 400, "data": {}, "message": "%s" % error.message})
        return APIResponse({"data": conf})


class GetVersionList(BaseAPI):
    """
    关于滚动升级，应用名称和命名空间一致就允许升级
    """

    def get_tmpl_info(self, name, category):
        """通过名称获取模板信息"""
        tmpl_id_list = MODULE_DICT[category].objects.filter(name=name).values("id")
        return [info["id"] for info in tmpl_id_list]

    def get_version_info(self, muster_id, tmpl_id_list, category):
        """获取版本信息"""
        all_show_vers = ShowVersion.objects.filter(template_id=muster_id, is_deleted=False).order_by("-updated")
        # 根据ID进行过滤
        ret_data = {"results": [], "count": 0}
        for show_ver in all_show_vers:
            real_version_id = show_ver.real_version_id
            info = VersionedEntity.objects.get(id=real_version_id)
            conf = info.get_entity()
            if not conf:
                continue
            category_id_info = conf.get(category) or None
            if not category_id_info:
                continue
            category_id_list = [int(_info) for _info in category_id_info.split(",")]
            if set(tmpl_id_list) & set(category_id_list):
                ret_data["results"].append(
                    {
                        "id": info.id,
                        "show_version_id": show_ver.id,
                        "version": show_ver.name,
                        "muster_id": info.template_id,
                    }
                )
                ret_data["count"] += 1
        return ret_data

    def get(self, request, project_id, instance_id):
        """获取针对当前实例的版本列表
        TODO: 现阶段也返回当前版本
        """
        instance_info = self.get_instance_info(instance_id)
        curr_inst = instance_info[0]

        conf = self.get_common_instance_conf(curr_inst)
        # 获取当前实例使用的版本
        category = curr_inst.category
        name = curr_inst.name
        # 获取模板ID
        tmpl_id_list = self.get_tmpl_info(name, category)
        metadata = conf.get("metadata") or {}
        labels = metadata.get("labels")
        muster_id = labels.get("io.tencent.paas.templateid")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), curr_inst.namespace
        )

        # 根据模板集ID及版本过滤相应分类的信息
        version_data = self.get_version_info(muster_id, tmpl_id_list, category)
        return APIResponse({"data": version_data})


class QueryContainerInfo(BaseAPI):
    def match_container_info(self, container_data):
        """匹配容器信息"""
        return {info.get("name"): info for info in container_data}

    def k8s_data_process(self, data, container_id):
        """根据container id处理数据"""
        ret_data = {}
        if not data:
            return ret_data
        for info in data:
            data_item = info.get("data") or {}
            status = data_item.get("status") or {}
            spec = data_item.get("spec") or {}
            container_status = status.get("containerStatuses") or []
            metadata = data_item.get("metadata") or {}
            # 匹配容器
            container_data = self.match_container_info(spec.get("containers") or [])
            for info in container_status:
                if not info.get("containerID"):
                    continue
                if info.get("containerID").split("//")[-1] == container_id:
                    container_info = container_data.get(info.get("name")) or {}
                    image = info.get("image", "")
                    image_split_str = utils.image_handler(image)
                    ret_data = {
                        "volumes": container_info.get("volumeMounts"),
                        "ports": container_info.get("ports", []),
                        "commands": {
                            "command": container_info.get("command", ""),
                            "args": ' '.join(container_info.get("args", "")),
                        },
                        "network_mode": spec.get("dnsPolicy", ""),
                        "labels": [{"key": key, "val": val} for key, val in metadata.get("labels", {}).items()],
                        "resources": container_info.get("resources", {}),
                        "health_check": container_info.get("healCheckStatus", []),
                        "env_args": container_info.get("env", {}),
                        "container_id": info.get("containerID", "").split("//")[-1],
                        "host_ip": status.get("hostIP", ""),
                        "image": image_split_str,
                        "container_ip": status.get("podIP", ""),
                        "host_name": spec.get("nodeName", ""),
                        "namespace": metadata.get("namespace", ""),
                        "container_name": info.get("name", ""),
                    }
                    break
        return ret_data

    def data_process(self, data, container_id):
        """根据container id处理数据"""
        ret_data = {}
        if not data:
            return ret_data
        for container_info in data:
            curr_data = container_info.get("data") or {}
            container_status = curr_data.get("containerStatuses") or []
            for info in container_status:
                if not info.get("containerID"):
                    continue
                if info.get("containerID") == container_id:
                    image = info.get("image", "")
                    image_split_str = utils.image_handler(image)
                    ret_data = {
                        "volumes": info.get("volumes", []),
                        "ports": info.get("containerPort", []),
                        "commands": {
                            "command": info.get("command", ""),
                            "args": ' '.join(info.get("args", "")),
                        },
                        "network_mode": info.get("networkMode", ""),
                        "labels": [{"key": key, "val": val} for key, val in info.get("labels", {}).items()],
                        "resources": info.get("resources", {}),
                        "health_check": info.get("healCheckStatus", []),
                        "env_args": info.get("env", {}),
                        "container_id": info.get("containerID", ""),
                        "host_ip": curr_data.get("hostIP", ""),
                        "image": image_split_str,
                        "container_ip": curr_data.get("pod_ip", ""),
                        "host_name": curr_data.get("hostName", ""),
                        "container_name": info.get("name", ""),
                    }
                    break
        return ret_data

    def get(self, request, project_id, cluster_id):
        """通过container id查询容器详情"""
        container_id = request.GET.get("container_id")
        # 获取kind
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        field = ["data"]
        # 通过集群查询下面所有的container信息
        flag, resp = self.get_pod_or_taskgroup(request, project_id, cluster_id, field=field, kind=project_kind)
        if not flag:
            return resp
        data = self.k8s_data_process(resp, container_id)
        return APIResponse({"data": data})

    def post(self, request, project_id, cluster_id):
        """针对k8s container id的处理
        由于container id的格式是这样docker://6c3c516bb80ba564f187a8f7c0ce011dd69d68b6f373a33a642aae5a8a0ebc39
        获取时有问题
        """
        container_id = request.data.get("container_id")
        # 获取kind
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        field = ["data"]
        # 通过集群查询下面所有的container信息
        flag, resp = self.get_pod_or_taskgroup(request, project_id, cluster_id, field=field, kind=project_kind)
        if not flag:
            return resp
        data = self.k8s_data_process(resp, container_id)
        return APIResponse({"data": data})


class BatchInstances(BaseAPI):
    def get_enforce(self, request):
        # 是否强制删除
        enforce = request.GET.get("enforce")
        if enforce not in [None, "0"]:
            enforce = 1
        else:
            enforce = 0
        return enforce

    def save_inst_event(self, data):
        """保存实例操作的事件"""
        try:
            InstanceEvent.objects.bulk_create([InstanceEvent(**info) for info in data])
        except Exception as error:
            logger.error(u"记录实例操作事件失败，详情: %s" % error)

    def del_oper(self, request, project_id, need_delete_info, kind, enforce):
        """通过bcs调用删除"""
        error_message = []
        error_resp = []
        deleted_id_list = []
        for info in need_delete_info:
            curr_info = info.pop("info")
            # 针对0/0的情况先查询一次
            if not self.get_category_info(
                request, project_id, info["cluster_id"], kind, info["inst_name"], info["namespace"], info["category"]
            ):
                deleted_id_list.append(info["inst_id"])
                continue
            with client.ContextActivityLogClient(
                project_id=project_id,
                user=request.user.username,
                resource_type="instance",
                resource=info.get("inst_name"),
                resource_id=info.get("inst_id"),
                extra=json.dumps(info),
                description=_("应用删除操作"),
            ).log_delete():
                resp = self.delete_instance(
                    request,
                    project_id,
                    info["cluster_id"],
                    info["namespace"],
                    info["inst_name"],
                    category=info["category"],
                    kind=kind,
                    inst_id_list=[info["inst_id"]],
                    enforce=enforce,
                )
                if resp.data.get("code") == ErrorCode.NoError:
                    self.update_instance_record_status(curr_info, app_constants.DELETE_INSTANCE, status="Deleting")
                else:
                    error_resp.append(
                        {
                            "instance_id": info["inst_version_id"],
                            "instance_config_id": info["inst_id"],
                            "resp_snapshot": json.dumps(resp.data),
                            "msg": resp.data.get("message"),
                            "category": info["category"],
                        }
                    )
                    error_message.append("%s: %s" % (info.get("inst_name"), resp.data.get("message")))
        self.save_inst_event(error_resp)
        # 如果存在0/0的情况
        if deleted_id_list:
            InstanceConfig.objects.filter(id__in=deleted_id_list).update(is_deleted=True, deleted_time=datetime.now())
        # 如果有错误，则返回异常信息
        if error_message:
            return APIResponse({"code": ErrorCode.SysError, "message": ";".join(error_message)})
        else:
            return APIResponse({"message": _("删除成功!")})

    def del_for_client(self, request, project_id, project_kind, category_name_list, namespace, enforce):
        """删除"""
        cluster_id = self.get_cluster_by_ns_name(request, project_id, namespace)
        err_msg = []
        for info in category_name_list:
            resp = self.delete_instance(
                request,
                project_id,
                cluster_id,
                namespace,
                info["name"],
                category=info["category"],
                kind=project_kind,
                enforce=enforce,
            )
            if resp.get("code") != ErrorCode.NoError:
                err_msg.append(resp.get("message"))
        if not err_msg:
            raise error_codes.CheckFailed(_("部分删除失败，详情: {}").format(",".join(err_msg)))

    def del_by_name_and_ns(self, request, project_id, project_kind, enforce, data):
        """通过name+namespace+cluster_id+resource_kind"""
        err_msg = []
        for res in data:
            resp = self.delete_instance(
                request,
                project_id,
                res["cluster_id"],
                res["namespace"],
                res["name"],
                category=res["resource_kind"],
                kind=project_kind,
                enforce=enforce,
            )
            if resp.get("code") != ErrorCode.NoError:
                err_msg.append(resp.get("message"))
        if not err_msg:
            raise error_codes.APIError(_("部分删除失败，详情: {}").format(",".join(err_msg)))

    def delete(self, request, project_id):
        """批量删除实例接口"""
        req_data = request.data
        enforce = self.get_enforce(request)
        # 获取kind
        flag, project_kind = self.get_project_kind(request, project_id)
        if not flag:
            return project_kind
        # 获取分类参数
        slz = BatchDeleteResourceSLZ(data=req_data)
        slz.is_valid(raise_exception=True)
        slz_data = slz.validated_data
        # 针对非模板集表单模式创建的需要传递应用名称、命名空间和集群
        if slz_data.get("resource_list"):
            self.del_by_name_and_ns(request, project_id, project_kind, enforce, slz_data["resource_list"])
        inst_id_list = slz_data.get("inst_id_list")
        if not inst_id_list:
            return APIResponse({"message": _("任务下发成功")})
        # 剔除为0的实例
        inst_id_list = [info for info in inst_id_list if info not in ["0", 0]]
        inst_info = InstanceConfig.objects.filter(id__in=inst_id_list, is_deleted=False)
        if not inst_info:
            return APIResponse({"code": 400, "message": _("没有查询到实例信息")})
        # 判断已经创建失败的实例，则只更新状态
        need_delete_inst_ids = []
        for info in inst_info:
            conf = json.loads(info.config)
            labels = (conf.get("metadata") or {}).get("labels") or {}
            # 添加权限
            self.bcs_single_app_perm_handler(
                request, project_id, labels.get("io.tencent.paas.templateid"), info.namespace
            )

            if info.is_deleted or info.oper_type == app_constants.DELETE_INSTANCE:
                continue
            if info.category in [
                K8SDAEMONSET_CATEGORY,
                K8SJOB_CATEGORY,
                K8SDEPLOYMENT_CATEGORY,
                K8SSTATEFULSET_CATEGORY,
            ]:
                # 包含cluster_id, namespace, inst_name, category
                conf = self.get_common_instance_conf(info)
                metadata = conf.get("metadata", {})
                labels = metadata.get("labels", {})
                cluster_id = labels.get("io.tencent.bcs.clusterid")
                namespace = metadata.get("namespace")
                name = metadata.get("name")
                item = {
                    "info": info,
                    "cluster_id": cluster_id,
                    "namespace": namespace,
                    "inst_name": name,
                    "category": info.category,
                    "inst_id": info.id,
                    "inst_version_id": info.instance_id,
                }
                need_delete_inst_ids.append(item)
        if not need_delete_inst_ids:
            return APIResponse({"message": _("删除成功!")})
        return self.del_oper(request, project_id, need_delete_inst_ids, project_kind, enforce)


class UpdateVersionConfig:
    def exclude_by_key_prefix(self, data):
        if not data:
            return data
        exclude_prefix_list = ['io.tencent.paas', 'io.tencent.bcs', 'io.tencent.bkdata']
        # list() can force a copy of the keys
        for key in list(data.keys()):
            for prefix in exclude_prefix_list:
                if key.startswith(prefix):
                    data.pop(key)
                    break
        return data

    def exclude_platform_injected_keys(self, config):
        # labels and annotations from metadata
        metadata = config.get('metadata', {})
        metadata_labels = metadata.get('labels', {})
        metadata_annotations = metadata.get('annotations', {})
        config['metadata']['labels'] = self.exclude_by_key_prefix(metadata_labels)
        config['metadata']['annotations'] = self.exclude_by_key_prefix(metadata_annotations)
        # labels from spec
        spec_labels = getitems(config, ['spec', 'template', 'metadata', 'labels'], default={})
        config['spec']['template']['metadata']['labels'] = self.exclude_by_key_prefix(spec_labels)

        return config


class GetInstanceVersionConf(UpdateInstanceNew, UpdateVersionConfig, InstanceAPI):
    def get_show_version_info(self, show_version_id):
        """获取展示版本ID"""
        show_version = ShowVersion.objects.filter(id=show_version_id)
        if not show_version:
            raise error_codes.CheckFailed(_("没有查询到展示版本信息"))
        return show_version[0]

    def get_version_info(self, category, version_id):
        """获取真正的版本信息"""
        info = VersionedEntity.objects.filter(id=version_id)
        if not info:
            raise error_codes.CheckFailed(_("没有查询到展示版本信息"))
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
                    if _obj.category == 'custom':
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

    def get_default_instance_value(self, variable_list, key, pre_instance_num):
        ret_val = ""
        for info in variable_list:
            if info["key"] == key:
                return info.get("value") or pre_instance_num
        return ret_val or pre_instance_num

    def generate_ns_config_info(self, request, ns_id, inst_entity, params, is_save=False):
        """生成某一个命名空间下的配置"""
        inst_conf = inst_utils.generate_namespace_config(ns_id, inst_entity, is_save, is_validate=False, **params)
        return inst_conf

    def get_online_app_conf(self, request, project_id, project_kind):
        """针对非模板创建的应用，获取线上的配置"""
        cluster_id, namespace, name, category = self.get_instance_resource(request, project_id)
        return self.online_app_conf(request, project_id, project_kind, cluster_id, name, namespace, category)

    def get(self, request, project_id, inst_id):
        """获取当前实例的版本配置信息"""
        project_kind = self.project_kind(request)
        if not self._from_template(inst_id):
            json_conf = self.get_online_app_conf(request, project_id, project_kind)
            return APIResponse({"data": {"json": json_conf, "yaml": self.json2yaml(json_conf)}})
        inst_info = self.get_instance_info(inst_id)
        category = inst_info.category
        # inst_name = inst_info.name
        inst_version_id = inst_info.instance_id
        inst_conf = json.loads(inst_info.config)
        labels = inst_conf.get("metadata", {}).get("labels", {})
        cluster_id = labels.get("io.tencent.bcs.clusterid")
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), inst_info.namespace
        )
        inst_version_info = VersionInstance.objects.filter(id=inst_version_id)
        if not inst_version_info:
            raise error_codes.APIError(_("没有查询到实例版本"))
        show_version_id = request.GET.get("show_version_id")
        app_category_id_list = []
        if show_version_id:
            real_version_id = self.get_show_version_info(show_version_id).real_version_id
            category_id_list = self.get_version_info(category, real_version_id)
        else:
            instance_entity = json.loads(inst_version_info[0].instance_entity)
            # 获取模板
            category_id_list = instance_entity.get(category) or []

        all_info = self.get_tmpl_info(category, category_id_list, inst_info.name)

        if not all_info:
            raise error_codes.CheckFailed(_("没有查询到版本信息"))
        # 通过展示版本获取真正版本
        curr_info = all_info[0]
        version_conf = curr_info.config
        instance_num_var_flag = False
        pre_spec_info = inst_conf.get("spec") or {}
        pre_instance_num = pre_spec_info.get("replicas") or pre_spec_info.get("instance") or DEFAULT_INSTANCE_NUM
        if show_version_id:
            variables = self.get_variables(request, project_id, version_conf, cluster_id, inst_info.namespace)
            spec_info = json.loads(version_conf).get("spec") or {}
            instance_num = spec_info.get("replicas") or spec_info.get("instance") or DEFAULT_INSTANCE_NUM
        else:
            variables = []
            instance_num = pre_instance_num
            variables_dict = json.loads(inst_info.variables)
            if isinstance(variables_dict, dict):
                variables = [{"key": key, "value": val} for key, val in variables_dict.items()]
        instance_num_key = ""
        if instance_num and not str(instance_num).isdigit():
            instance_num_var_flag = True
            instance_num_key = re.findall(r"[^\{\}]+", instance_num)
            if not instance_num_key:
                instance_num_key = ""
                instance_num = DEFAULT_INSTANCE_NUM
            else:
                instance_num_key = instance_num_key[-1]
                instance_num = self.get_default_instance_value(variables, instance_num_key, pre_instance_num)
            if show_version_id:
                for info in variables:
                    if info["key"] == instance_num_key and not info["value"]:
                        info["value"] = pre_instance_num

        if show_version_id:
            show_version_info = self.get_show_version_info(show_version_id)
            ids = self.get_tmpl_ids(show_version_info.real_version_id, category)
            tmpl_entity = self.get_tmpl_entity(category, ids, inst_info.name)
            variables_dict_info = {}
            # 当key为实例数量的key时，如果当前实例数量为0或空，则需要设置为滚动升级前的实例数量
            for info in variables:
                value = info["value"]
                if info["key"] == instance_num_key:
                    value = value or pre_instance_num
                variables_dict_info[info["key"]] = value
            # 渲染最终配置
            config_profile = self.generate_conf(
                request,
                project_id,
                inst_info,
                show_version_info,
                inst_info.namespace,
                tmpl_entity,
                category,
                instance_num or 0,
                project_kind,
                variables_dict_info,
            )
        else:
            config_profile = inst_conf
        # exclude the keys from platform
        config_profile = self.exclude_platform_injected_keys(config_profile)
        yaml_conf = self.json2yaml(config_profile)
        return APIResponse(
            {
                "data": {
                    "json": config_profile,
                    "yaml": yaml_conf,
                    "variable": variables,
                    "instance_num": instance_num,
                    "instance_num_key": instance_num_key,
                    "instance_num_var_flag": instance_num_var_flag,
                }
            }
        )


class GetInstanceVersions(BaseAPI):
    def get_inst_version_info(self, inst_version_id):
        """获取实例版本信息"""
        inst_version_info = VersionInstance.objects.filter(id=inst_version_id)
        if not inst_version_info:
            raise error_codes.APIError(_("没有查询到实例版本"))
        return inst_version_info[0]

    def get_show_version_info(self, id_list):
        """获取展示的版本信息"""
        show_version_info = ShowVersion.objects.filter(real_version_id__in=id_list, is_deleted=False).order_by(
            "-updated"
        )
        if not show_version_info:
            raise error_codes.APIError(_("没有查询到实例版本"))
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

    def get(self, request, project_id, inst_id):
        """获取实例对应的版本"""
        inst_info = self.get_instance_info(inst_id)[0]
        conf = json.loads(inst_info.config)
        labels = (conf.get("metadata") or {}).get("labels") or {}
        # 添加权限
        self.bcs_single_app_perm_handler(
            request, project_id, labels.get("io.tencent.paas.templateid"), inst_info.namespace
        )

        category = inst_info.category
        inst_name = inst_info.name
        inst_version_id = inst_info.instance_id
        inst_version_info = self.get_inst_version_info(inst_version_id)
        muster_id = inst_version_info.template_id
        category_tmpl_id_list = self.get_category_tmpl(category, inst_name)
        version_id_list = self.get_version_entity(muster_id, category, category_tmpl_id_list)
        # 查询显示的版本
        ret_data = self.get_show_version_info(version_id_list)

        return APIResponse({"data": ret_data})
