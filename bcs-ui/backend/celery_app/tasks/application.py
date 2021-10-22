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
import time
from datetime import datetime, timedelta

from celery import shared_task
from django.conf import settings
from django.utils.translation import ugettext_lazy as _

from backend.components.bcs import k8s
from backend.templatesets.legacy_apps.instance.constants import EventType, InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, InstanceEvent, VersionInstance
from backend.uniapps.application.constants import FUNC_MAP
from backend.utils.errcodes import ErrorCode

DEFAULT_RESPONSE = {"code": 0}
POLLING_INTERVAL_SECONDS = getattr(settings, "POLLING_INTERVAL_SECONDS", 5)
POLLING_TIMEOUT = timedelta(seconds=getattr(settings, "POLLING_TIMEOUT_SECONDS", 600))
logger = logging.getLogger(__name__)


def get_k8s_category_status(
    access_token, cluster_id, instance_name, project_id=None, category="application", field=None, namespace=None
):
    """查询mesos下application和deployment的状态"""
    client = k8s.K8SClient(access_token, project_id, cluster_id, None)
    curr_func = getattr(client, FUNC_MAP[category] % "get")
    resp = curr_func(
        {
            "name": instance_name,
            "field": field or "data.status",
            "namespace": namespace,
        }
    )
    return resp


def create_instance(access_token, cluster_id, ns, data, project_id=None, category="application", kind=2):
    """创建实例"""
    client = k8s.K8SClient(access_token, project_id, cluster_id, None)
    curr_func = getattr(client, FUNC_MAP[category] % "create")
    resp = curr_func(ns, data)
    resp = DEFAULT_RESPONSE
    return resp


def update_instance_record_status(info, oper_type, status="Running", is_bcs_success=True):
    """更新单条记录状态"""
    info.oper_type = oper_type
    info.status = status
    info.is_bcs_success = is_bcs_success
    info.save()


@shared_task
def application_polling_task(
    access_token, inst_id, cluster_id, instance_name, category, kind, ns_name, project_id, username=None, conf=None
):
    """轮训任务状态，并启动创建任务"""
    is_polling = True
    while is_polling:
        result = get_k8s_category_status(
            access_token, cluster_id, instance_name, category=category, namespace=ns_name, project_id=project_id
        )
        if result.get("code") == 0 and not result.get("data"):
            is_polling = False
        time.sleep(POLLING_INTERVAL_SECONDS)
    if str(inst_id) != "0":
        # 执行创建任务
        info = InstanceConfig.objects.get(id=inst_id)
        conf = json.loads(info.config)
    resp = create_instance(
        access_token, cluster_id, ns_name, conf, category=category, kind=kind, project_id=project_id
    )
    if str(inst_id) == "0":
        return
    # 更新instance状态
    if resp.get("code") != ErrorCode.NoError:
        update_instance_record_status(info, "rebuild", status="Error", is_bcs_success=False)
        # 记录失败事件
        conf_instance_id = conf.get("metadata", {}).get("labels", {}).get("io.tencent.paas.instanceid")
        err_msg = resp.get("message") or _("实例化失败，已通知管理员!")
        logger.error("实例化失败, 实例ID: %s, 详细:%s" % (inst_id, err_msg))
        try:
            InstanceEvent(
                instance_config_id=inst_id,
                category=category,
                msg_type=EventType.REQ_FAILED.value,
                instance_id=conf_instance_id,
                msg=err_msg,
                creator=username,
                updator=username,
                resp_snapshot=json.dumps(resp),
            ).save()
        except Exception as error:
            logger.error(u"存储实例化失败消息失败，详情: %s" % error)
    else:
        update_instance_record_status(info, "rebuild", status="Running", is_bcs_success=True)


@shared_task
def delete_instance_task(access_token, inst_id_list, project_kind):
    """后台更新删除实例是否被删除成功"""
    # 通过instance id获取到相应的记录，然后查询mesos/k8s的实例状态
    inst_info = InstanceConfig.objects.filter(id__in=inst_id_list)
    is_polling = True
    all_count = len(inst_info)
    end_time = datetime.now() + POLLING_TIMEOUT
    while is_polling:
        deleted_id_list = []
        time.sleep(POLLING_INTERVAL_SECONDS)
        for info in inst_info:
            inst_conf = json.loads(info.config)
            metadata = inst_conf.get("metadata") or {}
            labels = metadata.get("labels") or {}
            cluster_id = labels.get("io.tencent.bcs.clusterid")
            namespace = labels.get("io.tencent.bcs.namespace")
            project_id = labels.get("io.tencent.paas.projectid")
            category = info.category
            name = metadata.get("name")
            # 根据类型获取查询
            client = k8s.K8SClient(access_token, project_id, cluster_id, None)
            curr_func = getattr(client, FUNC_MAP[category] % "get")
            resp = curr_func({"name": name, "namespace": namespace})
            if not resp.get("data"):
                deleted_id_list.append(info.id)
                # 删除名称+命名空间+类型
                InstanceConfig.objects.filter(name=info.name, namespace=info.namespace, category=info.category).update(
                    is_deleted=True, deleted_time=datetime.now()
                )

        if len(deleted_id_list) == all_count or datetime.now() > end_time:
            is_polling = False


@shared_task
def update_create_error_record(id_list):
    records = InstanceConfig.objects.filter(id__in=id_list)
    records.update(ins_state=InsState.INS_SUCCESS.value, is_bcs_success=True)
    # 更新version instance
    version_instance_id_list = records.values_list("instance_id", flat=True)
    VersionInstance.objects.filter(id__in=version_instance_id_list).update(is_bcs_success=True)
