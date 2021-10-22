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

from celery import shared_task
from django.utils.translation import ugettext_lazy as _

from backend.components import paas_cc
from backend.components.bcs import k8s
from backend.container_service.clusters import models
from backend.container_service.clusters.constants import K8S_SKIP_NS_LIST
from backend.utils.errcodes import ErrorCode


@shared_task
def reschedule_node_pods(access_token, project_id, project_kind, cluster_id, data, log_id):
    log = models.NodeUpdateLog.objects.filter(id=log_id).first()
    if not log:
        update_node(access_token, project_id, data["id"], status=models.CommonStatus.ScheduleFailed)
    pod_taskgroup = get_pod_info(access_token, project_id, cluster_id, data, log)
    if not pod_taskgroup:
        resp = update_node(access_token, project_id, data["id"], status=models.CommonStatus.ScheduleFailed)
        if resp.get("code") != ErrorCode.NoError:
            update_log_info(log, models.CommonStatus.ScheduleFailed, resp.get("message"))
        return
    ok = reschedule_pod_taskgroup(access_token, project_id, pod_taskgroup, log)
    if not ok:
        status = models.CommonStatus.ScheduleFailed
    else:
        status = models.NodeStatus.ToRemoved
    resp = update_node(access_token, project_id, data["id"], status=status)
    if resp.get("code") != ErrorCode.NoError:
        update_log_info(log, models.CommonStatus.ScheduleFailed, resp.get("message"))


def update_node(access_token, project_id, node_id, status=None):
    if not status:
        status = models.CommonStatus.Scheduling
    resp = paas_cc.update_node(access_token, project_id, node_id, {"status": status})
    return resp


def get_pod_info(access_token, project_id, cluster_id, data, log):
    """获取pod下数量"""
    k8s_client = k8s.K8SClient(access_token, project_id, cluster_id, None)
    rsp_bcs = k8s_client.get_pod(host_ips=[data["inner_ip"]], field="namespace,resourceName,clusterId")
    if rsp_bcs.get("code") != ErrorCode.NoError:
        update_log_info(log, models.CommonStatus.ScheduleFailed, rsp_bcs.get("message"))
        return None
    cluster_ns_pod_data = {}
    for i in rsp_bcs["data"]:
        cluster_id = i.get("clusterId")
        namespace = i.get("namespace")
        if namespace not in K8S_SKIP_NS_LIST:
            pod_name = i.get("resourceName")
            item = {"namespace": namespace, "pod_name": pod_name}
            if cluster_id in cluster_ns_pod_data:
                cluster_ns_pod_data[cluster_id].append(item)
            else:
                cluster_ns_pod_data[cluster_id] = [item]
    return cluster_ns_pod_data


def reschedule_pod_taskgroup(access_token, project_id, data, log):
    """重新调度pod or taskgroup"""
    for cluster_id, pod_info in data.items():
        client = k8s.K8SClient(access_token, project_id, cluster_id, None)
        for info in pod_info:
            resp = client.delete_pod(info["namespace"], info["pod_name"])
            if resp.get("code") != ErrorCode.NoError:
                update_log_info(log, models.CommonStatus.ScheduleFailed, resp.get("message"))
                return False
    update_log_info(log, models.CommonStatus.Normal, _("调度结束!"))
    return True


def update_log_info(log, status, message):
    log.is_polling = False
    log.is_finished = True
    if status == models.CommonStatus.ScheduleFailed:
        log.status = models.CommonStatus.ScheduleFailed
        log.log = json.dumps({"state": "FAILURE", "node_tasks": [{"state": "FAILURE", "name": message}]})
    else:
        log.status = models.CommonStatus.ScheduleFailed
        log.log = json.dumps({"state": "SUCCESS", "node_tasks": [{"state": "SUCCESS", "name": _("调度结束!")}]})
    log.save()
