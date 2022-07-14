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

from backend.bcs_web.audit_log.audit.context import AuditContext
from backend.bcs_web.audit_log.constants import ActivityStatus, ActivityType
from backend.container_service.clusters.base.models import CtxCluster
from backend.resources.namespace.utils import get_namespace_by_id
from backend.resources.workloads.deployment import Deployment
from backend.templatesets.legacy_apps.configuration.auditor import TemplatesetAuditor
from backend.templatesets.legacy_apps.configuration.models import get_model_class_by_resource_name
from backend.templatesets.legacy_apps.instance.auditor import InstanceAuditor
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.drivers import get_scheduler_driver
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.templatesets.legacy_apps.instance.utils import generate_namespace_config, save_all_config
from backend.uniapps.application.constants import ROLLING_UPDATE_INSTANCE


def instantiate_resources(access_token, username, release_data, project_kind):
    project_id = release_data["project_id"]
    is_start = release_data["is_start"]

    configuration = save_all_config(release_data, access_token, username)
    # 调用 bcs API
    if is_start:
        driver = get_scheduler_driver(access_token, project_id, configuration, project_kind)
        instantiation_result = driver.instantiation()
        return instantiation_result


def generate_manifest(access_token, username, release_data):
    release_id = release_data["release_id"]
    namespace_id = release_data["namespace_id"]
    template_id = release_data["template_id"]

    v_inst = VersionInstance.objects.get(id=release_id, ns_id=namespace_id, template_id=template_id)
    params = {
        "instance_id": release_id,
        "version_id": v_inst.version_id,
        "show_version_id": v_inst.show_version_id,
        "template_id": template_id,
        "project_id": release_data["project_id"],
        "access_token": access_token,
        "username": username,
        "lb_info": {},
        "variable_dict": release_data["variable_info"].get(namespace_id) or {},
        "is_preview": False,
    }

    resource_name = release_data["resource_name"]
    instance_entity = json.loads(v_inst.instance_entity)
    resource_id_list = instance_entity.get(resource_name)
    res_model_cls = get_model_class_by_resource_name(resource_name)
    res_qsets = res_model_cls.objects.filter(name=release_data["name"], id__in=resource_id_list)

    configuration = generate_namespace_config(
        namespace_id, {resource_name: [res_qsets[0].id]}, is_save=False, **params
    )
    return configuration[resource_name][0]["config"]


def _update_resources(access_token, release_data, namespace_info, manifest):
    ctx_cluster = CtxCluster.create(
        token=access_token, id=namespace_info['cluster_id'], project_id=release_data['project_id']
    )
    return (
        Deployment(ctx_cluster)
        .replace(body=manifest, name=release_data['name'], namespace=namespace_info['name'])
        .data.to_dict()
    )


def update_resources(access_token, username, release_data, namespace_info):
    manifest = generate_manifest(access_token, username, release_data)
    return _update_resources(access_token, release_data, namespace_info, manifest)


class DeployController:
    def __init__(self, user, project_kind):
        self.access_token = user.token.access_token
        self.username = user.username
        self.project_kind = project_kind

    def create_release(self, release_data):
        release_result = instantiate_resources(
            self.access_token, self.username, release_data, project_kind=self.project_kind
        )

        template_name = release_data["template_name"]

        log_params = {
            "project_id": release_data["project_id"],
            "user": self.username,
            "resource": template_name,
            "resource_id": release_data["template_id"],
            "extra": {
                "variable_info": release_data["variable_info"],
                "instance_entity": release_data["instance_entity"],
                "ns_list": release_data["ns_list"],
            },
            'activity_type': ActivityType.Instantiate
        }

        # only one namespace
        if release_result["success"]:
            ret = release_result["success"][0]
            release_id = ret["instance_id"]
            log_params["description"] = "实例化模板集[{}]到命名空间[{}]".format(template_name, ret["ns_name"])
            log_params['activity_status'] = ActivityStatus.Succeed
            TemplatesetAuditor(AuditContext(**log_params)).log_raw()
            return release_id

        ret = release_result["failed"][0]
        release_id = ret["instance_id"]
        if ret["res_type"]:
            description = "实例化模板集[{template_name}]到命名空间[{namespace}]时，实例化{res_name}失败，" "错误消息：{err_msg}".format(
                template_name=template_name, namespace=ret["ns_name"], res_name=ret["res_type"], err_msg=ret["err_msg"]
            )
        else:
            description = "实例化模板集[{template_name}]到命名空间[{namespace}]失败，错误消息：{err_msg}".format(
                template_name=template_name, namespace=ret["ns_name"], err_msg=ret["err_msg"]
            )

        log_params["description"] = description
        log_params['activity_status'] = ActivityStatus.Failed
        TemplatesetAuditor(AuditContext(**log_params)).log_raw()
        return release_id

    def update_release(self, release_data):
        project_id = release_data["project_id"]
        namespace_id = release_data["namespace_id"]
        namespace_info = get_namespace_by_id(self.access_token, project_id, namespace_id)

        release_id = release_data["release_id"]
        log_params = {
            "project_id": project_id,
            "user": self.username,
            "resource": release_data["name"],
            "resource_id": release_id,
            "extra": {"namespace": namespace_info["name"], "variable_info": release_data["variable_info"]},
            'activity_type': ActivityType.Modify,
        }
        try:
            update_resources(self.access_token, self.username, release_data, namespace_info)
        except Exception as e:
            log_params["description"] = f"rollupdate failed: {e}"
            log_params['activity_status'] = ActivityStatus.Failed
            InstanceAuditor(AuditContext(**log_params)).log_raw()
            update_inst_params = {"ins_state": InsState.UPDATE_FAILED.value}
        else:
            log_params["description"] = f"rollupdate success"
            log_params['activity_status'] = ActivityStatus.Succeed
            InstanceAuditor(AuditContext(**log_params)).log_raw()
            update_inst_params = {"ins_state": InsState.UPDATE_SUCCESS.value}

        update_inst_params.update(
            {"oper_type": ROLLING_UPDATE_INSTANCE, "variables": release_data["variable_info"].get(namespace_id) or {}}
        )

        InstanceConfig.objects.filter(
            instance_id=release_id,
            category=release_data["resource_name"],
            name=release_data["name"],
            namespace=namespace_id,
        ).update(**update_inst_params)

        return release_id
