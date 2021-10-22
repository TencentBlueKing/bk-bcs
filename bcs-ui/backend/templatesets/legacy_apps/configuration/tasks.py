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
import logging
import time
from datetime import datetime, timedelta

from celery import shared_task

from backend.components.paas_cc import get_namespace_list
from backend.templatesets.legacy_apps.instance.constants import InsState
from backend.templatesets.legacy_apps.instance.models import InstanceConfig, VersionInstance
from backend.templatesets.legacy_apps.instance.utils import get_app_status

logger = logging.getLogger(__name__)
POLLING_TIMEOUT = timedelta(seconds=30)
POLLING_INTERVAL_SECONDS = 5


@shared_task
def check_instance_status(access_token, project_id, project_kind, tmpl_name_dict, ns_info):
    # 通过命名空间获取集群信息
    all_ns = get_namespace_list(access_token, project_id, desire_all_data=True)
    all_ns_list = all_ns.get("data", {}).get("results") or []
    # 进行匹配命名空间和集群及模板
    ns_id_info_map = {info["ns_id"]: info for info in ns_info}
    ns_cluster = {}
    ns_name_id = {}
    for info in all_ns_list:
        if str(info["id"]) in ns_id_info_map:
            ns_name_id[info["name"]] = info["id"]
            if info["cluster_id"] not in ns_cluster:
                ns_cluster[info["cluster_id"]] = [info["name"]]
            else:
                ns_cluster[info["cluster_id"]].append(info["name"])
    end_time = datetime.now() + POLLING_TIMEOUT
    # 查询状态
    while datetime.now() < end_time:
        time.sleep(POLLING_INTERVAL_SECONDS)
        for cluster_id, ns in ns_cluster.items():
            ns_name_str = ",".join(set(ns))
            for category, tmpl in tmpl_name_dict.items():
                tmpl_name_str = ",".join(set(tmpl))
                results = get_app_status(
                    access_token, project_id, project_kind, cluster_id, tmpl_name_str, ns_name_str, category
                )
                # 更新db
                for key, val in results.items():
                    if val:
                        ns_id = ns_name_id.get(key[1])
                        update_data_record(key[0], ns_id, key[2])


def update_data_record(name, ns_id, category):
    """更新db中相应记录状态"""
    records = InstanceConfig.objects.filter(name=name, namespace=ns_id, category=category)
    # 更新instance configure
    records.update(ins_state=InsState.INS_SUCCESS.value, is_bcs_success=True)
    # 更新version instance
    version_instance_id_list = records.values_list("instance_id", flat=True)
    VersionInstance.objects.filter(id__in=version_instance_id_list).update(is_bcs_success=True)
